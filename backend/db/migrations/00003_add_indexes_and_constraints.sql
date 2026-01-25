-- +goose Up
-- Add indexes for venues
CREATE INDEX venues_owner_uuid_idx ON venues(owner_uuid);
CREATE INDEX venues_visibility_idx ON venues(visibility);
CREATE INDEX venues_private_link_token_idx ON venues(private_link_token) WHERE private_link_token IS NOT NULL;

-- Add indexes for event_lists
CREATE INDEX event_lists_venue_uuid_idx ON event_lists(venue_uuid);
CREATE INDEX event_lists_visibility_idx ON event_lists(visibility);
CREATE INDEX event_lists_private_link_token_idx ON event_lists(private_link_token) WHERE private_link_token IS NOT NULL;
CREATE INDEX event_lists_venue_uuid_sort_order_idx ON event_lists(venue_uuid, sort_order);

-- Add indexes for events
CREATE INDEX events_event_list_uuid_idx ON events(event_list_uuid);
CREATE INDEX events_event_list_uuid_sort_order_idx ON events(event_list_uuid, sort_order);

-- Add indexes for refresh_tokens
CREATE INDEX refresh_tokens_owner_uuid_idx ON refresh_tokens(owner_uuid);
CREATE INDEX refresh_tokens_token_hash_idx ON refresh_tokens(token_hash);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS refresh_tokens_token_hash_idx;
DROP INDEX IF EXISTS refresh_tokens_owner_uuid_idx;
DROP INDEX IF EXISTS events_event_list_uuid_sort_order_idx;
DROP INDEX IF EXISTS events_event_list_uuid_idx;
DROP INDEX IF EXISTS event_lists_venue_uuid_sort_order_idx;
DROP INDEX IF EXISTS event_lists_private_link_token_idx;
DROP INDEX IF EXISTS event_lists_visibility_idx;
DROP INDEX IF EXISTS event_lists_venue_uuid_idx;
DROP INDEX IF EXISTS venues_private_link_token_idx;
DROP INDEX IF EXISTS venues_visibility_idx;
DROP INDEX IF EXISTS venues_owner_uuid_idx;
