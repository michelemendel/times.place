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

- Fixed external database connection issues (pgAdmin, Warp terminal) by using Docker's native port mapping.
- Direct postgres port (5432) works reliably from host, so proxy is not needed.

#### Notes

- **Dev container**:
  - Postgres service exposes port `5432:5432` directly (host:container).
  - External tools (pgAdmin, CLI) connect via `localhost:5432`.
  - Docker's native port mapping works reliably, so no proxy needed.
  - (Previously used socat proxy on port 5433, but removed as unnecessary.)

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