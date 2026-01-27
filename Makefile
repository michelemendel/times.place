.PHONY: help dev build preview install-frontend install-backend f-dev f-build f-preview f-install b-build b-run b-install
.PHONY: devcontainerup devcontainerdown devcontainerrebuild
.PHONY: dbgooseup dbgoosedown dbgoosestatus dbgoosecreate dbverify
.PHONY: sqlcgenerate brun dbconnect devshell

help:
	@echo "Available commands:"
	@echo ""
	@echo "Frontend:"
	@echo "  make fbuild    - Build frontend for production"
	@echo "  make fdev      - Start frontend development server"
	@echo "  make fpreview  - Preview frontend production build"
	@echo "  make finstall  - Install frontend dependencies"
	@echo ""
	@echo "Backend:"
	@echo "  make bbuild    - Build backend"
	@echo "  make brun      - Run backend server"
	@echo "  make binstall  - Install backend dependencies"
	@echo ""
	@echo "Backend Dev Container Shell:"
	@echo "  make devshell            - Open shell in devcontainer (run all commands from here)"
	@echo ""
	@echo "Backend Database (goose) - works from host or inside devcontainer:"
	@echo "  make dbgooseup      - Run all pending migrations"
	@echo "  make dbgoosedown    - Rollback last migration (one at a time)"
	@echo "  make dbgoosereset   - Rollback ALL migrations (drops all tables)"
	@echo "  make dbgoosestatus  - Show migration status"
	@echo "  make dbgoosecreate  - Create new migration (usage: make dbgoosecreate NAME=migration_name)"
	@echo "  make dbverify       - Verify schema: show migration status and list tables"
	@echo ""
	@echo "Backend Code Generation (sqlc):"
	@echo "  make sqlcgenerate - Generate sqlc code from queries"
	@echo ""
	@echo "Backend Test Data:"
	@echo "  make dbseed       - Seed test data into database"
	@echo "  make dbseedclear  - Clear existing data and seed test data"
	@echo ""
	@echo "Backend Testing:"
	@echo "  make btest        - Run backend tests"
	@echo "  make btestcover   - Run backend tests with coverage"
	@echo ""
	@echo "Backend Database Access (works from host or inside devcontainer):"
	@echo "  make dbconnect     - Connect to database with psql"
	@echo "  make dbhost        - Connect to database from host (direct connection)"
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

fdev:
	cd frontend && npm run dev

fpreview:
	cd frontend && npm run preview

finstall:
	cd frontend && npm install


# Backend targets

bbuild:
	cd backend && go build ./...

brun:
	cd backend && go run ./cmd/api/main.go

binstall:
	cd backend && go mod download


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
	echo "From here you can run: make dbgooseup, make dbgoosestatus, make dbconnect, etc."; \
	docker exec -it $$CONTAINER_NAME /bin/bash


# Backend Database (goose) targets
# These work both from host (via docker exec) and inside the devcontainer (direct execution)

dbgooseup:
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

dbgoosedown:
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

dbgoosereset:
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

dbgoosestatus:
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

dbgoosecreate:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make dbgoosecreate NAME=migration_name"; \
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
