# go-jobs — Architecture

## Tech Stack

| Component       | Choice                                          | Rationale                                                                                                                                                                                    |
| --------------- | ----------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Language        | Go                                              | Single binary, great concurrency for scraping, matches your stack                                                                                                                            |
| HTTP router     | Chi                                             | Lightweight, idiomatic Go router; composable middleware                                                                                                                                      |
| Database        | PostgreSQL                                      | Your standard stack, handles structured job data + user state well                                                                                                                           |
| SQL             | sqlc + golang-migrate                           | Compile-time type-safe queries; versioned migrations                                                                                                                                         |
| LLM integration | openai-go + anthropic-sdk-go + generative-ai-go | Official provider SDKs. Per-user API key passed at client construction per call. Native structured output on all three. No framework layer — direct SDK calls behind the `JobEnricher` port. |
| Web UI          | templ + htmx + Tailwind CSS + DaisyUI           | Type-safe server-rendered templates with htmx for interactivity. Tailwind standalone binary — no Node required.                                                                              |
| CLI             | cobra                                           | Standard Go CLI patterns; coexists with Chi HTTP server in the same binary                                                                                                                   |
| Tooling         | mise                                            | Unified tool version management and task runner (dev, build, migrate, sqlc)                                                                                                                  |
| Deployment      | Dokku                                           | Self-hosted PaaS on a single VPS. Git-push deploy, manages Postgres as an addon.                                                                                                             |

## Data Model

```
companies
├── id            (uuid)
├── name          (text)
├── careers_url   (text)       -- career page to scrape
├── ats_type      (text)       -- "greenhouse" | "lever" | "ashby" | "custom"
├── scrape_type   (text)       -- "http" | "headless" (headless = post-MVP)
├── active        (bool)
└── created_at    (timestamp)

jobs
├── id            (uuid)
├── company_id    (fk → companies)
├── external_id   (text)       -- dedup key from source
├── title         (text)
├── url           (text)       -- direct link to apply
├── location      (text)       -- raw location string
├── description   (text)
├── source        (text)       -- "greenhouse" | "lever" | "ashby"
├── raw_html      (text)       -- original posting for re-processing
├── first_seen    (timestamp)  -- when we first scraped this posting; shown as "added X days ago"
├── last_seen     (timestamp)  -- last successful scrape; used to detect inactive postings
└── active        (bool)       -- false when posting disappears from source

scrape_runs
├── id            (uuid)
├── started_at    (timestamp)
├── finished_at   (timestamp, nullable)
├── status        (text)       -- "running" | "completed" | "failed"
├── jobs_added    (int)        -- new jobs inserted this run
├── jobs_updated  (int)        -- existing jobs whose last_seen was refreshed
├── jobs_removed  (int)        -- jobs marked inactive this run
└── error         (text, nullable)  -- error message if status = "failed"

job_tags (enriched structured metadata)
├── job_id             (fk → jobs)
├── role_type          (text)       -- "engineering" | "analyst" | "design" | "pm" | ...
├── seniority          (text)       -- "intern" | "junior" | "mid" | "senior" | "staff" | "lead"
├── remote_policy      (text)       -- "remote" | "hybrid" | "onsite"
├── location_norm      (text)       -- normalized location
├── country            (text)       -- "US" | "UK" | "DE" | ...
├── tech_stack         (text[])     -- ["go", "postgres", "kubernetes"]
├── enrichment_source  (text)       -- "ats" | "rules" | "llm"
└── enriched_at        (timestamp)

users
├── id                (uuid)
├── email             (text, unique)
├── password_hash     (text)              -- bcrypt
├── llm_api_key       (text)              -- user-provided, encrypted at rest
├── llm_provider      (text)              -- "openai" | "anthropic" | "google"
├── last_visited_at   (timestamp)         -- updated on each session; drives "new since last visit"
└── created_at        (timestamp)
-- email_verified_at (timestamp) -- post-MVP; email verification not required for MVP

user_jobs
├── user_id        (fk → users)
├── job_id         (fk → jobs)
├── status         (text)                 -- see ApplicationStatus below
├── status_at      (timestamp)            -- when status last changed
├── applied_at     (timestamp, nullable)  -- set when status first becomes "applied"; never overwritten
├── notes          (text, nullable)       -- freeform user notes
└── unique(user_id, job_id)

-- ApplicationStatus values (ordered pipeline):
-- "interested"   -- flagged as want-to-apply
-- "applied"      -- application submitted; applied_at captured
-- "interviewing" -- in active interview process
-- "offer"        -- received an offer
-- "rejected"     -- rejected by company
-- "withdrawn"    -- withdrawn by user

user_companies
├── user_id       (fk → users)
├── company_id    (fk → companies)
└── hidden        (bool, default false)  -- user opts out of a company
-- absence of a row = company is visible (opt-out model, default all visible)
```

