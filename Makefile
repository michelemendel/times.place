.PHONY: help dev build preview install-frontend install-backend f-dev f-build f-preview f-install b-build b-run b-install
.PHONY: bdevcontainerup bdevcontainerdown bdevcontainerrebuild
.PHONY: bgooseup bgoosedown bgoosestatus bgoosecreate
.PHONY: bsqlcgenerate brun dbconnect dburl bshell dbhost dbports dbtest dbproxy

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
	@echo "Backend Dev Container:"
	@echo "  make bdevcontainerup      - Start devcontainer (Postgres + backend)"
	@echo "  make bdevcontainerdown    - Stop devcontainer"
	@echo "  make bdevcontainerrebuild  - Rebuild devcontainer"
	@echo ""
	@echo "Backend Database (goose):"
	@echo "  make bgooseup      - Run all pending migrations"
	@echo "  make bgoosedown    - Rollback last migration"
	@echo "  make bgoosestatus  - Show migration status"
	@echo "  make bgoosecreate  - Create new migration (usage: make bgoosecreate NAME=migration_name)"
	@echo ""
	@echo "Backend Code Generation (sqlc):"
	@echo "  make bsqlcgenerate - Generate sqlc code from queries"
	@echo ""
	@echo "Backend Database Access:"
	@echo "  make dbconnect   - Connect to database with psql (inside container)"
	@echo "  make dbhost      - Connect to database from host (for Warp/external terminals)"
	@echo "  make dburl       - Show database connection URL"
	@echo "  make dbports     - Show port mapping for pgAdmin/external tools"
	@echo "  make dbtest      - Test database connection from host"
	@echo "  make dbproxy     - Test connection via proxy port (for pgAdmin)"
	@echo "  make bshell      - Open shell in devcontainer (for Warp/external terminals)"


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
	cd backend && go run ./cmd/server/main.go

binstall:
	cd backend && go mod download


# Backend Dev Container targets
# Uses docker compose (requires dockerd mode in Rancher Desktop or Docker Desktop)

bdevcontainerup:
	@echo "Starting devcontainer..."
	@docker compose -f .devcontainer/docker-compose.yml up -d
	@echo "Waiting for Postgres to be ready..."
	@sleep 3

bdevcontainerdown:
	@echo "Stopping devcontainer..."
	@docker compose -f .devcontainer/docker-compose.yml down

bdevcontainerrebuild:
	@echo "Rebuilding devcontainer..."
	@docker compose -f .devcontainer/docker-compose.yml build --no-cache
	@docker compose -f .devcontainer/docker-compose.yml up -d
	@echo "Waiting for Postgres to be ready..."
	@sleep 3


# Backend Database (goose) targets
# Note: These assume you're running inside the devcontainer or have DATABASE_URL set

bgooseup:
	@cd backend && goose -dir db/migrations postgres "$${DATABASE_URL:-postgres://timesplace:timesplace@localhost:5432/timesplace?sslmode=disable}" up

bgoosedown:
	@cd backend && goose -dir db/migrations postgres "$${DATABASE_URL:-postgres://timesplace:timesplace@localhost:5432/timesplace?sslmode=disable}" down

bgoosestatus:
	@cd backend && goose -dir db/migrations postgres "$${DATABASE_URL:-postgres://timesplace:timesplace@localhost:5432/timesplace?sslmode=disable}" status

bgoosecreate:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make bgoosecreate NAME=migration_name"; \
		exit 1; \
	fi
	@cd backend && goose -dir db/migrations create $(NAME) sql


# Backend Code Generation (sqlc) targets

bsqlcgenerate:
	@cd backend && sqlc generate

# Backend Dev Container shell targets
# Useful for external terminals like Warp

dbshell:
	@echo "Opening shell in devcontainer backend service..."
	@CONTAINER_NAME=$$(docker ps --filter "name=backend" --filter "status=running" --format "{{.Names}}" | head -1); \
	if [ -z "$$CONTAINER_NAME" ]; then \
		echo "Error: Backend container not found. Is the devcontainer running?"; \
		echo "Try: make bdevcontainerup or use Cursor's 'Reopen in Container'"; \
		exit 1; \
	fi; \
	echo "Connecting to container: $$CONTAINER_NAME"; \
	docker exec -it $$CONTAINER_NAME /bin/bash

# Backend Database access targets
# Note: When running inside devcontainer, uses 'postgres' hostname (service name)
# When running outside, uses 'localhost' (port-forwarded)

dbconnect:
	@echo "Connecting to Postgres database..."
	@if [ -n "$$DATABASE_URL" ]; then \
		psql "$$DATABASE_URL"; \
	else \
		echo "Error: DATABASE_URL not set. Using default connection..."; \
		PGPASSWORD=timesplace psql -h localhost -p 5432 -U timesplace -d timesplace; \
	fi

dburl:
	@if [ -n "$$DATABASE_URL" ]; then \
		echo "$$DATABASE_URL"; \
	else \
		echo "postgres://timesplace:timesplace@localhost:5432/timesplace?sslmode=disable"; \
	fi

