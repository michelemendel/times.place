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
  - Clarified build order: frontend must be built before backend (Go embed reads files during compilation).
  - Same binary works in both environments; no separate build targets needed.
- **Environment variables**:
  - Documented local dev setup: `.env` file (gitignored), `.env.example` template, devcontainer support.
  - Documented production setup: Render dashboard configuration, secret vs non-secret variables, cookie settings.
  - Listed required variables: `DATABASE_URL`, `JWT_SECRET`, `REFRESH_TOKEN_SECRET` (secrets), `LOG_LEVEL`, cookie settings (non-secrets).

### Connection issues

#### Summary

- Fixed external database connection issues (pgAdmin, Warp terminal) by using Docker's native port mapping.
- Direct postgres port (5432) works reliably from host, so proxy is not needed.

#### Notes

- **Dev container**:
  - Postgres service exposes port `5432:5432` directly (host:container).
  - External tools (pgAdmin, CLI) connect via `localhost:5433` (proxy) or `localhost:5434` (direct to postgres).
  - Docker's native port mapping works reliably, so no proxy needed.
  - (Previously used socat proxy on port 5433, but removed as unnecessary.)

## 2026-01-25

### Summary

- **Database schema & migrations**: Created complete database schema with 4 migrations covering all core tables, indexes, constraints, and triggers.
- **Makefile database commands**: Set up comprehensive Makefile targets for database operations (migrations, seeding, verification, connection).
- **Test data infrastructure**: Created seed data system for both CLI usage and automated testing with transaction isolation.
- **Migration rollback fix**: Discovered that `dbdown` only rolls back one migration at a time, added `dbreset` command to rollback all migrations at once.

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
  - `dbup`: Apply all pending migrations
  - `dbdown`: Rollback last migration (one at a time)
  - `dbreset`: Rollback ALL migrations (drops all tables) - added today
  - `dbstatus`: Show migration status
  - `goosecreate`: Create new migration file
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
  - After running `dbup` (applies migrations 1-3) and then `dbdown` (rolls back only migration 3), tables from migration 2 remain in the database
  - Added `make dbreset` target that uses `goose reset` to rollback all migrations at once
  - Updated Makefile help text to clarify that `dbdown` rolls back one migration at a time
  - Use `make dbdown` for incremental rollbacks, `make dbreset` for full reset

## 2026-01-27

### Summary

- **Cursor skill creation**: Created `update-backend-implement-log` skill to automate updating this implementation log with backend work.

### Notes

- **Cursor skill**:
  - Created `.cursor/skills/update-backend-implement-log/SKILL.md` skill file
  - Skill guides the agent through identifying recent backend changes, organizing them by category, and adding properly formatted entries to this log
  - Skill automatically triggers when backend code changes need documentation or when explicitly requested
  - Follows the established format and template structure for consistency

### Summary- **sqlc configuration and queries**: Set up complete sqlc infrastructure with configuration, consolidated schema, and all required query files for CRUD operations, public endpoints, and token lookups.
- **Makefile test database setup**: Updated `btest` and `btestcover` targets to automatically create and migrate a separate test database before running tests.

### Notes

- **sqlc**:
  - Created `backend/sqlc.yaml` configuration file with PostgreSQL settings, pointing to `db/queries/` and `db/schema.sql`, outputting to `db/sqlc/` with pgx/v5 driver
  - Created `backend/db/schema.sql` consolidating all table definitions, indexes, and constraints from migrations (serves as schema source for sqlc)
  - Created 6 query files in `backend/db/queries/`:
    - `owners.sql`: CreateOwner, GetOwnerByID, GetOwnerByEmail (case-insensitive)
    - `venues.sql`: ListVenuesByOwner, GetVenueByIDAndOwner, CreateVenue, UpdateVenue, DeleteVenue
    - `event_lists.sql`: ListEventListsByVenueAndOwner, GetEventListByIDAndOwner, CreateEventList, UpdateEventList, DeleteEventList
    - `events.sql`: ListEventsByEventListAndOwner, GetEventByIDAndOwner, CreateEvent, UpdateEvent, DeleteEvent
    - `public.sql`: ListPublicVenues, SearchPublicVenues, GetPublicEventListsByVenue, GetVenueByToken, GetEventListByToken, GetVenueWithEventListsByToken
    - `refresh_tokens.sql`: CreateRefreshToken, GetRefreshTokenByHash, RevokeRefreshToken, RevokeRefreshTokenByHash, RotateRefreshToken, RevokeAllTokensForOwner
  - Generated sqlc code successfully: 8 Go files in `backend/db/sqlc/` with 31 query functions
  - All queries include proper authorization checks (owner-scoped queries verify ownership through JOINs)
  - Public queries exclude owner contact info (email, mobile)
  - Added `github.com/jackc/pgx/v5` dependency to `go.mod`
  - Generated code compiles successfully
