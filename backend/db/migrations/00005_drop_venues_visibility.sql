-- +goose Up
-- Visibility is only meaningful at event_list level (controllable in GUI). Drop from venues.
DROP INDEX IF EXISTS venues_visibility_idx;
ALTER TABLE venues DROP COLUMN IF EXISTS visibility;

-- +goose Down
ALTER TABLE venues ADD COLUMN visibility text NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private'));
CREATE INDEX venues_visibility_idx ON venues(visibility);
