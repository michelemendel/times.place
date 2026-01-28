import { api } from './client.js';

/**
 * Events API client
 *
 * Endpoints:
 * - GET    /api/event-lists/:event_list_uuid/events
 * - POST   /api/event-lists/:event_list_uuid/events
 * - GET    /api/events/:event_uuid
 * - PATCH  /api/events/:event_uuid
 * - DELETE /api/events/:event_uuid
 */

/**
 * List events for an event list (owned by current owner).
 * @param {string} eventListUuid
 * @returns {Promise<import('../types').Event[]>}
 */
export async function listEventsForEventList(eventListUuid) {
  return /** @type {Promise<import('../types').Event[]>} */ (
    api.getJSON(
      `/api/event-lists/${encodeURIComponent(eventListUuid)}/events`,
    )
  );
}

/**
 * Get a single event by UUID (owned by current owner).
 * @param {string} eventUuid
 * @returns {Promise<import('../types').Event>}
 */
export async function getEvent(eventUuid) {
  return /** @type {Promise<import('../types').Event>} */ (
    api.getJSON(`/api/events/${encodeURIComponent(eventUuid)}`)
  );
}

/**
 * Create a new event under an event list.
 *
 * Backend expects:
 * - datetime: RFC3339 string
 * - duration_minutes: integer or null
 * - sort_order: int32 or omitted
 *
 * @param {string} eventListUuid
 * @param {Object} data
 * @param {string} data.event_name
 * @param {string} data.datetime
 * @param {string} [data.comment]
 * @param {number | null} [data.duration_minutes]
 * @param {number} [data.sort_order]
 * @returns {Promise<import('../types').Event>}
 */
export async function createEvent(eventListUuid, data) {
  const payload = {
    event_name: data.event_name,
    datetime: data.datetime,
    comment: data.comment ?? '',
    duration_minutes:
      data.duration_minutes === null || data.duration_minutes === undefined
        ? null
        : Number(data.duration_minutes),
    sort_order: data.sort_order ?? 0,
  };

  return /** @type {Promise<import('../types').Event>} */ (
    api.postJSON(
      `/api/event-lists/${encodeURIComponent(eventListUuid)}/events`,
      payload,
    )
  );
}

/**
 * Partially update an event.
 *
 * @param {string} eventUuid
 * @param {Partial<{
 *   event_name: string;
 *   datetime: string;
 *   comment: string;
 *   duration_minutes: number | null;
 *   sort_order: number;
 * }>} data
 * @returns {Promise<import('../types').Event>}
 */
export async function updateEvent(eventUuid, data) {
  const payload = {};

  if (data.event_name !== undefined) payload.event_name = data.event_name;
  if (data.datetime !== undefined) payload.datetime = data.datetime;
  if (data.comment !== undefined) payload.comment = data.comment;
  if (data.duration_minutes !== undefined) {
    payload.duration_minutes =
      data.duration_minutes === null
        ? null
        : Number(data.duration_minutes);
  }
  if (data.sort_order !== undefined) payload.sort_order = data.sort_order;

  return /** @type {Promise<import('../types').Event>} */ (
    api.patchJSON(`/api/events/${encodeURIComponent(eventUuid)}`, payload)
  );
}

/**
 * Delete an event.
 * @param {string} eventUuid
 * @returns {Promise<void>}
 */
export async function deleteEvent(eventUuid) {
  await api.delete(`/api/events/${encodeURIComponent(eventUuid)}`);
}

