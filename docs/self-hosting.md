# Self-hosting go-jobs

The easiest way to run go-jobs is with the published Docker image.

## Docker Compose (recommended)

**1. Copy and fill in the env file**

```bash
cp scripts/env.production.example .env.production
$EDITOR .env.production
```

Required variables:

| Variable | How to generate |
|---|---|
| `POSTGRES_PASSWORD` | `openssl rand -hex 16` |
| `BASE_URL` | e.g. `https://jobs.yourdomain.com` |
| `CSRF_KEY` | `openssl rand -hex 16` |
| `SESSION_SECRET` | `openssl rand -hex 16` |
| `ENCRYPTION_KEY` | `openssl rand -hex 32` |

**2. Start the stack**

```bash
docker compose -f docker-compose.prod.yml --env-file .env.production up -d
```

The app runs on port `3000`. Put Caddy or nginx in front for TLS.

**To update to the latest image:**

```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml --env-file .env.production up -d
```

## Plain Docker

If you already have a Postgres database:

```bash
docker pull ghcr.io/justestif/go-jobs:latest

docker run -d \
  --name go-jobs \
  --restart unless-stopped \
  -p 3000:3000 \
  -e DATABASE_URL="postgres://user:pass@host:5432/go-jobs?sslmode=disable" \
  -e PORT=3000 \
  -e BASE_URL="https://jobs.yourdomain.com" \
  -e CSRF_KEY="..." \
  -e SESSION_SECRET="..." \
  -e ENCRYPTION_KEY="..." \
  ghcr.io/justestif/go-jobs:latest
```

Migrations run automatically on startup.

## Dokku

See [`docs/DEPLOY.md`](./DEPLOY.md#3-dokku) for the full Dokku deploy flow.

## Related docs

- [Full deploy guide](./DEPLOY.md) — Docker Compose, plain Docker, and Dokku
- [Architecture](./ARCHITECTURE.md)
- [Interfaces](./interfaces.md)
