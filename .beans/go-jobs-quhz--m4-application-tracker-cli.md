---
# go-jobs-quhz
title: 'M4: Application Tracker (CLI)'
status: todo
type: milestone
created_at: 2026-03-18T11:01:15Z
updated_at: 2026-03-18T11:01:15Z
---

CLI commands for managing job application pipeline state. Per-user tracking via stored session token. Commands: interested, apply, status, notes, pipeline.

## Tasks

- [ ] Implement ApplicationService (internal/core/services/application.go)
- [ ] Implement ApplicationService postgres adapters (user_repo, userjob_repo already exist — verify completeness)
- [ ] CLI: go-jobs interested <job-id> — mark job as interested
- [ ] CLI: go-jobs apply <job-id> — mark as applied (auto-sets interested, captures AppliedAt)
- [ ] CLI: go-jobs status <job-id> <status> — set interviewing/offer/rejected/withdrawn
- [ ] CLI: go-jobs notes <job-id> "<text>" — set notes
- [ ] CLI: go-jobs interested (no args) — list interested jobs as JSON
- [ ] CLI: go-jobs applied — list applied jobs as JSON
- [ ] CLI: go-jobs pipeline — list all tracked jobs grouped by status as JSON
- [ ] Auth token storage: read/write token from $XDG_CONFIG_HOME/go-jobs/token (used for all tracker commands)
- [ ] Wire ApplicationService in main.go and cli.Services
- [ ] go build ./... passes clean
