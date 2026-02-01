-- name: CreateOwner :one
INSERT INTO venue_owners (name, email, mobile, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOwnerByID :one
SELECT * FROM venue_owners
WHERE owner_uuid = $1;

-- name: GetOwnerByEmail :one
SELECT * FROM venue_owners
WHERE LOWER(email) = LOWER($1);

-- name: SetOwnerEmailVerified :exec
UPDATE venue_owners
SET email_verified_at = now(), modified_at = now()
WHERE owner_uuid = $1;

-- name: DeleteOwner :exec
DELETE FROM venue_owners
WHERE owner_uuid = $1;
