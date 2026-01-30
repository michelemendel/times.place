# Task List (Backend)

## Documentation & alignment

- [x] Split blueprint into `frontend/` and `backend/` sections
- [x] Define backend spec: API + schema + migrations + Go model
- [x] Define backend technical plan: Echo + JWT + goose + sqlc + dev container

## Tooling: local development & deployed/production environments

- [x] Add `.devcontainer/` config for backend development
- [x] Add Postgres service for local dev (devcontainer compose or equivalent)
- [x] Add Makefile targets for backend dev (devcontainer up/down, goose up/down/status, sqlc generate, run server)
- [x] Document local workflow: start containers, run migrations, run API
- [x] Set up local environment variables:
  - [x] Create `backend/.env.example` with all required variables (secrets and non-secrets) and placeholder values
  - [x] Add `backend/.env` to `.gitignore`
  - [x] Document how to load `.env` in Go (e.g. using `godotenv` or similar)
  - [x] Configure devcontainer to use local `.env` or set non-secret vars in `devcontainer.json`
- [ ] Set up production environment variables on Render.com:
  - [ ] Configure non-secret variables in Render Web Service dashboard (LOG_LEVEL, cookie settings)
  - [ ] Configure secret variables in Render with "Secret" toggle enabled (DATABASE_URL, JWT_SECRET, REFRESH_TOKEN_SECRET)
  - [ ] Link Render Postgres instance to Web Service (auto-provides DATABASE_URL) or set manually
  - [x] Document all required production environment variables and their purposes

## Database schema & migrations (goose)

- [x] Create migrations directory `backend/db/migrations/`
- [x] Add initial migration: enable `pgcrypto`
- [x] Add migration: create tables `venue_owners`, `venues`, `event_lists`, `events`, `refresh_tokens`
- [x] Add constraints + indexes (visibility checks, unique tokens, foreign keys)
- [x] Add migration: triggers or explicit update strategy for `modified_at` (choose one)
- [x] Run migrations locally: execute `make dbup` (works from host or inside devcontainer) and verify with `make dbverify` (shows migration status, tables, indexes). Alternatively, use `make devshell` to open an interactive shell in the devcontainer.

## sqlc

- [x] Add `sqlc.yaml` configuration
- [x] Add query files under `backend/db/queries/`:
  - [x] Owners: create, get-by-id, get-by-email
  - [x] Venues: list-by-owner, get-by-id+owner, create, update, delete
  - [x] Event lists: list-by-venue+owner, get-by-id+owner, create, update, delete
  - [x] Events: list-by-event-list+owner, get-by-id+owner, create, update, delete
  - [x] Public browse/search queries
  - [x] Token lookup queries (venue token / event list token)
- [x] Generate sqlc code and ensure builds

## API server (Echo)

- [x] Define project layout under `backend/internal/` (http/service/db wiring)
- [x] Implement JWT auth middleware (extract `owner_uuid` from token)
- [x] Implement request validation + consistent error responses

## Auth endpoints

- [x] `POST /api/auth/register` (hash password; enforce email unique)
- [x] `POST /api/auth/login` (verify password; issue JWT)
- [x] `POST /api/auth/refresh` (validate refresh token; rotate; issue new access JWT)
- [x] `POST /api/auth/logout` (revoke refresh token)
- [x] `GET /api/auth/me`

## Owner-scoped CRUD endpoints

- [x] Venues CRUD (`/api/venues...`)
- [x] Event lists CRUD (`/api/venues/:venue_uuid/event-lists`, `/api/event-lists/:event_list_uuid`)
- [x] Events CRUD (`/api/event-lists/:event_list_uuid/events`, `/api/events/:event_uuid`)
- [x] Cascade deletion behavior matches spec (venue delete removes event lists/events)
- [x] Ordering: support `sort_order` updates for event lists and events

## Public endpoints

- [x] `GET /api/public/venues` (+ optional `?query=` search)
- [x] `GET /api/public/venues/:venue_uuid/event-lists` (public-only)
- [x] `GET /api/public/venues/by-token/:token`
- [x] `GET /api/public/event-lists/by-token/:token`

## Testing

- [x] Add unit tests for auth components (AuthService, middleware, handler helpers)
- [x] Add minimal integration tests for auth + one CRUD path
- [x] Add tests for token-based access control

## Deployment & production (Render.com + GitHub Actions)

- [ ] Add GitHub Actions workflow:
  - [ ] Run backend build + tests on PRs
  - [ ] On merge to `main`, trigger Render Deploy Hook
- [x] Create Render Postgres instance and store connection details securely
- [x] Create Render Web Service for Go backend (Option B: serves `/` + `/api`)
- [x] Configure production environment variables in Render (DB URL, JWT secret, cookie settings)
- [x] Ensure database migrations run on deploy:
  - [x] Choose deployment migration strategy (Render deploy command vs GitHub Actions step)
  - [x] Run `goose up` against production DB as part of deploy
- [ ] Verify production routing:
  - [x] `/` serves frontend `index.html` (SPA fallback)
  - [x] `/api/...` serves API
    **How to check:** Call a public API endpoint (no auth). Expect 200 + JSON, not HTML.
    - **Local:** `curl -s http://localhost:8080/api/public/venues` → JSON array (e.g. `[]`). Or `make bverify-api`.
    - **Production:** `curl -s https://<your-render-url>/api/public/venues` → same. If you get HTML, the `/api` prefix is not routed to the API.
