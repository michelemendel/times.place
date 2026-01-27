FROM mcr.microsoft.com/devcontainers/go:1.25

# Install postgresql-client for psql and socat for port forwarding
RUN apt-get update && apt-get install -y postgresql-client socat && rm -rf /var/lib/apt/lists/*

# Install goose and sqlc
RUN go install github.com/pressly/goose/v3/cmd/goose@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Ensure tools are in PATH
ENV PATH="${PATH}:/root/go/bin"

# Add convenient aliases (system-wide for all users)
RUN echo '' >> /etc/bash.bashrc && \
    echo '# Convenient aliases' >> /etc/bash.bashrc && \
    echo 'alias ll="ls -lah"' >> /etc/bash.bashrc && \
    echo 'alias la="ls -A"' >> /etc/bash.bashrc && \
    echo 'alias l="ls -CF"' >> /etc/bash.bashrc && \
    echo 'alias ..="cd .."' >> /etc/bash.bashrc && \
    echo 'alias ...="cd ../.."' >> /etc/bash.bashrc && \
    echo 'alias ....="cd ../../.."' >> /etc/bash.bashrc && \
    echo 'alias grep="grep --color=auto"' >> /etc/bash.bashrc && \
    echo 'alias fgrep="fgrep --color=auto"' >> /etc/bash.bashrc && \
    echo 'alias egrep="egrep --color=auto"' >> /etc/bash.bashrc && \
    echo 'alias h="history"' >> /etc/bash.bashrc && \
    echo 'alias c="clear"' >> /etc/bash.bashrc && \
    echo 'alias gs="git status"' >> /etc/bash.bashrc && \
    echo 'alias gd="git diff"' >> /etc/bash.bashrc && \
    echo 'alias ga="git add"' >> /etc/bash.bashrc && \
    echo 'alias gc="git commit"' >> /etc/bash.bashrc && \
    echo 'alias gp="git push"' >> /etc/bash.bashrc

WORKDIR /workspace
