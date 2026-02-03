-- +goose Up
ALTER TABLE events ADD COLUMN event_date DATE;
ALTER TABLE events ADD COLUMN event_time TIME;

-- Update existing data splitting datetime into date and time in the venue's timezone
UPDATE events e
SET 
  event_date = (e.datetime AT TIME ZONE COALESCE(NULLIF(v.timezone, ''), 'UTC'))::date,
  event_time = (e.datetime AT TIME ZONE COALESCE(NULLIF(v.timezone, ''), 'UTC'))::time
FROM event_lists el
JOIN venues v ON el.venue_uuid = v.venue_uuid
WHERE e.event_list_uuid = el.event_list_uuid;

-- Drop the old column and set NOT NULL on the new time column
ALTER TABLE events ALTER COLUMN event_time SET NOT NULL;
ALTER TABLE events DROP COLUMN datetime;

-- Update indexes
DROP INDEX IF EXISTS events_event_list_uuid_sort_order_idx;
CREATE INDEX events_event_list_uuid_sort_order_idx ON events(event_list_uuid, sort_order, event_date, event_time);

-- +goose Down
ALTER TABLE events ADD COLUMN datetime TIMESTAMPTZ;

-- Reconstruct datetime from event_date and event_time
-- Use 2000-01-01 as default date if event_date is NULL (sentinel logic)
UPDATE events e
SET datetime = (COALESCE(e.event_date, '2000-01-01')::text || ' ' || e.event_time::text)::timestamp AT TIME ZONE COALESCE(NULLIF(v.timezone, ''), 'UTC')
FROM event_lists el
JOIN venues v ON el.venue_uuid = v.venue_uuid
WHERE e.event_list_uuid = el.event_list_uuid;

ALTER TABLE events ALTER COLUMN datetime SET NOT NULL;
ALTER TABLE events DROP COLUMN event_date;
ALTER TABLE events DROP COLUMN event_time;

DROP INDEX IF EXISTS events_event_list_uuid_sort_order_idx;
CREATE INDEX events_event_list_uuid_sort_order_idx ON events(event_list_uuid, sort_order, datetime);
