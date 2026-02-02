-- Consolidated schema for sqlc
-- This file mirrors the migrations in db/migrations/

-- Enable pgcrypto extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create venue_owners table
CREATE TABLE venue_owners (
    owner_uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    mobile text NOT NULL,
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    is_admin boolean NOT NULL DEFAULT false,
    is_demo boolean NOT NULL DEFAULT false,
    venue_limit integer NOT NULL DEFAULT 2,
    email_verified_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    modified_at timestamptz NOT NULL DEFAULT now()
);

-- Create email_verification_tokens table (separate table for token hygiene and cleanup)
CREATE TABLE email_verification_tokens (
    token_uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_uuid uuid NOT NULL REFERENCES venue_owners(owner_uuid) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
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

-- Indexes for venues
CREATE INDEX venues_owner_uuid_idx ON venues(owner_uuid);
CREATE INDEX venues_private_link_token_idx ON venues(private_link_token) WHERE private_link_token IS NOT NULL;

-- Indexes for event_lists
CREATE INDEX event_lists_venue_uuid_idx ON event_lists(venue_uuid);
CREATE INDEX event_lists_visibility_idx ON event_lists(visibility);
CREATE INDEX event_lists_private_link_token_idx ON event_lists(private_link_token) WHERE private_link_token IS NOT NULL;
CREATE INDEX event_lists_venue_uuid_sort_order_idx ON event_lists(venue_uuid, sort_order);

-- Indexes for events
CREATE INDEX events_event_list_uuid_idx ON events(event_list_uuid);
CREATE INDEX events_event_list_uuid_sort_order_idx ON events(event_list_uuid, sort_order);

-- Indexes for refresh_tokens
CREATE INDEX refresh_tokens_owner_uuid_idx ON refresh_tokens(owner_uuid);
CREATE INDEX refresh_tokens_token_hash_idx ON refresh_tokens(token_hash);

-- Indexes for email_verification_tokens
CREATE INDEX email_verification_tokens_token_hash_idx ON email_verification_tokens(token_hash);
CREATE INDEX email_verification_tokens_owner_uuid_idx ON email_verification_tokens(owner_uuid);
CREATE INDEX email_verification_tokens_expires_at_idx ON email_verification_tokens(expires_at);
