---
# go-jobs-sycc
title: 'M1: Data Foundation'
status: completed
type: milestone
priority: normal
created_at: 2026-03-18T10:23:14Z
updated_at: 2026-03-18T10:33:45Z
---

Full M1 implementation: domain types, port interfaces, DB migrations, Postgres adapter (sqlc), Greenhouse/Lever/Ashby scrapers, ScrapeService, CompanySeeder, and CLI scrape command.

## Tasks

- [x] Domain types (internal/core/domain/job.go)
- [x] Driving ports (internal/core/ports/driving.go)
- [x] Driven ports (internal/core/ports/driven.go)
- [x] DB migrations (full schema: companies, jobs, scrape_runs, job_tags, users, user_jobs, user_companies, sessions)
- [x] sqlc query definitions (queries.sql)
- [x] Postgres adapter: company_repo.go
- [x] Postgres adapter: job_repo.go
- [x] Postgres adapter: user_repo.go + session_repo.go
- [x] Postgres adapter: userjob_repo.go + scraperun_repo.go + usercompany_repo.go
- [x] Greenhouse scraper adapter
- [x] Lever scraper adapter
- [x] Ashby scraper adapter
- [x] CompanySeeder adapter (Simplify README parser)
- [x] ScrapeService (internal/core/services/scrape.go)
- [x] CLI scrape command (internal/cli/scrape.go + root.go)
- [x] Wire everything in cmd/jobs/main.go
- [x] go build ./... passes clean

## Summary of Changes

- **Domain types** (): all types from interfaces.md verbatim — Company, Job, JobTags, RawJob, User, UserJob, ScrapeRun, SearchFilters, UserSearchContext, plus all typed string enums
- **Ports** (, ): all six driving ports and all nine driven ports; zero imports from adapters
- **Migrations** (001–004): replaced placeholder with full schema — users+sessions, companies, jobs+job_tags+scrape_runs, user_jobs+user_companies
- **sqlc**: full query set for all tables;  runs clean
- **Postgres adapters**: company_repo, job_repo, user_repo (doubles as session_repo), userjob_repo, scraperun_repo, usercompany_repo + mapping.go helpers
- **Scrapers**: greenhouse.go (JSON boards API), lever.go (v0 API), ashby.go (POST API)
- **CompanySeeder**: simplify README parser with Greenhouse/Lever/Ashby URL pattern extraction
- **ScrapeService**: orchestrates seed→scrape→upsert→mark-inactive pipeline with per-company error isolation
- **CLI**:  command wired via cobra;  flag noted for M3+
- **main.go**: composition root wires all adapters → services → CLI/HTTP; CLI branch runs when args present, HTTP server otherwise
- **Architecture**:  has zero imports from ;  package not imported outside ;  and  pass clean