## Ports

Ports are interfaces defined by the core. The core owns them — adapters implement them. Nothing in the core references a concrete technology.

### Driving ports (callers → core)

These define how the outside world uses the core. CLI, HTTP handlers, the scheduler, and tests all call through these interfaces.

| Port                 | Methods (sketch)                                                                                                                            | Callers                                    |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| `JobSearchService`   | `Search(filters, userCtx) ([]Job, error)`, `GetByID(id) (Job, error)`                                                                      | CLI (local + remote), HTML UI, JSON API    |
| `ApplicationService` | `SetStatus(userID, jobID, status)`, `SetNotes(userID, jobID, notes)`, `GetUserJob(userID, jobID)`, `ListByStatus(userID, status)`, `ListPipeline(userID)` | CLI (local + remote), HTML UI, JSON API   |
| `ScrapeService`      | `Run(ctx) error`, `SeedCompanies(ctx) error`, `LatestRun(ctx) (ScrapeRun, error)`                                                           | Scheduler adapter, CLI (local only)        |
| `AuthService`        | `Register(email, password)`, `Login(email, password) token`, `Logout(token)`, `Authenticate(token) (User, error)`                          | HTTP middleware, CLI (local + remote)      |
| `UserService`        | `SetLLMKey(userID, provider, key)`, `GetByID(userID) (User, error)`, `TouchLastVisited(userID)`                                             | HTML UI, CLI                               |
| `CompanyService`     | `HideCompany(userID, companyID)`, `ShowCompany(userID, companyID)`, `ListCompanies(userID) ([]Company, error)`                              | HTML UI                                    |

### Driven ports (core → outside world)

These define what the core needs from infrastructure. The core calls them; adapters implement them.

| Port                    | Methods (sketch)                                                                                                                               | Adapter(s)                                    |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| `JobRepository`         | `Upsert(companyID, job) (JobID, error)`, `GetByID(id) (Job, error)`, `GetByIDs(ids) ([]Job, error)`, `Search(filters, userCtx) ([]Job, error)`, `ListUnenriched(limit) ([]Job, error)`, `MarkInactive(companyID, activeIDs) error`, `SaveTags(tags) error` | Postgres (sqlc) |
| `CompanyRepository`     | `Upsert(company) (CompanyID, error)`, `ListActive() ([]Company, error)`, `GetByID(id) (Company, error)`, `GetByBoardToken(atsType, token) (Company, error)` | Postgres (sqlc)  |
| `UserRepository`        | `Create(email, passwordHash) (User, error)`, `GetByEmail(email) (User, error)`, `GetByID(id) (User, error)`, `Update(user) error`              | Postgres (sqlc)                               |
| `SessionRepository`     | `SaveToken(userID, token) error`, `DeleteToken(token) error`, `GetUserByToken(token) (User, error)`                                            | Postgres (sqlc)                               |
| `UserJobRepository`     | `Upsert(userJob) error`, `GetUserJob(userID, jobID) (UserJob, error)`, `ListByStatus(userID, status) ([]JobID, error)`, `ListAll(userID) ([]UserJob, error)` | Postgres (sqlc)              |
| `ScrapeRunRepository`   | `Create(run) error`, `Update(run) error`, `Latest() (ScrapeRun, error)`                                                                        | Postgres (sqlc)                               |
| `UserCompanyRepository` | `SetHidden(userID, companyID, hidden bool) error`, `ListHidden(userID) ([]CompanyID, error)`                                                   | Postgres (sqlc)                               |
| `JobScraper`            | `Scrape(company) ([]RawJob, error)`                                                                                                            | Greenhouse, Lever, Ashby adapters             |
| `JobEnricher`           | `Enrich(job) (JobTags, error)`                                                                                                                 | Tiered enrichment adapter (ATS → rules → LLM) |
| `CompanySeeder`         | `Seed(ctx) ([]Company, error)`                                                                                                                 | Simplify README parser                        |

### Port decision rationale

