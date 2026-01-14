/**
 * Generate a UUID (v4).
 *
 * Uses `crypto.randomUUID()` when available, otherwise falls back to a
 * Math.random-based implementation (sufficient for demo/prototype usage).
 *
 * @returns {string}
 */
export function generateUUID() {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) return crypto.randomUUID();

  // Fallback for older browsers
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