- **Makefile**:
  - Updated `btest` and `btestcover` targets to work from host (via docker exec) or inside devcontainer
  - Both targets now automatically:
    - Create/reset `timesplace_test` database before running tests
    - Run migrations on test database using goose
    - Execute tests against the isolated test database
  - Uses `TEST_DATABASE_URL` environment variable (defaults to `timesplace_test` database)
  - Ensures clean test environment for each test run### Summary

- **API server infrastructure**: Implemented complete Echo-based API server with project layout, database connection, middleware, error handling, and request validation.
- **Authentication endpoints**: Implemented all 5 auth endpoints (register, login, refresh, logout, me) with password hashing, JWT tokens, refresh token management, and HttpOnly cookies.

### Notes

- **Project layout** (created `backend/internal/` structure):
  - `internal/store/store.go`: Database connection wrapper using pgxpool, wraps sqlc Queries
  - `internal/service/auth.go`: Auth service with password hashing (bcrypt), JWT generation/parsing, refresh token management
  - `internal/http/`: HTTP layer with Echo server setup:
    - `server.go`: Echo server initialization, middleware setup, graceful shutdown, environment variable loading
    - `routes.go`: Route registration for all API endpoints
    - `middleware.go`: JWT authentication middleware that extracts `owner_uuid` from token and injects into context
    - `errors.go`: Consistent error response format with helper functions (ValidationError, UnauthorizedError, etc.)
    - `validator.go`: Custom request validator with email format validation
    - `auth_handlers.go`: All authentication endpoint handlers
  - `cmd/api/main.go`: Updated to initialize and run the server
- **Dependencies** (added to `go.mod`):
  - `github.com/labstack/echo/v4`: Echo web framework
  - `github.com/joho/godotenv`: Environment variable loading
  - `github.com/golang-jwt/jwt/v5`: JWT token handling
  - `github.com/go-playground/validator/v10`: Request validation
  - `github.com/jackc/pgx/v5/pgxpool`: Database connection pooling
- **Auth/JWT**:
  - Access tokens: Short-lived JWT (15 minutes) with `sub` claim containing `owner_uuid`
  - Refresh tokens: Long-lived (30 days), stored hashed in database, delivered via HttpOnly cookies
  - Token rotation: Refresh endpoint rotates tokens (revokes old, creates new)
  - Password hashing: bcrypt with cost 12
  - JWT middleware: Extracts token from `Authorization: Bearer <token>` header, validates signature, injects `owner_uuid` into Echo context
- **API endpoints** (all implemented in `auth_handlers.go`):
  - `POST /api/auth/register`: Validates input, checks email uniqueness, hashes password, creates owner, generates tokens, stores refresh token, sets cookie
  - `POST /api/auth/login`: Validates credentials, generates tokens, stores refresh token, sets cookie
  - `POST /api/auth/refresh`: Validates refresh token, rotates tokens, generates new access token, sets new cookie
  - `POST /api/auth/logout`: Revokes refresh token, clears cookie
  - `GET /api/auth/me`: Protected endpoint that returns current owner profile (requires JWT middleware)
- **Error handling**:
  - Consistent error response format: `{ "error": { "code": "...", "message": "..." } }`
  - Error codes: `validation_error`, `unauthorized`, `forbidden`, `not_found`, `conflict`, `internal`
  - Proper HTTP status codes (400, 401, 403, 404, 409, 500)
- **Request validation**:
  - Custom validator with email format validation
  - Field-level validation using struct tags (`validate:"required,email"`)
  - Returns validation errors in consistent format
- **Cookie configuration**:
  - Refresh tokens stored in HttpOnly cookies (prevents XSS)
  - Configurable via environment variables: `COOKIE_DOMAIN`, `COOKIE_SECURE`, `COOKIE_SAME_SITE`
  - Secure cookies in production, lax SameSite by default
