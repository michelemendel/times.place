import { api } from './client.js';

/**
 * Event Lists API client
 *
 * Endpoints:
 * - GET    /api/venues/:venue_uuid/event-lists
 * - POST   /api/venues/:venue_uuid/event-lists
 * - GET    /api/event-lists/:event_list_uuid
 * - PATCH  /api/event-lists/:event_list_uuid
 * - DELETE /api/event-lists/:event_list_uuid
 */

/**
 * List event lists for a venue (owned by current owner).
 * @param {string} venueUuid
 * @returns {Promise<import('../types').EventList[]>}
 */
export async function listEventListsForVenue(venueUuid) {
  return /** @type {Promise<import('../types').EventList[]>} */ (
    api.getJSON(`/api/venues/${encodeURIComponent(venueUuid)}/event-lists`)
  );
}

/**
 * Get a single event list by UUID (owned by current owner).
 * @param {string} eventListUuid
 * @returns {Promise<import('../types').EventList>}
 */
export async function getEventList(eventListUuid) {
  return /** @type {Promise<import('../types').EventList>} */ (
    api.getJSON(`/api/event-lists/${encodeURIComponent(eventListUuid)}`)
  );
}

/**
 * Create a new event list under a venue.
 *
 * Backend:
 * - Parses `date` as "YYYY-MM-DD" or empty
 * - Generates `private_link_token` when not provided
 * - Default `sort_order` is 0 when omitted
 *
 * @param {string} venueUuid
 * @param {Object} data
 * @param {string} [data.name]
 * @param {string} [data.date] - "YYYY-MM-DD" or empty
 * @param {string} [data.comment]
 * @param {'public'|'private'} data.visibility
 * @param {number} [data.sort_order]
 * @returns {Promise<import('../types').EventList>}
 */
export async function createEventList(venueUuid, data) {
  const payload = {
    name: data.name ?? '',
    date: data.date ?? '',
    comment: data.comment ?? '',
    visibility: data.visibility,
    sort_order: data.sort_order ?? 0,
    // private_link_token omitted so backend generates one
  };

  return /** @type {Promise<import('../types').EventList>} */ (
    api.postJSON(
      `/api/venues/${encodeURIComponent(venueUuid)}/event-lists`,
      payload,
    )
  );
}

/**
 * Partially update an event list.
 *
 * @param {string} eventListUuid
 * @param {Partial<{
 *   name: string;
 *   date: string;
 *   comment: string;
 *   visibility: 'public' | 'private';
 *   private_link_token: string | null;
 *   sort_order: number;
 * }>} data
 * @returns {Promise<import('../types').EventList>}
 */
export async function updateEventList(eventListUuid, data) {
  const payload = {};

  if (data.name !== undefined) payload.name = data.name;
  if (data.date !== undefined) payload.date = data.date;
  if (data.comment !== undefined) payload.comment = data.comment;
  if (data.visibility !== undefined) payload.visibility = data.visibility;
  if (data.private_link_token !== undefined) {
    // Backend interprets empty string as "clear token", any non-empty as explicit UUID
    payload.private_link_token = data.private_link_token ?? '';
  }
  if (data.sort_order !== undefined) payload.sort_order = data.sort_order;

  return /** @type {Promise<import('../types').EventList>} */ (
    api.patchJSON(
      `/api/event-lists/${encodeURIComponent(eventListUuid)}`,
      payload,
    )
  );
}

/**
 * Delete an event list (and its events via DB cascade).
 * @param {string} eventListUuid
 * @returns {Promise<void>}
 */
export async function deleteEventList(eventListUuid) {
  await api.delete(`/api/event-lists/${encodeURIComponent(eventListUuid)}`);
}

