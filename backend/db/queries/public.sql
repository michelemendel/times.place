-- name: ListPublicVenues :many
-- Show any venue that has at least one public event list.
SELECT 
    v.venue_uuid,
    v.name,
    v.banner_image,
    v.address,
    v.geolocation,
    v.comment,
    v.timezone,
    v.private_link_token,
    v.created_at,
    v.modified_at,
    o.name AS owner_name,
    o.email AS owner_email
FROM venues v
INNER JOIN venue_owners o ON o.owner_uuid = v.owner_uuid
WHERE EXISTS (
    SELECT 1 FROM event_lists
    WHERE event_lists.venue_uuid = v.venue_uuid
      AND event_lists.visibility = 'public'
  )
ORDER BY v.created_at DESC;

-- name: ListPublicVenuesWithDistance :many
-- Show any venue that has at least one public event list, optionally annotated with distance from a user point.
-- Params:
--   user_lat (double precision)
--   user_lng (double precision)
--   radius_km (double precision, nullable)
WITH venues_with_distance AS (
  SELECT
    v.venue_uuid,
    v.name,
    v.banner_image,
    v.address,
    v.geolocation,
    v.comment,
    v.timezone,
    v.private_link_token,
    v.created_at,
    v.modified_at,
    o.name AS owner_name,
    o.email AS owner_email,
    CASE
      WHEN btrim(split_part(v.geolocation, ',', 1)) ~ '^-?[0-9]+(\\.[0-9]+)?$'
       AND btrim(split_part(v.geolocation, ',', 2)) ~ '^-?[0-9]+(\\.[0-9]+)?$' THEN
        ROUND(
          (
            2 * 6371 * ASIN(
              SQRT(
                POWER(SIN(RADIANS((btrim(split_part(v.geolocation, ',', 1))::double precision - sqlc.arg(user_lat)::double precision) / 2)), 2) +
                COS(RADIANS(sqlc.arg(user_lat)::double precision)) * COS(RADIANS(btrim(split_part(v.geolocation, ',', 1))::double precision)) *
                POWER(SIN(RADIANS((btrim(split_part(v.geolocation, ',', 2))::double precision - sqlc.arg(user_lng)::double precision) / 2)), 2)
              )
            )
          )::numeric,
          4
        )::double precision
      ELSE NULL
    END AS distance_km
  FROM venues v
  INNER JOIN venue_owners o ON o.owner_uuid = v.owner_uuid
  WHERE EXISTS (
      SELECT 1 FROM event_lists
      WHERE event_lists.venue_uuid = v.venue_uuid
        AND event_lists.visibility = 'public'
    )
)
SELECT
  venue_uuid,
  name,
  banner_image,
  address,
  geolocation,
  comment,
  timezone,
  private_link_token,
  created_at,
  modified_at,
  owner_name,
  owner_email,
  distance_km
FROM venues_with_distance
WHERE (sqlc.narg(radius_km)::double precision IS NULL OR (distance_km IS NOT NULL AND distance_km <= sqlc.narg(radius_km)::double precision))
ORDER BY (distance_km IS NULL) ASC, distance_km ASC, created_at DESC;

-- name: SearchPublicVenues :many
SELECT DISTINCT
    v.venue_uuid,
    v.name,
    v.banner_image,
    v.address,
    v.geolocation,
    v.comment,
    v.timezone,
    v.private_link_token,
    v.created_at,
    v.modified_at,
    o.name AS owner_name,
    o.email AS owner_email
FROM venues v
INNER JOIN venue_owners o ON o.owner_uuid = v.owner_uuid
WHERE EXISTS (
    SELECT 1 FROM event_lists el
    WHERE el.venue_uuid = v.venue_uuid
      AND el.visibility = 'public'
  )
  AND (
    v.name ILIKE '%' || $1 || '%'
    OR v.address ILIKE '%' || $1 || '%'
    OR v.comment ILIKE '%' || $1 || '%'
    OR EXISTS (
      SELECT 1 FROM event_lists el
      WHERE el.venue_uuid = v.venue_uuid
        AND (el.name ILIKE '%' || $1 || '%' OR el.comment ILIKE '%' || $1 || '%')
    )
    OR EXISTS (
      SELECT 1 FROM events e
      INNER JOIN event_lists el ON e.event_list_uuid = el.event_list_uuid
      WHERE el.venue_uuid = v.venue_uuid
        AND (e.event_name ILIKE '%' || $1 || '%' OR e.comment ILIKE '%' || $1 || '%')
    )
  )
ORDER BY v.created_at DESC;

