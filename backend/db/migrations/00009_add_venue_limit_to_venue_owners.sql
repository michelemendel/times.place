-- +goose Up
ALTER TABLE venue_owners
ADD COLUMN venue_limit integer NOT NULL DEFAULT 2;

-- +goose Down
ALTER TABLE venue_owners
DROP COLUMN venue_limit;
