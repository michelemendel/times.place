-- name: ListEventListsByVenueAndOwner :many
SELECT el.* FROM event_lists el
INNER JOIN venues v ON el.venue_uuid = v.venue_uuid
WHERE el.venue_uuid = $1 AND v.owner_uuid = $2
ORDER BY el.sort_order ASC, el.created_at ASC;

-- name: GetEventListByIDAndOwner :one
SELECT el.* FROM event_lists el
INNER JOIN venues v ON el.venue_uuid = v.venue_uuid
WHERE el.event_list_uuid = $1 AND v.owner_uuid = $2;

-- name: CreateEventList :one
INSERT INTO event_lists (
    venue_uuid,
    name,
    date,
    comment,
    visibility,
    private_link_token,
    sort_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateEventList :one
UPDATE event_lists
SET
    name = COALESCE($3, name),
    date = $4,
    comment = $5,
    visibility = COALESCE($6, visibility),
    private_link_token = $7,
    sort_order = COALESCE($8, sort_order)
WHERE event_list_uuid = $1
  AND EXISTS (
    SELECT 1 FROM venues v
    WHERE v.venue_uuid = event_lists.venue_uuid
      AND v.owner_uuid = $2
  )
RETURNING *;

-- name: DeleteEventList :exec
DELETE FROM event_lists
WHERE event_list_uuid = $1
  AND EXISTS (
    SELECT 1 FROM venues v
    WHERE v.venue_uuid = event_lists.venue_uuid
      AND v.owner_uuid = $2
  );
