-- +goose Up
ALTER TABLE venue_owners
ADD COLUMN is_admin boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE venue_owners
DROP COLUMN is_admin;
