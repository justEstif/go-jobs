# go-jobs — MVP

## What

A self-hosted job aggregator that pulls from curated company career pages and HackerNews, with fast filters, a CLI interface, a web UI, and a lightweight application tracker. LLM enrichment tags and classifies postings so you can search by what actually matters.

### Core Features

- **Job aggregation** — Scrape/pull job postings from a curated list of company career pages and HN "Who's Hiring" threads
- **LLM enrichment** — Tag postings with structured metadata (role type, seniority, tech stack, remote/hybrid/onsite, location) using GenKit + an LLM API
- **Filters** — Role type, location (including US/non-US), seniority, tech stack, company. Filter state lives in the URL (web) and CLI flags (terminal)
- **Application tracker** — Per-user saved/applied states. Saved = interested. Applied = submitted, with auto-captured date
- **CLI** — JSON output, agent-friendly. Search, filter, save, mark applied
- **Web UI** — Browse, filter, save, track. URL-based filter state (shareable, bookmarkable, no login)
- **Per-user identity** — Cookie/session based, no login required

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

You and a small group of people actively job searching who want signal over volume. Not limited to engineers — analysts, designers, PMs, anyone targeting the curated company list. Self-hosted by you, easy enough for friends to run their own instance.

### Why Existing Solutions Fall Short

| | EchoJobs | Simplify | go-jobs |
|---|---|---|---|
| Cost | Paid | Freemium | Free, self-hosted |
| Reliability | Goes down | SaaS dependency | You control uptime |
| Company list | 5,000+ (noisy) | N/A | Curated (~200–500) |
| Workday jobs | Included | N/A | Excluded by design |
| Filter UX | Clunky | N/A | URL-based, bookmarkable |
| Application tracking | No | Yes | Yes (built-in) |
| Discovery + tracking | No | No | Yes |
| CLI / agent-friendly | No | No | Yes |
| Role types | Engineering-focused | N/A | Any role at curated companies |

## How

### Tech Stack

| Component | Choice | Rationale |
|---|---|---|
| Language | Go | Single binary, great concurrency for scraping, matches your stack |
| Database | PostgreSQL | Your standard stack, handles structured job data + user state well |
| LLM integration | Firebase GenKit (Go SDK) | Unified interface for LLM calls, structured output, provider-agnostic |
| Web UI | TBD (server-rendered or SPA) | Keep it simple — could be templ + htmx, or a lightweight React app |
| CLI | cobra or built into the Go binary | Standard Go CLI patterns |

### Data Model

```
companies
├── id            (uuid)
├── name          (text)
├── careers_url   (text)       -- career page to scrape
├── ats_type      (text)       -- "greenhouse" | "lever" | "ashby" | "custom" | "hackernews"
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
├── source        (text)       -- "career_page" | "hackernews"
├── raw_html      (text)       -- original posting for re-processing
├── first_seen    (timestamp)
├── last_seen     (timestamp)
└── active        (bool)       -- false when posting disappears

job_tags (LLM-enriched structured metadata)
├── job_id        (fk → jobs)
├── role_type     (text)       -- "engineering" | "analyst" | "design" | "pm" | ...
├── seniority     (text)       -- "intern" | "junior" | "mid" | "senior" | "staff" | "lead"
├── remote_policy (text)       -- "remote" | "hybrid" | "onsite"
├── location_norm (text)       -- normalized location
├── country       (text)       -- "US" | "UK" | "DE" | ...
├── tech_stack    (text[])     -- ["go", "postgres", "kubernetes"]
└── enriched_at   (timestamp)

users
├── id            (uuid)
├── session_token (text, unique)
├── llm_api_key   (text)       -- user-provided, encrypted at rest
├── llm_provider  (text)       -- "openai" | "anthropic" | "google" (GenKit provider)
└── created_at    (timestamp)

user_jobs
├── user_id       (fk → users)
├── job_id        (fk → jobs)
├── status        (text)       -- "saved" | "applied"
├── status_at     (timestamp)  -- when status was set/changed
└── unique(user_id, job_id)

user_companies
├── user_id       (fk → users)
├── company_id    (fk → companies)
└── hidden        (bool, default false)  -- user opts out of a company
-- absence of a row = company is visible (opt-out model, default all visible)
```

### Architecture

