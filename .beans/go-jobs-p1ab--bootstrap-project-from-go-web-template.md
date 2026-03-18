---
# go-jobs-p1ab
title: Bootstrap project from go-web-template
status: completed
type: task
priority: normal
created_at: 2026-03-18T10:09:03Z
updated_at: 2026-03-18T10:15:53Z
---

Use gonew to scaffold the project from github.com/justEstif/go-web-template, then adapt the structure to match the go-jobs architecture (hexagonal layout, cobra CLI, correct module name).

## Tasks

- [x] Run gonew to scaffold template
- [x] Copy files into project directory
- [x] Rename cmd/web → cmd/jobs (composition root)
- [x] Create hexagonal folder structure (core/domain, core/ports, core/services, adapters/*, cli/*)
- [x] Move template's internal/database → adapters/postgres skeleton
- [x] Move template's internal/handlers → adapters/http skeleton
- [x] Move template's internal/middleware → adapters/http/middleware
- [x] Move template's components/ → components/ (keep as-is, templ stays at root)
- [x] Update mise.toml: rename binary
- [x] Add cobra dependency to go.mod
- [x] Verify project builds

## Summary of Changes

- Scaffolded project from `go-web-template` using `gonew`
- Renamed `cmd/web` → `cmd/jobs` (composition root)
- Restructured `internal/` to hexagonal layout: `core/{domain,ports,services}`, `adapters/{postgres,http,scrapers,enrichment,scheduler}`, `cli/`
- Moved sqlc-generated code to `internal/adapters/postgres/queries/` (package `queries`)
- Moved session/auth/middleware to `internal/adapters/http/middleware/` (all in one `middleware` package)
- Moved handlers to `internal/adapters/http/` (package `httphandlers`)
- Updated `sqlc.yaml`, `.air.toml`, `mise.toml` to reference new paths and binary name
- Added `github.com/spf13/cobra` to `go.mod`
- Build passes clean: `go build ./...`
