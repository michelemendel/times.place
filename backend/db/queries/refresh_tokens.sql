-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    owner_uuid,
    token_hash,
    expires_at,
    user_agent,
    ip_address
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > now();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE refresh_token_uuid = $1;

-- name: RevokeRefreshTokenByHash :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = $1;

-- name: RotateRefreshToken :exec
UPDATE refresh_tokens
SET replaced_by_token_uuid = $2
WHERE refresh_token_uuid = $1;

-- name: RevokeAllTokensForOwner :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE owner_uuid = $1
  AND revoked_at IS NULL;