-- name: SearchPublicVenuesWithDistance :many
-- Search public venues, optionally annotated with distance from a user point.
-- Params:
--   query text
--   user_lat (double precision)
--   user_lng (double precision)
--   radius_km (double precision, nullable)
WITH venues_with_distance AS (
  SELECT DISTINCT
    v.venue_uuid,
    v.name,
    v.banner_image,
    v.address,
    v.geolocation,
    v.comment,
    v.timezone,
    v.private_link_token,
    v.created_at,
    v.modified_at,
    o.name AS owner_name,
    o.email AS owner_email,
    CASE
      WHEN btrim(split_part(v.geolocation, ',', 1)) ~ '^-?[0-9]+(\\.[0-9]+)?$'
       AND btrim(split_part(v.geolocation, ',', 2)) ~ '^-?[0-9]+(\\.[0-9]+)?$' THEN
        ROUND(
          (
            2 * 6371 * ASIN(
              SQRT(
                POWER(SIN(RADIANS((btrim(split_part(v.geolocation, ',', 1))::double precision - sqlc.arg(user_lat)::double precision) / 2)), 2) +
                COS(RADIANS(sqlc.arg(user_lat)::double precision)) * COS(RADIANS(btrim(split_part(v.geolocation, ',', 1))::double precision)) *
                POWER(SIN(RADIANS((btrim(split_part(v.geolocation, ',', 2))::double precision - sqlc.arg(user_lng)::double precision) / 2)), 2)
              )
            )
          )::numeric,
          4
        )::double precision
      ELSE NULL
    END AS distance_km
  FROM venues v
  INNER JOIN venue_owners o ON o.owner_uuid = v.owner_uuid
  WHERE EXISTS (
      SELECT 1 FROM event_lists el
      WHERE el.venue_uuid = v.venue_uuid
        AND el.visibility = 'public'
    )
    AND (
      v.name ILIKE '%' || $1 || '%'
      OR v.address ILIKE '%' || $1 || '%'
      OR v.comment ILIKE '%' || $1 || '%'
      OR EXISTS (
        SELECT 1 FROM event_lists el
        WHERE el.venue_uuid = v.venue_uuid
          AND (el.name ILIKE '%' || $1 || '%' OR el.comment ILIKE '%' || $1 || '%')
      )
      OR EXISTS (
        SELECT 1 FROM events e
        INNER JOIN event_lists el ON e.event_list_uuid = el.event_list_uuid
        WHERE el.venue_uuid = v.venue_uuid
          AND (e.event_name ILIKE '%' || $1 || '%' OR e.comment ILIKE '%' || $1 || '%')
      )
    )
)
SELECT
  venue_uuid,
  name,
  banner_image,
  address,
  geolocation,
  comment,
  timezone,
  private_link_token,
  created_at,
  modified_at,
  owner_name,
  owner_email,
  distance_km
FROM venues_with_distance
WHERE (sqlc.narg(radius_km)::double precision IS NULL OR (distance_km IS NOT NULL AND distance_km <= sqlc.narg(radius_km)::double precision))
ORDER BY (distance_km IS NULL) ASC, distance_km ASC, created_at DESC;

-- name: GetPublicEventListsByVenue :many
SELECT 
    event_list_uuid,
    venue_uuid,
    name,
    date,
    comment,
    visibility,
    private_link_token,
    sort_order,
    created_at,
    modified_at
FROM event_lists
WHERE venue_uuid = $1 AND visibility = 'public'
ORDER BY sort_order ASC, created_at ASC;

-- name: GetVenueByToken :one
SELECT 
    v.venue_uuid,
    v.name,
    v.banner_image,
    v.address,
    v.geolocation,
    v.comment,
    v.timezone,
    v.private_link_token,
    v.created_at,
    v.modified_at,
    o.name AS owner_name,
    o.email AS owner_email
FROM venues v
INNER JOIN venue_owners o ON o.owner_uuid = v.owner_uuid
WHERE v.private_link_token = $1;

-- name: GetEventListByToken :one
SELECT 
    el.event_list_uuid,
    el.venue_uuid,
    el.name,
    el.date,
    el.comment,
    el.visibility,
    el.private_link_token,
    el.sort_order,
    el.created_at,
    el.modified_at
FROM event_lists el
WHERE el.private_link_token = $1;

-- name: GetVenueWithEventListsByToken :many
-- Returns venue and all its event lists (public + private if venue token matches)
SELECT 
    v.venue_uuid,
    v.name,
    v.banner_image,
    v.address,
    v.geolocation,
    v.comment,
    v.timezone,
    v.private_link_token,
    v.created_at,
    v.modified_at,
    o.name AS owner_name,
    o.email AS owner_email,
    el.event_list_uuid,
    el.name as event_list_name,
    el.date as event_list_date,
    el.comment as event_list_comment,
    el.visibility as event_list_visibility,
    el.private_link_token as event_list_private_link_token,
    el.sort_order as event_list_sort_order,
    el.created_at as event_list_created_at,
    el.modified_at as event_list_modified_at
FROM venues v
INNER JOIN venue_owners o ON o.owner_uuid = v.owner_uuid
LEFT JOIN event_lists el ON el.venue_uuid = v.venue_uuid
WHERE v.private_link_token = $1
  AND (el.visibility = 'public' OR el.venue_uuid = v.venue_uuid)
ORDER BY el.sort_order ASC, el.created_at ASC;
