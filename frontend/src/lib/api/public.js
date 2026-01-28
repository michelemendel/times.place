import { api } from './client.js';

/**
 * Public (unauthenticated) API client
 *
 * Endpoints:
 * - GET /api/public/venues?query=...
 * - GET /api/public/venues/:venue_uuid/event-lists
 * - GET /api/public/venues/by-token/:token
 * - GET /api/public/event-lists/by-token/:token
 */

/**
 * List public venues, optionally filtered by a search query.
 * @param {string} [query]
 * @returns {Promise<import('../types').Venue[]>}
 */
export async function listPublicVenues(query) {
  const qs = query && query.trim() ? `?query=${encodeURIComponent(query.trim())}` : '';
  return /** @type {Promise<import('../types').Venue[]>} */ (
    api.getJSON(`/api/public/venues${qs}`)
  );
}

/**
 * Get public event lists for a venue.
 * @param {string} venueUuid
 * @returns {Promise<import('../types').EventList[]>}
 */
export async function getPublicEventListsForVenue(venueUuid) {
  return /** @type {Promise<import('../types').EventList[]>} */ (
    api.getJSON(
      `/api/public/venues/${encodeURIComponent(venueUuid)}/event-lists`,
    )
  );
}

/**
 * Access a private venue (and its event lists) via token.
 *
 * Response shape:
 * {
 *   venue: Venue,
 *   event_lists: EventList[]
 * }
 *
 * @param {string} token
 * @returns {Promise<{ venue: import('../types').Venue; event_lists: import('../types').EventList[] }>}
 */
export async function getPrivateVenueByToken(token) {
  return api.getJSON(
    `/api/public/venues/by-token/${encodeURIComponent(token)}`,
  );
}

/**
 * Access a private event list (with its venue and events) via token.
 *
 * Response shape:
 * {
 *   venue: Venue,
 *   event_list: EventList,
 *   events: Event[]
 * }
 *
 * @param {string} token
 * @returns {Promise<{ venue: import('../types').Venue; event_list: import('../types').EventList; events: import('../types').Event[] }>}
 */
export async function getPrivateEventListByToken(token) {
  return api.getJSON(
    `/api/public/event-lists/by-token/${encodeURIComponent(token)}`,
  );
}

/**
 * Get events for a public event list.
 * @param {string} eventListUuid
 * @returns {Promise<import('../types').Event[]>}
 */
export async function getPublicEventsForEventList(eventListUuid) {
  return /** @type {Promise<import('../types').Event[]>} */ (
    api.getJSON(
      `/api/public/event-lists/${encodeURIComponent(eventListUuid)}/events`,
    )
  );
}

