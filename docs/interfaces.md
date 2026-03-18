# go-jobs — Interface Definitions

This is the contract layer. All code is written against these interfaces — adapters implement them, services consume them. Nothing here references a concrete technology.

Go module path assumed: `github.com/justestif/go-jobs`

---

## Domain Types (`internal/core/domain`)

```go
// job.go

type JobID = uuid.UUID
type CompanyID = uuid.UUID
type UserID = uuid.UUID

type ATSType string

const (
    ATSGreenhouse ATSType = "greenhouse"
    ATSLever      ATSType = "lever"
    ATSAshby      ATSType = "ashby"
    ATSCustom     ATSType = "custom"
)

type ScrapeType string

const (
    ScrapeHTTP     ScrapeType = "http"
    ScrapeHeadless ScrapeType = "headless"
)

type WorkplaceType string

const (
    WorkplaceRemote WorkplaceType = "remote"
    WorkplaceHybrid WorkplaceType = "hybrid"
    WorkplaceOnsite WorkplaceType = "onsite"
)

type EmploymentType string

const (
    EmploymentFullTime  EmploymentType = "fulltime"
    EmploymentPartTime  EmploymentType = "parttime"
    EmploymentIntern    EmploymentType = "intern"
    EmploymentContract  EmploymentType = "contract"
)

type RoleType string

const (
    RoleEngineering RoleType = "engineering"
    RoleDesign      RoleType = "design"
    RoleProduct     RoleType = "product"
    RoleData        RoleType = "data"
    RoleMarketing   RoleType = "marketing"
    RoleSales       RoleType = "sales"
    RoleOperations  RoleType = "operations"
    RoleFinance     RoleType = "finance"
    RoleOther       RoleType = "other"
)

type Seniority string

const (
    SeniorityIntern  Seniority = "intern"
    SeniorityJunior  Seniority = "junior"
    SeniorityMid     Seniority = "mid"
    SenioritySenior  Seniority = "senior"
    SeniorityStaff   Seniority = "staff"
    SeniorityLead    Seniority = "lead"
)

type EnrichmentSource string

const (
    EnrichmentATS   EnrichmentSource = "ats"
    EnrichmentRules EnrichmentSource = "rules"
    EnrichmentLLM   EnrichmentSource = "llm"
)

type LLMProvider string

const (
    LLMOpenAI    LLMProvider = "openai"
    LLMAnthropic LLMProvider = "anthropic"
    LLMGoogle    LLMProvider = "google"
)

// Company is a company whose job board we scrape.
type Company struct {
    ID         CompanyID
    Name       string
    CareersURL string
    ATSType    ATSType
    ScrapeType ScrapeType
    BoardToken string    // platform-specific slug/token (board_token, company_slug, board_name)
    Active     bool
    CreatedAt  time.Time
}

// Job is a single job posting.
type Job struct {
    ID          JobID
    CompanyID   CompanyID
    CompanyName string    // denormalised for display — set on read
    ExternalID  string    // dedup key from source (platform posting ID)
    Title       string
    URL         string    // direct apply link
    Location    string    // raw location string from source
    Description string    // plain text
    RawHTML     string    // original HTML — stored for re-enrichment
    Source      ATSType
    FirstSeen   time.Time
    LastSeen    time.Time
    Active      bool
    Tags        *JobTags  // nil if not yet enriched
}

// JobTags is the enriched structured metadata for a job.
type JobTags struct {
    JobID             JobID
    RoleType          RoleType
    Seniority         Seniority
    RemotePolicy      WorkplaceType
    LocationNorm      string
    Country           string           // ISO 2-letter code ("US", "DE", "BR")
    TechStack         []string
    EnrichmentSource  EnrichmentSource
    EnrichedAt        time.Time
}

// RawJob is the normalised intermediate type returned by JobScraper adapters.
// Platform differences are resolved inside each adapter before returning.
type RawJob struct {
    ExternalID     string
    Title          string
    URL            string        // direct apply link
    Location       string        // raw location string
    Description    string        // plain text
    RawHTML        string        // original HTML
    FirstSeen      time.Time

    // ATS metadata — zero value means unknown/unavailable
    Department     string        // maps to role_type
    Country        string        // ISO 2-letter if available
    WorkplaceType  WorkplaceType // empty string = unknown
    EmploymentType EmploymentType
}

// ApplicationStatus represents where a job is in the user's pipeline.
type ApplicationStatus string

const (
    StatusInterested   ApplicationStatus = "interested"
    StatusApplied      ApplicationStatus = "applied"
    StatusInterviewing ApplicationStatus = "interviewing"
    StatusOffer        ApplicationStatus = "offer"
    StatusRejected     ApplicationStatus = "rejected"
    StatusWithdrawn    ApplicationStatus = "withdrawn"
)

// IsTerminal returns true for states that close the application loop.
func (s ApplicationStatus) IsTerminal() bool {
    return s == StatusOffer || s == StatusRejected || s == StatusWithdrawn
}

// User is an authenticated identity.
type User struct {
    ID              UserID
    Email           string
    PasswordHash    string      // bcrypt; never returned to callers outside auth
    LLMAPIKey       string      // AES-256-GCM encrypted at rest; empty if not configured
    LLMProvider     LLMProvider
    LastVisitedAt   *time.Time  // updated on each session; nil on first visit
    CreatedAt       time.Time
}

// UserJob is a user's pipeline state for a specific job.
type UserJob struct {
    UserID    UserID
    JobID     JobID
    Status    ApplicationStatus
    StatusAt  time.Time   // when status last changed
    AppliedAt *time.Time  // set when status first becomes Applied; never overwritten
    Notes     string      // freeform; empty if not set
}

type ScrapeStatus string

const (
    ScrapeStatusRunning   ScrapeStatus = "running"
    ScrapeStatusCompleted ScrapeStatus = "completed"
    ScrapeStatusFailed    ScrapeStatus = "failed"
)

// ScrapeRun is a record of a single scrape pipeline execution.
type ScrapeRun struct {
    ID          uuid.UUID
    StartedAt   time.Time
    FinishedAt  *time.Time
    Status      ScrapeStatus
    JobsAdded   int
    JobsUpdated int
    JobsRemoved int
    Error       string // non-empty if Status = ScrapeStatusFailed
}

// SearchFilters defines the structural filter dimensions for job search.
// All slice fields use OR semantics within the field (e.g. seniority=senior OR mid).
// TechStack is the exception: AND semantics — job must mention all specified terms.
// URL encoding: repeated params (?seniority=senior&seniority=mid).
type SearchFilters struct {
    Query        string            // free text (title / company name)
    RoleTypes    []RoleType        // OR — match any
    Seniorities  []Seniority       // OR — match any
    RemotePolicy []WorkplaceType   // OR — match any
    Countries    []string          // OR — ISO 2-letter codes
    TechStack    []string          // AND — job must mention all terms
    CompanyIDs   []CompanyID       // OR — match any
    Limit        int
    Offset       int
}

// UserSearchContext carries identity-dependent search options.
// When passed to JobRepository.Search, the adapter JOINs user_jobs to annotate
// each returned Job with the user's pipeline state, and optionally restricts
// results to jobs added since the user's last visit.
// Kept separate from SearchFilters so callers without a user context
// (e.g. anonymous browse, CLI non-authenticated) never touch these fields,
// and the implicit OnlyNew↔UserID precondition becomes structurally impossible to violate.
type UserSearchContext struct {
    UserID  UserID
    OnlyNew bool // if true, only jobs where first_seen > user.LastVisitedAt
}
```

