import { api } from './client.js';

/**
 * Venues API client
 *
 * All functions talk to the authenticated owner endpoints:
 * - GET    /api/venues
 * - POST   /api/venues
 * - GET    /api/venues/:venue_uuid
 * - PATCH  /api/venues/:venue_uuid
 * - DELETE /api/venues/:venue_uuid
 */

/**
 * List all venues for current authenticated owner.
 * @returns {Promise<import('../types').Venue[]>}
 */
export async function listVenues() {
  return /** @type {Promise<import('../types').Venue[]>} */ (api.getJSON('/api/venues'));
}

/**
 * Get a single venue by UUID (scoped to current owner).
 * @param {string} venueUuid
 * @returns {Promise<import('../types').Venue>}
 */
export async function getVenue(venueUuid) {
  return /** @type {Promise<import('../types').Venue>} */ (
    api.getJSON(`/api/venues/${encodeURIComponent(venueUuid)}`)
  );
}

/**
 * Create a new venue.
 * The backend derives owner UUID from the JWT and will generate
 * a private_link_token when not provided.
 *
 * @param {Object} data
 * @param {string} data.name
 * @param {string} [data.banner_image]
 * @param {string} [data.address]
 * @param {string} [data.geolocation]
 * @param {string} [data.comment]
 * @param {string} [data.timezone]
 * @param {'public'|'private'} [data.visibility] - defaults to "private" if omitted
 * @returns {Promise<import('../types').Venue>}
 */
export async function createVenue(data) {
  const payload = {
    name: data.name,
    banner_image: data.banner_image ?? '',
    address: data.address ?? '',
    geolocation: data.geolocation ?? '',
    comment: data.comment ?? '',
    timezone: data.timezone ?? '',
    visibility: data.visibility ?? 'private',
    // Let backend generate private_link_token when not provided
  };

  return /** @type {Promise<import('../types').Venue>} */ (
    api.postJSON('/api/venues', payload)
  );
}

/**
 * Partially update an existing venue.
 *
 * Only provided fields are updated; others keep their current values.
 *
 * @param {string} venueUuid
 * @param {Partial<{
 *   name: string;
 *   banner_image: string;
 *   address: string;
 *   geolocation: string;
 *   comment: string;
 *   timezone: string;
 *   visibility: 'public' | 'private';
 * }>} data
 * @returns {Promise<import('../types').Venue>}
 */
export async function updateVenue(venueUuid, data) {
  const payload = {};

  if (data.name !== undefined) payload.name = data.name;
  if (data.banner_image !== undefined) payload.banner_image = data.banner_image;
  if (data.address !== undefined) payload.address = data.address;
  if (data.geolocation !== undefined) payload.geolocation = data.geolocation;
  if (data.comment !== undefined) payload.comment = data.comment;
  if (data.timezone !== undefined) payload.timezone = data.timezone;
  if (data.visibility !== undefined) payload.visibility = data.visibility;

  return /** @type {Promise<import('../types').Venue>} */ (
    api.patchJSON(`/api/venues/${encodeURIComponent(venueUuid)}`, payload)
  );
}

/**
 * Delete a venue (and all related event lists/events via DB cascade).
 * @param {string} venueUuid
 * @returns {Promise<void>}
 */
export async function deleteVenue(venueUuid) {
  await api.delete(`/api/venues/${encodeURIComponent(venueUuid)}`);
}

