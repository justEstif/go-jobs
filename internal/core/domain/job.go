package domain

import (
	"time"

	"github.com/google/uuid"
)

type JobID = uuid.UUID
type CompanyID = uuid.UUID
type UserID = uuid.UUID

// ATSType identifies which applicant tracking system hosts a company's job board.
type ATSType string

const (
	ATSGreenhouse ATSType = "greenhouse"
	ATSLever      ATSType = "lever"
	ATSAshby      ATSType = "ashby"
	ATSCustom     ATSType = "custom"
)

// ScrapeType controls how a company's job board is fetched.
type ScrapeType string

const (
	ScrapeHTTP     ScrapeType = "http"
	ScrapeHeadless ScrapeType = "headless"
)

// WorkplaceType describes the remote-work policy of a job posting.
type WorkplaceType string

const (
	WorkplaceRemote WorkplaceType = "remote"
	WorkplaceHybrid WorkplaceType = "hybrid"
	WorkplaceOnsite WorkplaceType = "onsite"
)

// EmploymentType describes the engagement type of a job posting.
type EmploymentType string

const (
	EmploymentFullTime EmploymentType = "fulltime"
	EmploymentPartTime EmploymentType = "parttime"
	EmploymentIntern   EmploymentType = "intern"
	EmploymentContract EmploymentType = "contract"
)

// RoleType is a broad functional category for a job posting.
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

// Seniority is the experience level associated with a job posting.
type Seniority string

const (
	SeniorityIntern Seniority = "intern"
	SeniorityJunior Seniority = "junior"
	SeniorityMid    Seniority = "mid"
	SenioritySenior Seniority = "senior"
	SeniorityStaff  Seniority = "staff"
	SeniorityLead   Seniority = "lead"
)

// EnrichmentSource tracks which pipeline tier produced a set of JobTags.
type EnrichmentSource string

const (
	EnrichmentATS   EnrichmentSource = "ats"
	EnrichmentRules EnrichmentSource = "rules"
	EnrichmentLLM   EnrichmentSource = "llm"
)

// LLMProvider identifies which LLM provider a user has configured.
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
	BoardToken string // platform-specific slug/token (board_token, company_slug, board_name)
	Active     bool
	CreatedAt  time.Time
}

// Job is a single job posting.
type Job struct {
	ID          JobID
	CompanyID   CompanyID
	CompanyName string // denormalised for display — set on read
	ExternalID  string // dedup key from source (platform posting ID)
	Title       string
	URL         string // direct apply link
	Location    string // raw location string from source
	Description string // plain text
	RawHTML     string // original HTML — stored for re-enrichment
	Source      ATSType
	FirstSeen   time.Time
	LastSeen    time.Time
	Active      bool
	Tags        *JobTags // nil if not yet enriched
}

// JobTags is the enriched structured metadata for a job.
type JobTags struct {
	JobID            JobID
	RoleType         RoleType
	Seniority        Seniority
	RemotePolicy     WorkplaceType
	LocationNorm     string
	Country          string // ISO 2-letter code ("US", "DE", "BR")
	TechStack        []string
	EnrichmentSource EnrichmentSource
	EnrichedAt       time.Time
}

// RawJob is the normalised intermediate type returned by JobScraper adapters.
// Platform differences are resolved inside each adapter before returning.
type RawJob struct {
	ExternalID  string
	Title       string
	URL         string // direct apply link
	Location    string // raw location string
	Description string // plain text
	RawHTML     string // original HTML
	FirstSeen   time.Time

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
	ID            UserID
	Email         string
	PasswordHash  string // bcrypt; never returned to callers outside auth
	LLMAPIKey     string // AES-256-GCM encrypted at rest; empty if not configured
	LLMProvider   LLMProvider
	LastVisitedAt *time.Time // updated on each session; nil on first visit
	CreatedAt     time.Time
}

// UserJob is a user's pipeline state for a specific job.
type UserJob struct {
	UserID    UserID
	JobID     JobID
	Status    ApplicationStatus
	StatusAt  time.Time  // when status last changed
	AppliedAt *time.Time // set when status first becomes Applied; never overwritten
	Notes     string     // freeform; empty if not set
}

// ScrapeStatus is the lifecycle state of a ScrapeRun.
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
//
// All slice fields use OR semantics within the field (e.g. Seniorities = [senior, mid]
// matches jobs tagged senior OR mid). TechStack is the exception: AND semantics —
// the job must mention all specified terms.
//
// URL encoding: repeated params (?seniority=senior&seniority=mid).
type SearchFilters struct {
	Query            string          // free text (title / company name)
	RoleTypes        []RoleType      // OR — match any
	Seniorities      []Seniority     // OR — match any
	RemotePolicy     []WorkplaceType // OR — match any
	Countries        []string        // OR — ISO 2-letter codes
	TechStack        []string        // AND — job must mention all terms
	CompanyIDs       []CompanyID     // OR — match any
	PostedWithinDays int             // only jobs first_seen within N days; <=0 disables
	Limit            int
	Offset           int
}

// UserSearchContext carries identity-dependent search options.
//
// When passed to JobRepository.Search, the adapter JOINs user_jobs to annotate
// each returned Job with the user's pipeline state, and optionally restricts
// results to jobs added since the user's last visit.
//
// Kept separate from SearchFilters so callers without a user context
// (e.g. anonymous browse, CLI non-authenticated) never touch these fields,
// and the implicit OnlyNew↔UserID precondition becomes structurally impossible to violate.
type UserSearchContext struct {
	UserID  UserID
	OnlyNew bool // if true, only jobs where first_seen > user.LastVisitedAt
}
