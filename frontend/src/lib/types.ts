export interface VenueOwner {
  owner_uuid: string;
  name: string;
  mobile: string;
  email: string;
  /**
   * Prototype-only password stored in localStorage.
   * Not secure; for demo purposes only.
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
  comment?: string;
  event_list_uuids: string[];
  timezone: string;
  visibility: 'public' | 'private';
  private_link_token?: string;
  created_at: string;
  modified_at: string;
}

export interface EventList {
  event_list_uuid: string;
  venue_uuid: string;
  event_uuids: string[];
  name: string;
  date: string; // ISO 8601 date string (e.g., "2024-12-25")
  comment?: string;
  private_link_token?: string;
  created_at: string;
  modified_at: string;
}

export interface Event {
  event_uuid: string;
  event_list_uuid: string;
  event_name: string;
  datetime: string; // RFC3339
  comment?: string;
  duration_minutes?: number;
  created_at: string;
  modified_at: string;
}
