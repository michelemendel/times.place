.PHONY: help dev build preview install-frontend install-backend f-dev f-build f-preview f-install b-build b-run b-install

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

# Frontend targets
fbuild:
	cd frontend && npm run build

fdev:
	cd frontend && npm run dev

fpreview:
	cd frontend && npm run preview

finstall:
	cd frontend && npm install

# Backend targets (placeholder for future implementation)
bbuild:
	cd backend && go build ./...

brun:
	cd backend && go run ./cmd/...

binstall:
	cd backend && go mod download
