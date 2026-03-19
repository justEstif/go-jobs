package ports

import (
	"context"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// JobRepository persists and retrieves job postings.
type JobRepository interface {
	// Upsert inserts or updates a job by (company_id, external_id).
	// The dedup key is a business rule — enforced here, not in the scraper adapter.
	// source identifies which ATS platform the job came from (mirrors company.ATSType).
	Upsert(ctx context.Context, companyID domain.CompanyID, job domain.RawJob, source domain.ATSType) (domain.JobID, error)
	GetByID(ctx context.Context, id domain.JobID) (domain.Job, error)
	// GetByIDs fetches multiple jobs by ID in a single query. Order of results
	// is not guaranteed to match the order of ids. Used by ApplicationService
	// to hydrate job lists returned from UserJobRepository.ListByStatus.
	GetByIDs(ctx context.Context, ids []domain.JobID) ([]domain.Job, error)
	// Search returns jobs matching filters. If userCtx is non-nil, results are
	// annotated with the user's pipeline state (avoids N+1) and OnlyNew
	// filtering is applied against userCtx.UserID's LastVisitedAt.
	Search(ctx context.Context, filters domain.SearchFilters, userCtx *domain.UserSearchContext) ([]domain.Job, error)
	// ListUnenriched returns jobs without a job_tags row, up to limit.
	ListUnenriched(ctx context.Context, limit int) ([]domain.Job, error)
	// MarkInactive sets active=false for jobs no longer present at source and
	// returns the number of jobs deactivated.
	MarkInactive(ctx context.Context, companyID domain.CompanyID, activeExternalIDs []string) (int, error)
	SaveTags(ctx context.Context, tags domain.JobTags) error
}

// CompanyRepository persists and retrieves companies.
type CompanyRepository interface {
	Upsert(ctx context.Context, company domain.Company) (domain.CompanyID, error)
	ListActive(ctx context.Context) ([]domain.Company, error)
	GetByID(ctx context.Context, id domain.CompanyID) (domain.Company, error)
	GetByBoardToken(ctx context.Context, atsType domain.ATSType, boardToken string) (domain.Company, error)
}

// UserRepository persists and retrieves user records.
// Concerned only with user identity and profile data — not with session tokens.
type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	// UpdatePassword sets a new bcrypt password hash for the user.
	UpdatePassword(ctx context.Context, userID domain.UserID, passwordHash string) error
	TouchLastVisited(ctx context.Context, userID domain.UserID) error
	// Delete permanently removes a user. All related data (sessions, user_jobs,
	// user_companies, coach_cache) is cascade-deleted by the database.
	Delete(ctx context.Context, userID domain.UserID) error
}

// SessionRepository manages opaque session tokens used for authentication.
//
// Separated from UserRepository so that changes to the auth mechanism
// (e.g. moving tokens from a column to a sessions table, or adopting JWT)
// do not require touching user persistence logic.
// AuthService depends on both UserRepository and SessionRepository.
// The composition root may wire both to the same postgres adapter package.
type SessionRepository interface {
	// SaveToken associates token with userID. Overwrites any existing token for this user.
	SaveToken(ctx context.Context, userID domain.UserID, token string) error
	// DeleteToken invalidates a token. No-op if the token does not exist.
	DeleteToken(ctx context.Context, token string) error
	// GetUserByToken returns the User associated with token, or an error if
	// the token is invalid or expired. Called on every authenticated request.
	GetUserByToken(ctx context.Context, token string) (domain.User, error)
}

// UserJobRepository manages per-user pipeline state.
type UserJobRepository interface {
	// Upsert writes all fields from userJob to the store.
	//
	// Callers must read the existing UserJob first (via GetUserJob) before calling
	// Upsert to avoid overwriting fields they did not intend to change — this is
	// a deliberate read-modify-write contract, not a partial-update mechanism.
	//
	// Semantics:
	//   - Status and StatusAt are always overwritten.
	//   - Notes: an empty string clears existing notes. Preserve by reading first.
	//   - AppliedAt: set by the adapter on the first transition to StatusApplied
	//     and never overwritten, regardless of the value in userJob.AppliedAt.
	//     Callers do not need to manage AppliedAt.
	Upsert(ctx context.Context, userJob domain.UserJob) error
	GetUserJob(ctx context.Context, userID domain.UserID, jobID domain.JobID) (domain.UserJob, error)
	ListByStatus(ctx context.Context, userID domain.UserID, status domain.ApplicationStatus) ([]domain.JobID, error)
	ListAll(ctx context.Context, userID domain.UserID) ([]domain.UserJob, error)
	// DeleteAll removes all pipeline state for a user (reset tracker).
	DeleteAll(ctx context.Context, userID domain.UserID) error
}

// ScrapeRunRepository persists scrape run records.
type ScrapeRunRepository interface {
	Create(ctx context.Context, run domain.ScrapeRun) error
	Update(ctx context.Context, run domain.ScrapeRun) error
	Latest(ctx context.Context) (domain.ScrapeRun, error)
}

// UserCompanyRepository manages per-user company visibility preferences.
type UserCompanyRepository interface {
	SetHidden(ctx context.Context, userID domain.UserID, companyID domain.CompanyID, hidden bool) error
	ListHidden(ctx context.Context, userID domain.UserID) ([]domain.CompanyID, error)
}

// JobScraper fetches raw job postings from a single ATS platform.
//
// Each ATS platform (Greenhouse, Lever, Ashby) has its own adapter.
// The adapter is responsible for normalising platform-specific fields
// into RawJob before returning — the core receives a clean, platform-agnostic struct.
type JobScraper interface {
	Scrape(ctx context.Context, company domain.Company) ([]domain.RawJob, error)
}

// JobEnricher extracts structured tags from a raw job posting.
//
// The tiered enrichment adapter (ATS metadata → rules → LLM) implements this.
// The core does not know which tier produced the result.
type JobEnricher interface {
	Enrich(ctx context.Context, job domain.Job) (domain.JobTags, error)
}

// CoachCacheRepository persists and retrieves cached LLM analysis results.
//
// Caching avoids re-spending on identical LLM calls. Users can force a refresh
// ("Re-analyze") which overwrites the cached row via Upsert.
type CoachCacheRepository interface {
	// Get returns the cached result for the (userID, jobID, kind) triple.
	// Returns an error wrapping pgx.ErrNoRows if no cache entry exists.
	Get(ctx context.Context, userID domain.UserID, jobID domain.JobID, kind domain.CoachCacheKind) (domain.CoachCache, error)
	// Upsert writes or overwrites a cache entry.
	Upsert(ctx context.Context, entry domain.CoachCache) error
}

// LLMClient sends prompts to an LLM provider and returns the response.
//
// Each adapter (Ollama, OpenAI) implements this interface. The adapter
// constructs a fresh SDK client per call using the provided credentials —
// keys are never retained beyond the call lifetime.
type LLMClient interface {
	// Complete sends a system prompt and user prompt to the model and returns
	// the full response text. The provider, model, API key, and base URL are
	// configured at adapter construction time.
	Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// CompanySeeder discovers company slugs/tokens from external sources
// (e.g. parsing the Simplify README files) and returns them as Company values
// ready for upsert.
//
// Does not persist — the ScrapeService calls this then hands results to
// CompanyRepository.Upsert.
type CompanySeeder interface {
	Seed(ctx context.Context) ([]domain.Company, error)
}
