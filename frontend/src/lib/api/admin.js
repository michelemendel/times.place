import { api } from './client.js';

/**
 * List all owners (admin only)
 * @returns {Promise<any[]>}
 */
export async function listOwners() {
  return api.getJSON('/api/admin/owners');
}

/**
 * Get owner details (admin only)
 * @param {string} uuid
 * @returns {Promise<any>}
 */
export async function getOwner(uuid) {
  return api.getJSON(`/api/admin/owners/${uuid}`);
}

/**
 * Delete owner (admin only)
 * @param {string} uuid
 * @returns {Promise<void>}
 */
export async function deleteOwner(uuid) {
  await api.delete(`/api/admin/owners/${uuid}`);
}

/**
 * List all venues (admin only)
 * @returns {Promise<any[]>}
 */
export async function listVenues() {
  return api.getJSON('/api/admin/venues');
}

/**
 * Update venue limit (admin only)
 * @param {string} uuid
 * @param {number} venueLimit
 * @returns {Promise<void>}
 */
export async function updateVenueLimit(uuid, venueLimit) {
  await api.patch(`/api/admin/owners/${uuid}/venue-limit`, {
    venue_limit: venueLimit,
  });
}