dbports:
	@echo "Checking Postgres port mapping for external tools (pgAdmin, etc.)..."
	@POSTGRES_CONTAINER=$$(docker ps --filter "name=postgres" --filter "status=running" --format "{{.Names}}" | head -1); \
	if [ -z "$$POSTGRES_CONTAINER" ]; then \
		echo "Error: Postgres container not found or not running."; \
		exit 1; \
	fi; \
	echo ""; \
	echo "Postgres container: $$POSTGRES_CONTAINER"; \
	echo ""; \
	echo "Port mappings:"; \
	docker port $$POSTGRES_CONTAINER 5432 || echo "  No port mapping found (port forwarding may not be configured)"; \
	echo ""; \
	echo "Connection options for pgAdmin:"; \
	echo ""; \
	echo "Option 1: Use port mapping above (if available)"; \
	echo "  Host: localhost (or 127.0.0.1)"; \
	echo "  Port: [use the mapped port from above]"; \
	echo "  Database: timesplace"; \
	echo "  Username: timesplace"; \
	echo "  Password: timesplace"; \
	echo ""; \
	echo "Option 2: Set up port forwarding manually"; \
	echo "  Run: docker port $$POSTGRES_CONTAINER 5432"; \
	echo "  Or forward manually: docker port $$POSTGRES_CONTAINER 5432/tcp"; \
	echo ""; \
	echo "Option 3: Connect via proxy port (RECOMMENDED for pgAdmin)"; \
	echo "  Host: localhost"; \
	echo "  Port: 5433"; \
	echo "  Database: timesplace"; \
	echo "  Username: timesplace"; \
	echo "  Password: timesplace"; \
	echo "  Note: This forwards through the backend container to postgres"; \
	echo "  Test with: make dbproxy"; \
	echo ""; \
	echo "Option 4: Connect via container IP (usually doesn't work from host)"; \
	CONTAINER_IP=$$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $$POSTGRES_CONTAINER); \
	if [ -n "$$CONTAINER_IP" ]; then \
		echo "  Host: $$CONTAINER_IP"; \
		echo "  Port: 5432"; \
		echo "  Database: timesplace"; \
		echo "  Username: timesplace"; \
		echo "  Password: timesplace"; \
	else \
		echo "  Could not determine container IP"; \
	fi

dbtest:
	@echo "Testing database connection from host..."
	@echo ""
	@POSTGRES_CONTAINER=$$(docker ps --filter "name=postgres" --filter "status=running" --format "{{.Names}}" | head -1); \
	if [ -z "$$POSTGRES_CONTAINER" ]; then \
		echo "Error: Postgres container not found or not running."; \
		exit 1; \
	fi; \
	echo "Postgres container: $$POSTGRES_CONTAINER"; \
	echo ""; \
	echo "Checking if postgres is listening inside container..."; \
	docker exec $$POSTGRES_CONTAINER netstat -tlnp 2>/dev/null | grep 5432 || docker exec $$POSTGRES_CONTAINER ss -tlnp 2>/dev/null | grep 5432 || echo "  Could not check listening ports"; \
	echo ""; \
	echo "Checking postgres configuration..."; \
	docker exec $$POSTGRES_CONTAINER psql -U timesplace -d timesplace -c "SHOW listen_addresses;" 2>&1 || echo "  Could not check config"; \
	echo ""; \
	echo "Test 1: Connection via localhost:5432"; \
	PGPASSWORD=timesplace psql -h localhost -p 5432 -U timesplace -d timesplace -c "SELECT version();" 2>&1 && echo "✓ localhost:5432 works!" || echo "✗ localhost:5432 failed"; \
	echo ""; \
	echo "Test 2: Connection via 127.0.0.1:5432"; \
	PGPASSWORD=timesplace psql -h 127.0.0.1 -p 5432 -U timesplace -d timesplace -c "SELECT version();" 2>&1 && echo "✓ 127.0.0.1:5432 works!" || echo "✗ 127.0.0.1:5432 failed"; \
	echo ""; \
	echo "Checking actual port mapping..."; \
	docker port $$POSTGRES_CONTAINER 5432 2>&1 || echo "  No port mapping found"; \
	echo ""; \
	echo "If both tests fail, the issue might be:"; \
	echo "  1. Cursor's port forwarding not working correctly"; \
	echo "  2. PostgreSQL not configured to accept external connections"; \
	echo "  3. Try using 'make dbhost' which connects directly to the container"
	echo "  4. Try using 'make dbproxy' to test the proxy port (5433)"

dbproxy:
	@echo "Testing database connection via proxy port (localhost:5433)..."
	@echo "This port forwards through the backend container to postgres"
	@echo ""
	@PGPASSWORD=timesplace psql -h localhost -p 5433 -U timesplace -d timesplace -c "SELECT version();" 2>&1 && echo "✓ Proxy port 5433 works! Use this for pgAdmin" || echo "✗ Proxy port 5433 failed - make sure devcontainer is rebuilt"

# Backend Dev Container shell and host access targets
# Useful for external terminals like Warp

dbhost:
	@echo "Connecting to Postgres from host..."
	@echo "Use this from Warp or other external terminals"
	@echo "Checking if postgres container is running..."
	@POSTGRES_CONTAINER=$$(docker ps --filter "name=postgres" --filter "status=running" --format "{{.Names}}" | head -1); \
	if [ -z "$$POSTGRES_CONTAINER" ]; then \
		echo "Error: Postgres container not found or not running."; \
		echo "Make sure the devcontainer is running (use 'Reopen in Container' in Cursor or 'make bdevcontainerup')"; \
		exit 1; \
	fi; \
	echo "Found postgres container: $$POSTGRES_CONTAINER"; \
	echo "Connecting directly to postgres container..."; \
	docker exec -it $$POSTGRES_CONTAINER psql -U timesplace -d timesplace
