.PHONY: help dev build preview install-frontend install-backend f-dev f-build f-preview f-install b-build b-run b-install

help:
	@echo "Available commands:"
	@echo ""
	@echo "Frontend:"
	@echo "  make f-dev      - Start frontend development server"
	@echo "  make f-build    - Build frontend for production"
	@echo "  make f-preview  - Preview frontend production build"
	@echo "  make f-install  - Install frontend dependencies"
	@echo ""
	@echo "Backend:"
	@echo "  make b-build    - Build backend"
	@echo "  make b-run       - Run backend server"
	@echo "  make b-install   - Install backend dependencies"
	@echo ""
	@echo "Convenience shortcuts (default to frontend):"
	@echo "  make dev              - Alias for f-dev"
	@echo "  make build            - Alias for f-build"
	@echo "  make preview          - Alias for f-preview"
	@echo "  make install-frontend - Alias for f-install"

# Frontend targets
f-dev:
	cd frontend && npm run dev

f-build:
	cd frontend && npm run build

f-preview:
	cd frontend && npm run preview

f-install:
	cd frontend && npm install

# Backend targets (placeholder for future implementation)
b-build:
	cd backend && go build ./...

b-run:
	cd backend && go run ./cmd/...

b-install:
	cd backend && go mod download

# Convenience shortcuts (default to frontend for now)
dev: f-dev
build: f-build
preview: f-preview
install-frontend: f-install