- **UUID conversion**:
  - Helper functions to convert between `pgtype.UUID` (database) and string UUIDs (API)
  - Helper functions to convert `pgtype.Timestamptz` to RFC3339 strings
- **Build fixes**:
  - Resolved circular import by moving handlers into `http` package (renamed `handlers/auth.go` to `auth_handlers.go`)
  - Fixed import aliases for `sqlc` package
  - All packages compile successfully

## 2026-01-28

### Summary

- **Auth unit tests**: Added focused unit tests covering `AuthService` (password hashing + JWT/refresh-token utilities), JWT middleware auth/header parsing, and refresh-token cookie helpers.
- **Why**: Lock in auth behavior/edge cases so we can refactor auth + HTTP layers safely.

### Notes

- **Auth/JWT (unit tests)**:
  - Added `backend/internal/service/auth_test.go`:
    - Password hashing + verification (`HashPassword`, `VerifyPassword`)
    - Access JWT generation/parsing (`GenerateAccessToken`, `ParseAccessToken`) including expiration and invalid signature cases
    - Refresh token generation + hashing (`GenerateRefreshToken`, `HashRefreshToken`)
    - `NewAuthService` env var behavior (`JWT_SECRET` required; `REFRESH_TOKEN_SECRET` optional fallback)
- **API / HTTP helpers (unit tests)**:
  - Added `backend/internal/http/auth_handlers_test.go`:
    - Refresh token cookie set/clear behavior, including env-controlled `COOKIE_DOMAIN`, `COOKIE_SECURE`, and `COOKIE_SAME_SITE`
    - Refresh token extraction precedence: cookie first, then JSON body fallback (`refresh_token`)
- **Middleware (unit tests)**:
  - Added `backend/internal/http/middleware_test.go`:
    - `JWTAuthMiddleware` behavior for valid/missing/invalid `Authorization` header formats
    - Context propagation + `GetOwnerUUIDFromContext` validation/error cases

### Summary

- **Owner-scoped CRUD endpoints**: Implemented protected CRUD for venues, event lists, and events (including nested list/create routes and `sort_order` updates).
- **Public endpoints**: Implemented public browse/search plus token-based access endpoints for venues and event lists.
- **Why**: Unblock the frontend from localStorage by providing the core read/write API surface with ownership enforcement and public sharing flows.

### Notes

- **API (owner-scoped CRUD)**:
  - Added venue handlers in `backend/internal/http/venue_handlers.go`:
    - `GET /api/venues`, `POST /api/venues`, `GET /api/venues/:venue_uuid`, `PATCH /api/venues/:venue_uuid`, `DELETE /api/venues/:venue_uuid`
  - Added event list handlers in `backend/internal/http/event_list_handlers.go`:
    - `GET /api/venues/:venue_uuid/event-lists`, `POST /api/venues/:venue_uuid/event-lists`
    - `GET /api/event-lists/:event_list_uuid`, `PATCH /api/event-lists/:event_list_uuid`, `DELETE /api/event-lists/:event_list_uuid`
  - Added event handlers in `backend/internal/http/event_handlers.go`:
    - `GET /api/event-lists/:event_list_uuid/events`, `POST /api/event-lists/:event_list_uuid/events`
    - `GET /api/events/:event_uuid`, `PATCH /api/events/:event_uuid`, `DELETE /api/events/:event_uuid`
  - Ownership is enforced via `JWTAuthMiddleware` + owner-scoped sqlc queries (JOIN-based checks); handlers return `404` on ownership mismatch to avoid leaking resource existence.
  - Cascade deletion relies on DB `ON DELETE CASCADE` constraints (venue → event_lists → events).
  - Ordering uses `sort_order` fields and existing ORDER BY behavior in sqlc queries.

- **API (public)**:
  - Added `backend/internal/http/public_handlers.go`:
    - `GET /api/public/venues` (supports `?query=` using `SearchPublicVenues`, otherwise `ListPublicVenues`)
    - `GET /api/public/venues/:venue_uuid/event-lists` (public-only)
    - `GET /api/public/venues/by-token/:token` (venue token returns venue + event lists)
    - `GET /api/public/event-lists/by-token/:token` (event list token returns venue + event list + events)
  - Public responses omit owner contact information (email/mobile).

- **Routing**:
  - Updated `backend/internal/http/routes.go` to register new route groups:
    - Protected: `/api/venues`, `/api/event-lists`, `/api/events`
    - Public: `/api/public/*`

