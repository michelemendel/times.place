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

