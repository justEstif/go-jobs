# AGENTS.md — go-jobs

Agent instructions for the go-jobs repository. Read this before writing any code.

## Project Overview

Self-hosted job aggregator built in Go. Hexagonal (ports and adapters) architecture.
Module: `github.com/justestif/go-jobs`

Architecture docs are authoritative — read them before implementing anything new:
- `docs/ARCHITECTURE.md` — system design, data model, folder structure
- `docs/interfaces.md` — all domain types and port interfaces (copy these verbatim)
- `docs/MVP.md` — product scope and milestones

---

## CLI

The CLI binary is named `jobs`. Users install it via npm:

```bash
npm install -g @justestif/go-jobs
jobs search --query "backend go"
jobs login
jobs pipeline
```

By default the CLI targets `http://127.0.0.1:3000` (via `BASE_URL` env in mise).
To target the hosted server: `jobs --base-url https://jobs.estifanos.cc <command>`

---

## Commands

### Development
```bash
mise run dev          # Live reload server (air: CSS + templ + Go on every save)
mise run dev:cli      # Run CLI against local dev server (BASE_URL already set in mise)
mise run setup        # First-time setup: migrate + templ generate + sqlc generate
```

### Build
```bash
mise run build        # Production: tailwindcss --minify + templ generate + go build → bin/jobs
go build ./...        # Compile check (no output = clean)
templ generate        # Regenerate _templ.go files after editing .templ files
sqlc generate         # Regenerate queries after editing migrations/ or queries.sql
```

### Release
```bash
git tag v0.1.0 && git push origin v0.1.0   # triggers .github/workflows/release.yml
```
The workflow: builds `jobs` for 5 platforms → creates GitHub Release with archives →
publishes `@justestif/go-jobs` to npm.

**One-time manual setup required before first publish:**
1. `npm login` → create an automation token at npmjs.com → add as `NPM_TOKEN` secret in GitHub repo settings
2. First publish: the package must exist on npm before provenance works — `cd npm/go-jobs && npm publish --access public` manually for v0.0.1, then let the workflow handle subsequent releases

### Database
```bash
mise run db-migrate   # Apply all pending migrations
mise run db-rollback  # Roll back last migration
docker-compose up -d  # Start PostgreSQL (required for dev)
```

### Tests
```bash
go test ./...                        # Run all tests
go test ./internal/core/services/... # Run tests in a specific package
go test -run TestFunctionName ./...  # Run a single test by name
go test -v -run TestFunctionName ./internal/core/services/search_test.go
```
No tests exist yet. When writing tests, use in-memory fakes of driven ports — not the real Postgres adapter.

### Lint / Format
```bash
gofmt -w .            # Format all Go files (required before committing)
go vet ./...          # Static analysis
```
No golangci-lint is configured yet.

---

## Environment Variables

Defined in `mise.toml` for local dev. All required at runtime:

| Variable         | Description                                     |
|------------------|-------------------------------------------------|
| `DATABASE_URL`   | Postgres connection string                      |
| `PORT`           | HTTP listen port (default: 3000)                |
| `CSRF_KEY`       | Exactly 32 bytes — CSRF cookie signing key      |
| `SESSION_SECRET` | Cookie session signing secret                   |
| `ENCRYPTION_KEY` | 64 hex chars — AES-256-GCM key for LLM API keys |

---

## Folder Structure

```
cmd/jobs/main.go                   # Composition root — the only place that knows all concrete types
internal/
  core/
    domain/                        # Pure domain types (Job, Company, User, JobTags, …)
    ports/
      driving.go                   # Interfaces the outside world calls into the core
      driven.go                    # Interfaces the core calls out to (repos, scrapers, enricher)
    services/                      # Use case implementations — no adapter imports allowed
  adapters/
    postgres/
      conn.go                      # pgxpool init (package postgres)
      queries/                     # sqlc-generated — NEVER import outside this package
    http/
      *.go                         # HTTP handlers (package httphandlers)
      middleware/                  # session, CSRF, auth middleware (package middleware)
    scrapers/                      # Greenhouse, Lever, Ashby JobScraper impls
    enrichment/                    # Tiered JobEnricher: ats.go → rules.go → llm.go
    scheduler/                     # time.Ticker → ScrapeService.Run
  cli/                             # cobra commands (driving adapter)
components/                        # templ templates (package components)
migrations/                        # golang-migrate SQL files
```

**Hard constraints:**
- `internal/core/` has **zero imports** from `internal/adapters/`. The Go compiler enforces this.
- `internal/adapters/postgres/queries/` is **never imported outside** `internal/adapters/postgres/`. Map sqlc types to domain types at the adapter boundary before returning.
- `cmd/jobs/main.go` is the **only** place that instantiates concrete types and wires them together.

---

## Code Style

### Formatting
Standard `gofmt`. Run `gofmt -w .` before every commit. No exceptions.

### Imports
Three groups, separated by blank lines: stdlib → third-party → project-internal.

