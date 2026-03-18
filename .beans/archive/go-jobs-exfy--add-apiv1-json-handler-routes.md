---
# go-jobs-exfy
title: Add /api/v1/ JSON handler routes
status: completed
type: task
priority: normal
created_at: 2026-03-18T21:36:57Z
updated_at: 2026-03-18T21:41:11Z
parent: go-jobs-qlpf
---

Implement JSON API endpoints under /api/v1/ as a new driving adapter alongside the existing HTML routes. Handlers are thin — they call the same core service ports the HTML handlers already use.

## Endpoints to implement
- POST /api/v1/auth/register
- POST /api/v1/auth/login  → returns {"token": "..."}
- POST /api/v1/auth/logout
- GET  /api/v1/jobs        → search (query params mirror CLI flags)
- GET  /api/v1/jobs/interested
- GET  /api/v1/jobs/applied
- POST /api/v1/jobs/{id}/interested
- POST /api/v1/jobs/{id}/apply
- POST /api/v1/jobs/{id}/status  → body: {"status": "..."}
- POST /api/v1/jobs/{id}/notes   → body: {"notes": "..."}
- GET  /api/v1/pipeline

## Auth
Protected endpoints use Authorization: Bearer <token> header. No CSRF needed for API routes.

## Summary of Changes

Added /api/v1/ JSON route tree as a second driving adapter.

### New files
- `internal/adapters/http/middleware/bearer.go` — BearerAuth middleware + TokenFromContext helper
- `internal/adapters/http/api/handler.go` — Handler struct, writeJSON/writeError helpers
- `internal/adapters/http/api/auth.go` — POST /api/v1/auth/{register,login,logout}
- `internal/adapters/http/api/jobs.go` — GET /api/v1/jobs, /jobs/interested, /jobs/applied
- `internal/adapters/http/api/tracker.go` — POST /api/v1/jobs/{id}/{interested,apply,status,notes}
- `internal/adapters/http/api/pipeline.go` — GET /api/v1/pipeline

### Modified files
- `cmd/jobs/main.go` — wired api.Handler and registered /api/v1/ route group

Auth: public routes (register, login, search) are open. All tracker/pipeline routes require Authorization: Bearer <token>.
