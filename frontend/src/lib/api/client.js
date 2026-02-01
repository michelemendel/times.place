/**
 * API Client
 * 
 * Provides a fetch wrapper with:
 * - JWT access token management (memory-based)
 * - Automatic token refresh on 401 responses
 * - Error handling and parsing
 * - Loading state management
 * - Network error handling
 */

// Memory-based access token storage (not localStorage for security)
/** @type {string | null} */
let accessToken = null;

// Loading state management
const loadingCallbacks = new Set();

/**
 * Add a callback to be notified of loading state changes
 * @param {Function} callback - Function called with (isLoading: boolean)
 * @returns {Function} - Unsubscribe function
 */
export function onLoadingChange(callback) {
  loadingCallbacks.add(callback);
  return () => {
    loadingCallbacks.delete(callback);
  };
}

/**
 * Notify all loading callbacks of state change
 * @param {boolean} isLoading - Whether a request is in progress
 */
function setLoading(isLoading) {
  loadingCallbacks.forEach((callback) => callback(isLoading));
}

/**
 * Set the access token (memory-based)
 * @param {string | null} token - JWT access token or null to clear
 */
export function setAccessToken(token) {
  accessToken = token;
}

/**
 * Get the current access token
 */
export function getAccessToken() {
  return accessToken;
}

/**
 * Clear the access token
 */
export function clearAccessToken() {
  accessToken = null;
}

/**
 * Check if we're currently refreshing the token (to prevent refresh loops)
 */
let isRefreshing = false;
/** @type {Promise<string | null> | null} */
let refreshPromise = null;

/**
 * Refresh the access token using the refresh token cookie
 */
async function refreshAccessToken() {
  // If already refreshing, return the existing promise
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      const response = await fetch('/api/auth/refresh', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include', // Include cookies (refresh token)
      });

      if (!response.ok) {
        // Refresh failed - clear token and throw
        accessToken = null;
        throw new Error('Token refresh failed');
      }

      const data = await response.json();
      accessToken = data.access_token;
      return accessToken;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

/**
 * Parse error response from API
 * @param {Response} response - Fetch response object
 */
async function parseErrorResponse(response) {
  try {
    const data = await response.json();
    if (data.error && data.error.message) {
      return {
        code: data.error.code || 'unknown',
        message: data.error.message,
      };
    }
    return {
      code: 'unknown',
      message: data.message || `HTTP ${response.status}: ${response.statusText}`,
    };
  } catch {
    return {
      code: 'unknown',
      message: `HTTP ${response.status}: ${response.statusText}`,
    };
  }
}

/**
 * API Client class
 */
class ApiClient {
  /**
   * Make an API request
   * @param {string} endpoint - API endpoint (e.g., '/api/venues')
   * @param {RequestInit} options - Fetch options
   * @returns {Promise<Response>}
   */
  async request(endpoint, options = {}) {
    const url = endpoint.startsWith('/api/') ? endpoint : `/api/${endpoint}`;
    
    // Set loading state
    setLoading(true);

    try {
      // Prepare headers: normalize to plain object so type assertion is valid, then allow adding Authorization
      const baseHeaders = options.headers instanceof Headers
        ? Object.fromEntries(options.headers.entries())
        : Array.isArray(options.headers)
          ? Object.fromEntries(options.headers)
          : (options.headers ?? {});
      /** @type {Record<string, string>} */
      const headers = {
        'Content-Type': 'application/json',
        ...baseHeaders,
      };

      // Add access token if available
      if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
      }

      // Make request
      let response = await fetch(url, {
        ...options,
        headers,
        credentials: 'include', // Always include cookies for refresh token
      });

      // Handle 401 Unauthorized - try to refresh token
      if (response.status === 401 && accessToken) {
        try {
          // Try to refresh the token
          await refreshAccessToken();
          
          // Retry the original request with new token
          if (accessToken) {
            headers['Authorization'] = `Bearer ${accessToken}`;
            response = await fetch(url, {
              ...options,
              headers,
              credentials: 'include',
            });
          }
        } catch (refreshError) {
          // Refresh failed - clear token
          accessToken = null;
          // Return the original 401 response
        }
      }

      // Handle errors
      if (!response.ok) {
        const error = await parseErrorResponse(response);
        throw new ApiError(error.message, response.status, error.code);
      }

      return response;
    } catch (error) {
      // Check for offline state first
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        throw new ApiError('You are currently offline. Please check your internet connection.', 0, 'offline');
      }
      
      // Handle network errors
      if (error instanceof TypeError && error.message === 'Failed to fetch') {
        throw new ApiError('Network error. Please check your connection.', 0, 'network_error');
      }
      
      // Handle timeout errors (AbortError from AbortController)
      if (error instanceof Error && (error.name === 'AbortError' || error.message.includes('timeout'))) {
        throw new ApiError('Request timed out. Please try again.', 0, 'timeout');
      }
      
      // Re-throw ApiError as-is
      if (error instanceof ApiError) {
        throw error;
      }
      // Wrap other errors
      const message = error instanceof Error ? error.message : 'An unexpected error occurred';
      throw new ApiError(message, 0, 'unknown');
    } finally {
      setLoading(false);
    }
  }

  /**
   * GET request
   * @param {string} endpoint - API endpoint
   * @param {RequestInit} [options] - Fetch options
   */
  async get(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'GET' });
  }

  /**
   * POST request
   * @param {string} endpoint - API endpoint
   * @param {object} body - Request body (JSON-serialized)
   * @param {RequestInit} [options] - Fetch options
   */
  async post(endpoint, body, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'POST',
      body: JSON.stringify(body),
    });
  }

  /**
   * PATCH request
   * @param {string} endpoint - API endpoint
   * @param {object} body - Request body (JSON-serialized)
   * @param {RequestInit} [options] - Fetch options
   */
  async patch(endpoint, body, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'PATCH',
      body: JSON.stringify(body),
    });
  }

  /**
   * DELETE request
   * @param {string} endpoint - API endpoint
   * @param {RequestInit} [options] - Fetch options
   */
  async delete(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'DELETE' });
  }

  /**
   * GET request and parse JSON
   * @param {string} endpoint - API endpoint
   * @param {RequestInit} [options] - Fetch options
   */
  async getJSON(endpoint, options = {}) {
    const response = await this.get(endpoint, options);
    return response.json();
  }

  /**
   * POST request and parse JSON
   * @param {string} endpoint - API endpoint
   * @param {object} body - Request body (JSON-serialized)
   * @param {RequestInit} [options] - Fetch options
   */
  async postJSON(endpoint, body, options = {}) {
    const response = await this.post(endpoint, body, options);
    return response.json();
  }

  /**
   * PATCH request and parse JSON
   * @param {string} endpoint - API endpoint
   * @param {object} body - Request body (JSON-serialized)
   * @param {RequestInit} [options] - Fetch options
   */
  async patchJSON(endpoint, body, options = {}) {
    const response = await this.patch(endpoint, body, options);
    return response.json();
  }
}

/**
 * Custom API Error class
 */
export class ApiError extends Error {
  /**
   * @param {string} message - Error message
   * @param {number} [status=0] - HTTP status code
   * @param {string} [code='unknown'] - Error code
   */
  constructor(message, status = 0, code = 'unknown') {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

// Export singleton instance
export const api = new ApiClient();