---

## Driving Ports (`internal/core/ports/driving.go`)

Interfaces the outside world calls into the core. CLI, HTTP handlers, and the scheduler all depend on these — never on concrete implementations.

```go
// JobSearchService handles job discovery and browsing.
type JobSearchService interface {
    // Search returns jobs matching filters. If userCtx is non-nil, results are
    // annotated with the user's pipeline state and OnlyNew filtering is applied.
    Search(ctx context.Context, filters SearchFilters, userCtx *UserSearchContext) ([]Job, error)
    GetByID(ctx context.Context, id JobID) (Job, error)
}

// ApplicationService manages a user's pipeline state for jobs.
type ApplicationService interface {
    // SetStatus transitions a job to the given status.
    // Business rules enforced here:
    //   - Setting Applied also sets Interested (if not already) and captures AppliedAt.
    //   - SetStatus to Interested on an already-Applied job is a no-op (can't go backwards).
    SetStatus(ctx context.Context, userID UserID, jobID JobID, status ApplicationStatus) error
    SetNotes(ctx context.Context, userID UserID, jobID JobID, notes string) error
    GetUserJob(ctx context.Context, userID UserID, jobID JobID) (UserJob, error)
    // ListByStatus returns jobs in a given status, annotated with UserJob state.
    ListByStatus(ctx context.Context, userID UserID, status ApplicationStatus) ([]Job, error)
    // ListPipeline returns all tracked jobs for a user, grouped by status.
    ListPipeline(ctx context.Context, userID UserID) (map[ApplicationStatus][]Job, error)
}

// ScrapeService orchestrates the full scrape → enrich → persist pipeline.
// The scheduler calls Run; the service owns the pipeline order and business rules
// (e.g. skip headless companies, dedup on upsert).
type ScrapeService interface {
    Run(ctx context.Context) error
    // SeedCompanies fetches company slugs from the Simplify README sources
    // and upserts them into the company list.
    SeedCompanies(ctx context.Context) error
    // LatestRun returns the most recent ScrapeRun record (for UI status display).
    LatestRun(ctx context.Context) (ScrapeRun, error)
}

// AuthService manages user registration and authentication.
// Depends on UserRepository (user records) and SessionRepository (tokens).
// Authenticate is called by HTTP middleware and CLI on every request to resolve
// a session cookie/file token to a User — it is the composition point between
// the two repositories.
type AuthService interface {
    Register(ctx context.Context, email, password string) (User, error)
    Login(ctx context.Context, email, password string) (token string, err error)
    Logout(ctx context.Context, token string) error
    // Authenticate validates a session token and returns the associated user.
    Authenticate(ctx context.Context, token string) (User, error)
}

// UserService manages user settings.
type UserService interface {
    SetLLMKey(ctx context.Context, userID UserID, provider LLMProvider, apiKey string) error
    GetByID(ctx context.Context, id UserID) (User, error)
    // TouchLastVisited updates LastVisitedAt to now. Called on each authenticated session.
    TouchLastVisited(ctx context.Context, userID UserID) error
}

// CompanyService manages per-user company visibility.
type CompanyService interface {
    HideCompany(ctx context.Context, userID UserID, companyID CompanyID) error
    ShowCompany(ctx context.Context, userID UserID, companyID CompanyID) error
    ListCompanies(ctx context.Context, userID UserID) ([]Company, error) // excludes hidden
}
```

