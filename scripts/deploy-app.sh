#!/bin/bash
set -euo pipefail

# ─────────────────────────────────────────────
# go-jobs Dokku provisioning script
# Run once from your dev machine to set up the
# app on your Dokku server.
#
# Usage:
#   ./scripts/deploy-app.sh [env-file]
#
# Reads DOKKU_HOST, APP_NAME, DOMAIN, and app secrets from
# .env.production (or the path passed as the first argument).
# Shell env vars take precedence over the file.
# ─────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="${1:-${SCRIPT_DIR}/../.env.production}"

# Load env file early so DOKKU_HOST, APP_NAME, DOMAIN, and secrets are available.
# Shell env vars already set take precedence (${VAR:-} pattern preserves them).
if [ -f "$ENV_FILE" ]; then
  while IFS= read -r line || [ -n "$line" ]; do
    [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
    key="${line%%=*}"
    value="${line#*=}"
    # Only set if not already set in the environment
    if [ -z "${!key+x}" ]; then
      export "$key"="$value"
    fi
  done < "$ENV_FILE"
fi

: "${DOKKU_HOST:?DOKKU_HOST is required (set in ${ENV_FILE} or shell env)}"
: "${APP_NAME:?APP_NAME is required (set in ${ENV_FILE} or shell env)}"
: "${DOMAIN:?DOMAIN is required (set in ${ENV_FILE} or shell env)}"

DB_NAME="${APP_NAME}-db"

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

echo "▶ Configuring port mapping: http:80 → 3000"
dokku_cmd ports:add "$APP_NAME" http:80:3000 || echo "  (already set)"

echo "▶ Configuring log rotation"
dokku_cmd docker-options:add "$APP_NAME" deploy,run \
  "--log-opt max-size=50m --log-opt max-file=10"

echo "▶ Setting base env vars"
dokku_cmd config:set --no-restart "$APP_NAME" \
  PORT=3000 \
  BASE_URL="https://${DOMAIN}"

if [ -f "$ENV_FILE" ]; then
  echo "▶ Loading app secrets from ${ENV_FILE}"
  # Skip vars that are for the deploy script itself or Docker Compose only
  skip_keys="DOKKU_HOST|APP_NAME|DOMAIN|POSTGRES_PASSWORD|BASE_URL"
  env_args=()
  while IFS= read -r line || [ -n "$line" ]; do
    [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
    key="${line%%=*}"
    [[ "$key" =~ ^($skip_keys)$ ]] && continue
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
