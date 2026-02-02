-- +goose Up
-- Create password_reset_tokens table
CREATE TABLE password_reset_tokens (
    token_hash text PRIMARY KEY,
    owner_uuid uuid NOT NULL REFERENCES venue_owners(owner_uuid) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Index for looking up tokens by owner
CREATE INDEX idx_password_reset_tokens_owner_uuid ON password_reset_tokens(owner_uuid);

-- +goose Down
DROP TABLE IF EXISTS password_reset_tokens;