- `JobRepository.Upsert` — dedup key (`external_id + company_id`) is a business rule; lives in the core, executed by the adapter via upsert semantics
- `JobRepository.GetByIDs` — `ApplicationService.ListByStatus` hydrates `[]JobID` from `UserJobRepository` into `[]Job`; a bulk fetch avoids N+1 without requiring a JOIN across repository boundaries
- `JobScraper` — will be swapped/extended (new ATS platforms post-MVP, headless post-MVP); tested with fakes in isolation
- `JobEnricher` — decoupled from LLM; tiers 1–2 work without any external service; LLM tier is just another implementation detail of the adapter
- `ScrapeService` — owns the headless-skip rule and pipeline order; scheduler is a dumb timer that calls `ScrapeService.Run`
- `SessionRepository` — separated from `UserRepository` so token storage strategy (column vs. table, or JWT migration) can change independently of user record persistence; `AuthService` depends on both; composition root wires both to the same postgres adapter
- `AuthService` — driving port called by HTTP middleware and CLI to resolve a session token to a `User` before any business use case runs; depends on `UserRepository` + `SessionRepository`
- `CompanySeeder` — seeding company slugs from external READMEs is a driven port so the seeder can be replaced or faked in tests without touching `ScrapeService`

## System Overview

```
         DRIVING ADAPTERS                CORE                    DRIVEN ADAPTERS
         (call into core)           (the fortress)            (called by core)

    ┌─────────────────┐                                      ┌──────────────────────┐
    │   CLI (cobra)   │                                      │  Postgres adapter    │
    │   local mode    │──── AuthService ────────────────────▶│  (sqlc)              │
    │                 │──── JobSearchService ───────────────▶│                      │
    │  search/pipeline│──── ApplicationService ─────────────▶│  JobRepository       │
    │  scrape/enrich  │──── ScrapeService ──────────────────▶│  CompanyRepository   │
    │  serve          │                                      │  UserRepository      │
    └─────────────────┘                                      │  SessionRepository   │
                              ┌──────────────────┐           │  UserJobRepository   │
    ┌─────────────────┐       │                  │           │  ScrapeRunRepository │
    │   CLI (cobra)   │       │   C O R E        │           │  UserCompanyRepo     │
    │   remote mode   │       │                  │           └──────────────────────┘
    │  (httpclient    │──────▶│  AuthService     │
    │   adapter)      │       │  JobSearchService│           ┌──────────────────────┐
    │                 │       │  ApplicationSvc  │──────────▶│  Scraper adapters    │
    │  search/pipeline│       │  ScrapeService   │           │                      │
    │  login/logout   │       │  UserService     │           │  GreenhouseAdapter   │
    └─────────────────┘       │  CompanyService  │           │  LeverAdapter        │
                              │                  │           │  AshbyAdapter        │
    ┌─────────────────┐       └──────────────────┘           │                      │
    │  HTTP handlers  │              ▲                       │  impl: JobScraper    │
    │                 │              │                       └──────────────────────┘
    │  HTML routes    │──────────────┤
    │  (templ + htmx) │              │                       ┌──────────────────────┐
    │                 │              │                       │  Enrichment adapter  │
    │  /api/v1/ routes│──────────────┘                      │                      │
    │  (JSON API)     │                                      │  tier 1: ATS fields  │
    └─────────────────┘                                      │  tier 2: rules       │
                                                             │  tier 3: LLM/GenKit  │
    ┌─────────────────┐                                      │                      │
    │  HTTP middleware │                                      │  impl: JobEnricher   │
    │  (session auth  │                                      └──────────────────────┘
    │   bearer auth)  │
    └─────────────────┘

    ┌─────────────────┐
    │  Scheduler      │
    │  (time.Ticker)  │──── ScrapeService.Run
    └─────────────────┘
```

**Dependency rule:** arrows point inward. Adapters depend on core interfaces. The core depends on nothing outside itself.

**Two driving adapters for HTTP:** The HTML handlers (`/`) and the JSON API (`/api/v1/`) are two separate driving adapters that call the same core services. The HTML adapter renders templ templates; the JSON API returns structured data for the CLI and other clients. Neither knows about the other.

**Two CLI modes:** The CLI cobra commands are wired differently depending on whether a base URL is configured. In local mode the composition root injects in-process service implementations that talk directly to Postgres. In remote mode it injects `httpclient` adapters that call the `/api/v1/` endpoints over HTTP — no database connection required.

