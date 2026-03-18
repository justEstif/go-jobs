# Self-hosting go-jobs

This guide helps you run `go-jobs` on your own infrastructure with PostgreSQL.

## What you need

- A machine that can run Go binaries (local laptop, VM, or VPS)
- PostgreSQL database
- `mise` installed for task/runtime setup
- Docker + Docker Compose (for local Postgres)

## 1) Clone and bootstrap

```bash
git clone https://github.com/justestif/go-jobs.git
cd go-jobs
mise install
```

## 2) Start PostgreSQL

For local setup:

```bash
docker-compose up -d
```

If you already have managed Postgres, skip this and use that `DATABASE_URL`.

## 3) Configure environment variables

These are required at runtime:

- `DATABASE_URL` — Postgres connection string
- `PORT` — HTTP listen port (default `3000`)
- `CSRF_KEY` — exactly 32 bytes
- `SESSION_SECRET` — cookie session signing secret
- `ENCRYPTION_KEY` — 64 hex chars (AES-256-GCM key)

Example values for local development:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/go_jobs?sslmode=disable'
export PORT='3000'
export CSRF_KEY='0123456789abcdef0123456789abcdef'
export SESSION_SECRET='change-this-in-real-deployments'
export ENCRYPTION_KEY='0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
```

## 4) Run setup once

```bash
mise run setup
```

This runs migrations and code generation tasks (`templ` + `sqlc`) needed before serving.

## 5) Run the app

Development mode (live reload):

```bash
mise run dev
```

Production build:

```bash
mise run build
./go-jobs
```

The default server URL is `http://localhost:3000`.

## 6) Verify background behavior

- Running `go-jobs` starts the web server and the scrape/enrich scheduler
- `SCRAPE_INTERVAL` controls scheduler frequency (default `6h`)
- `ENRICH_LIMIT` controls enrichment batch size per run (default `1000`)

Optional override example:

```bash
export SCRAPE_INTERVAL='1h'
export ENRICH_LIMIT='500'
```

## Deployment notes

- Keep secrets in your host's secret manager or environment, not in git
- Put app and DB behind regular backups
- Run behind a reverse proxy (Caddy, Nginx, Traefik) for TLS and domain routing
- Use a process supervisor (systemd, Nomad, Fly machine, etc.) for restarts

## Useful commands

```bash
go build ./...
go test ./...
mise run db-migrate
mise run db-rollback
```

## Related docs

- [Project landing page](./index.html)
- [MVP scope](./MVP.md)
- [Architecture](./ARCHITECTURE.md)
- [Interfaces](./interfaces.md)
