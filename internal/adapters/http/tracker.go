package httphandlers

import (
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

// TrackerHandler handles htmx tracker action endpoints.
type TrackerHandler struct {
	application ports.ApplicationService
	search      ports.JobSearchService
}

// NewTrackerHandler constructs a TrackerHandler.
func NewTrackerHandler(application ports.ApplicationService, search ports.JobSearchService) *TrackerHandler {
	return &TrackerHandler{application: application, search: search}
}

// requireUser extracts the authenticated user ID from the context.
// Returns false and writes an error response if not authenticated.
func requireUser(w http.ResponseWriter, r *http.Request) (domain.UserID, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return domain.UserID{}, false
	}
	return userID, true
}

// requireJobID parses the job ID from the URL. Returns false and writes an
// error response on failure.
func requireJobID(w http.ResponseWriter, r *http.Request) (domain.JobID, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return domain.JobID{}, false
	}
	return id, true
}

// Interested handles POST /jobs/{id}/interested.
func (h *TrackerHandler) Interested(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	jobID, ok := requireJobID(w, r)
	if !ok {
		return
	}

	if err := h.application.SetStatus(r.Context(), userID, jobID, domain.StatusInterested); err != nil {
		log.Printf("set interested for job %s: %v", jobID, err)
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	h.renderActionBar(w, r, userID, jobID)
}

// Apply handles POST /jobs/{id}/apply.
func (h *TrackerHandler) Apply(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	jobID, ok := requireJobID(w, r)
	if !ok {
		return
	}

	if err := h.application.SetStatus(r.Context(), userID, jobID, domain.StatusApplied); err != nil {
		log.Printf("set applied for job %s: %v", jobID, err)
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	h.renderActionBar(w, r, userID, jobID)
}

// SetStatus handles POST /jobs/{id}/status.
func (h *TrackerHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	jobID, ok := requireJobID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	status := domain.ApplicationStatus(r.FormValue("status"))
	switch status {
	case domain.StatusInterested, domain.StatusApplied, domain.StatusInterviewing,
		domain.StatusOffer, domain.StatusRejected, domain.StatusWithdrawn:
		// valid
	default:
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	if err := h.application.SetStatus(r.Context(), userID, jobID, status); err != nil {
		log.Printf("set status %s for job %s: %v", status, jobID, err)
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	h.renderActionBar(w, r, userID, jobID)
}

// SetNotes handles POST /jobs/{id}/notes.
func (h *TrackerHandler) SetNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	jobID, ok := requireJobID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	notes := r.FormValue("notes")
	if err := h.application.SetNotes(r.Context(), userID, jobID, notes); err != nil {
		log.Printf("set notes for job %s: %v", jobID, err)
		http.Error(w, "Failed to save notes", http.StatusInternalServerError)
		return
	}

	uj, err := h.application.GetUserJob(r.Context(), userID, jobID)
	if err != nil {
		http.Error(w, "Failed to load updated state", http.StatusInternalServerError)
		return
	}
	csrfToken := csrf.Token(r)
	components.NotesPartial(jobID, uj.Notes, csrfToken).Render(r.Context(), w)
}

// renderActionBar fetches updated UserJob state and renders the action bar partial.
func (h *TrackerHandler) renderActionBar(w http.ResponseWriter, r *http.Request, userID domain.UserID, jobID domain.JobID) {
	uj, err := h.application.GetUserJob(r.Context(), userID, jobID)
	if err != nil {
		http.Error(w, "Failed to load updated state", http.StatusInternalServerError)
		return
	}
	csrfToken := csrf.Token(r)
	components.ActionBarPartial(jobID, &uj, csrfToken).Render(r.Context(), w)
}
