# go-jobs

go-jobs is a self-hosted job aggregator for people who want a focused startup job search without relying on a SaaS job board.

It pulls postings from ATS platforms like Greenhouse, Lever, and Ashby, stores them locally, and gives you a fast web UI plus a CLI to search, filter, and track applications in one place.

## What the app does

- Aggregates jobs from supported startup ATS platforms
- Enriches postings with structured metadata like role type, seniority, remote policy, location, and tech stack
- Lets users search and filter jobs in the browser or from the terminal
- Includes a lightweight application tracker for interested, applied, interviewing, offer, rejected, and withdrawn states
- Runs as a self-hosted Go app backed by PostgreSQL

## Who it is for

- Job seekers who want better signal than broad job boards
- People who prefer owning their own tooling and data
- Users who want both a web interface and an agent-friendly CLI

## Tech stack

- Go
- PostgreSQL
- Chi
- templ
- HTMX
- Tailwind CSS + DaisyUI
- sqlc
- golang-migrate

## Project docs

- `docs/MVP.md` — product scope and milestones
- `docs/ARCHITECTURE.md` — system design and data model
- `docs/interfaces.md` — domain types and ports

## Local development

```bash
mise install
docker-compose up -d
mise run setup
mise run dev
```

Open http://localhost:3000

## Common commands

```bash
mise run dev
mise run build
go test ./...
go build ./...
mise run db-migrate
mise run db-rollback
```
