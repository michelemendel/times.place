.PHONY: help dev build preview install-frontend install-backend f-dev f-build f-preview f-install finstall-clean b-build b-run b-install b-health pstart
.PHONY: devcontainerup devcontainerdown devcontainerrebuild
.PHONY: dbup dbdown dbreset dbstatus goosecreate dbverify
.PHONY: sqlcgenerate bstart bstop brestart dbconnect dbconnect-renderdotcom devshell

help:
	@echo "Available commands:"
	@echo ""
	@echo "Frontend:"
	@echo "  make fbuild    - Build frontend for production"
	@echo "  make fstart    - Start frontend development server (with HMR)"
	@echo "  make fpreview  - Preview frontend production build"
	@echo "  make finstall  - Install frontend dependencies"
	@echo "  make finstall-clean - Reinstall frontend deps (fixes rollup on host macOS/Windows after devcontainer)"
	@echo ""
	@echo "Backend:"
	@echo "  make bbuild    - Build backend"
	@echo "  make bstart    - Start backend server (serves API; serves built frontend if present)"
	@echo "  make bstop     - Stop backend server"
	@echo "  make brestart  - Restart backend server"
	@echo "  make binstall  - Install backend dependencies"
	@echo "  make bhealth   - Test healthcheck endpoint (requires server running)"
	@echo ""
	@echo "Production Mode (Backend serves frontend):"
	@echo "  make pstart    - Start backend serving frontend (assumes frontend is already built)"
	@echo ""
	@echo "Development Mode (Separate servers):"
	@echo "  make dev       - Start both frontend (Vite) and backend (API only) in separate terminals"
	@echo "  Note: Use 'make fstart' and 'make bstart' in separate terminals for full control"
	@echo ""
	@echo "Backend Dev Container Shell:"
	@echo "  make devshell            - Open shell in devcontainer (run all commands from here)"
	@echo ""
	@echo "Backend Database - works from host or inside devcontainer:"
	@echo "  make dbup         - Run all pending migrations"
	@echo "  make dbdown       - Rollback last migration (one at a time)"
	@echo "  make dbreset      - Rollback ALL migrations (drops all tables)"
	@echo "  make dbstatus     - Show migration status"
	@echo "  make dbverify     - Verify schema: show migration status and list tables"
	@echo "  make goosecreate  - Create new migration (usage: make goosecreate NAME=migration_name)"
	@echo ""
	@echo "Backend Code Generation (sqlc):"
	@echo "  make sqlcgenerate - Generate sqlc code from queries"
	@echo ""
	@echo "Backend Test Data:"
	@echo "  make dbseed       - Seed test data into database"
	@echo "  make dbseedclear  - Clear existing data and seed test data"
	@echo "  make dbseedrc     - Seed Render.com DB from local machine (DATABASE_URL_RENDER_COM)"
	@echo "  make dbseedrcclear - Clear and seed Render.com DB from local machine (DATABASE_URL_RENDER_COM)"
	@echo ""
	@echo "Backend Testing:"
	@echo "  make btest        - Run backend tests"
	@echo "  make btestcover   - Run backend tests with coverage"
	@echo ""
	@echo "Backend Database Access (works from host or inside devcontainer):"
	@echo "  make dbconnect        - Connect to database with psql"
	@echo "  make dbconnect-renderdotcom - Connect to Render.com Postgres (uses DATABASE_URL_RENDER_COM from backend/.env)"
	@echo "  make dbhost           - Connect to database from host (direct connection)"
	@echo "  make dburl         - Show database connection URLs"
	@echo "  make dbports       - Show port mapping info for GUI tools"
	@echo ""
	@echo "Devcontainer Management (optional - Cursor manages this automatically):"
	@echo "  make devcontainerup      - Start devcontainer (Postgres + backend)"
	@echo "  make devcontainerdown    - Stop devcontainer"
	@echo "  make devcontainerrebuild - Rebuild devcontainer"
	@echo ""
	@echo "  Note: These targets are useful when:"
	@echo "    - Using Warp terminal outside the devcontainer"
	@echo "    - Need to rebuild container without restarting Cursor"
	@echo "    - Running CI/CD scripts"
	@echo "    - Prefer command-line control over Cursor's automatic management"


