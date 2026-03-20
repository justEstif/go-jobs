package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/domain"

	"github.com/google/uuid"
)

// Search handles GET /api/v1/jobs.
// Query params mirror the CLI search flags:
//
//	q, role, seniority, remote, country, tech, company, posted_within, limit, offset, page, per_page
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	filters := parseFilters(r)

	var userCtx *domain.UserSearchContext
	if userID, ok := middleware.UserIDFromContext(r.Context()); ok {
		userCtx = &domain.UserSearchContext{UserID: userID}
	}

	jobs, err := h.search.Search(r.Context(), filters, userCtx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	writeJSON(w, http.StatusOK, jobs)
}

// Interested handles GET /api/v1/jobs/interested.
// Requires Bearer auth. Returns jobs the user has marked as interested.
func (h *Handler) Interested(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobs, err := h.application.ListByStatus(r.Context(), userID, domain.StatusInterested)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list interested jobs")
		return
	}

	writeJSON(w, http.StatusOK, jobs)
}

// Applied handles GET /api/v1/jobs/applied.
// Requires Bearer auth. Returns jobs the user has applied to.
func (h *Handler) Applied(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobs, err := h.application.ListByStatus(r.Context(), userID, domain.StatusApplied)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list applied jobs")
		return
	}

	writeJSON(w, http.StatusOK, jobs)
}

// parseFilters extracts domain.SearchFilters from URL query params.
func parseFilters(r *http.Request) domain.SearchFilters {
	q := r.URL.Query()

	limit := apiPerPage(q.Get("per_page"), q.Get("limit"))
	page := apiPage(q.Get("page"))
	offset := 0
	if offsetRaw := q.Get("offset"); offsetRaw != "" {
		if v, err := strconv.Atoi(offsetRaw); err == nil && v >= 0 {
			offset = v
		}
	} else {
		offset = (page - 1) * limit
	}

	f := domain.SearchFilters{
		Query:            q.Get("q"),
		PostedWithinDays: apiPostedWithin(q.Get("posted_within")),
		Limit:            limit,
		Offset:           offset,
	}

	for _, v := range q["role"] {
		f.RoleTypes = append(f.RoleTypes, domain.RoleType(v))
	}
	for _, v := range q["seniority"] {
		f.Seniorities = append(f.Seniorities, domain.Seniority(v))
	}
	for _, v := range q["remote"] {
		f.RemotePolicy = append(f.RemotePolicy, domain.WorkplaceType(v))
	}
	for _, v := range q["country"] {
		f.Countries = append(f.Countries, v)
	}
	for _, v := range q["tech"] {
		if t := strings.TrimSpace(v); t != "" {
			f.TechStack = append(f.TechStack, t)
		}
	}
	for _, raw := range q["company"] {
		if id, err := uuid.Parse(raw); err == nil {
			f.CompanyIDs = append(f.CompanyIDs, id)
		}
	}

	return f
}

func apiPage(raw string) int {
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return 1
	}
	return v
}

func apiPerPage(perPageRaw, limitRaw string) int {
	for _, raw := range []string{perPageRaw, limitRaw} {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			return v
		}
	}
	return 50
}

func apiPostedWithin(raw string) int {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "24h":
		return 1
	case "7d":
		return 7
	case "14d":
		return 14
	case "30d":
		return 30
	case "all", "0":
		return 0
	default:
		return 90
	}
}
