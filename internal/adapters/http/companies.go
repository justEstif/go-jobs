package httphandlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"

	"github.com/justestif/go-jobs/components"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// CompanyHandler renders the companies list and handles hide/show toggles.
type CompanyHandler struct {
	companies ports.CompanyService
}

// NewCompanyHandler constructs a CompanyHandler.
func NewCompanyHandler(companies ports.CompanyService) *CompanyHandler {
	return &CompanyHandler{companies: companies}
}

// List handles GET /companies.
func (h *CompanyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	all, hidden, err := h.companies.ListAllWithHiddenIDs(r.Context(), userID)
	if err != nil {
		log.Printf("list companies for user %s: %v", userID, err)
		http.Error(w, "Failed to load companies", http.StatusInternalServerError)
		return
	}

	csrfToken := csrf.Token(r)
	components.CompaniesPage(all, hidden, csrfToken).Render(r.Context(), w)
}

// Hide handles POST /companies/{id}/hide.
func (h *CompanyHandler) Hide(w http.ResponseWriter, r *http.Request) {
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
	components.CompanyTogglePartial(companyID, true, csrfToken).Render(r.Context(), w)
}

// Show handles POST /companies/{id}/show.
func (h *CompanyHandler) Show(w http.ResponseWriter, r *http.Request) {
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

	csrfToken := csrf.Token(r)
	components.CompanyTogglePartial(companyID, false, csrfToken).Render(r.Context(), w)
}

// requireCompanyID parses the company ID from the URL.
func (h *CompanyHandler) requireCompanyID(w http.ResponseWriter, r *http.Request) ([16]byte, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return [16]byte{}, false
	}
	return id, true
}
