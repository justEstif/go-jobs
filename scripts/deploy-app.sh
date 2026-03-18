#!/bin/bash
set -euo pipefail

# ─────────────────────────────────────────────
# go-jobs Dokku provisioning script
# Run once from your dev machine to set up the
# app on your Dokku server.
#
# Required env vars (set in your shell or .env.production):
#   DOKKU_HOST   — server IP or hostname (e.g. 1.2.3.4 or dokku.myserver.com)
#   APP_NAME     — Dokku app name (e.g. go-jobs)
#   DOMAIN       — public domain (e.g. jobs.yourdomain.com)
#
# Usage:
#   export DOKKU_HOST=... APP_NAME=... DOMAIN=...
#   ./scripts/deploy-app.sh
#
# Or with an env file:
#   env $(grep -v '^#' .env.production | xargs) ./scripts/deploy-app.sh
# ─────────────────────────────────────────────

: "${DOKKU_HOST:?DOKKU_HOST is required}"
: "${APP_NAME:?APP_NAME is required}"
: "${DOMAIN:?DOMAIN is required}"

DB_NAME="${APP_NAME}-db"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="${1:-${SCRIPT_DIR}/../.env.production}"

dokku_cmd() {
  ssh "dokku@${DOKKU_HOST}" "$@"
}

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Deploying: $APP_NAME → $DOMAIN"
echo "  Server:    $DOKKU_HOST"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "▶ Creating app: $APP_NAME"
dokku_cmd apps:create "$APP_NAME" || echo "  (already exists)"

echo "▶ Setting domain: $DOMAIN"
dokku_cmd domains:set "$APP_NAME" "$DOMAIN"

echo "▶ Creating Postgres database: $DB_NAME"
dokku_cmd postgres:create "$DB_NAME" || echo "  (already exists)"
echo "▶ Linking database"
dokku_cmd postgres:link "$DB_NAME" "$APP_NAME" || echo "  (already linked)"

echo "▶ Configuring log rotation"
dokku_cmd docker-options:add "$APP_NAME" deploy,run \
  "--log-opt max-size=50m --log-opt max-file=10"

echo "▶ Setting base env vars"
dokku_cmd config:set --no-restart "$APP_NAME" \
  PORT=3000 \
  BASE_URL="https://${DOMAIN}"

if [ -f "$ENV_FILE" ]; then
  echo "▶ Loading env vars from ${ENV_FILE}"
  env_args=()
  while IFS= read -r line || [ -n "$line" ]; do
    [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
    env_args+=("$line")
  done < "$ENV_FILE"
  if [ ${#env_args[@]} -gt 0 ]; then
    dokku_cmd config:set --no-restart "$APP_NAME" "${env_args[@]}"
    echo "  Set ${#env_args[@]} variable(s)"
  fi
else
  echo "  ⚠ No ${ENV_FILE} found — skipping app secrets"
  echo "  Set CSRF_KEY, SESSION_SECRET, ENCRYPTION_KEY manually:"
  echo "    dokku config:set $APP_NAME CSRF_KEY=... SESSION_SECRET=... ENCRYPTION_KEY=..."
fi

# Add git remote if not already set
if ! git remote get-url dokku &>/dev/null 2>&1; then
  echo "▶ Adding git remote"
  git remote add dokku "dokku@${DOKKU_HOST}:${APP_NAME}"
  echo "  ✓ Remote added"
else
  echo "▶ Updating git remote"
  git remote set-url dokku "dokku@${DOKKU_HOST}:${APP_NAME}"
  echo "  ✓ Remote updated"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅  App provisioned!"
echo ""
echo "  Deploy with:"
echo "    git push dokku main"
echo ""
echo "  Live at: https://${DOMAIN}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