**Composition root** (`cmd/jobs/main.go`): the only place that knows about all concrete types. Detects the mode (local vs. remote), instantiates the appropriate adapters, wires them into core services, and hands services to delivery adapters.

## Folder Structure

```
go-jobs/
├── cmd/
│   └── jobs/
│       └── main.go               -- composition root: wire adapters → core → delivery
│
├── internal/
│   ├── core/                     -- the fortress; no imports from adapters/
│   │   ├── domain/               -- pure domain types (Job, Company, User, JobTags, ...)
│   │   ├── ports/                -- interface definitions (driving + driven)
│   │   │   ├── driving.go        -- JobSearchService, ApplicationService, ScrapeService, AuthService, ...
│   │   │   └── driven.go         -- JobRepository, JobScraper, JobEnricher, SessionRepository, ...
│   │   └── services/             -- use case implementations
│   │       ├── search.go         -- JobSearchService impl
│   │       ├── application.go    -- ApplicationService impl
│   │       ├── scrape.go         -- ScrapeService impl (owns pipeline order + headless-skip rule)
│   │       ├── user.go           -- UserService impl
│   │       └── company.go        -- CompanyService impl
│   │
│   ├── adapters/
│   │   ├── postgres/             -- driven: JobRepository, CompanyRepository, UserRepository, ...
│   │   │   ├── queries/          -- sqlc-generated (DO NOT import outside this package)
│   │   │   ├── job_repo.go       -- maps db.Job → core/domain.Job
│   │   │   ├── company_repo.go   -- maps db.Company → core/domain.Company
│   │   │   ├── user_repo.go      -- maps db.User → core/domain.User
│   │   │   ├── userjob_repo.go
│   │   │   └── usercompany_repo.go
│   │   │
│   │   ├── scrapers/             -- driven: JobScraper implementations
│   │   │   ├── greenhouse.go
│   │   │   ├── lever.go
│   │   │   └── ashby.go
│   │   │
│   │   ├── enrichment/           -- driven: JobEnricher (tiered pipeline)
│   │   │   ├── enricher.go       -- orchestrates tiers 1→2→3
│   │   │   ├── ats.go            -- tier 1: extract from raw ATS payload
│   │   │   ├── rules.go          -- tier 2: keyword/regex matching
│   │   │   └── llm.go            -- tier 3: GenKit LLM call
│   │   │
│   │   ├── http/                 -- driving: HTML routes (templ + htmx) + JSON API (/api/v1/)
│   │   │   ├── auth.go           -- HTML: /register, /login, /logout
│   │   │   ├── jobs.go           -- HTML: /, /jobs/{id}
│   │   │   ├── tracker.go        -- HTML: /jobs/{id}/interested|apply|status|notes
│   │   │   ├── pipeline.go       -- HTML: /pipeline
│   │   │   ├── companies.go      -- HTML: /companies
│   │   │   ├── api/              -- JSON API driving adapter (package api)
│   │   │   │   ├── handler.go    -- Handler struct, writeJSON/writeError helpers
│   │   │   │   ├── auth.go       -- POST /api/v1/auth/{register,login,logout,me}
│   │   │   │   ├── jobs.go       -- GET  /api/v1/jobs, /jobs/interested, /jobs/applied
│   │   │   │   ├── tracker.go    -- POST /api/v1/jobs/{id}/{interested,apply,status,notes}
│   │   │   │   └── pipeline.go   -- GET  /api/v1/pipeline
│   │   │   └── middleware/
│   │   │       ├── session.go    -- scs session management + UserID context helpers
│   │   │       ├── auth.go       -- OptionalAuth, RequireAuth (session-based, for HTML routes)
│   │   │       ├── bearer.go     -- BearerAuth middleware (token-based, for /api/v1/ routes)
│   │   │       └── csrf.go
│   │   │
│   │   ├── httpclient/           -- driving: CLI remote mode — implements ports over HTTP
│   │   │   ├── client.go         -- shared Client (baseURL, token, http.Client) + request helpers
│   │   │   ├── auth.go           -- AuthClient: implements AuthService + SessionRepository
│   │   │   ├── search.go         -- SearchClient: implements JobSearchService
│   │   │   └── application.go    -- ApplicationClient: implements ApplicationService
│   │   │
│   │   └── scheduler/            -- driving: time.Ticker → ScrapeService.Run
│   │       └── scheduler.go
│   │
│   └── cli/                      -- driving: cobra commands → core services (local or remote)
│       ├── root.go               -- Services struct, NewRootCmd, --base-url persistent flag
│       ├── search.go
│       ├── tracker.go            -- interested, apply, status, notes, applied, pipeline
│       ├── auth.go               -- register, login, logout
│       ├── scrape.go             -- local only
│       ├── enrich.go             -- local only
│       ├── serve.go              -- local only
│       └── token.go              -- XDG token file helpers
│
├── components/                   -- templ templates
├── migrations/                   -- golang-migrate SQL files
├── static/                       -- generated CSS (gitignored)
├── styles/                       -- Tailwind input.css
└── mise.toml
```