# Frontend targets

fbuild:
	cd frontend && npm run build

fstart:
	cd frontend && npm run dev

fpreview:
	cd frontend && npm run preview

finstall:
	cd frontend && npm install

# Reinstall frontend deps from scratch. Use when building on host (e.g. macOS) after
# node_modules was installed in devcontainer (Linux); npm optional deps are platform-specific.
# Uses npx rimraf so node_modules is removed reliably (avoids "Directory not empty" on macOS).
finstall-clean:
	cd frontend && npx --yes rimraf node_modules && rm -f package-lock.json && npm install


# Backend targets
# When run from host, build inside backend container so the binary matches container arch (avoids Exec format error)

bbuild:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		cd backend && mkdir -p bin && go build -buildvcs=false -o bin/api ./cmd/api; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Run from inside the devcontainer or start it with: make devcontainerup"; \
			exit 1; \
		fi; \
		echo "Building backend inside container..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && mkdir -p bin && go build -buildvcs=false -o bin/api ./cmd/api"; \
	fi

bstart: bbuild
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Starting backend server..."; \
		cd /workspace/backend && ./bin/api; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Starting backend server..."; \
		docker exec -it $$CONTAINER_NAME bash -c "cd /workspace/backend && ./bin/api"; \
	fi

# Production mode: backend serves frontend
pstart: bbuild
	@if [ ! -d "frontend/build" ]; then \
		echo "Error: Frontend build directory not found. Run 'make fbuild' first."; \
		exit 1; \
	fi
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Starting backend server (serving built frontend)..."; \
		cd /workspace/backend && ./bin/api; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Starting backend server (serving built frontend)..."; \
		docker exec -it $$CONTAINER_NAME bash -c "cd /workspace/backend && ./bin/api"; \
	fi

binstall:
	cd backend && go mod download

bstop:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Stopping backend server (inside devcontainer)..."; \
		pkill -f "backend/bin/api" || pkill -f "times.place" || echo "No running server process found."; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			exit 1; \
		fi; \
		echo "Stopping backend server (from host via docker exec)..."; \
		docker exec $$CONTAINER_NAME pkill -f "backend/bin/api" || docker exec $$CONTAINER_NAME pkill -f "times.place" || echo "No running server process found."; \
	fi

brestart:
	@echo "Restarting backend server..."
	@$(MAKE) bstop || true
	@sleep 1
	@$(MAKE) bstart

bhealth:
	@echo "Testing healthcheck endpoint..."
	@if ! command -v curl > /dev/null 2>&1; then \
		echo "Error: curl is not installed. Please install curl to use this command."; \
		exit 1; \
	fi; \
	RESPONSE=$$(curl -s -w "\n%{http_code}" http://localhost:8080/health 2>&1); \
	HTTP_CODE=$$(echo "$$RESPONSE" | tail -n1); \
	BODY=$$(echo "$$RESPONSE" | sed '$$d'); \
	if [ -z "$$HTTP_CODE" ] || [ "$$HTTP_CODE" = "000" ]; then \
		echo ""; \
		echo "Error: Cannot connect to backend server at http://localhost:8080"; \
		echo "  - Is the server running? Try: make bstart"; \
		echo "  - Is the devcontainer running with port forwarding enabled?"; \
		exit 1; \
	elif [ "$$HTTP_CODE" != "200" ]; then \
		echo ""; \
		echo "Error: Server returned HTTP $$HTTP_CODE"; \
		echo "Response: $$BODY"; \
		exit 1; \
	else \
		if command -v jq > /dev/null 2>&1; then \
			echo "$$BODY" | jq .; \
		else \
			echo "$$BODY"; \
		fi; \
	fi


# Backend Dev Container shell
# All database commands should be run from inside the devcontainer

devshell:
	@CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
	if [ -z "$$CONTAINER_NAME" ]; then \
		echo "Error: Backend container not found. Is the devcontainer running?"; \
		echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
		exit 1; \
	fi; \
	echo "Opening shell in devcontainer backend service..."; \
	echo "From here you can run: make dbup, make dbstatus, make dbconnect, etc."; \
	docker exec -it $$CONTAINER_NAME /bin/bash


# Backend Database (goose) targets
# These work both from host (via docker exec) and inside the devcontainer (direct execution)

dbup:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Running migrations (inside devcontainer)..."; \
		cd /workspace/backend && goose -dir db/migrations postgres "$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}" up; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Running migrations (from host via docker exec)..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && goose -dir db/migrations postgres \"$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\" up"; \
	fi

