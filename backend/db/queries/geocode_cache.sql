-- name: GetGeocodeCache :one
SELECT normalized_address, lat, lng, display_name, created_at, updated_at
FROM geocode_cache
WHERE normalized_address = $1;

-- name: UpsertGeocodeCache :exec
INSERT INTO geocode_cache (
  normalized_address,
  lat,
  lng,
  display_name,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, now(), now()
)
ON CONFLICT (normalized_address) DO UPDATE
SET
  lat = EXCLUDED.lat,
  lng = EXCLUDED.lng,
  display_name = EXCLUDED.display_name,
  updated_at = now();

-- name: ListVenuesNeedingGeocode :many
SELECT venue_uuid, address
FROM venues
WHERE geolocation = '' AND address <> ''
ORDER BY created_at ASC;

-- name: SetVenueGeolocation :exec
UPDATE venues
SET geolocation = $2, modified_at = now()
WHERE venue_uuid = $1;