Key constraints:

- `internal/core/` has **zero imports** from `internal/adapters/`. The Go import graph enforces the dependency rule at compile time.
- `internal/adapters/postgres/queries/` is sqlc-generated and **never imported outside `internal/adapters/postgres/`**. The postgres adapter maps sqlc types (`db.Job`, `db.Company`, etc.) to core domain types before returning them. sqlc types are an implementation detail of the adapter, not domain types.

## CLI Remote Mode

The CLI can operate in two modes depending on whether a base URL is configured.

### Mode detection (composition root)

`cmd/jobs/main.go` pre-scans `os.Args` and `BASE_URL` env before booting any adapters:

1. `--base-url=<url>` flag in `os.Args` (highest precedence)
2. `--base-url <url>` flag in `os.Args`
3. `BASE_URL` environment variable
4. Empty → local mode

### Command split

| Command | Local mode | Remote mode |
|---|---|---|
| `serve` | ✅ starts the HTTP server | ❌ local only |
| `scrape` | ✅ runs scrape pipeline | ❌ local only |
| `enrich` | ✅ runs enrichment pipeline | ❌ local only |
| `search` | in-process service | `GET /api/v1/jobs` |
| `interested` | in-process service | `GET /api/v1/jobs/interested` |
| `applied` | in-process service | `GET /api/v1/jobs/applied` |
| `pipeline` | in-process service | `GET /api/v1/pipeline` |
| `apply` | in-process service | `POST /api/v1/jobs/{id}/apply` |
| `status` | in-process service | `POST /api/v1/jobs/{id}/status` |
| `notes` | in-process service | `POST /api/v1/jobs/{id}/notes` |
| `register` | in-process service | `POST /api/v1/auth/register` |
| `login` | in-process service | `POST /api/v1/auth/login` |
| `logout` | in-process service | `POST /api/v1/auth/logout` |

`serve`, `scrape`, and `enrich` always use local mode regardless of base URL — they require direct DB access.

### Auth flow (remote mode)

1. `go-jobs login --base-url https://myjobs.io` → `POST /api/v1/auth/login` → token written to `~/.config/go-jobs/token`
2. Subsequent commands read the token file and pass `Authorization: Bearer <token>` on every API request
3. `go-jobs logout` → `POST /api/v1/auth/logout` → server invalidates the token; local file deleted

### JSON API routes

All routes are under `/api/v1/`. Public routes require no auth; protected routes require `Authorization: Bearer <token>`.

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | public | Create account |
| `POST` | `/api/v1/auth/login` | public | Returns `{"token": "..."}` |
| `GET`  | `/api/v1/auth/me` | Bearer | Returns current user |
| `POST` | `/api/v1/auth/logout` | Bearer | Invalidates token |
| `GET`  | `/api/v1/jobs` | optional Bearer | Search jobs (query params mirror CLI flags) |
| `GET`  | `/api/v1/jobs/interested` | Bearer | List interested jobs |
| `GET`  | `/api/v1/jobs/applied` | Bearer | List applied jobs |
| `POST` | `/api/v1/jobs/{id}/interested` | Bearer | Mark as interested |
| `POST` | `/api/v1/jobs/{id}/apply` | Bearer | Mark as applied |
| `POST` | `/api/v1/jobs/{id}/status` | Bearer | Set status (`{"status": "..."}`) |
| `POST` | `/api/v1/jobs/{id}/notes` | Bearer | Set notes (`{"notes": "..."}`) |
| `GET`  | `/api/v1/pipeline` | Bearer | All tracked jobs grouped by status |

### httpclient adapter