dbdown:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Rolling back last migration (inside devcontainer)..."; \
		cd /workspace/backend && goose -dir db/migrations postgres "$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}" down; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Rolling back last migration (from host via docker exec)..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && goose -dir db/migrations postgres \"$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\" down"; \
	fi

dbreset:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Rolling back ALL migrations (inside devcontainer)..."; \
		cd /workspace/backend && goose -dir db/migrations postgres "$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}" reset; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Rolling back ALL migrations (from host via docker exec)..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && goose -dir db/migrations postgres \"$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\" reset"; \
	fi

dbstatus:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Migration status (inside devcontainer):"; \
		cd /workspace/backend && goose -dir db/migrations postgres "$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}" status; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Migration status (from host via docker exec):"; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && goose -dir db/migrations postgres \"$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\" status"; \
	fi

dbverify:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Verifying database schema (inside devcontainer)..."; \
		echo ""; \
		echo "=== Migration Status ==="; \
		cd /workspace/backend && goose -dir db/migrations postgres "$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}" status; \
		echo ""; \
		echo "=== Tables ==="; \
		psql "$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}" -c '\dt'; \
		echo ""; \
		echo "=== Indexes ==="; \
		psql "$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}" -c 'SELECT schemaname, tablename, indexname FROM pg_indexes WHERE schemaname = '\''public'\'' ORDER BY tablename, indexname;'; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Verifying database schema (from host via docker exec)..."; \
		echo ""; \
		echo "=== Migration Status ==="; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && goose -dir db/migrations postgres \"$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\" status"; \
		echo ""; \
		echo "=== Tables ==="; \
		docker exec $$CONTAINER_NAME bash -c "psql \"$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\" -c '\dt'"; \
		echo ""; \
		echo "=== Indexes ==="; \
		docker exec $$CONTAINER_NAME bash -c "psql \"$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\" -c 'SELECT schemaname, tablename, indexname FROM pg_indexes WHERE schemaname = '\''public'\'' ORDER BY tablename, indexname;'"; \
	fi

goosecreate:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make goosecreate NAME=migration_name"; \
		exit 1; \
	fi
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Creating new migration: $(NAME) (inside devcontainer)"; \
		cd /workspace/backend && goose -dir db/migrations create $(NAME) sql; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Creating new migration: $(NAME) (from host via docker exec)"; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && goose -dir db/migrations create $(NAME) sql"; \
	fi

# Backend Code Generation (sqlc) targets

sqlcgenerate:
	@CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
	if [ -z "$$CONTAINER_NAME" ]; then \
		echo "Error: Backend container not found. Is the devcontainer running?"; \
		echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
		echo "Then run: make devshell"; \
		exit 1; \
	fi; \
	echo "Generating sqlc code..."; \
	docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && sqlc generate"

# Backend Test Data targets
# These work both from host (via docker exec) and inside the devcontainer (direct execution)

dbseed:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Seeding test data (inside devcontainer)..."; \
		cd /workspace/backend && go run ./cmd/cli/seed/main.go; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Seeding test data (from host via docker exec)..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && go run ./cmd/cli/seed/main.go"; \
	fi

dbseedclear:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Clearing and seeding test data (inside devcontainer)..."; \
		cd /workspace/backend && go run ./cmd/cli/seed/main.go -clear; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Clearing and seeding test data (from host via docker exec)..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && go run ./cmd/cli/seed/main.go -clear"; \
	fi

