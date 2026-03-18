package api

import (
	"net/http"

	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// pipelineGroup is the JSON shape for a single pipeline status group.
type pipelineGroup struct {
	Status string       `json:"status"`
	Jobs   []domain.Job `json:"jobs"`
}

// pipelineOrder defines the canonical display order for pipeline statuses.
var pipelineOrder = []domain.ApplicationStatus{
	domain.StatusInterested,
	domain.StatusApplied,
	domain.StatusInterviewing,
	domain.StatusOffer,
	domain.StatusRejected,
	domain.StatusWithdrawn,
}

// Pipeline handles GET /api/v1/pipeline.
// Requires Bearer auth. Returns all tracked jobs grouped by status.
func (h *Handler) Pipeline(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	grouped, err := h.application.ListPipeline(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load pipeline")
		return
	}

	out := make([]pipelineGroup, 0, len(pipelineOrder))
	for _, status := range pipelineOrder {
		jobs := grouped[status]
		if len(jobs) == 0 {
			continue
		}
		out = append(out, pipelineGroup{
			Status: string(status),
			Jobs:   jobs,
		})
	}

	writeJSON(w, http.StatusOK, out)
}
