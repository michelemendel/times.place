export interface VenueOwner {
  owner_uuid: string;
  name: string;
  mobile: string;
  email: string;
  is_admin?: boolean;
  /** Whether the owner has verified their email; required for creating/editing venues and events. */
  email_verified?: boolean;
  /**
   * Password field (optional, used during registration/login).
   * Passwords are handled by the backend and never stored client-side.
   */
  password?: string;
  created_at: string;
  modified_at: string;
}

export interface Venue {
  venue_uuid: string;
  owner_uuid: string;
  name: string;
  banner_image: string;
  address: string;
  geolocation: string;
  /** Owner display name; set by public API, not by owner API */
  owner_name?: string;
  /** Owner email; set by public API, not by owner API */
  owner_email?: string;
  comment?: string;
  /** Client-side only; API does not return this. Derived from listing event-lists. */
  event_list_uuids?: string[];
  /** Returned by GET /api/venues (owner list) so My Venues can show event lists without extra calls. */
  event_lists?: EventList[];
  timezone: string;
  private_link_token?: string;
  created_at: string;
  modified_at: string;
}

export interface EventList {
  event_list_uuid: string;
  venue_uuid: string;
  /** Client-side only; API does not return this. Derived from listing events. */
  event_uuids?: string[];
  name: string;
  date: string; // ISO 8601 date string (e.g., "2024-12-25")
  comment?: string;
  visibility: 'public' | 'private';
  private_link_token?: string;
  sort_order: number;
  /** Client-side only; added when loading venue data for convenience */
  events: Event[];
  created_at: string;
  modified_at: string;
}

export interface Event {
  event_uuid: string;
  event_list_uuid: string;
  event_name: string;
  event_date: string; // YYYY-MM-DD
  event_time: string; // HH:MM:SS
  comment?: string;
  duration_minutes?: number | null;
  /** UI-only property: HH:MM for input binding */
  time?: string;
  /** UI-only property: YYYY-MM-DD for input binding */
  date?: string;
  sort_order: number;
  created_at: string;
  modified_at: string;
}