# Seed Render.com database from local machine only; uses DATABASE_URL_RENDER_COM from backend/.env
# Requires Go on the host (no container).
dbseedrc:
	@if [ ! -f backend/.env ]; then \
		echo "Error: backend/.env not found. Add DATABASE_URL_RENDER_COM with your Render external Postgres URL."; \
		exit 1; \
	fi; \
	set -a && . backend/.env && set +a && \
	if [ -z "$$DATABASE_URL_RENDER_COM" ]; then \
		echo "Error: DATABASE_URL_RENDER_COM not set in backend/.env"; \
		exit 1; \
	fi && \
	echo "Seeding Render.com database..."; \
	cd backend && DATABASE_URL="$$DATABASE_URL_RENDER_COM" go run ./cmd/cli/seed/main.go

dbseedrcclear:
	@if [ ! -f backend/.env ]; then \
		echo "Error: backend/.env not found. Add DATABASE_URL_RENDER_COM with your Render external Postgres URL."; \
		exit 1; \
	fi; \
	set -a && . backend/.env && set +a && \
	if [ -z "$$DATABASE_URL_RENDER_COM" ]; then \
		echo "Error: DATABASE_URL_RENDER_COM not set in backend/.env"; \
		exit 1; \
	fi && \
	echo "Clearing and seeding Render.com database..."; \
	cd backend && DATABASE_URL="$$DATABASE_URL_RENDER_COM" go run ./cmd/cli/seed/main.go -clear

# Backend Database access targets
# These work both from host (via docker exec) and inside the devcontainer (direct execution)

dburl:
	@echo "Database connection URL:"
	@echo "  From host: postgres://timesplace:timesplace@localhost:5432/timesplace?sslmode=disable"
	@echo "  From container: postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable"

dbports:
	@echo "Database port mappings:"
	@echo "  Host port 5432 → Postgres container port 5432 (direct connection)"
	@echo "  Host port 5433 → Backend container port 5432 (proxy to postgres)"
	@echo ""
	@echo "Connection details for GUI tools (pgAdmin, DBeaver, etc.):"
	@echo "  Recommended: Use port 5433 (proxy, works reliably with Cursor)"
	@echo "  Host: localhost"
	@echo "  Port: 5433"
	@echo "  Database: timesplace"
	@echo "  Username: timesplace"
	@echo "  Password: timesplace"
	@echo ""
	@echo "  Alternative: Port 5432 (direct, may work from CLI but not pgAdmin)"

dbhost:
	@echo "Connecting to database from host (direct connection via localhost:5432)..."
	@psql "postgres://timesplace:timesplace@localhost:5432/timesplace?sslmode=disable"

dbconnect:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Connecting to Postgres database (inside devcontainer)..."; \
		psql "$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}"; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Connecting to Postgres database (from host via docker exec)..."; \
		docker exec -it $$CONTAINER_NAME bash -c "psql \"$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\""; \
	fi

# Connect to Render.com Postgres. Requires DATABASE_URL_RENDER_COM in backend/.env (external URL from Render dashboard).
dbconnect-renderdotcom:
	@if [ ! -f backend/.env ]; then \
		echo "Error: backend/.env not found. Add DATABASE_URL_RENDER_COM with your Render external Postgres URL."; \
		exit 1; \
	fi; \
	set -a && . backend/.env && set +a && \
	if [ -z "$$DATABASE_URL_RENDER_COM" ]; then \
		echo "Error: DATABASE_URL_RENDER_COM not set in backend/.env"; \
		exit 1; \
	fi && \
	echo "Connecting to Render.com Postgres..." && \
	psql "$$DATABASE_URL_RENDER_COM"

# Backend Testing targets
# These work both from host (via docker exec) and inside the devcontainer (direct execution)
# Tests use a separate test database that is reset before each test run

