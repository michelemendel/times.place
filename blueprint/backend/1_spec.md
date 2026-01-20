# Specification (Backend)

## Purpose

Implement the backend (HTTP API + PostgreSQL) for `times.place` to replace the frontend-only localStorage prototype. The backend must support:

- Venue owner registration and login (JWT).
- Venue owners managing their own venues, event lists, and events.
- Public read access to **public** event lists, plus access to **private** event lists/venues via unguessable tokens.
- A database schema and migrations that match the data the current frontend already uses.

## Definitions

- **Owner**: An authenticated venue owner account.
- **Public**: Any visitor without authentication.
- **Visibility**: `public` or `private` for both venues and event lists.
- **Private Link Token**: An unguessable token used for sharing a private venue or private event list (currently used by the frontend as `/?token=...`).

## Requirements

### Functional Requirements

- **Authentication**:
  - Owners can register with `name`, `email`, `mobile`, `password`.
  - Owners can log in with `email`, `password` and receive a **short-lived JWT access token**.
  - Owners receive a **refresh token** that can be used to obtain new access tokens without re-entering credentials.
  - Owners can log out using a server endpoint that **invalidates the refresh token**; the client then discards the access token.
  - Owners can fetch their own profile using the access token.

- **Authorization**:
  - Owners can only read/modify/delete data they own.
  - Public endpoints must not disclose owner `email` / `mobile` (unless explicitly required later).

- **Core CRUD**:
  - Owners can create/update/delete venues.
  - Owners can create/update/delete event lists under their venues, and set event list ordering.
  - Owners can create/update/delete events under event lists, and set event ordering.

- **Public read**:
  - Public visitors can browse venues that have at least one **public** event list.
  - Public visitors can view **public** event lists and their events.
  - Public visitors can view a **private** venue or event list only by presenting a valid `private_link_token`.

### Data Contract Compatibility (Frontend)

The backend API must use the same field names used in the frontend types (`frontend/src/lib/types.ts`):

- `owner_uuid`, `venue_uuid`, `event_list_uuid`, `event_uuid` (UUID strings)
- Timestamps: `created_at`, `modified_at` (RFC3339 strings)
- Event datetime: `datetime` (RFC3339 string)
- Tokens: `private_link_token` (string; stored as UUID in DB, returned as string)

Notes:

- The frontend currently stores a plaintext `password` field on `VenueOwner` for prototype purposes. The backend **must not** store plaintext passwords. Store a password hash (bcrypt/argon2) and never return it in API responses.

## API Contract (HTTP JSON)

### Conventions

- **Base path**: `/api`
- **Auth**: `Authorization: Bearer <jwt>`
- **Content type**: `application/json`
- **IDs**: UUID strings
- **Time**: RFC3339 in responses; DB stores `timestamptz` for event datetimes.

### Error format

All errors return:

```json
{
  "error": {
    "code": "string",
    "message": "string"
  }
}
```

Example codes: `validation_error`, `unauthorized`, `forbidden`, `not_found`, `conflict`, `internal`.

### Auth endpoints

- `POST /api/auth/register`
  - **Request**:
    - `name` (string)
    - `email` (string, case-insensitive unique)
    - `mobile` (string)
    - `password` (string)
  - **Response**: `201` with `{ owner, access_token }` and sets a refresh token (cookie or response field; see below)

- `POST /api/auth/login`
  - **Request**: `{ email, password }`
  - **Response**: `200` with `{ owner, access_token }` and sets a refresh token (cookie or response field; see below)

- `POST /api/auth/refresh`
  - **Request**: empty JSON body; refresh token is provided via cookie (recommended) or request body.
  - **Response**: `200` with `{ access_token }` (and refresh token rotation if enabled)

- `POST /api/auth/logout`
  - **Request**: empty JSON body; refresh token is provided via cookie (recommended) or request body.
  - **Response**: `204` and invalidates the refresh token server-side (revocation/rotation strategy)

- `GET /api/auth/me`
  - **Auth**: required
  - **Response**: `200` with `{ owner }`

Owner response shape:

- `owner_uuid`, `name`, `email`, `mobile`, `created_at`, `modified_at`

### Owner-scoped endpoints (CRUD)

Venues:

- `GET /api/venues` (auth required) → list venues owned by current owner
- `POST /api/venues` (auth required) → create venue
- `GET /api/venues/:venue_uuid` (auth required) → get one venue (must be owned)
- `PATCH /api/venues/:venue_uuid` (auth required) → update venue fields
- `DELETE /api/venues/:venue_uuid` (auth required) → delete venue and cascade delete event lists/events

Event lists:

- `GET /api/venues/:venue_uuid/event-lists` (auth required) → list event lists for venue (owned)
- `POST /api/venues/:venue_uuid/event-lists` (auth required) → create event list
- `GET /api/event-lists/:event_list_uuid` (auth required) → get one event list (via ownership through venue)
- `PATCH /api/event-lists/:event_list_uuid` (auth required) → update event list fields (including `visibility`, `private_link_token`, `sort_order`, `date`)
- `DELETE /api/event-lists/:event_list_uuid` (auth required) → delete event list and cascade delete events

Events:

- `GET /api/event-lists/:event_list_uuid/events` (auth required) → list events for event list (owned)
- `POST /api/event-lists/:event_list_uuid/events` (auth required) → create event
- `GET /api/events/:event_uuid` (auth required) → get event (owned through event_list → venue)
- `PATCH /api/events/:event_uuid` (auth required) → update event fields (including `datetime`, `sort_order`)
- `DELETE /api/events/:event_uuid` (auth required) → delete event

