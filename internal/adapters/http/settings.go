package httphandlers

import (
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/justestif/go-jobs/components"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// SettingsHandler handles the user settings pages.
type SettingsHandler struct {
	user ports.UserService
}

// NewSettingsHandler constructs a SettingsHandler.
func NewSettingsHandler(user ports.UserService) *SettingsHandler {
	return &SettingsHandler{user: user}
}

// Show handles GET /settings. Renders the settings page.
func (h *SettingsHandler) Show(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.user.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to load user", http.StatusInternalServerError)
		return
	}

	saved := r.URL.Query().Get("saved")
	csrfToken := csrf.Token(r)
	components.SettingsPage(user, saved, true, csrfToken).Render(r.Context(), w)
}

// SaveResume handles POST /settings/resume.
func (h *SettingsHandler) SaveResume(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	resume := r.FormValue("resume")

	if err := h.user.SetResume(r.Context(), userID, resume); err != nil {
		http.Error(w, "Failed to save resume", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?saved=Resume+saved", http.StatusSeeOther)
}

// SaveLLM handles POST /settings/llm.
func (h *SettingsHandler) SaveLLM(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	provider := domain.LLMProvider(r.FormValue("provider"))
	model := r.FormValue("model")
	baseURL := r.FormValue("base_url")
	apiKey := r.FormValue("api_key")

	if err := h.user.SetLLMConfig(r.Context(), userID, provider, apiKey, model, baseURL); err != nil {
		http.Error(w, "Failed to save LLM settings", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?saved=LLM+settings+saved", http.StatusSeeOther)
}
