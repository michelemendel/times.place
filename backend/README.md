# Backend Development Guide

This guide covers local development setup and workflow for the times.place backend.

## Prerequisites

- **Go 1.25.5+** (installed in devcontainer)
- **PostgreSQL** (provided via devcontainer)
- **goose** (database migrations) - installed in devcontainer
- **sqlc** (SQL code generation) - installed in devcontainer
- **Docker or Docker-compatible runtime** (for devcontainer)
  - Must expose the Docker API for VS Code/Cursor Dev Containers (e.g. dockerd mode)
  - **Note**: Some runtimes support containerd/nerdctl; VS Code/Cursor's Dev Containers extension requires the Docker API socket (typically dockerd mode)

## Local Development Setup

### 1. Environment Variables

**Important**: The `.env.example` file is a template only and is not automatically used. You must create your own `.env` file.

Copy the example environment file and configure it:

```bash
cp backend/.env.example backend/.env
```

**Note**: The `backend/.env` file is gitignored and will not be committed. Each developer creates their own local copy.

Edit `backend/.env` and set your values:

- **DATABASE_URL**: PostgreSQL connection string
- **JWT_SECRET**: Strong random secret for JWT signing (generate with `openssl rand -base64 32`)
- **REFRESH_TOKEN_SECRET**: Strong random secret for refresh tokens (generate with `openssl rand -base64 32`)
- **RESEND_API_KEY**: API key from [Resend](https://resend.com) for sending verification emails (required for email verification; leave empty to skip sending)
- **PORT**: Backend API port (default: `8080`)
- **LOG_LEVEL**: Logging verbosity (`debug`, `info`, `warn`, `error`)

See `backend/.env.example` for all available options and their descriptions.

### 2. Dev Container Setup

The project uses a Dev Container for consistent development environments.

#### Starting the Dev Container

**Recommended: Using VS Code/Cursor**

1. **Ensure your container runtime exposes the Docker API** (e.g. dockerd mode; see your runtime’s preferences)
2. Open the project in VS Code/Cursor
3. When prompted, click "Reopen in Container" (or use Command Palette: "Dev Containers: Reopen in Container")
4. The devcontainer will start automatically, including the Postgres service

**Note**: If you're using Cursor, you don't need to run `make bdevcontainerup` - the devcontainer starts automatically when you "Reopen in Container".

**Alternative: Using Makefile (for external terminals/scripts)**

```bash
make bdevcontainerup
```

This starts:

- Backend dev container (Go + tools)
- Postgres service container

**Note**: This works with both containerd and dockerd modes. However, if you want to use "Reopen in Container" (Option A), you need dockerd mode. Use this option if you're working from an external terminal or need to start containers via scripts.

#### Stopping the Dev Container

```bash
make bdevcontainerdown
```

#### Rebuilding the Dev Container

If you need to rebuild (e.g., after changing Dockerfile):

```bash
make bdevcontainerrebuild
```

### 3. Database Migrations

**Important**: All database commands run inside the devcontainer (via `docker exec`). You can run these Makefile targets directly from your host terminal - they automatically execute inside the container.

Apply database migrations:

```bash
make dbup
```

Check migration status:

```bash
make dbstatus
```

Verify schema (shows migration status, tables, and indexes):

```bash
make dbverify
```

Rollback last migration (if needed):

```bash
make dbdown
```

Create a new migration:

```bash
make goosecreate NAME=your_migration_name
```

This creates a new migration file in `backend/db/migrations/`.

Connect to database with psql:

```bash
make dbconnect
```

**Optional: Interactive shell**

If you want to run multiple commands interactively or explore the container:

```bash
make devshell
```

Once inside the shell, you can run commands directly (e.g., `goose status`, `psql`, etc.) without the Makefile wrappers.

### 4. Code Generation (sqlc)

After writing or updating SQL queries in `backend/db/queries/`, generate Go code:

```bash
make bsqlcgenerate
```

This reads SQL from `backend/db/queries/` and generates type-safe Go code in `backend/db/sqlc/`.

### 5. Running the Server

Start the backend API server:

```bash
make brun
```

Or manually:

```bash
cd backend && go run ./cmd/api/main.go
```

The server will:

- Load environment variables from `backend/.env` (if using `godotenv`)
- Start on the port specified in `PORT` (default: `8080`)
- Serve `/api/*` routes, and (if `frontend/build/` exists) also serve the frontend at `/` with SPA fallback routing

## Local Development Workflow

### Typical Workflow

1. **Start devcontainer** (if not already running)
   - **If using Cursor**: Click "Reopen in Container" when prompted (no need to run `make bdevcontainerup`)
   - **If using external terminal/scripts**: Run `make bdevcontainerup`

2. **Set up environment variables**

   ```bash
   cp backend/.env.example backend/.env
   # Edit backend/.env with your values
   ```

3. **Apply database migrations**

   ```bash
   make bgooseup
   ```

4. **Seed test data** (optional, for development)

   ```bash
   make dbseed
   ```

   This seeds test data matching the frontend demo data. Test credentials:
   - Owner 1: `abe@demo.org` / `demo`
   - Owner 2: `ben@demo.org` / `demo`

5. **Generate sqlc code** (after writing queries)

   ```bash
   make bsqlcgenerate
   ```

6. **Run the server**

   You can run the application in two modes:

   **Development Mode (Recommended for local development):**
   - Frontend runs on Vite dev server with Hot Module Replacement (HMR)
   - Backend runs separately serving only API routes
   - Best for active frontend development

   ```bash
   # Terminal 1: Start backend (API only)
   make bstart

   # Terminal 2: Start frontend dev server
   make fstart
   ```

   The frontend dev server (on `http://localhost:5173`) will proxy `/api/*` requests to the backend (on `http://localhost:8080`).

   **Production Mode (Backend serves frontend):**
   - Frontend is built and served by the backend
   - Single server on port 8080
   - Matches production deployment setup
   - Useful for testing production-like behavior locally

   ```bash
   # Build frontend, then start backend serving it
   make fbuild
   make pstart
   ```

   The application will be available at `http://localhost:8080` with both API (`/api/*`) and frontend routes served by the backend.

   **Note:** `make pstart` expects the frontend to already be built (run `make fbuild` first) and then starts the backend, which will serve the built frontend from `frontend/build/`.

### Database Connection

**Connection Details:**

- **Host**: `localhost` (from host) or `postgres` (from inside devcontainer)
- **Port**: `5434` (host; postgres container uses 5432 internally)
- **Database**: `timesplace`
- **User**: `timesplace`
- **Password**: `timesplace`

**From inside devcontainer:**

- Use `DATABASE_URL=postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable`
- The `postgres` hostname resolves to the Postgres service container
- Connect with psql: `psql -h postgres -U timesplace -d timesplace`

**From host machine:**

- Use `make dbconnect` to connect via docker exec (works from any terminal)
- Or connect directly: `psql "postgres://timesplace:timesplace@localhost:5434/timesplace?sslmode=disable"`

**Using GUI Tools (pgAdmin, DBeaver, TablePlus, etc.):**

**Recommended: Use proxy port (works reliably with Cursor port forwarding):**

- Host: `localhost`
- Port: `5433` (proxy port through backend container)
- Database: `timesplace`
- Username: `timesplace`
- Password: `timesplace`

**Alternative: Direct port (may work from host CLI, but pgAdmin may have issues):**

- Host: `localhost`
- Port: `5434` (direct postgres port on host)
- Database: `timesplace`
- Username: `timesplace`
- Password: `timesplace`

### Loading Environment Variables in Go

The application should load `.env` files using a library like `github.com/joho/godotenv`:

```go
import "github.com/joho/godotenv"

func init() {
    // Load .env file from backend directory
    // This will NOT automatically fall back to .env.example
    // Developers must create .env from .env.example
    if err := godotenv.Load("backend/.env"); err != nil {
        // Handle error - .env file is required for local development
        log.Fatal("Error loading .env file: ", err)
    }
}
```

Then access variables via `os.Getenv()`:

```go
databaseURL := os.Getenv("DATABASE_URL")
jwtSecret := os.Getenv("JWT_SECRET")
```

**Important**: The `.env` file must exist - the application will not automatically use `.env.example`. Always create `backend/.env` from `backend/.env.example` before running the application.

## Project Structure

```
backend/
├── cmd/
│   ├── api/             # API server entry point
│   └── cli/             # CLI commands (e.g., seed)
├── db/
│   ├── migrations/      # Goose migration files
│   ├── queries/         # SQL query files for sqlc
│   └── sqlc/            # Generated sqlc code (do not edit manually)
├── internal/
│   ├── http/            # Echo routes, handlers, middleware
│   ├── service/         # Business logic layer
│   └── store/           # Data access layer (wraps sqlc)
├── utils/               # Utility functions
├── .env                 # Local environment variables (gitignored)
├── .env.example         # Example environment variables
└── go.mod               # Go module definition
```

## Makefile Targets

All backend-related Makefile targets:

- `make bbuild` - Build backend
- `make brun` - Run backend server
- `make binstall` - Install Go dependencies
- `make bdevcontainerup` - Start devcontainer
- `make bdevcontainerdown` - Stop devcontainer
- `make bdevcontainerrebuild` - Rebuild devcontainer
- `make dbup` - Apply all pending migrations
- `make dbdown` - Rollback last migration
- `make dbreset` - Rollback ALL migrations (drops all tables)
- `make dbstatus` - Show migration status
- `make goosecreate NAME=name` - Create new migration
- `make dbverify` - Verify schema: show migration status and list tables
- `make bsqlcgenerate` - Generate sqlc code
- `make dbseed` - Seed test data into database
- `make dbseedclear` - Clear demo data only (does not seed; run `make dbseed` to re-seed)
- `make devshell` - Open shell in devcontainer (for Warp/external terminals)
- `make dbconnect` - Connect to database with psql (works from host or inside container)
- `make dbhost` - Connect to database from host (direct connection via localhost:5434)
- `make dburl` - Show database connection URLs
- `make dbports` - Show port mapping info for GUI tools (pgAdmin, etc.)

## Troubleshooting

### Postgres not ready

If migrations fail with connection errors, wait a few seconds for Postgres to start:

```bash
# Check if Postgres is healthy
docker compose -f .devcontainer/docker-compose.yml ps
```

### Port already in use

If port `8080` is already in use, change `PORT` in `backend/.env`:

```bash
PORT=8081
```

### Environment variables not loading

Ensure:

1. `backend/.env` exists (copy from `.env.example`)
2. Your Go code uses `godotenv.Load()` or similar
3. You're running from the `backend/` directory or setting the path correctly

### Devcontainer not starting

Check:

1. **Docker (or your container runtime) is running**
   - Make sure it's fully started (some runtimes need 10–30 seconds after launching)
   - Verify Docker is accessible: `docker info`
   - **Note**: `make bdevcontainerup` works with both containerd and dockerd modes, but VS Code/Cursor "Reopen in Container" requires the Docker API (e.g. dockerd mode)
   - **Fix Docker CLI plugin symlinks** (if your runtime reports incorrect symlinks, e.g. pointing to a different installation):
     - Remove the old symlinks and create new ones pointing to your current Docker runtime (paths depend on your installation; check your runtime’s docs).
     - If removal fails due to permissions, quit any other Docker/container app, use `sudo` if needed, or use your runtime’s built-in option to fix symlinks.

2. **Container orchestration tool**
   - The Makefile uses `docker compose` by default
   - **Note on containerd vs dockerd mode**:
     - `make bdevcontainerup` works with both containerd (using nerdctl) and dockerd modes
     - However, **VS Code/Cursor's "Reopen in Container" requires dockerd mode** because it needs the Docker API socket, which is only available in dockerd mode
     - If you need containerd for k8s, you can still use `make bdevcontainerup` (you may need to alias `docker` to `nerdctl` or modify the Makefile), but "Reopen in Container" won't work
     - For full Dev Containers support (including "Reopen in Container"), use dockerd mode

3. **Socket permission issues**
   - If you see "permission denied" errors:
     - Make sure your container runtime is fully initialized (wait a bit longer)
     - Try restarting the runtime
     - Verify Docker is running: `docker info`

4. **Ports `8080`, `5433`, and `5434` are not in use by other services**
   - Check: `lsof -i :8080`, `lsof -i :5433`, `lsof -i :5434`

5. **Try rebuilding**: `make bdevcontainerrebuild`

6. **VS Code/Cursor "Reopen in Container"**
   - **Requirement**: Your runtime must expose the Docker API (e.g. `dockerd (moby)` mode, not containerd-only).
   - Configure your runtime: enable dockerd / Docker API (exact steps depend on the runtime).
   - Restart the runtime after changing the mode.
   - **Why?** VS Code/Cursor's Dev Containers extension needs the Docker API socket; containerd-only/nerdctl setups often don’t expose it.
   - **Important distinction**:
     - `make bdevcontainerup` works with both containerd and dockerd modes
     - Only "Reopen in Container" requires the Docker API (dockerd mode)
     - If you need containerd for k8s, you can use `make bdevcontainerup` but won't be able to use "Reopen in Container"
   - **Common issue**: If you see warnings about Docker CLI plugin symlinks pointing to the wrong installation, see troubleshooting section #1 above

## Backoffice Administration

The platform includes a Backoffice interface for manual administration of owners and venues, accessible at `/backoffice` for admin users.

### Promoting a User to Admin

By design, there is no public sign-up for admin accounts. To grant admin privileges, you must manually update the `is_admin` flag in the database for an existing user.

1.  **Register** a standard account via the frontend.
2.  **Run the SQL command** to promote the user:

    ```bash
    # If running with devcontainer:
    make devshell
    psql "$DATABASE_URL" -c "UPDATE venue_owners SET is_admin = true WHERE email = 'YOUR_EMAIL@example.com';"
    ```

    Or using a local `psql` connection:

    ```sql
    UPDATE venue_owners SET is_admin = true WHERE email = 'YOUR_EMAIL@example.com';
    ```

3.  **Access the Backoffice**: Log out and log back in (or refresh). A **Backoffice** link will appear in the user menu.

## CI/CD Pipeline - GitHub Actions & Render Deploy Hook

The CI workflow (`.github/workflows/ci.yml`) runs build and tests on every push and PR. On **push to `main`**, it also triggers a Render deploy via a deploy hook.

### 1. GitHub repository settings

1. Go to **https://github.com/michelemendel/times.place/settings/actions** (or: repo → **Settings** → **Actions** → **General**).
2. Under **Actions permissions**, ensure **"Allow all actions and reusable workflows"** (or at least allow the actions used in `ci.yml`).
3. Under **Workflow permissions**, choose **"Read and write permissions"** if you need artifacts; **"Read repository contents"** is enough for this workflow.
4. Save if you changed anything.

### 2. Add the Render deploy hook secret

1. In Render: open your **Web Service** → **Settings** → **Deploy Hook**.
2. Copy the deploy hook URL (e.g. `https://api.render.com/deploy/srv-xxxxx?key=yyyy`).
3. In GitHub: repo → **Settings** → **Secrets and variables** → **Actions**.
4. Click **New repository secret**.
5. Name: `RENDER_DEPLOY_HOOK_URL`, Value: paste the full deploy hook URL.
6. Save.

After that, every push to `main` that passes CI will run the **Deploy to Render** job and trigger a new deploy. The workflow does **not** trigger the deploy on pull requests, only on pushes to `main`.

### Pre-Deploy (database migrations)

In Render, set **Pre-Deploy Command** so migrations run before each deploy:

- **Pre-Deploy Command:** `./scripts/render-pre-deploy.sh`

The script runs `goose up` against `DATABASE_URL`. It also strips surrounding double or single quotes from `DATABASE_URL` so migrations work even when the URL is pasted with quotes (which would otherwise cause goose to fail with "failed to parse as keyword/value"). If you see that error, either use this script as Pre-Deploy or ensure `DATABASE_URL` in Render has **no** leading/trailing quote characters.

## Production Environment Variables (Render.com)

When deploying to Render.com, configure the following environment variables in the Render Web Service dashboard:

### Non-Secret Variables

Set these in the Render dashboard (visible values):

- **LOG_LEVEL**: `info` or `warn` (production logging verbosity)
- **PORT**: Usually set automatically by Render (default: `10000`), but can be overridden
- **COOKIE_DOMAIN**: Domain for refresh token cookies (e.g., `.times.place` for subdomain support, or leave empty for same-origin)
- **COOKIE_SECURE**: `true` (require HTTPS for cookies in production)
- **COOKIE_SAME_SITE**: `lax` (or `strict`, `none` - use `lax` for most cases)

### Secret Variables

Set these in the Render dashboard with the **"Secret" toggle enabled** (values hidden in UI/logs):

- **DATABASE_URL**: PostgreSQL connection string (no surrounding quotes)
  - If you link a Render Postgres instance to the Web Service, Render automatically provides this
  - Otherwise, set manually: `postgres://user:password@host:port/database?sslmode=require` or `postgresql://...`
  - Do not include literal double or single quotes around the value; the Pre-Deploy script strips them if present
- **JWT_SECRET**: Strong random secret for signing JWT access tokens
  - Generate with: `openssl rand -base64 32`
  - Must be a strong, random string
- **REFRESH_TOKEN_SECRET**: Strong random secret for generating refresh tokens
  - Generate with: `openssl rand -base64 32`
  - Can be the same as `JWT_SECRET` or different

### Setting Up in Render

1. Go to your Render Web Service dashboard
2. Navigate to **Environment** section
3. Add each variable:
   - Click **"Add Environment Variable"**
   - Enter the variable name and value
   - Toggle **"Secret"** for sensitive values (JWT_SECRET, REFRESH_TOKEN_SECRET, DATABASE_URL)
4. If using Render Postgres:
   - Link the Postgres instance to your Web Service (Render will auto-provide `DATABASE_URL`)
   - Or manually set `DATABASE_URL` as a secret variable

### Variable Reference

| Variable               | Type       | Required | Default | Description                                      |
| ---------------------- | ---------- | -------- | ------- | ------------------------------------------------ |
| `DATABASE_URL`         | Secret     | Yes      | -       | PostgreSQL connection string                     |
| `JWT_SECRET`           | Secret     | Yes      | -       | Secret for JWT signing                           |
| `REFRESH_TOKEN_SECRET` | Secret     | Yes      | -       | Secret for refresh tokens                        |
| `PORT`                 | Non-secret | No       | `10000` | Server port (Render sets automatically)          |
| `LOG_LEVEL`            | Non-secret | No       | `info`  | Logging level (`debug`, `info`, `warn`, `error`) |
| `COOKIE_DOMAIN`        | Non-secret | No       | -       | Cookie domain (empty for same-origin)            |
| `COOKIE_SECURE`        | Non-secret | No       | `true`  | Require HTTPS for cookies                        |
| `COOKIE_SAME_SITE`     | Non-secret | No       | `lax`   | SameSite attribute (`lax`, `strict`, `none`)     |

## Testing

For detailed information about testing, test data management, and database testing strategies, see [TESTING_STRATEGY.md](./TESTING_STRATEGY.md).

**Quick Start:**

```bash
# Run all tests
cd backend && go test ./...

# Run tests with coverage
go test ./... -cover
```

**Key Points:**

- Tests run against a live database (not mocks)
- Test isolation via database transactions (automatic rollback)
- Test data seeding utilities in `backend/internal/testdata/`
- See `TESTING_STRATEGY.md` for full documentation

## Next Steps

After setting up the development environment:

1. Create database schema migrations (see `blueprint/backend/3_tasks.md`)
2. Write SQL queries for sqlc (see `blueprint/backend/3_tasks.md`)
3. Implement API endpoints (see `blueprint/backend/3_tasks.md`)
4. Write tests using the test helpers (see `TESTING_STRATEGY.md`)

Refer to `blueprint/backend/` for detailed specifications and implementation plans.
