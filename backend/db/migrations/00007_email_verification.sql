-- +goose Up
ALTER TABLE venue_owners
ADD COLUMN email_verified_at timestamptz;

CREATE TABLE email_verification_tokens (
    token_uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_uuid uuid NOT NULL REFERENCES venue_owners(owner_uuid) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX email_verification_tokens_token_hash_idx ON email_verification_tokens(token_hash);
CREATE INDEX email_verification_tokens_owner_uuid_idx ON email_verification_tokens(owner_uuid);
CREATE INDEX email_verification_tokens_expires_at_idx ON email_verification_tokens(expires_at);

-- +goose Down
DROP TABLE IF EXISTS email_verification_tokens;
ALTER TABLE venue_owners DROP COLUMN IF EXISTS email_verified_at;
