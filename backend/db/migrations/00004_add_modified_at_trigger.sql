-- +goose Up
-- Create function to update modified_at timestamp
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_modified_at()
RETURNS TRIGGER AS $BODY$
BEGIN
    NEW.modified_at = now();
    RETURN NEW;
END;
$BODY$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Create triggers for all tables with modified_at
CREATE TRIGGER update_venue_owners_modified_at
    BEFORE UPDATE ON venue_owners
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

CREATE TRIGGER update_venues_modified_at
    BEFORE UPDATE ON venues
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

CREATE TRIGGER update_event_lists_modified_at
    BEFORE UPDATE ON event_lists
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

CREATE TRIGGER update_events_modified_at
    BEFORE UPDATE ON events
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

-- +goose Down
-- Drop triggers
DROP TRIGGER IF EXISTS update_events_modified_at ON events;
DROP TRIGGER IF EXISTS update_event_lists_modified_at ON event_lists;
DROP TRIGGER IF EXISTS update_venues_modified_at ON venues;
DROP TRIGGER IF EXISTS update_venue_owners_modified_at ON venue_owners;

-- Drop function
DROP FUNCTION IF EXISTS update_modified_at();
