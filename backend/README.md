# Backend Development Guide

This guide covers local development setup and workflow for the times.place backend.

## Prerequisites

- **Go 1.25.5+** (installed in devcontainer)
- **PostgreSQL** (provided via devcontainer)
- **goose** (database migrations) - installed in devcontainer
- **sqlc** (SQL code generation) - installed in devcontainer
- **Docker/Rancher Desktop** (for devcontainer)
  - **Rancher Desktop**: Must be in `dockerd (moby)` mode for VS Code/Cursor Dev Containers support
  - **Note**: While Rancher Desktop supports containerd mode with `nerdctl`, VS Code/Cursor's Dev Containers extension requires the Docker API socket, which is only available in dockerd mode

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
- **SERVE_FRONTEND**: Set to `false` for local dev (frontend runs separately)
- **PORT**: Backend API port (default: `8080`)
- **LOG_LEVEL**: Logging verbosity (`debug`, `info`, `warn`, `error`)

See `backend/.env.example` for all available options and their descriptions.

### 2. Dev Container Setup

The project uses a Dev Container for consistent development environments.

#### Starting the Dev Container

**Option A: Using VS Code/Cursor**

1. **Ensure Rancher Desktop is in dockerd mode**: Preferences → Container Engine → Select "dockerd (moby)"
2. Open the project in VS Code/Cursor
3. When prompted, click "Reopen in Container" (or use Command Palette: "Dev Containers: Reopen in Container")
4. The devcontainer will start automatically, including the Postgres service

**Option B: Using Makefile**

```bash
make bdevcontainerup
```

This starts:

- Backend dev container (Go + tools)
- Postgres service container

**Note**: This works with both containerd and dockerd modes. However, if you want to use "Reopen in Container" (Option A), you need dockerd mode.

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

Once the devcontainer and Postgres are running, apply database migrations:

```bash
make bgooseup
```

Check migration status:

```bash
make bgoosestatus
```

Rollback last migration (if needed):

```bash
make bgoosedown
```

Create a new migration:

```bash
make bgoosecreate NAME=your_migration_name
```

