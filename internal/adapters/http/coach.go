package httphandlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"

	"github.com/justestif/go-jobs/components"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// CoachHandler handles the unified Job Coach page: resume config, LLM provider
// config, case study generation, and per-job analysis.
type CoachHandler struct {
	coach ports.JobCoachService
	user  ports.UserService
}

// NewCoachHandler constructs a CoachHandler.
func NewCoachHandler(coach ports.JobCoachService, user ports.UserService) *CoachHandler {
	return &CoachHandler{coach: coach, user: user}
}

// Show handles GET /coach. Renders the unified coach page.
func (h *CoachHandler) Show(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.user.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to load user", http.StatusInternalServerError)
		return
	}

	saved := r.URL.Query().Get("saved")
	csrfToken := csrf.Token(r)
	components.CoachPage(user, saved, "", "", true, csrfToken).Render(r.Context(), w)
}

// SaveResume handles POST /coach/resume.
func (h *CoachHandler) SaveResume(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	resume := r.FormValue("resume")

	if len(resume) > domain.MaxResumeLength {
		http.Error(w, fmt.Sprintf("Resume too long: %d characters exceeds maximum of %d", len(resume), domain.MaxResumeLength), http.StatusBadRequest)
		return
	}

	if err := h.user.SetResume(r.Context(), userID, resume); err != nil {
		http.Error(w, "Failed to save resume", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/coach?saved=Resume+saved", http.StatusSeeOther)
}

// SaveLLM handles POST /coach/llm.
func (h *CoachHandler) SaveLLM(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, "/coach?saved=LLM+settings+saved", http.StatusSeeOther)
}

// Analyze handles POST /jobs/{id}/analyze. Runs the Job Coach analysis
// and returns an htmx partial for the coach panel.
func (h *CoachHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	refresh := r.URL.Query().Get("refresh") == "true"

	result, err := h.coach.AnalyzeJob(r.Context(), userID, jobID, refresh)
	if err != nil {
		log.Printf("coach analyze: %v", err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	csrfToken := csrf.Token(r)
	components.CoachResultPanel(result, "", idStr, csrfToken).Render(r.Context(), w)
}

// CaseStudyGenerate handles POST /coach/case-study. Generates a case study and
// renders the full coach page with the result.
func (h *CoachHandler) CaseStudyGenerate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	description := r.FormValue("description")

	if description == "" {
		http.Error(w, "Project description is required", http.StatusBadRequest)
		return
	}

	user, err := h.user.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to load user", http.StatusInternalServerError)
		return
	}

	result, err := h.coach.GenerateCaseStudy(r.Context(), userID, description)
	if err != nil {
		log.Printf("coach case study: %v", err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	csrfToken := csrf.Token(r)
	components.CoachPage(user, "", description, result, true, csrfToken).Render(r.Context(), w)
}
