-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (owner_uuid, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEmailVerificationTokenByHash :one
SELECT token_uuid, owner_uuid, expires_at FROM email_verification_tokens
WHERE token_hash = $1
  AND expires_at > now();

-- name: DeleteEmailVerificationTokensByOwner :exec
DELETE FROM email_verification_tokens
WHERE owner_uuid = $1;

-- name: DeleteEmailVerificationTokenByHash :exec
DELETE FROM email_verification_tokens
WHERE token_hash = $1;

-- name: GetLatestVerificationCreatedAtByOwner :one
SELECT created_at FROM email_verification_tokens WHERE owner_uuid = $1 ORDER BY created_at DESC LIMIT 1;
