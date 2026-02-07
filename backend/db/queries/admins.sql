-- name: ListDetailsAllOwners :many
SELECT 
    o.owner_uuid, 
    o.name, 
    o.email, 
    o.is_admin, 
    o.is_demo, 
    o.venue_limit,
    o.created_at,
    (SELECT COUNT(*) FROM venues v WHERE v.owner_uuid = o.owner_uuid) AS venue_count
FROM venue_owners o
ORDER BY o.created_at DESC;

-- name: GetOwnerDetails :one
SELECT 
    owner_uuid, name, email, is_admin, is_demo, venue_limit, created_at, modified_at, mobile
FROM venue_owners
WHERE owner_uuid = $1;

-- name: ListAllVenues :many
SELECT 
    v.venue_uuid, 
    v.name, 
    v.address, 
    v.owner_uuid,
    o.name AS owner_name,
    o.email AS owner_email,
    (SELECT COUNT(*) FROM events e
     INNER JOIN event_lists el ON e.event_list_uuid = el.event_list_uuid
     WHERE el.venue_uuid = v.venue_uuid AND el.visibility = 'public')::bigint AS public_events_count,
    (SELECT COUNT(*) FROM events e
     INNER JOIN event_lists el ON e.event_list_uuid = el.event_list_uuid
     WHERE el.venue_uuid = v.venue_uuid AND el.visibility = 'private')::bigint AS private_events_count
FROM venues v
JOIN venue_owners o ON v.owner_uuid = o.owner_uuid
ORDER BY v.created_at DESC;

-- name: AdminDeleteOwner :exec
DELETE FROM venue_owners
WHERE owner_uuid = $1;

-- name: SetOwnerAdmin :exec
UPDATE venue_owners
SET is_admin = $2
WHERE owner_uuid = $1;

-- name: UpdateOwnerVenueLimit :exec
UPDATE venue_owners
SET venue_limit = $2, modified_at = now()
WHERE owner_uuid = $1;
