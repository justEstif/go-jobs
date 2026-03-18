package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// MarkInterested handles POST /api/v1/jobs/{id}/interested.
// Requires Bearer auth. Marks the job as interested in the user's pipeline.
func (h *Handler) MarkInterested(w http.ResponseWriter, r *http.Request) {
	userID, jobID, ok := requireUserAndJob(w, r)
	if !ok {
		return
	}

	if err := h.application.SetStatus(r.Context(), userID, jobID, domain.StatusInterested); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MarkApplied handles POST /api/v1/jobs/{id}/apply.
// Requires Bearer auth. Marks the job as applied (also sets interested if not already set).
func (h *Handler) MarkApplied(w http.ResponseWriter, r *http.Request) {
	userID, jobID, ok := requireUserAndJob(w, r)
	if !ok {
		return
	}

	if err := h.application.SetStatus(r.Context(), userID, jobID, domain.StatusApplied); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type setStatusRequest struct {
	Status string `json:"status"`
}

// SetStatus handles POST /api/v1/jobs/{id}/status.
// Requires Bearer auth. Body: {"status": "<status>"}.
func (h *Handler) SetStatus(w http.ResponseWriter, r *http.Request) {
	userID, jobID, ok := requireUserAndJob(w, r)
	if !ok {
		return
	}

	var req setStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	status := domain.ApplicationStatus(req.Status)
	switch status {
	case domain.StatusInterested, domain.StatusApplied, domain.StatusInterviewing,
		domain.StatusOffer, domain.StatusRejected, domain.StatusWithdrawn:
		// valid
	default:
		writeError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	if err := h.application.SetStatus(r.Context(), userID, jobID, status); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type setNotesRequest struct {
	Notes string `json:"notes"`
}

// SetNotes handles POST /api/v1/jobs/{id}/notes.
// Requires Bearer auth. Body: {"notes": "<text>"}.
func (h *Handler) SetNotes(w http.ResponseWriter, r *http.Request) {
	userID, jobID, ok := requireUserAndJob(w, r)
	if !ok {
		return
	}

	var req setNotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := h.application.SetNotes(r.Context(), userID, jobID, req.Notes); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save notes")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// requireUserAndJob extracts the authenticated user ID and the job ID URL param.
// Returns false and writes an error response on any failure.
func requireUserAndJob(w http.ResponseWriter, r *http.Request) (domain.UserID, domain.JobID, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return domain.UserID{}, domain.JobID{}, false
	}

	idStr := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job ID")
		return domain.UserID{}, domain.JobID{}, false
	}

	return userID, jobID, true
}
