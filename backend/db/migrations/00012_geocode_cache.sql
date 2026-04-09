-- +goose Up
CREATE TABLE IF NOT EXISTS geocode_cache (
    normalized_address text PRIMARY KEY,
    lat double precision NOT NULL,
    lng double precision NOT NULL,
    display_name text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS geocode_cache_updated_at_idx ON geocode_cache(updated_at);

-- +goose Down
DROP TABLE IF EXISTS geocode_cache;

