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
    await api.post('/api/auth/logout');
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
 * Get current authenticated owner
 * Attempts to refresh token if no access token is available but refresh token cookie exists
 * @returns {Promise<VenueOwner>}
 */
export async function getCurrentOwner() {
  const { getAccessToken, setAccessToken } = await import('./client.js');
  
  // If no access token but we might have a refresh token cookie, try to refresh first
  if (!getAccessToken()) {
    try {
      // Try to refresh the token using the refresh token cookie
      // The refresh endpoint will use the HttpOnly cookie automatically
      const refreshResponse = await api.postJSON('/api/auth/refresh', {});
      setAccessToken(refreshResponse.access_token);
    } catch (refreshError) {
      // Refresh failed - user is not authenticated, re-throw to be handled by caller
      throw refreshError;
    }
  }
  
  const response = await api.getJSON('/api/auth/me');
  
  // Update owner data in store
  currentOwnerStore.set(response.owner);

  return response.owner;
}
