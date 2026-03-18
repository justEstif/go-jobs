# go-jobs — MVP

## What

A self-hosted job aggregator that pulls from startup ATS platforms (Greenhouse, Lever, Ashby), with fast filters, a CLI interface, a web UI, and a lightweight application tracker. LLM enrichment tags and classifies postings so you can search by what actually matters.

### Core Features

- **Job aggregation** — Scrape job postings from startup ATS platforms (Greenhouse, Lever, Ashby). Companies are discovered automatically from each platform — no manual curation needed
- **LLM enrichment** — Tag postings with structured metadata (role type, seniority, tech stack, remote/hybrid/onsite, location) using raw provider SDKs (OpenAI, Anthropic, Gemini). Optional — system works without it.
- **Filters** — Role type, location (including US/non-US), seniority, tech stack, company. All categorical filters are multi-select (OR logic). Filter state lives in the URL (web) and CLI flags (terminal).
- **Application tracker** — Full pipeline: interested → applied → interviewing → offer / rejected / withdrawn. Applied auto-captures date. Notes field per job. Applying automatically sets interested.
- **CLI** — JSON output, agent-friendly. Search, filter, track. Output includes tracker state and `first_seen` for pipeline management.
- **Web UI** — Browse, filter, track. URL-based filter state (shareable, bookmarkable). Shows "added X days ago" and highlights jobs added since last visit.
- **Auth** — Email + password login. Email verification post-MVP.

### What It Is NOT (MVP)

- Not an auto-applier or form filler
- Not a resume builder (though a user may use job data to tailor resumes externally)
- Not a cover letter generator
- Not an email parser or inbox integration
- Not a LinkedIn scraper or recruiter finder
- Not an MCP server (post-MVP)
- No preference learning or query suggestions (post-MVP — revisit once enrichment pipeline is stable)

## Why

### The Problem

EchoJobs is the best job board for engineers, but it has real friction:

- **Paid** — subscription required
- **Unreliable** — goes down unpredictably
- **Noisy** — 5,000+ companies including Workday-style enterprise ATS postings that aren't relevant if you're targeting startups and product companies
- **Clunky filters** — you don't trust them to surface what you actually want

Application tracking means context-switching to Simplify, Notion, or a spreadsheet — none connected to where you found the job.

### Who It's For

You and a small group of people actively job searching who want signal over volume. Not limited to engineers — analysts, designers, PMs, any role at any company on the supported ATS platforms. Self-hosted by you, easy enough for friends to run their own instance.

### Why Existing Solutions Fall Short

|                      | EchoJobs            | Simplify        | go-jobs                                               |
| -------------------- | ------------------- | --------------- | ----------------------------------------------------- |
| Cost                 | Paid                | Freemium        | Free, self-hosted                                     |
| Reliability          | Goes down           | SaaS dependency | You control uptime                                    |
| Company list         | 5,000+ (noisy)      | N/A             | Startup ATS platforms only (Greenhouse, Lever, Ashby) |
| Workday jobs         | Included            | N/A             | Excluded by design — Workday not a supported source   |
| Filter UX            | Clunky              | N/A             | URL-based, bookmarkable                               |
| Application tracking | No                  | Yes             | Yes (built-in)                                        |
| Discovery + tracking | No                  | No              | Yes                                                   |
| CLI / agent-friendly | No                  | No              | Yes                                                   |
| Role types           | Engineering-focused | N/A             | Any role at any supported ATS company                 |

## How

See [ARCHITECTURE.md](ARCHITECTURE.md) for tech stack, data model, system design, and LLM enrichment pipeline details.

## Competition

| Tool                | What It Does                          | Strength                      | Weakness                                  | How go-jobs Differs                       |
| ------------------- | ------------------------------------- | ----------------------------- | ----------------------------------------- | ----------------------------------------- |
| **EchoJobs**        | Aggregates jobs from 5,000+ companies | Large coverage, established   | Paid, unreliable, noisy, clunky filters   | Curated list, free, self-hosted, reliable |
| **Simplify**        | Application tracker + autofill        | Good tracking UX              | Freemium, no job discovery                | Discovery + tracking in one place         |
| **LinkedIn Jobs**   | Massive job board                     | Huge coverage                 | Noisy, algorithmic feed, privacy concerns | Curated signal, no account needed         |
| **HN Who's Hiring** | Monthly job thread                    | High quality, startup-focused | Manual searching, no filters, no tracking | Structured + searchable + trackable       |
| **Otta**            | Curated startup job board             | Good curation                 | Paid/limited, SaaS                        | Self-hosted, you control the list         |
| **Wellfound**       | Startup job board                     | Startup-focused               | Requires account, limited filters         | No account, CLI-friendly, self-hosted     |

