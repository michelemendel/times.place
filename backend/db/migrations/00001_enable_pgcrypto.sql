-- +goose Up
-- Enable pgcrypto extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- +goose Down
-- Disable pgcrypto extension
DROP EXTENSION IF EXISTS "pgcrypto";
