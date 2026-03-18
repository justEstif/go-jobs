---
# go-jobs-quhz
title: 'M4: Application Tracker (CLI)'
status: completed
type: milestone
priority: normal
created_at: 2026-03-18T11:01:15Z
updated_at: 2026-03-18T11:03:56Z
---

CLI commands for managing job application pipeline state. Per-user tracking via stored session token. Commands: interested, apply, status, notes, pipeline.

## Tasks

- [x] Implement ApplicationService (internal/core/services/application.go)
- [x] Implement ApplicationService postgres adapters (user_repo, userjob_repo already exist — verify completeness)
- [x] CLI: go-jobs interested <job-id> — mark job as interested
- [x] CLI: go-jobs apply <job-id> — mark as applied (auto-sets interested, captures AppliedAt)
- [x] CLI: go-jobs status <job-id> <status> — set interviewing/offer/rejected/withdrawn
- [x] CLI: go-jobs notes <job-id> "<text>" — set notes
- [x] CLI: go-jobs interested (no args) — list interested jobs as JSON
- [x] CLI: go-jobs applied — list applied jobs as JSON
- [x] CLI: go-jobs pipeline — list all tracked jobs grouped by status as JSON
- [x] Auth token storage: read/write token from $XDG_CONFIG_HOME/go-jobs/token (used for all tracker commands)
- [x] Wire ApplicationService in main.go and cli.Services
- [x] go build ./... passes clean

## Summary of Changes

- `internal/core/services/application.go` — `ApplicationService` implementation with full business rules: auto-interested on apply, no-backwards-to-interested guard, read-modify-write for notes, batch GetByIDs for list hydration
- `internal/cli/token.go` — token file helpers: read/write/delete at `$XDG_CONFIG_HOME/go-jobs/token`, `requireToken()` guard with clear login prompt
- `internal/cli/tracker.go` — 6 tracker commands: `interested [job-id]`, `apply <job-id>`, `status <job-id> <status>`, `notes <job-id> <text>`, `applied`, `pipeline`
- `cli.Services` extended with `Application ports.ApplicationService` and `Session ports.SessionRepository`
- `main.go` instantiates `UserRepo`, `UserJobRepo`, `ApplicationService`; wires into `cliServices`
- `go build ./...` and `go vet ./...` pass clean
