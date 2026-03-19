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
	// ResetTracker deletes all pipeline state (user_jobs) for the user.
	// Lives here because it operates on UserJobRepository, which ApplicationService owns.
	ResetTracker(ctx context.Context, userID domain.UserID) error
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

// AuthService manages user registration, authentication, and password operations.
//
// Depends on UserRepository (user records) and SessionRepository (tokens).
// Authenticate is called by HTTP middleware and CLI on every request to resolve
// a session cookie/file token to a User — it is the composition point between
// the two repositories.
//
// ChangePassword and DeleteAccount live here (not UserService) because they
// verify the current password with bcrypt — that is authentication logic.
type AuthService interface {
	Register(ctx context.Context, email, password string) (domain.User, error)
	Login(ctx context.Context, email, password string) (token string, err error)
	Logout(ctx context.Context, token string) error
	// Authenticate validates a session token and returns the associated user.
	Authenticate(ctx context.Context, token string) (domain.User, error)
	// ChangePassword verifies the current password and sets a new one.
	// Returns ErrInvalidCredentials if the current password is wrong.
	ChangePassword(ctx context.Context, userID domain.UserID, currentPassword, newPassword string) error
	// DeleteAccount permanently removes the user and all associated data.
	// Requires the current password for confirmation.
	// Returns ErrInvalidCredentials if the password is wrong.
	DeleteAccount(ctx context.Context, userID domain.UserID, password string) error
}

// UserService manages user profile and settings (LLM config, resume).
//
// Does not handle password or account deletion — those are auth concerns
// owned by AuthService.
type UserService interface {
	// SetLLMConfig stores the user's LLM provider, model, base URL, and
	// encrypted API key. The apiKey is encrypted before storage — callers
	// pass the plaintext key.
	SetLLMConfig(ctx context.Context, userID domain.UserID, provider domain.LLMProvider, apiKey, model, baseURL string) error
	// SetResume stores the user's resume (markdown/plaintext) for Job Coach analysis.
	SetResume(ctx context.Context, userID domain.UserID, resume string) error
	GetByID(ctx context.Context, id domain.UserID) (domain.User, error)
	// TouchLastVisited updates LastVisitedAt to now. Called on each authenticated session.
	TouchLastVisited(ctx context.Context, userID domain.UserID) error
}

// CompanyService manages per-user company visibility.
type CompanyService interface {
	HideCompany(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) error
	ShowCompany(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) error
	// ListCompanies returns all active companies excluding those hidden by the user.
	// Used for job browse filtering.
	ListCompanies(ctx context.Context, userID domain.UserID) ([]domain.Company, error)
	// ListHiddenCompanies returns the full Company objects for companies the user
	// has hidden. Used to render the blocked-companies list on the settings page.
	ListHiddenCompanies(ctx context.Context, userID domain.UserID) ([]domain.Company, error)
	// IsCompanyHidden returns true if the user has hidden the specified company.
	// Efficient single-row check — avoids fetching the entire hidden list.
	IsCompanyHidden(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) (bool, error)
}

// JobCoachService provides LLM-powered resume analysis and optimization.
//
// Users upload a resume to their profile and can analyze it against a specific
// job posting for ATS optimization, fit analysis, and a tailored resume rewrite.
// They can also generate portfolio case studies from project descriptions.
type JobCoachService interface {
	// AnalyzeJob compares the user's resume against a job posting and returns
	// structured analysis: ATS keyword gaps, fit assessment, and an optimized
	// resume tailored to the role. Returns a cached result if available unless
	// refresh is true.
	AnalyzeJob(ctx context.Context, userID domain.UserID, jobID domain.JobID, refresh bool) (string, error)

	// GenerateCaseStudy expands a project description or resume bullet into a
	// structured portfolio case study (Problem → Process → Solution → Results → Learnings).
	GenerateCaseStudy(ctx context.Context, userID domain.UserID, projectDescription string) (string, error)

	// BuildAnalyzePrompt returns the raw system + user prompt for job analysis
	// without calling the LLM. Users can pipe this to their own tooling.
	// Requires a resume but does NOT require an LLM provider to be configured.
	BuildAnalyzePrompt(ctx context.Context, userID domain.UserID, jobID domain.JobID) (systemPrompt, userPrompt string, err error)

	// BuildCaseStudyPrompt returns the raw prompts for case study generation.
	BuildCaseStudyPrompt(ctx context.Context, userID domain.UserID, projectDescription string) (systemPrompt, userPrompt string, err error)
}