This creates a new migration file in `backend/db/migrations/`.

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
cd backend && go run ./cmd/server/main.go
```

The server will:

- Load environment variables from `backend/.env` (if using `godotenv`)
- Start on the port specified in `PORT` (default: `8080`)
- Serve only `/api/*` routes when `SERVE_FRONTEND=false`

## Local Development Workflow

### Typical Workflow

1. **Start devcontainer** (if not already running)

   ```bash
   make bdevcontainerup
   ```

   Or use VS Code/Cursor "Reopen in Container"

2. **Set up environment variables**

   ```bash
   cp backend/.env.example backend/.env
   # Edit backend/.env with your values
   ```

3. **Apply database migrations**

   ```bash
   make bgooseup
   ```

4. **Generate sqlc code** (after writing queries)

   ```bash
   make bsqlcgenerate
   ```

5. **Run the server**

   ```bash
   make brun
   ```

6. **Run frontend separately** (in another terminal, from project root)
   ```bash
   make fdev
   ```

The frontend dev server (on `http://localhost:5173`) will proxy `/api/*` requests to the backend (on `http://localhost:8080`).

### Database Connection

**Connection Details:**

- **Host**: `localhost` (from host) or `postgres` (from inside devcontainer)
- **Port**: `5432`
- **Database**: `timesplace`
- **User**: `timesplace`
- **Password**: `timesplace`

**From inside devcontainer:**

- Use `DATABASE_URL=postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable`
- The `postgres` hostname resolves to the Postgres service container
- Connect with psql: `psql -h postgres -U timesplace -d timesplace`

**From host machine:**

- Use `DATABASE_URL=postgres://timesplace:timesplace@localhost:5432/timesplace?sslmode=disable`
- Port `5432` is forwarded from the container to your host
- Connect with psql: `make bdb-psql` or `psql -h localhost -p 5432 -U timesplace -d timesplace`
- Get connection URL: `make bdb-psql-url`

**Using GUI Tools:**

You can connect using any PostgreSQL client (pgAdmin, DBeaver, TablePlus, etc.) with:

- Host: `localhost`
- Port: `5432`
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
│   └── server/          # Main application entry point
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
- `make bgooseup` - Apply all pending migrations
- `make bgoosedown` - Rollback last migration
- `make bgoosestatus` - Show migration status
- `make bgoosecreate NAME=name` - Create new migration
- `make bsqlcgenerate` - Generate sqlc code
- `make bdb` - Connect to database with psql
- `make bdburl` - Show database connection URL

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

1. **Docker/Rancher Desktop is running**
   - For Rancher Desktop: Make sure it's fully started (wait 10-30 seconds after launching)
   - Verify Docker is accessible: `docker info`
   - **Note**: `make bdevcontainerup` works with both containerd and dockerd modes, but VS Code/Cursor "Reopen in Container" requires dockerd mode
   - **Fix Docker CLI plugin symlinks** (if Rancher Desktop reports incorrect symlinks):
     - Rancher Desktop may show warnings that symlinks point to Docker Desktop instead of Rancher Desktop
     - This happens when Docker Desktop was previously installed
     - **Fix**: Remove the old symlinks and create new ones pointing to Rancher Desktop:
       ```bash
       # Remove old symlinks (may require sudo if protected)
       rm ~/.docker/cli-plugins/docker-buildx ~/.docker/cli-plugins/docker-compose
       # Create correct symlinks to Rancher Desktop
       ln -sf ~/.rd/bin/docker-buildx ~/.docker/cli-plugins/docker-buildx
       ln -sf ~/.rd/bin/docker-compose ~/.docker/cli-plugins/docker-compose
       ```
     - If removal fails due to permissions, you may need to:
       - Quit Docker Desktop completely (if still running)
       - Use `sudo` to remove the symlinks
       - Or let Rancher Desktop fix them automatically (check Rancher Desktop settings)

2. **Container orchestration tool**
   - The Makefile uses `docker compose` by default
   - **Note on containerd vs dockerd mode**:
     - `make bdevcontainerup` works with both containerd (using nerdctl) and dockerd modes
     - However, **VS Code/Cursor's "Reopen in Container" requires dockerd mode** because it needs the Docker API socket, which is only available in dockerd mode
     - If you need containerd for k8s, you can still use `make bdevcontainerup` (you may need to alias `docker` to `nerdctl` or modify the Makefile), but "Reopen in Container" won't work
     - For full Dev Containers support (including "Reopen in Container"), use dockerd mode

3. **Socket permission issues (Rancher Desktop)**
   - If you see "permission denied" errors:
     - Make sure Rancher Desktop is fully initialized (wait a bit longer)
     - Try restarting Rancher Desktop
     - Verify Docker is running: `docker info`

4. **Ports `8080` and `5432` are not in use by other services**
   - Check: `lsof -i :8080` and `lsof -i :5432`

5. **Try rebuilding**: `make bdevcontainerrebuild`

6. **VS Code/Cursor "Reopen in Container"**
   - **Requirement**: Rancher Desktop must be in `dockerd (moby)` mode (not containerd)
   - Configure Rancher Desktop: Preferences → Container Engine → Select "dockerd (moby)"
   - Restart Rancher Desktop after changing the mode
   - **Why dockerd?**: VS Code/Cursor's Dev Containers extension requires the Docker API socket, which is only available in dockerd mode. Containerd mode uses `nerdctl` which doesn't expose the Docker API socket that Dev Containers expects.
   - **Important distinction**: 
     - `make bdevcontainerup` works with both containerd and dockerd modes
     - Only "Reopen in Container" requires dockerd mode
     - If you need containerd for k8s, you can use `make bdevcontainerup` but won't be able to use "Reopen in Container"
   - **Common issue**: If you see warnings about Docker CLI plugin symlinks pointing to Docker Desktop instead of Rancher Desktop, see troubleshooting section #1 above

## Production Environment Variables (Render.com)

When deploying to Render.com, configure the following environment variables in the Render Web Service dashboard:

### Non-Secret Variables

Set these in the Render dashboard (visible values):

- **SERVE_FRONTEND**: `true` (enables Go to serve frontend static assets)
- **LOG_LEVEL**: `info` or `warn` (production logging verbosity)
- **PORT**: Usually set automatically by Render (default: `10000`), but can be overridden
- **COOKIE_DOMAIN**: Domain for refresh token cookies (e.g., `.times.place` for subdomain support, or leave empty for same-origin)
- **COOKIE_SECURE**: `true` (require HTTPS for cookies in production)
- **COOKIE_SAME_SITE**: `lax` (or `strict`, `none` - use `lax` for most cases)

### Secret Variables

Set these in the Render dashboard with the **"Secret" toggle enabled** (values hidden in UI/logs):

- **DATABASE_URL**: PostgreSQL connection string
  - If you link a Render Postgres instance to the Web Service, Render automatically provides this
  - Otherwise, set manually: `postgres://user:password@host:port/database?sslmode=require`
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
| `SERVE_FRONTEND`       | Non-secret | Yes      | `true`  | Enable frontend serving                          |
| `PORT`                 | Non-secret | No       | `10000` | Server port (Render sets automatically)          |
| `LOG_LEVEL`            | Non-secret | No       | `info`  | Logging level (`debug`, `info`, `warn`, `error`) |
| `COOKIE_DOMAIN`        | Non-secret | No       | -       | Cookie domain (empty for same-origin)            |
| `COOKIE_SECURE`        | Non-secret | No       | `true`  | Require HTTPS for cookies                        |
| `COOKIE_SAME_SITE`     | Non-secret | No       | `lax`   | SameSite attribute (`lax`, `strict`, `none`)     |

## Next Steps

After setting up the development environment:

1. Create database schema migrations (see `blueprint/backend/3_tasks.md`)
2. Write SQL queries for sqlc (see `blueprint/backend/3_tasks.md`)
3. Implement API endpoints (see `blueprint/backend/3_tasks.md`)

Refer to `blueprint/backend/` for detailed specifications and implementation plans.