---

## Driven Ports (`internal/core/ports/driven.go`)

Interfaces the core calls out to. Postgres adapters, scrapers, and the enrichment pipeline all implement these.

```go
// JobRepository persists and retrieves job postings.
type JobRepository interface {
    // Upsert inserts or updates a job by (company_id, external_id).
    // Dedup key is a business rule — enforced here, not in the scraper adapter.
    Upsert(ctx context.Context, companyID CompanyID, job RawJob) (JobID, error)
    GetByID(ctx context.Context, id JobID) (Job, error)
    // GetByIDs fetches multiple jobs by ID in a single query. Order of results
    // is not guaranteed to match the order of ids. Used by ApplicationService
    // to hydrate job lists returned from UserJobRepository.ListByStatus.
    GetByIDs(ctx context.Context, ids []JobID) ([]Job, error)
    // Search returns jobs matching filters. If userCtx is non-nil, results are
    // annotated with the user's pipeline state (avoids N+1) and OnlyNew
    // filtering is applied against userCtx.UserID's LastVisitedAt.
    Search(ctx context.Context, filters SearchFilters, userCtx *UserSearchContext) ([]Job, error)
    // ListUnenriched returns jobs without a job_tags row, up to limit.
    ListUnenriched(ctx context.Context, limit int) ([]Job, error)
    // MarkInactive sets active=false for jobs no longer present at source.
    MarkInactive(ctx context.Context, companyID CompanyID, activeExternalIDs []string) error
    SaveTags(ctx context.Context, tags JobTags) error
}

// CompanyRepository persists and retrieves companies.
type CompanyRepository interface {
    Upsert(ctx context.Context, company Company) (CompanyID, error)
    ListActive(ctx context.Context) ([]Company, error)
    GetByID(ctx context.Context, id CompanyID) (Company, error)
    GetByBoardToken(ctx context.Context, atsType ATSType, boardToken string) (Company, error)
}

// UserRepository persists and retrieves user records.
// Concerned only with user identity and profile data — not with session tokens.
type UserRepository interface {
    Create(ctx context.Context, email, passwordHash string) (User, error)
    GetByEmail(ctx context.Context, email string) (User, error)
    GetByID(ctx context.Context, id UserID) (User, error)
    Update(ctx context.Context, user User) error
}

// SessionRepository manages opaque session tokens used for authentication.
// Separated from UserRepository so that changes to the auth mechanism
// (e.g. moving tokens from a column to a sessions table, or adopting JWT)
// do not require touching user persistence logic.
// AuthService depends on both UserRepository and SessionRepository.
// The composition root may wire both to the same postgres adapter package.
type SessionRepository interface {
    // SaveToken associates token with userID. Overwrites any existing token for this user.
    SaveToken(ctx context.Context, userID UserID, token string) error
    // DeleteToken invalidates a token. No-op if the token does not exist.
    DeleteToken(ctx context.Context, token string) error
    // GetUserByToken returns the User associated with token, or an error if
    // the token is invalid or expired. Called on every authenticated request.
    GetUserByToken(ctx context.Context, token string) (User, error)
}

// UserJobRepository manages per-user pipeline state.
type UserJobRepository interface {
    // Upsert writes all fields from userJob to the store.
    // Callers must read the existing UserJob first (via GetUserJob) before calling
    // Upsert to avoid overwriting fields they did not intend to change — this is
    // a deliberate read-modify-write contract, not a partial-update mechanism.
    // Semantics:
    //   - Status and StatusAt are always overwritten.
    //   - Notes: an empty string clears existing notes. Preserve by reading first.
    //   - AppliedAt: set by the adapter on the first transition to StatusApplied
    //     and never overwritten, regardless of the value in userJob.AppliedAt.
    //     Callers do not need to manage AppliedAt.
    Upsert(ctx context.Context, userJob UserJob) error
    GetUserJob(ctx context.Context, userID UserID, jobID JobID) (UserJob, error)
    ListByStatus(ctx context.Context, userID UserID, status ApplicationStatus) ([]JobID, error)
    ListAll(ctx context.Context, userID UserID) ([]UserJob, error)
}

// ScrapeRunRepository persists scrape run records.
type ScrapeRunRepository interface {
    Create(ctx context.Context, run ScrapeRun) error
    Update(ctx context.Context, run ScrapeRun) error
    Latest(ctx context.Context) (ScrapeRun, error)
}

// UserCompanyRepository manages per-user company visibility preferences.
type UserCompanyRepository interface {
    SetHidden(ctx context.Context, userID UserID, companyID CompanyID, hidden bool) error
    ListHidden(ctx context.Context, userID UserID) ([]CompanyID, error)
}

// JobScraper fetches raw job postings from a single ATS platform.
// Each ATS platform (Greenhouse, Lever, Ashby) has its own adapter.
// The adapter is responsible for normalising platform-specific fields
// into RawJob before returning — the core receives a clean, platform-agnostic struct.
type JobScraper interface {
    Scrape(ctx context.Context, company Company) ([]RawJob, error)
}

// JobEnricher extracts structured tags from a raw job posting.
// The tiered enrichment adapter (ATS metadata → rules → LLM) implements this.
// The core does not know which tier produced the result.
type JobEnricher interface {
    Enrich(ctx context.Context, job Job) (JobTags, error)
}

// CompanySeeder discovers company slugs/tokens from external sources
// (e.g. parsing the Simplify README files) and returns them as Company values
// ready for upsert. Does not persist — the ScrapeService calls this then
// hands results to CompanyRepository.Upsert.
type CompanySeeder interface {
    Seed(ctx context.Context) ([]Company, error)
}
```

