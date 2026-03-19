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

// CoachHandler handles Job Coach endpoints.
type CoachHandler struct {
	coach ports.JobCoachService
}

// NewCoachHandler constructs a CoachHandler.
func NewCoachHandler(coach ports.JobCoachService) *CoachHandler {
	return &CoachHandler{coach: coach}
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
