---
# go-jobs-mt2u
title: Add 10,000-char max length on resume
status: completed
type: task
priority: normal
created_at: 2026-03-19T17:00:02Z
updated_at: 2026-03-19T17:01:30Z
---

Enforce a 10,000-character limit on resume text to prevent abuse. Validate in the service layer (authoritative), HTTP handler (user-friendly error), CLI (early feedback), and templ template (maxlength attribute + JS counter).

## Summary of Changes

Enforced a 10,000-character maximum on resume text at four layers:

1. **Domain constant** (`domain.MaxResumeLength = 10_000`) — single source of truth
2. **Service layer** (`userService.SetResume`) — authoritative validation, returns error before DB call
3. **HTTP handler** (`SettingsHandler.SaveResume`) — returns 400 with descriptive message
4. **CLI** (`resume set`) — early feedback before calling service
5. **Templ template** — `maxlength` attribute on textarea + live character counter showing current/max

### Files changed
- `internal/core/domain/job.go` — added `MaxResumeLength` constant
- `internal/core/services/user.go` — length check in `SetResume`
- `internal/adapters/http/settings.go` — 400 error on oversized resume
- `internal/cli/resume.go` — early error + changed output from bytes to chars
- `components/settings.templ` — maxlength attr, live char counter, `formatMaxResumeLength` helper