```
┌─────────────────────────────────────────────────────┐
│                    go-jobs binary                    │
│                                                     │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │   CLI    │  │  Web UI  │  │   Scrape Scheduler │  │
│  │ (cobra)  │  │ (http)   │  │   (cron/ticker)    │  │
│  └────┬─────┘  └────┬─────┘  └────────┬──────────┘  │
│       │              │                 │             │
│       └──────────┬───┘                 │             │
│                  │                     │             │
│           ┌──────▼──────┐    ┌─────────▼──────────┐  │
│           │  Core API   │    │    Scraper Engine   │  │
│           │             │    │                     │  │
│           │ - search    │    │ - career page       │  │
│           │ - filter    │    │   scrapers          │  │
│           │ - save/apply│    │ - HN API client     │  │
│           │ - user mgmt │    │ - dedup             │  │
│           └──────┬──────┘    └─────────┬──────────┘  │
│                  │                     │             │
│                  │           ┌─────────▼──────────┐  │
│                  │           │   LLM Enrichment   │  │
│                  │           │   (GenKit)          │  │
│                  │           │                     │  │
│                  │           │ - tag extraction    │  │
│                  │           │ - classification    │  │
│                  │           └─────────┬──────────┘  │
│                  │                     │             │
│                  └──────────┬──────────┘             │
│                             │                        │
│                    ┌────────▼────────┐                │
│                    │   PostgreSQL    │                │
│                    │                 │                │
│                    │ jobs, companies │                │
│                    │ tags, users     │                │
│                    └─────────────────┘                │
└─────────────────────────────────────────────────────┘
```

### Data Sources

| Source | Method | Difficulty | Notes |
|---|---|---|---|
| Company career pages | HTTP scrape / parse | Medium-High | Each company's page is different. Start with a few, add incrementally. Consider Greenhouse/Lever API patterns (many startups use these) |
| HackerNews "Who's Hiring" | HN API (Algolia) | Low | Monthly threads, well-structured, free API |

**Key insight on career pages:** Many startups use hosted ATS platforms (Greenhouse, Lever, Ashby) that have consistent HTML structures or even APIs. Prioritize these over fully custom career pages — you get more companies for less scraper work.

| ATS Platform | `ats_type` | `scrape_type` | Notes |
|---|---|---|---|
| Greenhouse | `greenhouse` | `http` | Consistent DOM, JSON API at `/embed/job_board?for=<company>` |
| Lever | `lever` | `http` | Consistent HTML structure |
| Ashby | `ashby` | `http` | Public API available |
| Custom (SSR) | `custom` | `http` | Case-by-case |
| Custom (JS SPA) | `custom` | `headless` | Post-MVP — defer until core is stable |
| HackerNews | `hackernews` | `http` | Algolia API, not a scrape |

MVP ships with `scrape_type = "http"` only. `"headless"` companies are stored in the DB but skipped at scrape time with a log warning.

### LLM Enrichment Pipeline

1. Job posting scraped → raw text/HTML stored
2. Background worker picks up un-enriched jobs
3. GenKit call extracts structured tags: role type, seniority, tech stack, location, remote policy
4. Tags stored in `job_tags`
5. Failed enrichments retried with backoff

The LLM call is the most expensive part. Budget ~$0.01-0.05 per posting depending on model. For 500 companies posting ~10 jobs each, that's ~$50-250 per full scrape cycle. Cache aggressively — only re-enrich when posting content changes.

**User-provided API keys:** Each user supplies their own LLM API key and chooses their provider. The app stores the key encrypted at rest and passes it to GenKit at enrichment time. This means:
- No shared API cost — each user pays for their own enrichment
- GenKit's provider-agnostic interface means switching models is a config change, not a code change
- The app should validate the key is working at setup time and give clear errors when it's missing or invalid

## Competition

| Tool | What It Does | Strength | Weakness | How go-jobs Differs |
|---|---|---|---|---|
| **EchoJobs** | Aggregates jobs from 5,000+ companies | Large coverage, established | Paid, unreliable, noisy, clunky filters | Curated list, free, self-hosted, reliable |
| **Simplify** | Application tracker + autofill | Good tracking UX | Freemium, no job discovery | Discovery + tracking in one place |
| **LinkedIn Jobs** | Massive job board | Huge coverage | Noisy, algorithmic feed, privacy concerns | Curated signal, no account needed |
| **HN Who's Hiring** | Monthly job thread | High quality, startup-focused | Manual searching, no filters, no tracking | Structured + searchable + trackable |
| **Otta** | Curated startup job board | Good curation | Paid/limited, SaaS | Self-hosted, you control the list |
| **Wellfound** | Startup job board | Startup-focused | Requires account, limited filters | No account, CLI-friendly, self-hosted |