```go
import (
    "context"
    "fmt"

    "github.com/go-chi/chi/v5"
    chimiddleware "github.com/go-chi/chi/v5/middleware"

    httphandlers "github.com/justestif/go-jobs/internal/adapters/http"
    "github.com/justestif/go-jobs/internal/adapters/http/middleware"
)
```

Use import aliases only to resolve name collisions. Current conventions:
- `chimiddleware` — chi's middleware package (avoids conflict with project middleware)
- `httphandlers` — `internal/adapters/http` package (avoids conflict with `net/http`)

### Package Names
- Match the last path segment, except when it collides with a stdlib package.
- `internal/adapters/http` → `package httphandlers` (not `http`)
- `internal/adapters/postgres/queries` → `package queries`
- `internal/adapters/postgres` → `package postgres`
- `internal/adapters/http/middleware` → `package middleware`

### Naming
- **Types, interfaces, exported functions:** PascalCase (`JobSearchService`, `NewSessionManager`)
- **Unexported:** camelCase (`userIDContextKey`, `csrfKey`)
- **Constants:** PascalCase for exported (`SessionName`), typed-constant idiom for unexported keys:
  ```go
  type contextKey string
  const userIDContextKey contextKey = "user_id"
  ```
- **Typed string enums:** define a named type, then `const` block with typed values:
  ```go
  type ATSType string
  const (
      ATSGreenhouse ATSType = "greenhouse"
      ATSLever      ATSType = "lever"
  )
  ```
- **ID type aliases** (in domain): `type JobID = uuid.UUID` (alias, not new type)

### Error Handling

**Fatal on unrecoverable startup errors** (missing config, DB unreachable):
```go
if err := postgres.InitDB(); err != nil {
    log.Fatalf("Failed to initialize database: %v", err)
}
```

**Wrap with context** when propagating errors up:
```go
return fmt.Errorf("unable to create connection pool: %w", err)
```
Use `%w` (not `%v`) when wrapping so callers can `errors.Is`/`errors.As`.

**HTTP handlers:** return HTTP error and early-return on bad input:
```go
if err := r.ParseForm(); err != nil {
    http.Error(w, "Invalid form data", http.StatusBadRequest)
    return
}
```

**Service layer:** log per-item errors and continue (e.g. one failing scraper does not abort the pipeline). Never panic in adapter or service code.

**Zero value + error pattern** for typed returns:
```go
return [16]byte{}, http.ErrNoCookie
return domain.Job{}, fmt.Errorf("job %s not found: %w", id, err)
```

### Context
Every interface method takes `context.Context` as the **first** parameter. Adapters must respect cancellation. Timeouts are set by callers (HTTP handlers, scheduler), not by interface implementations.

### Comments
Godoc-style on all exported symbols. Multi-paragraph with blank comment line between paragraphs. Use indented code blocks (4 spaces) for usage examples inside comments.

```go
// OptionalAuth loads the session user ID into the request context if a valid
// session exists, but does not block unauthenticated requests.
//
// Retrieve the user ID in a handler with:
//
//	id, ok := middleware.UserIDFromContext(r.Context())
func OptionalAuth(sm *SessionManager) func(http.Handler) http.Handler {
```

Inline comments explain non-obvious design choices, not what the code does.

Generated files carry standard headers — do not edit them:
```go
// Code generated by sqlc. DO NOT EDIT.
// Code generated by templ - DO NOT EDIT.
```

### Domain Types
Defined in `docs/interfaces.md`. Copy them **verbatim** into `internal/core/domain/` — do not invent types that diverge from the spec. Nullable fields use `*time.Time` (not `pgtype.Timestamp` — that's an adapter concern).

---

## Architecture Rules (Hexagonal)

1. **Dependency direction:** all arrows point inward. Adapters depend on core interfaces. The core depends on nothing outside itself.
2. **Ports are owned by the core.** `internal/core/ports/driving.go` and `driven.go` define what the world looks like to the core. Adapters implement driven ports; CLI/HTTP/scheduler call driving ports.
3. **No technology names in ports.** Port interfaces say what, not how. `JobRepository` not `PostgresJobRepository`; `JobScraper` not `GreenhouseAdapter`.
4. **Services are thin orchestrators.** A service method calls one or more driven port methods, enforces business rules, and returns a domain type. No SQL, no HTTP, no filesystem.
5. **Composition root wires everything.** `cmd/jobs/main.go` instantiates all concrete adapters, passes them to services, passes services to delivery adapters. Nothing else does this.

---

## Templ Components

Edit `.templ` files, then run `templ generate` to produce `_templ.go` files. Never edit `_templ.go` directly. Components live in `components/` at the project root (package `components`).

---

## Issue Tracking

This project uses **beans** (`beans` CLI) for issue tracking. Check for existing beans before starting work:
```bash
beans list --json --ready    # unblocked work ready to start
beans show --json <id>       # view a specific bean
```
Create a bean for any non-trivial work, mark it `in-progress` before starting, and `completed` when done.