btest:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Resetting test database..."; \
		TEST_DB_URL="$${TEST_DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace_test?sslmode=disable}"; \
		MAIN_DB_URL="$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}"; \
		psql "$$MAIN_DB_URL" -c "DROP DATABASE IF EXISTS timesplace_test;" || true; \
		psql "$$MAIN_DB_URL" -c "CREATE DATABASE timesplace_test;"; \
		echo "Running migrations on test database..."; \
		cd /workspace/backend && goose -dir db/migrations postgres "$$TEST_DB_URL" up; \
		echo "Running backend tests..."; \
		cd /workspace/backend && go test ./... -v; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Resetting test database..."; \
		docker exec $$CONTAINER_NAME bash -c "TEST_DB_URL=\"\$${TEST_DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace_test?sslmode=disable}\"; MAIN_DB_URL=\"\$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\"; psql \"\$$MAIN_DB_URL\" -c 'DROP DATABASE IF EXISTS timesplace_test;' || true; psql \"\$$MAIN_DB_URL\" -c 'CREATE DATABASE timesplace_test;'"; \
		echo "Running migrations on test database..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && TEST_DB_URL=\"\$${TEST_DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace_test?sslmode=disable}\"; goose -dir db/migrations postgres \"\$$TEST_DB_URL\" up"; \
		echo "Running backend tests..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && go test ./... -v"; \
	fi

btestcover:
	@if [ -f /.dockerenv ] || [ -n "$${DEVCONTAINER}" ]; then \
		echo "Resetting test database..."; \
		TEST_DB_URL="$${TEST_DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace_test?sslmode=disable}"; \
		MAIN_DB_URL="$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}"; \
		psql "$$MAIN_DB_URL" -c "DROP DATABASE IF EXISTS timesplace_test;" || true; \
		psql "$$MAIN_DB_URL" -c "CREATE DATABASE timesplace_test;"; \
		echo "Running migrations on test database..."; \
		cd /workspace/backend && goose -dir db/migrations postgres "$$TEST_DB_URL" up; \
		echo "Running backend tests with coverage..."; \
		cd /workspace/backend && go test ./... -cover; \
	else \
		CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
		if [ -z "$$CONTAINER_NAME" ]; then \
			echo "Error: Backend container not found. Is the devcontainer running?"; \
			echo "Try: make devcontainerup or use Cursor's 'Reopen in Container'"; \
			echo "Or run this command from inside the devcontainer"; \
			exit 1; \
		fi; \
		echo "Resetting test database..."; \
		docker exec $$CONTAINER_NAME bash -c "TEST_DB_URL=\"\$${TEST_DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace_test?sslmode=disable}\"; MAIN_DB_URL=\"\$${DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable}\"; psql \"\$$MAIN_DB_URL\" -c 'DROP DATABASE IF EXISTS timesplace_test;' || true; psql \"\$$MAIN_DB_URL\" -c 'CREATE DATABASE timesplace_test;'"; \
		echo "Running migrations on test database..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && TEST_DB_URL=\"\$${TEST_DATABASE_URL:-postgres://timesplace:timesplace@postgres:5432/timesplace_test?sslmode=disable}\"; goose -dir db/migrations postgres \"\$$TEST_DB_URL\" up"; \
		echo "Running backend tests with coverage..."; \
		docker exec $$CONTAINER_NAME bash -c "cd /workspace/backend && go test ./... -cover"; \
	fi


# Devcontainer Management targets
# These targets are OPTIONAL if you're using Cursor's devcontainer feature.
# Cursor automatically manages the devcontainer lifecycle when you use "Reopen in Container".
#
# These targets are useful when:
#   - Using Warp terminal outside the devcontainer and want manual control
#   - Need to rebuild the container without restarting Cursor
#   - Running CI/CD or scripts that need to start/stop containers
#   - Prefer command-line control over Cursor's automatic management
#
# Uses docker compose (requires dockerd mode in Rancher Desktop or Docker Desktop)

devcontainerup:
	@echo "Starting devcontainer..."
	@docker compose -f .devcontainer/docker-compose.yml up -d
	@echo "Waiting for Postgres to be ready..."
	@sleep 3

devcontainerdown:
	@echo "Stopping devcontainer..."
	@docker compose -f .devcontainer/docker-compose.yml down

devcontainerrebuild:
	@echo "Rebuilding devcontainer..."
	@docker compose -f .devcontainer/docker-compose.yml build --no-cache
	@docker compose -f .devcontainer/docker-compose.yml up -d
	@echo "Waiting for Postgres to be ready..."
	@sleep 3