**The gap:** No existing tool combines curated aggregation + structured enrichment + application tracking + CLI/agent interface in a self-hosted package.

## MVP Milestones

Each milestone is independently testable/demoable.

### M1: Data Foundation
- Set up Go project structure, PostgreSQL schema, migrations
- Curate initial company list (~20 companies using Greenhouse/Lever)
- Build scraper for Greenhouse + Lever career pages
- Store raw job postings in DB with dedup
- CLI command: `go-jobs scrape` — runs scrapers, prints summary
- **Demo:** Run scrape, see jobs in DB

### M2: HackerNews Source
- Add HN "Who's Hiring" scraper via Algolia API
- Parse individual postings from thread comments
- Same dedup + storage pipeline
- CLI command: `go-jobs scrape --source hn`
- **Demo:** Scrape latest HN thread, see jobs alongside career page jobs

### M3: LLM Enrichment
- Integrate GenKit with chosen LLM provider
- Build enrichment worker that processes un-tagged jobs
- Extract: role type, seniority, tech stack, remote policy, normalized location, country
- CLI command: `go-jobs enrich` — processes un-enriched jobs
- **Demo:** Run enrich, see structured tags on previously raw postings

### M4: CLI Search & Filters
- CLI command: `go-jobs search` with flags for all filter dimensions
- JSON output by default (agent-friendly)
- Human-readable table output with `--format table`
- Filter by: role type, seniority, location, country, remote policy, tech stack, company
- **Demo:** `go-jobs search --role engineering --seniority senior --remote remote --tech go` returns matching jobs as JSON

### M5: Application Tracker (CLI)
- `go-jobs save <job-id>` — mark as saved
- `go-jobs apply <job-id>` — mark as applied (auto-captures date)
- `go-jobs saved` — list saved jobs
- `go-jobs applied` — list applied jobs
- Per-user state (CLI uses a local config file for user identity)
- **Demo:** Save a job, apply to it, list both

### M6: Web UI
- HTTP server serving web UI
- Browse/search jobs with filters (filter state in URL query params)
- View job details
- Save/apply buttons (session-cookie identity)
- Saved/applied views
- Company toggle — show/hide specific companies (saved to `user_companies`)
- Shareable filter URLs
- **Demo:** Open browser, filter jobs, hide a company, share URL with a friend, they see same results

### M7: Scheduled Scraping + Enrichment
- Background scheduler runs scrape + enrich on a configurable interval
- Mark jobs as inactive when they disappear from source
- CLI command: `go-jobs serve` — starts web UI + scheduler
- **Demo:** Start server, jobs update automatically on schedule

## Open Questions

- [ ] Which LLM provider/model for enrichment? (cost vs quality tradeoff — GPT-4o-mini, Claude Haiku, Gemini Flash?)
- [ ] Web UI approach — server-rendered (templ + htmx) or SPA (React/Svelte)?
- [ ] How many companies in the initial curated list? Start with 20? 50?
- [ ] Should the curated company list be a config file (YAML/TOML) or managed in DB?
- [ ] Deployment strategy — Docker compose (Go binary + Postgres)?
- [ ] Rate limiting strategy for scraping — how polite do we need to be?
- [x] How to handle JS-rendered career pages? → **`scrape_type` field on companies. MVP ships `http` only. `headless` companies stored but skipped with warning. Post-MVP.**
- [x] Company list in config or DB? → **DB. Global curated list. Users toggle visibility per-company via `user_companies` (opt-out, default all visible).**
- [x] LLM API key ownership? → **User-provided. Each user stores their own key + provider. Encrypted at rest. GenKit handles provider switching.**
- [x] What's the name? → **go-jobs**
- [x] Per-user or global state? → **per-user**
- [x] Should filters persist across sessions? → **URL stores state**
- [x] How is "per-user" identity handled? → **cookie/session (web), local config (CLI)**
- [x] Tech stack? → **Go + PostgreSQL + GenKit**
- [x] Limited to engineers? → **No, any role at curated companies**
