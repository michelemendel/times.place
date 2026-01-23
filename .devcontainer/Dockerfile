FROM mcr.microsoft.com/devcontainers/go:1.25

# Install postgresql-client for psql and socat for port forwarding
RUN apt-get update && apt-get install -y postgresql-client socat && rm -rf /var/lib/apt/lists/*

# Install goose and sqlc
RUN go install github.com/pressly/goose/v3/cmd/goose@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Ensure tools are in PATH
ENV PATH="${PATH}:/root/go/bin"

WORKDIR /workspace
