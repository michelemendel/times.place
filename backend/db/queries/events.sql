-- name: ListEventsByEventListAndOwner :many
SELECT e.* FROM events e
INNER JOIN event_lists el ON e.event_list_uuid = el.event_list_uuid
INNER JOIN venues v ON el.venue_uuid = v.venue_uuid
WHERE e.event_list_uuid = $1 AND v.owner_uuid = $2
ORDER BY e.sort_order ASC, e.event_date ASC, e.event_time ASC;

-- name: GetEventByIDAndOwner :one
SELECT e.* FROM events e
INNER JOIN event_lists el ON e.event_list_uuid = el.event_list_uuid
INNER JOIN venues v ON el.venue_uuid = v.venue_uuid
WHERE e.event_uuid = $1 AND v.owner_uuid = $2;

-- name: CreateEvent :one
INSERT INTO events (
    event_list_uuid,
    event_name,
    event_date,
    event_time,
    comment,
    duration_minutes,
    sort_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateEvent :one
UPDATE events
SET
    event_name = COALESCE($3, event_name),
    event_date = $4,
    event_time = COALESCE($5, event_time),
    comment = $6,
    duration_minutes = $7,
    sort_order = COALESCE($8, sort_order)
WHERE event_uuid = $1
  AND EXISTS (
    SELECT 1 FROM event_lists el
    INNER JOIN venues v ON el.venue_uuid = v.venue_uuid
    WHERE el.event_list_uuid = events.event_list_uuid
      AND v.owner_uuid = $2
  )
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE event_uuid = $1
  AND EXISTS (
    SELECT 1 FROM event_lists el
    INNER JOIN venues v ON el.venue_uuid = v.venue_uuid
    WHERE el.event_list_uuid = events.event_list_uuid
      AND v.owner_uuid = $2
  );