---

## Notes

**On `RawJob` vs `Job`:**
`RawJob` is what scrapers return — a flat struct with only what the ATS gives us. `Job` is what the core works with after persistence — it has an ID, timestamps, and optionally enriched `Tags`. Adapters never see `Job`; the core never builds `Job` manually.

**On error handling:**
All interface methods return `error`. Callers decide whether a scraper error is fatal or loggable. `ScrapeService.Run` logs per-company errors and continues — a single failing scraper does not abort the pipeline.

**On `ctx`:**
Every method takes `context.Context` as the first argument. Adapters respect cancellation. Timeouts are set by the caller (scheduler, HTTP handler), not the interface.

**On `UserSearchContext`:**
When passed to `JobRepository.Search` (and `JobSearchService.Search`), the adapter JOINs `user_jobs` to annotate each returned `Job` with the user's pipeline state — avoiding N+1. When nil, results are unannotated (anonymous browse, unauthenticated CLI). `OnlyNew` lives here rather than in `SearchFilters` because it only makes sense when there is a user context; collocating them makes the dependency structurally obvious. The service reads `LastVisitedAt` from the user before building the `UserSearchContext`, then calls `UserService.TouchLastVisited` after the query returns — so the "new" window is jobs added since the previous visit, not this one.

**On `ApplicationService.SetStatus` and the pipeline order:**
The service enforces pipeline business rules: setting `StatusApplied` captures `AppliedAt` (first time only); setting `StatusInterested` on an already-applied job is a no-op. Terminal states (`offer`, `rejected`, `withdrawn`) can still be updated to each other (e.g. accepting an offer after initially declining). The repo's `Upsert` is dumb — it applies what it's given; the service owns the rules.

**On auth:**
`AuthService.Login` returns a random opaque token persisted via `SessionRepository`. The token is passed as a session cookie (web) or stored in a local config file (CLI). `AuthService.Authenticate` is called by HTTP middleware on every request to resolve the token to a `User` — it calls `SessionRepository.GetUserByToken` which returns the full `User` directly. Email verification is post-MVP — `Register` creates the user and immediately allows login.

`SessionRepository` is intentionally separate from `UserRepository` so that changes to the token storage strategy (column vs. dedicated table, or adopting JWT) do not require touching user persistence code. The composition root may wire both to the same postgres adapter package.

**On `ScrapeService.SeedCompanies`:**
Calls `CompanySeeder.Seed` → `CompanyRepository.Upsert`. Idempotent — safe to call on every scrape cycle. New companies appear automatically as the Simplify READMEs update.

**On `UserService.TouchLastVisited`:**
Called after every authenticated page load / CLI search. The previous value of `LastVisitedAt` is what drives `OnlyNew` filtering — read it before updating.
