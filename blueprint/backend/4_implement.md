# Implementation Log (Backend)

This file will track backend implementation work sessions, decisions made during coding, and any deviations from `backend/2_plan.md`.

## Template

### YYYY-MM-DD

#### Summary

- What changed?
- Why?

#### Notes

- Migrations:
- sqlc:
- Auth/JWT:
- API:
- Dev container:

## 2026-01-20

### Summary

- Updated backend blueprint to use **short-lived access JWT + refresh tokens** (instead of “discard JWT to logout”).
- Added refresh/logout endpoints and a `refresh_tokens` DB table specification to support real server-side logout.

### Notes

- **Migrations**: Backend schema spec now includes `refresh_tokens` (hashed token storage, rotation support).
- **sqlc**: Upcoming sqlc query set will need refresh-token CRUD (lookup by hash, revoke, rotate).
- **Auth/JWT**:
  - Access token: short-lived JWT used on API calls.
  - Refresh token: long-lived opaque secret, recommended via HttpOnly cookie, stored hashed in DB.
- **API**:
  - Added `POST /api/auth/refresh` and `POST /api/auth/logout` to spec.
  - Updated register/login responses to include refresh-token delivery (cookie or response field; cookie recommended).
- **Dev container**: No devcontainer changes yet (docs only).

## 2026-01-21

### Summary

- Expanded backend documentation around local development and production deployment.
- Documented Render.com production setup (Option B: Go serves frontend + `/api`) and a GitHub Actions CI/CD approach using Render deploy hooks.
- Updated repo `README.md` with developer prerequisites and clarified `goose`/`sqlc` as CLI tools.

### Notes

- **Dev workflow / HMR**:
  - Documented the two-process local dev setup: SvelteKit dev server (HMR) + Go API, with `/api` proxied during development.
  - Emphasized using relative `/api/...` URLs so the same frontend code works in dev (proxy) and prod (served by Go).
- **Migrations + sqlc**:
  - Documented the roles of `backend/db/migrations/`, `backend/db/queries/`, and `backend/db/sqlc/`, including which parts are authored vs generated and how the tools depend on them.
- **Production (Render.com)**:
  - Documented the Render resources (Web Service + Postgres), and migration timing (`goose up` as part of deploy).
  - Added tasks for Render configuration, environment variables, and verification of `/` + `/api` routing.

## 2026-01-22

### Summary

- Added technical implementation details for how Go serves frontend assets (embed directive, runtime flag).
- Documented environment variable management for both local development and production on Render.com.
- Clarified build order requirement (frontend must be built before backend due to Go embed directive).

### Notes

- **Frontend serving (technical)**:
  - Documented using Go's `embed` directive to embed frontend build assets at compile time.
  - Documented `SERVE_FRONTEND` runtime flag to control whether Go serves frontend (false in dev, true in prod).
  - Clarified build order: frontend must be built before backend (Go embed reads files during compilation).
  - Same binary works in both environments; no separate build targets needed.
- **Environment variables**:
  - Documented local dev setup: `.env` file (gitignored), `.env.example` template, devcontainer support.
  - Documented production setup: Render dashboard configuration, secret vs non-secret variables, cookie settings.
  - Listed required variables: `DATABASE_URL`, `JWT_SECRET`, `REFRESH_TOKEN_SECRET` (secrets), `SERVE_FRONTEND`, `LOG_LEVEL`, cookie settings (non-secrets).

### Connection issues

#### Summary

- Fixed external database connection issues (pgAdmin, Warp terminal) by adding a proxy port through the backend container.
- Cursor's port forwarding for docker-compose dependent services (postgres) doesn't work reliably, so we use Docker's native port mapping instead.

#### Notes

- **Dev container**:
  - Added `socat` to backend container to proxy connections from backend:5432 to postgres:5432.
  - Added port mapping `5433:5432` on backend service (host port 5433 → container port 5432).
  - Created `start-with-proxy.sh` script that runs socat in background, then sleeps.
  - External tools (pgAdmin) connect via `localhost:5433`, which routes through backend container to postgres.
  - Direct postgres port (5432) may not work due to Cursor port forwarding limitations.
  - Added `make bdbproxy` target to test proxy connection.