`internal/adapters/httpclient/` implements the same driving port interfaces (`AuthService`, `JobSearchService`, `ApplicationService`, `SessionRepository`) as the in-process services, but over HTTP. The CLI cobra commands are identical in both modes — only the injected implementations differ. The composition root is the only place that knows which mode is active.

## Data Sources

The company list is not hand-curated. Companies are seeded by parsing the Simplify/Pitt CSC community job repos, which embed direct ATS apply links in their README tables. This gives ~296 startup/growth-stage companies across all three platforms with zero manual curation and zero Workday noise.

**Seed sources (parsed on first run, re-parsed on each scrape cycle to pick up new companies):**

- `https://raw.githubusercontent.com/SimplifyJobs/Summer2026-Internships/dev/README.md`
- `https://raw.githubusercontent.com/SimplifyJobs/New-Grad-Positions/dev/README.md`

Company slugs/tokens are extracted by URL pattern per platform and upserted into the `companies` table.

| ATS Platform    | `ats_type`   | `scrape_type` | Seed count | Notes                                              |
| --------------- | ------------ | ------------- | ---------- | -------------------------------------------------- |
| Greenhouse      | `greenhouse` | `http`        | ~154       | JSON API at `boards-api.greenhouse.io/v1/boards/…` |
| Lever           | `lever`      | `http`        | ~49        | Public v0 API at `api.lever.co/v0/postings/…`      |
| Ashby           | `ashby`      | `http`        | ~93        | Public API at `api.ashbyhq.com/posting-api/…`      |
| Custom (SSR)    | `custom`     | `http`        | —          | Post-MVP — case-by-case                            |
| Custom (JS SPA) | `custom`     | `headless`    | —          | Post-MVP — requires headless browser               |
| HackerNews      | —            | —             | —          | Post-MVP — low priority given ATS coverage         |

MVP ships Greenhouse, Lever, and Ashby only. `scrape_type = "headless"` companies are stored but skipped with a log warning.

## Enrichment Pipeline

Enrichment is a `JobEnricher` port — the core doesn't care how tags are extracted, only that they arrive in a standard shape. The pipeline runs in tiers, cheapest first:

### Tier 1: ATS metadata (free, instant)

Greenhouse, Lever, and Ashby APIs return structured fields in the raw scrape payload — `department`, `offices`, `job_level`, `location`, and similar. Extract what's available before touching any external service. Covers 60-70% of what `job_tags` needs at zero cost.

### Tier 2: Rule-based (free, fast)

Pattern match on job title and description for fields ATS metadata doesn't cover:

- Seniority keywords in title (`senior`, `staff`, `lead`, `intern`)
- Tech stack terms in description against a known keyword list
- Remote policy phrases (`remote`, `hybrid`, `on-site`)

### Tier 3: LLM (optional, user-provided)

For fields that are still blank or low-confidence after tiers 1–2, an LLM call fills the gaps. LLM enrichment is **optional** — the system is fully functional without it. Users who want richer tag quality supply their own API key.

1. Job posting scraped → raw payload + text stored
2. Background worker picks up un-enriched jobs
3. Tier 1: extract from ATS metadata fields
4. Tier 2: rule-based extraction on title + description
5. Tier 3 (if LLM key configured): GenKit call for remaining gaps
6. Tags stored in `job_tags` with `enrichment_source`
7. Failed LLM enrichments retried with backoff; tiers 1–2 never fail

**`enrichment_source` field on `job_tags`:** tracks which tier produced each tag (`"ats"`, `"rules"`, `"llm"`). Useful for debugging tag quality and deciding what to re-enrich.

**LLM cost:** ~$0.01–0.05 per posting depending on model. For 500 companies × 10 jobs, that's ~$50–250 per full cycle — but tiers 1–2 reduce the LLM surface significantly. Cache aggressively — only re-enrich when posting content changes.

**User-provided API keys:** Each user supplies their own LLM API key and chooses their provider. The app stores the key encrypted at rest and passes it to GenKit at enrichment time.

- No shared API cost — each user pays for their own enrichment
- GenKit's provider-agnostic interface means switching models is a config change, not a code change
- Key is validated at setup time with a clear error when missing or invalid

**Supported LLM providers (MVP):** OpenAI, Google Gemini, Anthropic. Each is a ~60-line adapter file in `internal/adapters/enrichment/`. Adding a new provider = adding one file.

**Rate limiting:** Scraping uses conservative delays between requests. LLM enrichment uses exponential backoff on 429s, implemented per-provider adapter.