Payload notes:

- **Venue** fields align to frontend:
  - `name`, `banner_image`, `address`, `geolocation`, `comment`, `timezone`, `visibility`, `private_link_token`
- **EventList** fields align to frontend:
  - `name`, `comment`, `date` (nullable/empty allowed), `visibility`, `private_link_token`, `sort_order`
- **Event** fields align to frontend:
  - `event_name`, `datetime`, `comment`, `duration_minutes`, `sort_order`

### Public endpoints (read-only)

Browse:

- `GET /api/public/venues`
  - Returns venues that have at least one public event list.
  - Does not include owner contact information.

- `GET /api/public/venues/:venue_uuid/event-lists`
  - Returns **public** event lists for the venue, unless a valid token is presented (see below).

Token-based access:

- `GET /api/public/venues/by-token/:token`
  - If `token` matches `venues.private_link_token`, returns venue + all event lists the token grants access to:
    - Venue event lists: **public** lists, plus **private** lists if they belong to that venue (since the venue token is treated as granting access to the venue).

- `GET /api/public/event-lists/by-token/:token`
  - If `token` matches `event_lists.private_link_token`, returns:
    - the parent venue (public fields only)
    - that event list
    - its events

Search (optional initial implementation):

- `GET /api/public/venues?query=...`
  - Initial version may implement basic `ILIKE` matching across a small set of columns (venue name/address/comment, event list name/comment, event name/comment).
  - A later iteration may add full text search.

## Database (PostgreSQL) Schema + Migrations (goose)

### Migration tooling

- Use `goose` for migrations.
- Migrations live under `backend/db/migrations/` (proposed), applied in order.

### Schema overview

We model relationships with foreign keys (no UUID arrays in DB).

#### `venue_owners`

- `owner_uuid uuid primary key`
- `name text not null`
- `mobile text not null`
- `email text not null unique`
- `password_hash text not null`
- `created_at timestamptz not null default now()`
- `modified_at timestamptz not null default now()`

#### `venues`

- `venue_uuid uuid primary key`
- `owner_uuid uuid not null references venue_owners(owner_uuid) on delete cascade`
- `name text not null`
- `banner_image text not null default ''`
- `address text not null default ''`
- `geolocation text not null default ''`  (e.g. `"lat,lng"`)
- `comment text null`
- `timezone text not null default ''` (IANA tz, e.g. `"Asia/Jerusalem"`)
- `visibility text not null` (check: `visibility in ('public','private')`)
- `private_link_token uuid null unique`
- `created_at timestamptz not null default now()`
- `modified_at timestamptz not null default now()`

Indexes:

- `(owner_uuid)`
- `(visibility)`
- `(private_link_token)` unique

#### `event_lists`

- `event_list_uuid uuid primary key`
- `venue_uuid uuid not null references venues(venue_uuid) on delete cascade`
- `name text not null default ''`
- `date date null` (nullable supports “no date”)
- `comment text null`
- `visibility text not null` (check: `visibility in ('public','private')`)
- `private_link_token uuid null unique`
- `sort_order int not null default 0`
- `created_at timestamptz not null default now()`
- `modified_at timestamptz not null default now()`

Indexes:

- `(venue_uuid)`
- `(visibility)`
- `(private_link_token)` unique
- `(venue_uuid, sort_order)`

#### `events`

- `event_uuid uuid primary key`
- `event_list_uuid uuid not null references event_lists(event_list_uuid) on delete cascade`
- `event_name text not null default ''`
- `datetime timestamptz not null`
- `comment text null`
- `duration_minutes int null`
- `sort_order int not null default 0`
- `created_at timestamptz not null default now()`
- `modified_at timestamptz not null default now()`

Indexes:

- `(event_list_uuid)`
- `(event_list_uuid, sort_order)`

#### `refresh_tokens` (for access-token refresh + logout)

Store refresh tokens as **opaque secrets** (not JWTs). Persist only a hash so DB leakage does not leak active refresh tokens.

- `refresh_token_uuid uuid primary key`
- `owner_uuid uuid not null references venue_owners(owner_uuid) on delete cascade`
- `token_hash text not null unique`
- `issued_at timestamptz not null default now()`
- `expires_at timestamptz not null`
- `revoked_at timestamptz null`
- `replaced_by_token_uuid uuid null` (optional; supports rotation chains)
- `user_agent text null` (optional)
- `ip_address text null` (optional)

Indexes:

- `(owner_uuid)`
- `(token_hash)` unique

### UUID generation in DB

We will use Postgres `pgcrypto` (`gen_random_uuid()`) as the default UUID generator inside migrations, while the application continues to use `github.com/google/uuid` in API payloads and domain structs.

## Go Data Model (application) + sqlc mapping

### Design goals

- Domain/application structs use `github.com/google/uuid` and `time.Time`.
- Database boundary uses `pgx` and sqlc-generated code; avoid leaking `pgtype.UUID` into core domain.

### Proposed Go structs (API/domain)

- `VenueOwner` (no password hash in API responses)
- `Venue`
- `EventList`
- `Event`

JSON field names match the frontend (`owner_uuid`, `event_list_uuid`, etc.).

### sqlc

- Use sqlc to generate type-safe query methods.
- We will keep SQL in `backend/db/queries/*.sql` and generate into a package such as `backend/db/sqlc`.
- Queries to include (minimum):
  - Create/find owner by email, get owner by id
  - CRUD venues by owner
  - CRUD event lists by venue (and via owner)
  - CRUD events by event list (and via owner)
  - Public list venues with public event lists
  - Token lookups for venue/event list

