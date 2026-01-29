#!/bin/sh
# Render Pre-Deploy: run goose migrations.
# Strips surrounding double/single quotes from DATABASE_URL so goose can parse it
# (Render or env UIs sometimes paste the URL with quotes, which breaks goose).

set -e

URL="${DATABASE_URL:-}"
# Strip leading/trailing double quotes
case "$URL" in
  '"'*) URL="${URL#\"}" ;;
esac
case "$URL" in
  *'"') URL="${URL%\"}" ;;
esac
# Strip leading/trailing single quotes
case "$URL" in
  "'"*) URL="${URL#\'}" ;;
esac
case "$URL" in
  *"'") URL="${URL%\'}" ;;
esac

if [ -z "$URL" ]; then
  echo "render-pre-deploy: DATABASE_URL is not set"
  exit 1
fi

# Migration dir: in Docker image it is /app/backend/db/migrations (WORKDIR /app)
# Fallback to ./backend/db/migrations in case CWD is repo root (e.g. native build)
MIGRATIONS_DIR="/app/backend/db/migrations"
if [ ! -d "$MIGRATIONS_DIR" ]; then
  MIGRATIONS_DIR="./backend/db/migrations"
fi
if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "render-pre-deploy: migrations dir not found (tried /app/backend/db/migrations and ./backend/db/migrations)"
  exit 1
fi

echo "render-pre-deploy: running goose from $MIGRATIONS_DIR"
ls -la "$MIGRATIONS_DIR" || true
goose -dir "$MIGRATIONS_DIR" postgres "$URL" up
