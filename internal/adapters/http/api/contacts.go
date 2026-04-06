package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// ContactsAPI handles /api/v1/contacts endpoints.
type ContactsAPI struct {
	contacts ports.ContactService
	search   ports.JobSearchService
}

// NewContactsAPI constructs a ContactsAPI.
func NewContactsAPI(contacts ports.ContactService, search ports.JobSearchService) *ContactsAPI {
	return &ContactsAPI{contacts: contacts, search: search}
}

// ImportCSV handles POST /api/v1/contacts/import (multipart form).
func (a *ContactsAPI) ImportCSV(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "no file provided")
		return
	}
	defer file.Close()

	result, err := a.contacts.ImportCSV(r.Context(), userID, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Stats handles GET /api/v1/contacts/stats.
func (a *ContactsAPI) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	total, linked, companies, err := a.contacts.Stats(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{
		"total":     total,
		"linked":    linked,
		"companies": companies,
	})
}

// DeleteAll handles DELETE /api/v1/contacts.
func (a *ContactsAPI) DeleteAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	if err := a.contacts.DeleteContacts(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Referrals handles GET /api/v1/jobs/{id}/referrals.
func (a *ContactsAPI) Referrals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	jobID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	job, err := a.search.GetByID(r.Context(), jobID)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	contacts, err := a.contacts.ContactsAtCompany(r.Context(), userID, job.CompanyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type contactJSON struct {
		Name        string `json:"name"`
		Title       string `json:"title"`
		Email       string `json:"email"`
		LinkedInURL string `json:"linkedin_url"`
	}
	out := make([]contactJSON, len(contacts))
	for i, c := range contacts {
		out[i] = contactJSON{
			Name:        c.FullName,
			Title:       c.Title,
			Email:       c.Email,
			LinkedInURL: c.LinkedInURL,
		}
	}
	writeJSON(w, http.StatusOK, out)
}
