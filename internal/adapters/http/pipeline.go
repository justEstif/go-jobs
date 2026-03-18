package httphandlers

import (
	"log"
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/justestif/go-jobs/components"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// PipelineHandler renders the user's tracked jobs grouped by pipeline status.
type PipelineHandler struct {
	application ports.ApplicationService
}

// NewPipelineHandler constructs a PipelineHandler.
func NewPipelineHandler(application ports.ApplicationService) *PipelineHandler {
	return &PipelineHandler{application: application}
}

// List handles GET /pipeline.
func (h *PipelineHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pipeline, err := h.application.ListPipeline(r.Context(), userID)
	if err != nil {
		log.Printf("list pipeline for user %s: %v", userID, err)
		http.Error(w, "Failed to load pipeline", http.StatusInternalServerError)
		return
	}

	csrfToken := csrf.Token(r)
	components.PipelinePage(pipeline, csrfToken).Render(r.Context(), w)
}
