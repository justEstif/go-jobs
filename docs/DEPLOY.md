# Deploying go-jobs

Three ways to run go-jobs in production:

1. [Docker Compose](#1-docker-compose-recommended-for-self-hosters) — easiest, runs the published image with Postgres
2. [Plain Docker](#2-plain-docker) — if you already have a Postgres instance
3. [Dokku](#3-dokku) — git-push deploys on your own server

---

## Prerequisites

All deployment methods need the same set of secrets. Generate them once:

```bash
# Postgres password (any string — used by docker-compose only)
openssl rand -hex 16

# CSRF_KEY — 32 hex chars
openssl rand -hex 16

# SESSION_SECRET — 32 hex chars
openssl rand -hex 16

# ENCRYPTION_KEY — 64 hex chars
openssl rand -hex 32
```

---

## 1. Docker Compose (recommended for self-hosters)

This pulls the published image from GHCR and spins up Postgres alongside it.

**Step 1 — Create your env file**

```bash
cp scripts/env.production.example .env.production
# Fill in the values you generated above
$EDITOR .env.production
```

**Step 2 — Start the stack**

```bash
docker compose -f docker-compose.prod.yml --env-file .env.production up -d
```

The app is now running at `http://localhost:3000`.

**To update to the latest image:**

```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml --env-file .env.production up -d
```

**To stop:**

```bash
docker compose -f docker-compose.prod.yml down
```

> **Note:** Put a reverse proxy (nginx, Caddy) in front to terminate TLS and point your domain at port 3000.

---

## 2. Plain Docker

If you already have a Postgres database running elsewhere:

```bash
docker pull ghcr.io/justestif/go-jobs:latest

docker run -d \
  --name go-jobs \
  --restart unless-stopped \
  -p 3000:3000 \
  -e DATABASE_URL="postgres://user:pass@host:5432/go-jobs?sslmode=disable" \
  -e PORT=3000 \
  -e BASE_URL="https://jobs.yourdomain.com" \
  -e CSRF_KEY="<32 hex chars>" \
  -e SESSION_SECRET="<32 hex chars>" \
  -e ENCRYPTION_KEY="<64 hex chars>" \
  ghcr.io/justestif/go-jobs:latest
```

**Migrations** run automatically on startup via the container entrypoint. If you need to run them manually:

```bash
docker run --rm \
  -e DATABASE_URL="postgres://..." \
  ghcr.io/justestif/go-jobs:latest \
  migrate -path /app/migrations -database $DATABASE_URL up
```

---

## 3. Dokku

**Step 1 — Install Dokku plugins** (once, on the server)

```bash
sudo dokku plugin:install https://github.com/dokku/dokku-postgres.git postgres
```

**Step 2 — Provision the app** (once, from your dev machine)

```bash
export DOKKU_HOST=<your-server-ip-or-hostname>
export APP_NAME=go-jobs
export DOMAIN=jobs.yourdomain.com

# Copy and fill in your secrets
cp scripts/env.production.example .env.production
$EDITOR .env.production

# Run the provisioning script
./scripts/deploy-app.sh
```

This will:
- Create the Dokku app and set the domain
- Create and link a Postgres database
- Set `PORT`, `BASE_URL`, and all secrets from `.env.production`
- Add a `dokku` git remote to your local repo

**Step 3 — Deploy**

```bash
git push dokku main
```

Dokku will build the Docker image, run `migrate up` as a release phase (via `app.json`), then swap traffic to the new container.

**To redeploy after changes:**

```bash
git push dokku main
```

**To set or update a secret:**

```bash
ssh dokku@$DOKKU_HOST config:set go-jobs ENCRYPTION_KEY=<new-value>
```

**To enable HTTPS** (Let's Encrypt):

```bash
sudo dokku plugin:install https://github.com/dokku/dokku-letsencrypt.git
ssh dokku@$DOKKU_HOST letsencrypt:enable go-jobs
```

---

## Operations

### Process architecture

The app runs as two separate Dokku process types defined in `Procfile`:

| Process | Command | Role |
|---------|---------|------|
| `web` | `serve` | HTTP server only — no background work |
| `worker` | `scrape --loop --enrich` | Scrapes all companies (20 concurrent goroutines), then enriches un-tagged jobs, repeats on `SCRAPE_INTERVAL` |

Scale them independently:

```bash
ssh dokku@$DOKKU_HOST ps:scale go-jobs web=1 worker=1
```

### Force a manual scrape

`dokku run` spins up a one-off container from the current image without touching the running processes:

```bash
# Scrape + enrich (mirrors what the worker does each cycle)
ssh dokku@$DOKKU_HOST run go-jobs scrape --enrich

# Scrape only
ssh dokku@$DOKKU_HOST run go-jobs scrape

# Enrich only (e.g. after adding an LLM key)
ssh dokku@$DOKKU_HOST run go-jobs enrich
```

### Watch live logs

```bash
# All processes
ssh dokku@$DOKKU_HOST logs go-jobs -t

# Worker only
ssh dokku@$DOKKU_HOST logs go-jobs --ps worker -t
```

### Tuning the scrape interval and enrich limit

```bash
# Change scrape interval to 12 hours (takes effect on next worker restart)
ssh dokku@$DOKKU_HOST config:set go-jobs SCRAPE_INTERVAL=12h

# Raise enrichment limit per cycle
ssh dokku@$DOKKU_HOST config:set go-jobs ENRICH_LIMIT=2000
```

`config:set` triggers an automatic redeploy, so the new value is picked up immediately.

---

## Environment variable reference

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | ✅ | Postgres connection string. Set automatically by Dokku; set manually otherwise. |
| `PORT` | ✅ | HTTP listen port. Use `3000`. |
| `BASE_URL` | ✅ | Public URL of the app, no trailing slash. e.g. `https://jobs.yourdomain.com` |
| `CSRF_KEY` | ✅ | 32 hex chars (`openssl rand -hex 16`). Signs CSRF cookies. |
| `SESSION_SECRET` | ✅ | 32 hex chars (`openssl rand -hex 16`). Signs session cookies. |
| `ENCRYPTION_KEY` | ✅ | 64 hex chars (`openssl rand -hex 32`). AES-256-GCM key for stored LLM API keys. |
| `POSTGRES_PASSWORD` | docker-compose only | Postgres superuser password for the managed container. |
| `SCRAPE_INTERVAL` | ❌ | How often the worker re-scrapes. Default: `6h`. Any Go duration string (`1h`, `30m`, …). |
| `ENRICH_LIMIT` | ❌ | Max jobs enriched per cycle. Default: `1000`. |
