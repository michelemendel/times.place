/**
 * Authentication API
 * 
 * Functions for authentication endpoints:
 * - Register
 * - Login
 * - Logout
 * - Refresh token
 * - Get current user
 */

import { api, setAccessToken, clearAccessToken } from './client.js';
import { currentOwnerStore } from '../stores.js';

/**
 * Register a new venue owner
 * @param {Object} data - Registration data
 * @param {string} data.name - Owner name
 * @param {string} data.email - Owner email
 * @param {string} data.mobile - Owner mobile number
 * @param {string} data.password - Owner password
 * @returns {Promise<{owner: VenueOwner, access_token: string}>}
 */
export async function register(data) {
  const response = await api.postJSON('/api/auth/register', {
    name: data.name,
    email: data.email,
    mobile: data.mobile,
    password: data.password,
  });

  // Store access token in memory
  setAccessToken(response.access_token);

  // Store owner data in store
  currentOwnerStore.set(response.owner);

  return response;
}

/**
 * Login with email and password
 * @param {string} email - Owner email
 * @param {string} password - Owner password
 * @returns {Promise<{owner: VenueOwner, access_token: string}>}
 */
export async function login(email, password) {
  const response = await api.postJSON('/api/auth/login', {
    email,
    password,
  });

  // Store access token in memory
  setAccessToken(response.access_token);

  // Store owner data in store
  currentOwnerStore.set(response.owner);

  return response;
}

/**
 * Logout and revoke refresh token
 * @returns {Promise<void>}
 */
export async function logout() {
  try {
    await api.post('/api/auth/logout', {});
  } catch (error) {
    // Continue with logout even if API call fails
    console.error('Logout API call failed:', error);
  } finally {
    // Clear access token from memory
    clearAccessToken();

    // Clear owner data from store
    currentOwnerStore.set(null);
  }
}

/**
 * Get full /api/auth/me response (owner, venue_count, venue_limit).
 * Use when you need venue limit info (e.g. to grey out "Add Venue").
 * @returns {Promise<{owner: import('$lib/types').VenueOwner, venue_count?: number, venue_limit?: number}>}
 */
export async function getAuthMe() {
  const { getAccessToken, setAccessToken } = await import('./client.js');

  if (!getAccessToken()) {
    try {
      const refreshResponse = await api.postJSON('/api/auth/refresh', {});
      setAccessToken(refreshResponse.access_token);
    } catch (refreshError) {
      throw refreshError;
    }
  }

  const response = await api.getJSON('/api/auth/me');
  currentOwnerStore.set(response.owner);
  return response;
}

/**
 * Get current authenticated owner
 * Attempts to refresh token if no access token is available but refresh token cookie exists
 * @returns {Promise<import('$lib/types').VenueOwner>}
 */
export async function getCurrentOwner() {
  const response = await getAuthMe();
  return response.owner;
}

/**
 * Permanently delete the current account and all associated data (venues, event lists, events).
 * On success, clears access token and owner store. Does not call logout; account and tokens are already gone.
 * @returns {Promise<void>}
 */
export async function deleteAccount() {
  await api.delete('/api/auth/me');
  clearAccessToken();
  currentOwnerStore.set(null);
}