**The gap:** No existing tool combines curated aggregation + structured enrichment + application tracking + CLI/agent interface in a self-hosted package.

## MVP Milestones

Each milestone is independently testable/demoable.

### M1: Data Foundation

- Set up Go project structure, PostgreSQL schema, migrations
- Build Greenhouse adapter — discovers companies + jobs via JSON API
- Build Lever adapter — discovers companies + jobs via HTML scrape
- Build Ashby adapter — discovers companies + jobs via public API
- Store raw job postings in DB with dedup
- CLI command: `go-jobs scrape` — runs all adapters, prints summary
- CLI command: `go-jobs scrape --source greenhouse` — run a single adapter
- **Demo:** Run scrape, see companies + jobs populated in DB automatically

### M2: LLM Enrichment

- Integrate GenKit with OpenAI, Gemini, and Anthropic providers
- Build enrichment worker that processes un-tagged jobs
- Extract: role type, seniority, tech stack, remote policy, normalized location, country
- CLI command: `go-jobs enrich` — processes un-enriched jobs
- **Demo:** Run enrich, see structured tags on previously raw postings

### M3: CLI Search & Filters

- CLI command: `go-jobs search` with flags for all filter dimensions
- JSON output by default (agent-friendly)
- Human-readable table output with `--format table`
- Filter by: role type, seniority, location, country, remote policy, tech stack, company
- **Demo:** `go-jobs search --role engineering --seniority senior --remote remote --tech go` returns matching jobs as JSON

### M4: Application Tracker (CLI)

- `go-jobs interested <job-id>` — mark as interested
- `go-jobs apply <job-id>` — mark as applied (auto-captures date; sets interested if not already set)
- `go-jobs status <job-id> <status>` — set status (interviewing / offer / rejected / withdrawn)
- `go-jobs notes <job-id> "<text>"` — set notes on a job
- `go-jobs interested` — list interested jobs (JSON, includes status + applied_at)
- `go-jobs applied` — list applied jobs
- `go-jobs pipeline` — list all tracked jobs grouped by status
- Per-user state (CLI authenticates via stored token in local config)
- **Demo:** Mark a job as interested, apply to it, update to interviewing, list pipeline

### M5: Auth

- `go-jobs register` — register with email + password
- `go-jobs login` — authenticate, store token in local config
- Web UI: register/login pages; session cookie on successful auth
- **Demo:** Register, log in via web and CLI, tracker state persists across devices

### M6: Web UI

- HTTP server serving web UI (templ + htmx)
- Browse/search jobs with filters (filter state in URL query params); all filters multi-select
- Job cards show "added X days ago" (from `first_seen`); jobs added since last visit are highlighted
- Job detail view with full description
- Interested / apply / status buttons
- Notes field per job
- Pipeline view — tracked jobs grouped by status
- Company toggle — show/hide specific companies; hidden list visible and reversible
- Shareable filter URLs
- Scraper status indicator ("last updated X hours ago" from `scrape_runs`)
- **Demo:** Open browser, filter jobs, see new-since-last-visit highlights, track a job through the pipeline

### M7: Scheduled Scraping + Enrichment

- Background scheduler runs scrape + enrich on a configurable interval
- Mark jobs as inactive when they disappear from source
- CLI command: `go-jobs serve` — starts web UI + scheduler
- **Demo:** Start server, jobs update automatically on schedule

## Open Questions

- [x] Which LLM providers? → **OpenAI, Gemini, Anthropic. Raw official SDKs — no framework layer. Per-user API key passed at client construction.**
- [x] Rate limiting strategy? → **Scraping uses conservative delays between requests. LLM enrichment uses exponential backoff on 429s per-provider.**
- [x] LLM API key ownership? → **User-provided. Each user stores their own key + provider. Encrypted at rest (AES-256-GCM).**
- [x] How is "per-user" identity handled? → **Email + password auth. Session cookie (web), stored token (CLI). Email verification post-MVP.**
- [x] Tech stack? → **Go + Chi + PostgreSQL + sqlc + templ + htmx + Tailwind/DaisyUI + mise**
- [x] Tracker states? → **interested → applied → interviewing → offer / rejected / withdrawn. Notes field per job.**
- [x] Filter multi-select? → **Yes — all categorical filters (role type, seniority, remote policy, country, tech stack) are multi-select OR logic.**
- [x] "New since last visit"? → **Yes — jobs added after `users.last_visited_at` are highlighted. `last_visited_at` updated on each session.**
- [x] Limited to engineers? → **No, any role at any supported ATS company**
- [x] Include HackerNews? → **Post-MVP. ATS platforms (Greenhouse, Lever, Ashby) cover the target space. HN adds complexity for marginal coverage gain.**
