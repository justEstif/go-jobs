---
# go-jobs-l7xu
title: 'M3: CLI Search & Filters'
status: completed
type: milestone
priority: normal
created_at: 2026-03-18T10:55:53Z
updated_at: 2026-03-18T11:00:17Z
---

CLI command `go-jobs search` with flags for all filter dimensions. JSON output by default (agent-friendly), human-readable table output with --format table. Filter by: role type, seniority, location, country, remote policy, tech stack, company.

## Tasks

- [x] Add EnrichService driving port (done in M2)
- [x] Implement JobSearchService driving port (internal/core/services/search.go)
- [x] Implement JobRepository.Search in postgres adapter (queries.sql + job_repo.go)
- [x] CLI command: go-jobs search (internal/cli/search.go)
- [x] --format table output renderer
- [x] Wire JobSearchService in main.go and cli.Services
- [x] go build ./... passes clean

## Summary of Changes

- Added `SearchJobs` SQL query with 9 parameters covering all filter dimensions (free-text, role_type, seniority, remote_policy, country, tech_stack AND semantics, company_ids, limit, offset)
- Ran `sqlc generate` to produce `SearchJobsRow` / `SearchJobsParams` types
- Added `domainJobFromSearchRow` mapping in `mapping.go` (includes nullable job_tags fields via LEFT JOIN)
- Replaced stub `JobRepo.Search` with a full implementation using the new query
- Implemented `internal/core/services/search.go` — thin `jobSearchService` delegating to `JobRepository`
- Added `internal/cli/search.go` — `go-jobs search` with flags: `--query`, `--role`, `--seniority`, `--remote`, `--country`, `--tech`, `--company`, `--limit`, `--offset`, `--format`
- `--format json` (default): machine-readable JSON via json v2
- `--format table`: human-readable tabwriter output
- Wired `Search` into `cli.Services` (`root.go`) and instantiated `NewJobSearchService` in `main.go`
- `go build ./...` and `go vet ./...` pass clean
