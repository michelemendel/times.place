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

# Run from image workdir /app; migrations are at backend/db/migrations
goose -dir /app/backend/db/migrations postgres "$URL" up
