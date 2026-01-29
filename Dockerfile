# Production Dockerfile for Render.com: build frontend + Go API, single image.
# Build context: repo root. Use: Dockerfile Path = ./Dockerfile

# -----------------------------------------------------------------------------
# Stage 1: Build frontend (SvelteKit)
# -----------------------------------------------------------------------------
FROM node:20-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN cd frontend && npm ci

COPY frontend/ ./frontend/
RUN cd frontend && npm run build

# -----------------------------------------------------------------------------
# Stage 2: Build Go binary and install goose (for Pre-Deploy migrations)
# -----------------------------------------------------------------------------
FROM golang:1.25-alpine AS go-builder

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /app

COPY backend/ ./backend/
RUN cd backend && go build -o bin/api ./cmd/api

# -----------------------------------------------------------------------------
# Stage 3: Minimal runtime image
# -----------------------------------------------------------------------------
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Frontend static assets (Go server serves these at /)
COPY --from=frontend-builder /app/frontend/build ./frontend/build

# Go binary, migrations, and Pre-Deploy script (Render: Pre-Deploy = ./scripts/render-pre-deploy.sh)
COPY --from=go-builder /app/backend/bin/api ./backend/bin/api
COPY --from=go-builder /app/backend/db/migrations ./backend/db/migrations
COPY --from=go-builder /go/bin/goose /usr/local/bin/goose
COPY scripts/render-pre-deploy.sh ./scripts/render-pre-deploy.sh
RUN chmod +x ./scripts/render-pre-deploy.sh

# Server listens on PORT (Render sets this)
EXPOSE 10000

CMD ["./backend/bin/api"]
