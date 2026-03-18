---
# go-jobs-kt7h
title: Set up Dokku deployment
status: todo
type: task
created_at: 2026-03-18T21:56:57Z
updated_at: 2026-03-18T21:56:57Z
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

- [ ] Set `APP_NAME`, `DOKKU_HOST`, `DOMAIN` for the real server
- [ ] Set `PORT=3000` (not 5000 — the script defaults to 5000)
- [ ] Confirm `NEEDS_POSTGRES=true`, `NEEDS_REDIS=false` (matches current stack)
- [ ] Create `scripts/deploy-app.sh` with project-specific config baked in
- [ ] Create `scripts/.env.production.example` listing all required env vars:
  - `DATABASE_URL` (set automatically by Dokku postgres:link — may be skippable)
  - `CSRF_KEY` (32-byte hex — `openssl rand -hex 16`)
  - `SESSION_SECRET` (`openssl rand -hex 16`)
  - `ENCRYPTION_KEY` (64-byte hex — `openssl rand -hex 32`)
- [ ] Add a `Dockerfile` (or `Procfile`) so Dokku knows how to build and run the app
- [ ] Verify migrations run on deploy (add a `release` phase or startup hook)
- [ ] Document deploy steps in `docs/DEPLOY.md` or `README.md`
