# Task List (Backend)

## Documentation & alignment

- [x] Split blueprint into `frontend/` and `backend/` sections
- [x] Define backend spec: API + schema + migrations + Go model
- [x] Define backend technical plan: Echo + JWT + goose + sqlc + dev container

## Tooling: local development & deployed/production environments

- [ ] Add `.devcontainer/` config for backend development
- [ ] Add Postgres service for local dev (devcontainer compose or equivalent)
- [ ] Add Makefile targets for backend dev (devcontainer up/down, goose up/down/status, sqlc generate, run server)
- [ ] Document local workflow: start containers, run migrations, run API
- [ ] Set up local environment variables:
  - [ ] Create `backend/.env.example` with all required variables (secrets and non-secrets) and placeholder values
  - [ ] Add `backend/.env` to `.gitignore`
  - [ ] Document how to load `.env` in Go (e.g. using `godotenv` or similar)
  - [ ] Configure devcontainer to use local `.env` or set non-secret vars in `devcontainer.json`
- [ ] Set up production environment variables on Render.com:
  - [ ] Configure non-secret variables in Render Web Service dashboard (SERVE_FRONTEND, LOG_LEVEL, cookie settings)
  - [ ] Configure secret variables in Render with "Secret" toggle enabled (DATABASE_URL, JWT_SECRET, REFRESH_TOKEN_SECRET)
  - [ ] Link Render Postgres instance to Web Service (auto-provides DATABASE_URL) or set manually
  - [ ] Document all required production environment variables and their purposes

## Database schema & migrations (goose)

- [ ] Create migrations directory `backend/db/migrations/`
- [ ] Add initial migration: enable `pgcrypto`
- [ ] Add migration: create tables `venue_owners`, `venues`, `event_lists`, `events`, `refresh_tokens`
- [ ] Add constraints + indexes (visibility checks, unique tokens, foreign keys)
- [ ] Add migration: triggers or explicit update strategy for `modified_at` (choose one)

## sqlc

- [ ] Add `sqlc.yaml` configuration
- [ ] Add query files under `backend/db/queries/`:
  - [ ] Owners: create, get-by-id, get-by-email
  - [ ] Venues: list-by-owner, get-by-id+owner, create, update, delete
  - [ ] Event lists: list-by-venue+owner, get-by-id+owner, create, update, delete
  - [ ] Events: list-by-event-list+owner, get-by-id+owner, create, update, delete
  - [ ] Public browse/search queries
  - [ ] Token lookup queries (venue token / event list token)
- [ ] Generate sqlc code and ensure builds

## API server (Echo)

- [ ] Define project layout under `backend/internal/` (http/service/db wiring)
- [ ] Implement JWT auth middleware (extract `owner_uuid` from token)
- [ ] Implement request validation + consistent error responses

## Auth endpoints

- [ ] `POST /api/auth/register` (hash password; enforce email unique)
- [ ] `POST /api/auth/login` (verify password; issue JWT)
- [ ] `POST /api/auth/refresh` (validate refresh token; rotate; issue new access JWT)
- [ ] `POST /api/auth/logout` (revoke refresh token)
- [ ] `GET /api/auth/me`

## Owner-scoped CRUD endpoints

- [ ] Venues CRUD (`/api/venues...`)
- [ ] Event lists CRUD (`/api/venues/:venue_uuid/event-lists`, `/api/event-lists/:event_list_uuid`)
- [ ] Events CRUD (`/api/event-lists/:event_list_uuid/events`, `/api/events/:event_uuid`)
- [ ] Cascade deletion behavior matches spec (venue delete removes event lists/events)
- [ ] Ordering: support `sort_order` updates for event lists and events

## Public endpoints

- [ ] `GET /api/public/venues` (+ optional `?query=` search)
- [ ] `GET /api/public/venues/:venue_uuid/event-lists` (public-only)
- [ ] `GET /api/public/venues/by-token/:token`
- [ ] `GET /api/public/event-lists/by-token/:token`

## Testing

- [ ] Add minimal integration tests for auth + one CRUD path
- [ ] Add tests for token-based access control

## Deployment & production (Render.com + GitHub Actions)

- [ ] Create Render Postgres instance and store connection details securely
- [ ] Create Render Web Service for Go backend (Option B: serves `/` + `/api`)
- [ ] Configure production environment variables in Render (DB URL, JWT secret, cookie settings)
- [ ] Add GitHub Actions workflow:
  - [ ] Run backend build + tests on PRs
  - [ ] On merge to `main`, trigger Render Deploy Hook
- [ ] Ensure database migrations run on deploy:
  - [ ] Choose deployment migration strategy (Render deploy command vs GitHub Actions step)
  - [ ] Run `goose up` against production DB as part of deploy
- [ ] Verify production routing:
  - [ ] `/` serves frontend `index.html` (SPA fallback)
  - [ ] `/api/...` serves API
