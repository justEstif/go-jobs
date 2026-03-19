package httphandlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"

	"github.com/justestif/go-jobs/components"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// SettingsHandler renders the settings page and handles account management.
//
// auth handles password verification (ChangePassword, DeleteAccount).
// application handles pipeline reset (ResetTracker).
// companies handles company visibility preferences.
type SettingsHandler struct {
	auth        ports.AuthService
	application ports.ApplicationService
	companies   ports.CompanyService
	sm          *middleware.SessionManager
}

// NewSettingsHandler constructs a SettingsHandler.
func NewSettingsHandler(auth ports.AuthService, application ports.ApplicationService, companies ports.CompanyService, sm *middleware.SessionManager) *SettingsHandler {
	return &SettingsHandler{auth: auth, application: application, companies: companies, sm: sm}
}

// Show handles GET /settings. Renders the settings page with blocked companies.
func (h *SettingsHandler) Show(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	hidden, err := h.companies.ListHiddenCompanies(r.Context(), userID)
	if err != nil {
		log.Printf("list hidden companies for user %s: %v", userID, err)
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	saved := r.URL.Query().Get("saved")
	errMsg := r.URL.Query().Get("error")
	csrfToken := csrf.Token(r)
	components.SettingsPage(hidden, saved, errMsg, csrfToken).Render(r.Context(), w)
}

// ChangePassword handles POST /settings/password.
func (h *SettingsHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	current := r.FormValue("current_password")
	newPass := r.FormValue("new_password")
	confirm := r.FormValue("confirm_password")

	if newPass != confirm {
		http.Redirect(w, r, "/settings?error=New+passwords+do+not+match", http.StatusSeeOther)
		return
	}
	if len(newPass) < 8 {
		http.Redirect(w, r, "/settings?error=New+password+must+be+at+least+8+characters", http.StatusSeeOther)
		return
	}

	if err := h.auth.ChangePassword(r.Context(), userID, current, newPass); err != nil {
		if errors.Is(err, ports.ErrInvalidCredentials) {
			http.Redirect(w, r, "/settings?error=Current+password+is+incorrect", http.StatusSeeOther)
			return
		}
		log.Printf("change password for user %s: %v", userID, err)
		http.Error(w, "Failed to change password", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?saved=Password+changed", http.StatusSeeOther)
}

// ResetTracker handles POST /settings/reset-tracker.
func (h *SettingsHandler) ResetTracker(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	if err := h.application.ResetTracker(r.Context(), userID); err != nil {
		log.Printf("reset tracker for user %s: %v", userID, err)
		http.Error(w, "Failed to reset tracker", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?saved=Pipeline+tracker+reset", http.StatusSeeOther)
}

// DeleteAccount handles POST /settings/delete-account.
func (h *SettingsHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	password := r.FormValue("password")
	if err := h.auth.DeleteAccount(r.Context(), userID, password); err != nil {
		if errors.Is(err, ports.ErrInvalidCredentials) {
			http.Redirect(w, r, "/settings?error=Incorrect+password", http.StatusSeeOther)
			return
		}
		log.Printf("delete account for user %s: %v", userID, err)
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}

	// Destroy the HTTP session after deleting the user.
	_ = h.sm.ClearSession(r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HideCompany handles POST /companies/{id}/hide.
// Returns an htmx partial for the job detail page.
func (h *SettingsHandler) HideCompany(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	companyID, ok := h.requireCompanyID(w, r)
	if !ok {
		return
	}

	if err := h.companies.HideCompany(r.Context(), userID, companyID); err != nil {
		log.Printf("hide company %s for user %s: %v", companyID, userID, err)
		http.Error(w, "Failed to hide company", http.StatusInternalServerError)
		return
	}

	csrfToken := csrf.Token(r)
	components.HideCompanyPartial(companyID, true, csrfToken).Render(r.Context(), w)
}

// ShowCompany handles POST /companies/{id}/show.
// Returns an htmx partial to flip the button back, or an empty response
// when called from the settings page (removes the row via hx-swap="outerHTML").
func (h *SettingsHandler) ShowCompany(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	companyID, ok := h.requireCompanyID(w, r)
	if !ok {
		return
	}

	if err := h.companies.ShowCompany(r.Context(), userID, companyID); err != nil {
		log.Printf("show company %s for user %s: %v", companyID, userID, err)
		http.Error(w, "Failed to show company", http.StatusInternalServerError)
		return
	}

	// If called from settings page (unblock), return empty to remove the row.
	if r.URL.Query().Get("from") == "settings" {
		w.WriteHeader(http.StatusOK)
		return
	}

	csrfToken := csrf.Token(r)
	components.HideCompanyPartial(companyID, false, csrfToken).Render(r.Context(), w)
}

// requireCompanyID parses the company ID from the URL.
func (h *SettingsHandler) requireCompanyID(w http.ResponseWriter, r *http.Request) ([16]byte, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return [16]byte{}, false
	}
	return id, true
}
