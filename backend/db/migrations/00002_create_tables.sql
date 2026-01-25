-- +goose Up
-- Create venue_owners table
CREATE TABLE venue_owners (
    owner_uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    mobile text NOT NULL,
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    modified_at timestamptz NOT NULL DEFAULT now()
);

-- Create venues table
CREATE TABLE venues (
    venue_uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_uuid uuid NOT NULL REFERENCES venue_owners(owner_uuid) ON DELETE CASCADE,
    name text NOT NULL,
    banner_image text NOT NULL DEFAULT '',
    address text NOT NULL DEFAULT '',
    geolocation text NOT NULL DEFAULT '',
    comment text,
    timezone text NOT NULL DEFAULT '',
    visibility text NOT NULL CHECK (visibility IN ('public', 'private')),
    private_link_token uuid UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    modified_at timestamptz NOT NULL DEFAULT now()
);

-- Create event_lists table
CREATE TABLE event_lists (
    event_list_uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_uuid uuid NOT NULL REFERENCES venues(venue_uuid) ON DELETE CASCADE,
    name text NOT NULL DEFAULT '',
    date date,
    comment text,
    visibility text NOT NULL CHECK (visibility IN ('public', 'private')),
    private_link_token uuid UNIQUE,
    sort_order int NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    modified_at timestamptz NOT NULL DEFAULT now()
);

-- Create events table
CREATE TABLE events (
    event_uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_list_uuid uuid NOT NULL REFERENCES event_lists(event_list_uuid) ON DELETE CASCADE,
    event_name text NOT NULL DEFAULT '',
    datetime timestamptz NOT NULL,
    comment text,
    duration_minutes int,
    sort_order int NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    modified_at timestamptz NOT NULL DEFAULT now()
);

-- Create refresh_tokens table
CREATE TABLE refresh_tokens (
    refresh_token_uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_uuid uuid NOT NULL REFERENCES venue_owners(owner_uuid) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    issued_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    replaced_by_token_uuid uuid,
    user_agent text,
    ip_address text
);

-- +goose Down
-- Drop tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS event_lists;
DROP TABLE IF EXISTS venues;
DROP TABLE IF EXISTS venue_owners;
