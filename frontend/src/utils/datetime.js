/**
 * Date and time utility functions for time.place
 * 
 * Handles timezone-aware formatting of dates and times based on user's browser settings.
 * All functions automatically use the user's locale and timezone.
 * 
 * Note: Locale and timezone are different:
 * - Locale (e.g., "en-US", "he-IL"): Controls language/cultural formatting (month names, day names, number formats)
 * - Timezone (e.g., "America/New_York", "Asia/Jerusalem"): Controls what time it actually is (geographic location)
 * These are independent - you can have Hebrew locale with New York timezone, or English locale with Jerusalem timezone.
 */

import config from '../config.json';

/**
 * Get the user's timezone from browser settings
 * @returns {string} IANA timezone name (e.g., "America/New_York", "Asia/Jerusalem")
 */
export function getUserTimezone() {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

/**
 * Get the user's locale from browser settings
 * @returns {string} Locale string (e.g., "en-US", "he-IL")
 */
export function getUserLocale() {
  return navigator.language || navigator.languages?.[0] || 'en-US';
}

/**
 * Get the default hour format from configuration
 * @returns {boolean} true for 12-hour format (AM/PM), false for 24-hour format
 */
function getDefaultHour12() {
  return config.hour12 ?? false;
}

/**
 * Format an event time (Unix epoch timestamp) to display only the time portion
 * The time is adjusted for the user's timezone automatically
 * 
 * @param {number} unixTimestamp - Unix epoch timestamp in seconds
 * @param {object} options - Optional formatting options
 * @param {boolean} options.hour12 - Use 12-hour format (default: from config file)
 * @param {string} options.locale - Override locale (default: user's browser locale)
 * @returns {string} Formatted time string (e.g., "14:30" for 24-hour or "2:30 PM" for 12-hour)
 */
export function formatEventTime(unixTimestamp, options = {}) {
  const {
    hour12 = getDefaultHour12(),
    locale = getUserLocale()
  } = options;

  const date = new Date(unixTimestamp * 1000); // Convert seconds to milliseconds
  const timeZone = getUserTimezone();

  return new Intl.DateTimeFormat(locale, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: hour12,
    timeZone: timeZone
  }).format(date);
}

/**
 * Format an event list date (ISO 8601 date string) for display
 * 
 * @param {string} isoDateString - ISO 8601 date string (e.g., "2024-12-25")
 * @param {object} options - Optional formatting options
 * @param {string} options.locale - Override locale (default: user's browser locale)
 * @returns {string} Formatted date string (e.g., "December 25, 2024" or "25/12/2024")
 */
export function formatEventListDate(isoDateString, options = {}) {
  const { locale = getUserLocale() } = options;

  const date = new Date(isoDateString + 'T00:00:00'); // Parse ISO date string
  const timeZone = getUserTimezone();

  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    timeZone: timeZone
  }).format(date);
}

/**
 * Format a full date and time from Unix epoch timestamp
 * Useful for internal operations, debugging, or admin views
 * 
 * @param {number} unixTimestamp - Unix epoch timestamp in seconds
 * @param {object} options - Optional formatting options
 * @param {string} options.locale - Override locale (default: user's browser locale)
 * @returns {string} Formatted date and time string in ISO 8601 format adjusted for timezone
 */
export function formatFullDateTime(unixTimestamp, options = {}) {
  const { locale = getUserLocale() } = options;
  const date = new Date(unixTimestamp * 1000);
  const timeZone = getUserTimezone();

  // Format as ISO 8601 with timezone offset
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    timeZone: timeZone,
    timeZoneName: 'short'
  }).format(date);
}

/**
 * Validate and parse an ISO 8601 date string
 * 
 * @param {string} isoDateString - ISO 8601 date string (e.g., "2024-12-25")
 * @returns {Date|null} Parsed Date object or null if invalid
 */
export function parseISODate(isoDateString) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(isoDateString)) {
    return null;
  }
  const date = new Date(isoDateString + 'T00:00:00');
  return isNaN(date.getTime()) ? null : date;
}

/**
 * Convert a time input (hours, minutes) combined with an event list date to Unix epoch timestamp
 * 
 * @param {string} eventListDate - ISO 8601 date string from event list (e.g., "2024-12-25")
 * @param {number} hours - Hours (0-23)
 * @param {number} minutes - Minutes (0-59)
 * @returns {number} Unix epoch timestamp in seconds
 */
export function createEventTimestamp(eventListDate, hours, minutes) {
  const date = new Date(`${eventListDate}T${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:00`);
  return Math.floor(date.getTime() / 1000); // Convert to seconds
}

