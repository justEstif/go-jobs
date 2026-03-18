---
# go-jobs-kt7h
title: Set up Dokku deployment
status: todo
type: task
priority: normal
created_at: 2026-03-18T21:56:57Z
updated_at: 2026-03-18T22:04:39Z
---

Set up Dokku deployment for go-jobs using the provisioning script as a base.

## Reference Script

https://gist.githubusercontent.com/justEstif/34a8f8931c763e2ef350ec319f55dfe3/raw/b9f7f3912128c3ab877f7260259da292462c4f25/deploy-app.sh

## What the script does

1. Creates the Dokku app
2. Sets domain
3. Creates and links a Postgres DB
4. Configures log rotation
5. Sets base env vars (`PORT`, `BASE_URL`)
6. Loads extra vars from `.env.production`
7. Adds/updates the `dokku` git remote

## Adaptations needed for go-jobs

### Core files
- [x] Write multi-stage `Dockerfile` (tailwind → templ → go build → scratch final image)
- [x] Write `docker-compose.prod.yml` for self-hosters (pulls from GHCR, wires Postgres)
- [ ] Write `scripts/deploy-app.sh` with project-specific config (APP_NAME, DOKKU_HOST, DOMAIN, PORT=3000)
- [x] Create `scripts/env.production.example` listing all required env vars

### CI / publishing
- [x] Add `.github/workflows/docker-publish.yml` — builds and pushes to GHCR on push to main and on version tags
- [ ] Confirm GHCR package visibility is public after first push

### Dokku-specific
- [ ] Add `app.json` with a `release` phase command to run migrations before traffic switches
- [ ] Set `PORT=3000` in Dokku config (script defaults to 5000)
- [ ] Confirm `NEEDS_POSTGRES=true`, `NEEDS_REDIS=false`

### Env vars to document (in .env.production.example)
- `DATABASE_URL` — set automatically by Dokku postgres:link, needed manually otherwise
- `CSRF_KEY` — 32-byte hex: `openssl rand -hex 16`
- `SESSION_SECRET` — `openssl rand -hex 16`
- `ENCRYPTION_KEY` — 64-byte hex: `openssl rand -hex 32`
- `PORT` — 3000
- `BASE_URL` — e.g. https://go-jobs.your.domain

### Docs
- [ ] Write `docs/DEPLOY.md` covering: Docker pull & run, docker-compose self-host, and Dokku deploy
