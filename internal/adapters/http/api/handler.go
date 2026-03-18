// Package api implements the /api/v1/ JSON driving adapter.
//
// This is a second driving adapter alongside the HTML web UI handlers. Both
// call the same core service ports — no business logic lives here. API routes
// use Bearer token auth; no CSRF is required.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/justestif/go-jobs/internal/core/ports"
)

// Handler holds all services required by the /api/v1/ route group.
type Handler struct {
	auth        ports.AuthService
	search      ports.JobSearchService
	application ports.ApplicationService
}

// New constructs a Handler.
func New(
	auth ports.AuthService,
	search ports.JobSearchService,
	application ports.ApplicationService,
) *Handler {
	return &Handler{
		auth:        auth,
		search:      search,
		application: application,
	}
}

// writeJSON serialises v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error body: {"error": "message"}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
