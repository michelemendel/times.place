-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (token_hash, owner_uuid, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPasswordResetTokenByHash :one
SELECT * FROM password_reset_tokens
WHERE token_hash = $1;

-- name: DeletePasswordResetTokenByHash :exec
DELETE FROM password_reset_tokens
WHERE token_hash = $1;

-- name: DeletePasswordResetTokensByOwner :exec
DELETE FROM password_reset_tokens
WHERE owner_uuid = $1;
