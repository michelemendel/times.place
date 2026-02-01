-- +goose Up
ALTER TABLE venue_owners
ADD COLUMN is_demo boolean NOT NULL DEFAULT false;

-- Mark existing demo accounts (from seed) so they are locked and clearable
UPDATE venue_owners SET is_demo = true WHERE email IN ('abe@demo.org', 'ben@demo.org');

-- +goose Down
ALTER TABLE venue_owners DROP COLUMN is_demo;