- **Build fixes**:
  - Adjusted `pgtype.Date` handling to match pgx/v5 (`pgtype.Date.Time`), and removed a few unused imports/variables so `make bbuild` passes cleanly.

### Summary

- **Health check endpoint**: Added `/health` endpoint for monitoring and deployment health checks.
- **Public events endpoint**: Added endpoint to retrieve events for public event lists without requiring authentication.
- **Dev workflow improvements**: Enhanced environment variable loading to support both workspace root and backend directory locations.

### Notes

- **API (health check)**:
  - Added `backend/internal/http/healthcheck.go`: Health check handler with database connectivity check.
  - `GET /health`: Returns health status with timestamp and optional database connection status.
  - Returns `200 OK` with `{"status": "ok", "timestamp": "...", "database": "connected"}` when healthy.
  - Returns `503 Service Unavailable` with `{"status": "degraded", "database": "unavailable"}` when database is unreachable.
  - Public endpoint (no authentication required) for monitoring and deployment verification.
- **API (public)**:
  - Added `GET /api/public/event-lists/:event_list_uuid/events` endpoint in `backend/internal/http/public_handlers.go`:
    - Returns events for a public event list (only if event list visibility is "public").
    - Returns `404` if event list is private or not found (doesn't leak existence of private lists).
    - Enables frontend to load events separately from event list data for better performance.
- **Dev workflow**:
  - Updated `backend/internal/http/server.go`: Environment variable loading now tries both `.env` (workspace root) and `backend/.env` (backend directory) for flexibility when running from different locations.
- **Test data**:
  - Updated `backend/internal/testdata/seed.go`: Added "DEMO:" prefix to all venue names in test data to clearly distinguish demo venues from production data.

## 2026-01-29

### Summary

- **Frontend serving simplification**: Removed the `SERVE_FRONTEND` runtime flag; the backend now always attempts to serve the built frontend (when `frontend/build/` exists) and otherwise serves the API normally.
- **Dev workflow clarity**: Updated Makefile targets and documentation to remove `SERVE_FRONTEND` references while keeping the two-process HMR workflow intact (`make fstart` + `make bstart`).

### Notes

- **API / routing**:
  - Updated `backend/internal/http/routes.go` to always register frontend routes after `/api/*` routes, keeping the existing “warn but don’t fail” behavior when `frontend/build/` is missing.
- **Dev workflow / Makefile**:
  - Updated `Makefile` so `bstart` no longer sets `SERVE_FRONTEND=false`.
  - Updated `Makefile` so `pstart` no longer sets `SERVE_FRONTEND=true` (it remains a convenience wrapper that errors if `frontend/build/` is missing).
- **Dev container / env vars**:
  - Removed `SERVE_FRONTEND` from `.devcontainer/devcontainer.json` and `backend/.env.example`.
- **Docs**:
  - Updated `backend/README.md` and backend blueprint docs to remove `SERVE_FRONTEND` and describe the new “serve built assets if present” behavior.

### Summary

- **Frontend build path resolution**: Made serving the frontend from the backend robust across different working directories (workspace root, backend/, devcontainer vs host). Added `FRONTEND_BUILD_DIR` env support and CWD-based candidate paths, with startup logging when the build dir is found.

### Notes

- **API / frontend serving** (`backend/internal/http/frontend.go`):
  - `getFrontendFS()` now returns `(fs.FS, string, error)` (added absolute path used for logging).
  - **FRONTEND_BUILD_DIR**: If set (absolute or relative to CWD), that path is used first to locate `frontend/build`.
  - **CWD-based candidates**: Build path is resolved from `os.Getwd()`: `cwd/frontend/build`, `cwd/../frontend/build`, `cwd/../../frontend/build`, plus legacy relative paths, so the build is found whether the server runs from workspace root or backend/.
  - **Success logging**: On successful setup, server logs `Serving frontend from <absolute-path>` so operators can confirm which build is served.
- **Routing** (`backend/internal/http/routes.go`):
  - When frontend routes fail to setup, warning message now suggests setting `FRONTEND_BUILD_DIR` to the `frontend/build` path if needed.

### Summary

- **Venue visibility removed**: Dropped `venues.visibility` column and related index via migration. Visibility is only meaningful at event-list level (the only level controllable in the GUI); public venue listing is determined by "at least one public event list" per venue.

### Notes

- **Migrations** (`backend/db/migrations/00005_drop_venues_visibility.sql`):
  - **Up**: Drop index `venues_visibility_idx` and column `venues.visibility`.
  - **Down**: Restore column with default `'public'` and CHECK constraint, and recreate index.
- **Schema** (`backend/db/schema.sql`):
  - Removed `visibility` from `venues` table definition and removed `venues_visibility_idx` index.
- **sqlc / queries**:
  - **venues.sql**: Removed `visibility` from `CreateVenue` INSERT and from `UpdateVenue` SET clause; parameter counts adjusted ($8 for create, $9 for update).
  - **public.sql**: Removed `visibility` from `ListPublicVenues`, `SearchPublicVenues`, `GetVenueByToken`, and `GetVenueWithEventListsByToken` SELECT lists.
  - Ran `sqlc generate`; generated Go types no longer include `Visibility` on venue row/params structs.
- **API** (`backend/internal/http/venue_handlers.go`, `backend/internal/http/public_handlers.go`):
  - Removed `Visibility` from `CreateVenueRequest`, `UpdateVenueRequest`, and `VenueResponse`; removed from `venueToResponse` and from public row-to-response helpers (`venueRowToResponse`, `searchVenueRowToResponse`, `venueTokenRowToResponse`).
  - Create/Update venue handlers no longer pass or read visibility; raw SQL in `GetEventListByToken` (venue by `venue_uuid`) no longer selects or scans `visibility`.
- **Test data** (`backend/internal/testdata/seed.go`):
  - Removed `visibility` column and value from all four venue INSERTs.
- **Integration tests** (`backend/internal/http/integration_test.go`):
  - Removed `"visibility": "public"` from venue create payloads in `TestIntegration_AuthAndVenueCRUD_MinimalFlow` and `TestIntegration_TokenBasedAccessControl`.

## 2026-01-30

### Summary

- **Free-tier venue limit**: Enforced a maximum of 2 venues per owner for the free tier. Limit is configurable via `FREE_TIER_MAX_VENUES`; venue create returns 403 with a clear message when at limit. `GET /api/auth/me` now exposes `venue_count` and `venue_limit` so the frontend can show an upgrade prompt.

### Notes

- **Config** (`backend/internal/http/config.go`):
  - Added `FreeTierMaxVenues()` helper that reads `FREE_TIER_MAX_VENUES` from env (default 2); invalid or missing value falls back to 2.
- **Environment** (`backend/.env.example`):
  - Documented `FREE_TIER_MAX_VENUES=2` (free-tier max venues per owner).
- **sqlc** (`backend/db/queries/venues.sql`):
  - Added `CountVenuesByOwner :one` query; ran `sqlc generate` (new function in `backend/db/sqlc/venues.sql.go`).
- **API (venue create)** (`backend/internal/http/venue_handlers.go`):
  - Before creating a venue, handler calls `CountVenuesByOwner` and compares to `FreeTierMaxVenues()`; if count ≥ limit, returns `403 Forbidden` with message: "Free tier allows at most N venues. Upgrade to add more."
- **API (auth/me)** (`backend/internal/http/auth_handlers.go`):
  - `GET /api/auth/me` response now includes top-level `venue_count` (int64) and `venue_limit` (int64) in addition to `owner`, so the frontend can show upgrade prompts when at limit.
- **Integration test** (`backend/internal/http/integration_test.go`):
  - Added `TestIntegration_FreeTierVenueLimit`: sets `FREE_TIER_MAX_VENUES=2`, creates 2 venues (201), then 3rd create returns 403; verifies `/api/auth/me` returns `venue_count: 2` and `venue_limit: 2`.

## 2026-02-01

### Summary

- **Delete account**: Added `DELETE /api/auth/me` so an authenticated owner can delete their account; handler deletes the owner row and clears the refresh-token cookie.

### Notes

- **sqlc** (`backend/db/queries/owners.sql`):
  - Added `DeleteOwner :exec` query; ran `sqlc generate` (new function in `backend/db/sqlc/owners.sql.go`).
- **API** (`backend/internal/http/auth_handlers.go`):
  - New `DeleteMe` handler: reads owner UUID from JWT context, calls `DeleteOwner`, clears refresh-token cookie, returns `204 No Content`.
- **Routes** (`backend/internal/http/routes.go`):
  - Registered `auth.DELETE("/me", authHandler.DeleteMe, JWTAuthMiddleware(authService))`.

### Summary

- **Demo data locking (schema-based)**: Added `is_demo` to `venue_owners` so seeded demo accounts (Abe, Ben) and their data can be locked from mutation and cleared separately from real data.
- **Clear demo only**: New `ClearDemoDataOnly` and CLI flag `--clear-demo-only` (plus `make dbcleardemo`) remove only demo owners and their cascaded data, leaving real users' data intact.

### Notes

- **Migrations** (`backend/db/migrations/00006_add_is_demo_to_venue_owners.sql`):
  - **Up**: Add `is_demo boolean NOT NULL DEFAULT false` to `venue_owners`; backfill `is_demo = true` for `abe@demo.org` and `ben@demo.org`.
  - **Down**: Drop column `is_demo`.
- **Schema** (`backend/db/schema.sql`):
  - Added `is_demo` to `venue_owners` table definition for sqlc.
- **sqlc**:
  - Ran `sqlc generate`; `VenueOwner` now has `IsDemo bool`; `GetOwnerByID` / `GetOwnerByEmail` / `CreateOwner` return/scan `is_demo`.
- **Test data** (`backend/internal/testdata/seed.go`):
  - Abe and Ben inserts/upserts set `is_demo = true`; `ON CONFLICT (email) DO UPDATE SET is_demo = true` so reseeding keeps them marked as demo.
  - New `ClearDemoDataOnly(ctx, db)`: `DELETE FROM venue_owners WHERE is_demo = true` (CASCADE removes their venues, event_lists, events, refresh_tokens).
- **API (demo lock)** (`backend/internal/http/demo.go`, handlers):
  - New `IsDemoOwner(ctx, queries, ownerUUIDStr)` helper: loads owner by UUID, returns `owner.IsDemo`.
  - Mutation handlers return **403 Forbidden** when the authenticated owner has `is_demo = true`:
    - **Auth**: `DeleteMe` → "Demo accounts cannot be deleted"
    - **Venues**: `Update`, `Delete` → "Demo data cannot be modified"
    - **Event lists**: `Create`, `Update`, `Delete` → "Demo data cannot be modified"
    - **Events**: `Create`, `Update`, `Delete` → "Demo data cannot be modified"
  - Reads (GET, list, login) unchanged; only mutations are blocked for demo owners.
- **CLI** (`backend/cmd/cli/seed/main.go`):
  - New flag `--clear-demo-only`: runs `ClearDemoDataOnly` then seeds (no full wipe).
  - Existing `-clear`: unchanged (clear ALL data then seed); help text clarified.
- **Makefile**:
  - New target `dbcleardemo`: runs seed with `-clear-demo-only` (from host or devcontainer).
  - Help text updated: `dbseedclear` described as "Clear ALL data (destroys real data)", `dbcleardemo` as "Clear only demo data and reseed (preserves real data)".

### Summary

- **Email verification**: Owners must verify their email before creating or editing venues, event lists, or events. Verification tokens are stored in a separate table; Resend is used for sending verification emails. Unverified owners can log in and read data but receive 403 with code `email_not_verified` on mutations.

### Notes

- **Plan** (`blueprint/backend/2_plan.md`):
  - Under **Security → Email verification**: Documented that we require verification before writes; tokens in `email_verification_tokens`; Resend API key via `RESEND_API_KEY` in `.env`.
- **Migrations** (`backend/db/migrations/00007_email_verification.sql`):
  - **Up**: Add `email_verified_at timestamptz` to `venue_owners`; create `email_verification_tokens` table (token_uuid, owner_uuid, token_hash, expires_at, created_at) with indexes on token_hash, owner_uuid, expires_at.
  - **Down**: Drop `email_verification_tokens`; drop `email_verified_at` from `venue_owners`.
- **Schema** (`backend/db/schema.sql`):
  - Added `email_verified_at` to `venue_owners`; added `email_verification_tokens` table and indexes for sqlc.
- **sqlc**:
  - New `backend/db/queries/email_verification.sql`: CreateEmailVerificationToken, GetEmailVerificationTokenByHash, DeleteEmailVerificationTokensByOwner, DeleteEmailVerificationTokenByHash.
  - **owners.sql**: Added SetOwnerEmailVerified; CreateOwner / GetOwnerByID / GetOwnerByEmail now return `email_verified_at`. Ran `sqlc generate`.
- **Mailer** (`backend/internal/mailer/`):
  - Interface `Sender` with `SendVerificationEmail(to, verificationLink string) error`.
  - **resend.go**: ResendSender using Resend API (POST https://api.resend.com/emails); API key and optional from address from env (`RESEND_API_KEY`, `RESEND_FROM`).
- **Environment** (`backend/.env.example`):
  - `RESEND_API_KEY` (required for sending verification emails), optional `RESEND_FROM`, `VERIFICATION_BASE_URL` (base URL for verification links, default http://localhost:5173).
- **API (auth)** (`backend/internal/http/auth_handlers.go`):
  - Owner response and `ownerToResponse` now include `email_verified` (boolean from `email_verified_at.Valid`).
  - **Register**: After creating owner and refresh token, creates email verification token (24h expiry), builds link from VERIFICATION_BASE_URL, sends verification email via mailer (best-effort; registration does not fail if send fails).
  - **VerifyEmail** (GET `/api/auth/verify-email?token=...`): Validates token by hash, sets `email_verified_at`, deletes token; returns 200 with message.
  - **ResendVerification** (POST `/api/auth/resend-verification`, protected): Deletes existing tokens for owner, creates new token, sends email; returns 202. Returns 400 if already verified.
- **API (errors)** (`backend/internal/http/errors.go`):
  - New `ErrorCodeEmailNotVerified` and `EmailNotVerifiedError(c, message)` returning 403 with code `email_not_verified`.
- **API (verified gate)** (`backend/internal/http/demo.go`, venue/event_list/event handlers):
  - New `IsEmailVerified(ctx, queries, ownerUUIDStr)` helper (same pattern as IsDemoOwner).
  - All mutation handlers (venue Create/Update/Delete, event list Create/Update/Delete, event Create/Update/Delete) now check `IsEmailVerified` after demo check; if not verified return `EmailNotVerifiedError`.
- **Routes** (`backend/internal/http/routes.go`, server.go):
  - Auth handler now takes `mailer.Sender` (nil allowed for tests). Registered GET `/api/auth/verify-email`, POST `/api/auth/resend-verification` (protected). Server creates ResendSender and passes to RegisterRoutes.
- **Integration tests** (`backend/internal/http/integration_test.go`):
  - RegisterRoutes now takes mailer (nil in tests). New helper `setOwnerEmailVerified(t, s, ownerUUIDStr)`; after register in tests that do mutations, call `setOwnerEmailVerified` so create/update/delete succeed.
- **Unit tests** (`backend/internal/http/auth_handlers_test.go`):
  - NewAuthHandler call updated to pass nil mailer.

### Summary

- **Resend verification rate limit**: Added 60-second cooldown per owner on `POST /api/auth/resend-verification` to avoid sending a second email immediately after the first (improves deliverability; second email often lands in spam).

### Notes

- **sqlc** (`backend/db/queries/email_verification.sql`):
  - New `GetLatestVerificationCreatedAtByOwner :one`: returns most recent `created_at` for the owner (used to enforce cooldown before sending resend).
- **API (errors)** (`backend/internal/http/errors.go`):
  - New `ErrorCodeTooManyRequests` and `TooManyRequestsError(c, message)` for 429 Too Many Requests.
- **API (auth)** (`backend/internal/http/auth_handlers.go`):
  - **ResendVerification**: Before deleting tokens and sending, calls `GetLatestVerificationCreatedAtByOwner`; if a token was created within the last 60 seconds, returns 429 with message "Please wait a minute before requesting another verification email." so users cannot trigger back-to-back resends.

## 2026-02-07

### Summary

- **Admin venues list**: Extended `ListAllVenues` to return per-venue counts of public and private events so the backoffice Venues table can show them without extra requests.

### Notes

- **sqlc** (`backend/db/queries/admins.sql`):
  - **ListAllVenues**: Added two scalar subqueries: `public_events_count` (COUNT of events in event_lists with `visibility = 'public'`) and `private_events_count` (COUNT of events in event_lists with `visibility = 'private'`) per venue; both cast to `bigint`.
  - Ran `sqlc generate`; `ListAllVenuesRow` now includes `PublicEventsCount int64` and `PrivateEventsCount int64`.
- **API** (`backend/internal/http/admin_handlers.go`):
  - **ListVenues** (GET `/api/admin/venues`): Response for each venue now includes `public_events_count` and `private_events_count` in addition to existing fields.
