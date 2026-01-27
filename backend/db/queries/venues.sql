-- name: ListVenuesByOwner :many
SELECT * FROM venues
WHERE owner_uuid = $1
ORDER BY created_at DESC;

-- name: GetVenueByIDAndOwner :one
SELECT * FROM venues
WHERE venue_uuid = $1 AND owner_uuid = $2;

-- name: CreateVenue :one
INSERT INTO venues (
    owner_uuid,
    name,
    banner_image,
    address,
    geolocation,
    comment,
    timezone,
    visibility,
    private_link_token
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateVenue :one
UPDATE venues
SET
    name = COALESCE($3, name),
    banner_image = COALESCE($4, banner_image),
    address = COALESCE($5, address),
    geolocation = COALESCE($6, geolocation),
    comment = $7,
    timezone = COALESCE($8, timezone),
    visibility = COALESCE($9, visibility),
    private_link_token = $10
WHERE venue_uuid = $1 AND owner_uuid = $2
RETURNING *;

-- name: DeleteVenue :exec
DELETE FROM venues
WHERE venue_uuid = $1 AND owner_uuid = $2;
