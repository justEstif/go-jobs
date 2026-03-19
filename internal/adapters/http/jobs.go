package httphandlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"

	"github.com/justestif/go-jobs/components"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// JobSearchHandler handles the jobs browse and search pages.
type JobSearchHandler struct {
	search      ports.JobSearchService
	application ports.ApplicationService
	user        ports.UserService
	coach       ports.JobCoachService
}

// NewJobSearchHandler constructs a JobSearchHandler.
func NewJobSearchHandler(search ports.JobSearchService, application ports.ApplicationService, user ports.UserService, coach ports.JobCoachService) *JobSearchHandler {
	return &JobSearchHandler{
		search:      search,
		application: application,
		user:        user,
		coach:       coach,
	}
}

// List handles GET /. Renders the jobs list with search bar and filter panel.
func (h *JobSearchHandler) List(w http.ResponseWriter, r *http.Request) {
	filters := parseSearchFilters(r)

	var userCtx *domain.UserSearchContext
	userID, loggedIn := middleware.UserIDFromContext(r.Context())
	if loggedIn {
		// Touch last visited asynchronously — best-effort, don't block the request.
		// Use WithoutCancel so the goroutine's DB write survives handler return,
		// which cancels r.Context().
		touchCtx := context.WithoutCancel(r.Context())
		go func(uid [16]byte) {
			if err := h.user.TouchLastVisited(touchCtx, uid); err != nil {
				log.Printf("touch last visited: %v", err)
			}
		}(userID)

		onlyNew := r.URL.Query().Get("new") == "1"
		userCtx = &domain.UserSearchContext{
			UserID:  userID,
			OnlyNew: onlyNew,
		}
	}

	jobs, err := h.search.Search(r.Context(), filters, userCtx)
	if err != nil {
		log.Printf("search jobs: %v", err)
		http.Error(w, "Failed to load jobs", http.StatusInternalServerError)
		return
	}

	csrfToken := csrf.Token(r)
	components.JobsListPage(jobs, filters, loggedIn, csrfToken).Render(r.Context(), w)
}

// Detail handles GET /jobs/{id}. Renders the job detail page.
func (h *JobSearchHandler) Detail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	job, err := h.search.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	var userJob *domain.UserJob
	var hasResume, hasLLM bool
	userID, loggedIn := middleware.UserIDFromContext(r.Context())
	if loggedIn {
		uj, err := h.application.GetUserJob(r.Context(), userID, job.ID)
		if err == nil {
			userJob = &uj
		}

		// Check if user has resume and LLM configured for coach feature.
		user, err := h.user.GetByID(r.Context(), userID)
		if err == nil {
			hasResume = user.Resume != ""
			hasLLM = user.LLMProvider != ""
		}
	}

	csrfToken := csrf.Token(r)
	components.JobDetailPage(job, userJob, loggedIn, csrfToken, hasResume, hasLLM).Render(r.Context(), w)
}

// parseSearchFilters extracts domain.SearchFilters from URL query params.
// All slice fields use repeated params: ?role=engineering&role=data
func parseSearchFilters(r *http.Request) domain.SearchFilters {
	q := r.URL.Query()

	_, postedWithinPresent := q["posted_within"]
	f := domain.SearchFilters{
		Query:            q.Get("q"),
		PostedWithinDays: parsePostedWithin(q.Get("posted_within"), postedWithinPresent),
		Limit:            parsePerPage(q.Get("per_page")),
	}

	page := parsePage(q.Get("page"))
	f.Offset = (page - 1) * f.Limit

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

	return f
}

func parsePage(raw string) int {
	if raw == "" {
		return 1
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return 1
	}
	return v
}

func parsePerPage(raw string) int {
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 25
	}
	switch v {
	case 25, 50, 100:
		return v
	default:
		return 25
	}
}

// parsePostedWithin converts the posted_within query param to a number of days.
// present=false means the param was absent entirely — default to 1 day (24h).
// present=true with an empty/unrecognised value means the user explicitly chose
// "Any time", so no date filter is applied.
func parsePostedWithin(raw string, present bool) int {
	switch raw {
	case "24h":
		return 1
	case "7d":
		return 7
	case "14d":
		return 14
	case "90d":
		return 90
	default:
		if !present {
			return 1 // default: last 24 hours
		}
		return 0 // explicit "Any time"
	}
}
