import { api } from './client.js';

/**
 * List all owners (admin only)
 * @returns {Promise<any[]>}
 */
export async function listOwners() {
  return await api.getJSON('/api/admin/owners');
}

/**
 * Get owner details (admin only)
 * @param {string} uuid
 * @returns {Promise<any>}
 */
export async function getOwner(uuid) {
  return await api.getJSON(`/api/admin/owners/${uuid}`);
}

/**
 * Delete owner (admin only)
 * @param {string} uuid
 * @returns {Promise<void>}
 */
export async function deleteOwner(uuid) {
  return await api.delete(`/api/admin/owners/${uuid}`);
}

/**
 * List all venues (admin only)
 * @returns {Promise<any[]>}
 */
export async function listVenues() {
  return await api.getJSON('/api/admin/venues');
}
