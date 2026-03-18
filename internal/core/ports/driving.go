package ports

import (
	"context"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// JobSearchService handles job discovery and browsing.
type JobSearchService interface {
	// Search returns jobs matching filters. If userCtx is non-nil, results are
	// annotated with the user's pipeline state and OnlyNew filtering is applied.
	Search(ctx context.Context, filters domain.SearchFilters, userCtx *domain.UserSearchContext) ([]domain.Job, error)
	GetByID(ctx context.Context, id domain.JobID) (domain.Job, error)
}

// ApplicationService manages a user's pipeline state for jobs.
type ApplicationService interface {
	// SetStatus transitions a job to the given status.
	//
	// Business rules:
	//   - Setting Applied also sets Interested (if not already) and captures AppliedAt.
	//   - Setting Interested on an already-Applied job is a no-op (can't go backwards).
	SetStatus(ctx context.Context, userID domain.UserID, jobID domain.JobID, status domain.ApplicationStatus) error
	SetNotes(ctx context.Context, userID domain.UserID, jobID domain.JobID, notes string) error
	GetUserJob(ctx context.Context, userID domain.UserID, jobID domain.JobID) (domain.UserJob, error)
	// ListByStatus returns jobs in a given status, annotated with UserJob state.
	ListByStatus(ctx context.Context, userID domain.UserID, status domain.ApplicationStatus) ([]domain.Job, error)
	// ListPipeline returns all tracked jobs for a user, grouped by status.
	ListPipeline(ctx context.Context, userID domain.UserID) (map[domain.ApplicationStatus][]domain.Job, error)
}

// ScrapeService orchestrates the full scrape → enrich → persist pipeline.
//
// The scheduler calls Run; the service owns pipeline order and business rules
// (e.g. skip headless companies, dedup on upsert).
type ScrapeService interface {
	Run(ctx context.Context) error
	// SeedCompanies fetches company slugs from the Simplify README sources
	// and upserts them into the company list.
	SeedCompanies(ctx context.Context) error
	// LatestRun returns the most recent ScrapeRun record (for UI status display).
	LatestRun(ctx context.Context) (domain.ScrapeRun, error)
}

// AuthService manages user registration and authentication.
//
// Depends on UserRepository (user records) and SessionRepository (tokens).
// Authenticate is called by HTTP middleware and CLI on every request to resolve
// a session cookie/file token to a User — it is the composition point between
// the two repositories.
type AuthService interface {
	Register(ctx context.Context, email, password string) (domain.User, error)
	Login(ctx context.Context, email, password string) (token string, err error)
	Logout(ctx context.Context, token string) error
	// Authenticate validates a session token and returns the associated user.
	Authenticate(ctx context.Context, token string) (domain.User, error)
}

// UserService manages user settings.
type UserService interface {
	SetLLMKey(ctx context.Context, userID domain.UserID, provider domain.LLMProvider, apiKey string) error
	GetByID(ctx context.Context, id domain.UserID) (domain.User, error)
	// TouchLastVisited updates LastVisitedAt to now. Called on each authenticated session.
	TouchLastVisited(ctx context.Context, userID domain.UserID) error
}

// CompanyService manages per-user company visibility.
type CompanyService interface {
	HideCompany(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) error
	ShowCompany(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) error
	// ListCompanies returns all active companies excluding those hidden by the user.
	ListCompanies(ctx context.Context, userID domain.UserID) ([]domain.Company, error)
}

// EnrichService runs the enrichment pipeline over un-tagged jobs.
//
// It fetches up to limit un-enriched jobs, runs each through the JobEnricher,
// and persists the resulting tags. Errors on individual jobs are logged and
// skipped — a single failing enrichment does not abort the batch.
type EnrichService interface {
	// Run enriches up to limit un-tagged jobs. Returns the count of
	// successfully enriched jobs and jobs that failed enrichment.
	Run(ctx context.Context, limit int) (enriched, failed int, err error)
}
