import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { VenueOwner, Venue, EventList, Event } from './types';

// Persistence keys (for venues, event lists, events - still using localStorage until migration)
const STORAGE_KEY_OWNERS = 'times_place_owners';
const STORAGE_KEY_VENUES = 'times_place_venues';
const STORAGE_KEY_EVENT_LISTS = 'times_place_event_lists';
const STORAGE_KEY_EVENTS = 'times_place_events';

// Helper to load from storage
function load<T>(key: string, fallback: T): T {
  if (!browser) return fallback;
  const stored = localStorage.getItem(key);
  return stored ? JSON.parse(stored) : fallback;
}

// Stores
// currentOwnerStore: No longer persisted to localStorage - managed by API/auth
export const currentOwnerStore = writable<VenueOwner | null>(null);
// ownersStore: Still using localStorage (legacy, will be removed in future migration)
export const ownersStore = writable<VenueOwner[]>(load(STORAGE_KEY_OWNERS, []));
export const venueStore = writable<Venue[]>(load(STORAGE_KEY_VENUES, []));
export const eventListStore = writable<EventList[]>(load(STORAGE_KEY_EVENT_LISTS, []));
export const eventStore = writable<Event[]>(load(STORAGE_KEY_EVENTS, []));

// Subscribe and save to localStorage (excluding currentOwnerStore - managed by API)
if (browser) {
  ownersStore.subscribe((val) => localStorage.setItem(STORAGE_KEY_OWNERS, JSON.stringify(val)));
  venueStore.subscribe((val) => localStorage.setItem(STORAGE_KEY_VENUES, JSON.stringify(val)));
  eventListStore.subscribe((val) => localStorage.setItem(STORAGE_KEY_EVENT_LISTS, JSON.stringify(val)));
  eventStore.subscribe((val) => localStorage.setItem(STORAGE_KEY_EVENTS, JSON.stringify(val)));
}