## 2026-01-25

### Summary

- **Database schema & migrations**: Created complete database schema with 4 migrations covering all core tables, indexes, constraints, and triggers.
- **Makefile database commands**: Set up comprehensive Makefile targets for database operations (migrations, seeding, verification, connection).
- **Test data infrastructure**: Created seed data system for both CLI usage and automated testing with transaction isolation.
- **Migration rollback fix**: Discovered that `dbgoosedown` only rolls back one migration at a time, added `dbgoosereset` command to rollback all migrations at once.

### Notes

- **Migrations** (created 4 migration files in `backend/db/migrations/`):
  - `00001_enable_pgcrypto.sql`: Enables pgcrypto extension for UUID generation
  - `00002_create_tables.sql`: Creates core schema:
    - `venue_owners` (owner_uuid, name, mobile, email, password_hash, timestamps)
    - `venues` (venue_uuid, owner_uuid FK, name, banner_image, address, geolocation, comment, timezone, visibility, private_link_token, timestamps)
    - `event_lists` (event_list_uuid, venue_uuid FK, name, date, comment, visibility, private_link_token, sort_order, timestamps)
    - `events` (event_uuid, event_list_uuid FK, event_name, datetime, comment, duration_minutes, sort_order, timestamps)
    - `refresh_tokens` (refresh_token_uuid, owner_uuid FK, token_hash, issued_at, expires_at, revoked_at, replaced_by_token_uuid, user_agent, ip_address)
  - `00003_add_indexes_and_constraints.sql`: Adds performance indexes:
    - Foreign key indexes (venues.owner_uuid, event_lists.venue_uuid, events.event_list_uuid, refresh_tokens.owner_uuid)
    - Visibility indexes for public/private filtering
    - Partial indexes for private_link_token (WHERE token IS NOT NULL)
    - Composite indexes for sort_order queries (venue_uuid + sort_order, event_list_uuid + sort_order)
    - Unique index on refresh_tokens.token_hash
  - `00004_add_modified_at_trigger.sql`: Creates `update_modified_at()` function and triggers for automatic `modified_at` timestamp updates on all tables
  - All migrations include proper `-- +goose Down` sections for rollback support
- **Makefile database targets**:
  - `dbgooseup`: Apply all pending migrations
  - `dbgoosedown`: Rollback last migration (one at a time)
  - `dbgoosereset`: Rollback ALL migrations (drops all tables) - added today
  - `dbgoosestatus`: Show migration status
  - `dbgoosecreate`: Create new migration file
  - `dbverify`: Verify schema (shows migration status, tables, indexes)
  - `dbseed`: Seed test data into database
  - `dbseedclear`: Clear existing data and seed test data
  - `dbconnect`: Connect to database with psql
  - All commands work from host (via docker exec) or inside devcontainer
- **Test data infrastructure**:
  - `backend/internal/testdata/seed.go`: Test data seeding functions with realistic sample data
  - `backend/cmd/cli/seed/main.go`: CLI tool for seeding database (supports `-clear` flag)
  - `backend/internal/test/helpers.go`: Test database helpers with transaction isolation
  - Test helpers use database transactions for automatic cleanup (rollback after each test)
- **Migration rollback behavior**:
  - `goose down` command only rolls back the last applied migration (one at a time)
  - After running `dbgooseup` (applies migrations 1-3) and then `dbgoosedown` (rolls back only migration 3), tables from migration 2 remain in the database
  - Added `make dbgoosereset` target that uses `goose reset` to rollback all migrations at once
  - Updated Makefile help text to clarify that `dbgoosedown` rolls back one migration at a time
  - Use `make dbgoosedown` for incremental rollbacks, `make dbgoosereset` for full reset

