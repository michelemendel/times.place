<script>
  import { onMount, afterUpdate, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { get } from 'svelte/store';
  import { goto } from '$app/navigation';
  import { currentOwnerStore } from '$lib/stores';
  import { getVenue, createVenue, updateVenue } from '$lib/api/venues.js';
  import {
    listEventListsForVenue,
    createEventList,
    updateEventList,
    deleteEventList as deleteEventListApi,
  } from '$lib/api/eventLists.js';
  import {
    listEventsForEventList,
    createEvent,
    updateEvent,
    deleteEvent as deleteEventApi,
  } from '$lib/api/events.js';
  import { parseISODate } from '$lib/utils/datetime.js';
  import { formatEventTime } from '$lib/utils/datetime.js';
  import { ApiError } from '$lib/api/client.js';
  import BannerImage from '$lib/BannerImage.svelte';

  // Common timezones organized by region
  /** @type {any} */
  const timezones = [
    { value: '', label: 'No timezone' },
    {
      group: 'Americas',
      zones: [
        { value: 'America/New_York', label: 'Eastern Time (New York)' },
        { value: 'America/Chicago', label: 'Central Time (Chicago)' },
        { value: 'America/Denver', label: 'Mountain Time (Denver)' },
        { value: 'America/Los_Angeles', label: 'Pacific Time (Los Angeles)' },
        { value: 'America/Toronto', label: 'Eastern Time (Toronto)' },
        { value: 'America/Vancouver', label: 'Pacific Time (Vancouver)' },
        { value: 'America/Mexico_City', label: 'Central Time (Mexico City)' },
        { value: 'America/Sao_Paulo', label: 'Brasília Time (São Paulo)' },
        {
          value: 'America/Buenos_Aires',
          label: 'Argentina Time (Buenos Aires)',
        },
      ],
    },
    {
      group: 'Europe',
      zones: [
        { value: 'Europe/London', label: 'Greenwich Mean Time (London)' },
        { value: 'Europe/Paris', label: 'Central European Time (Paris)' },
        { value: 'Europe/Berlin', label: 'Central European Time (Berlin)' },
        { value: 'Europe/Rome', label: 'Central European Time (Rome)' },
        { value: 'Europe/Madrid', label: 'Central European Time (Madrid)' },
        {
          value: 'Europe/Amsterdam',
          label: 'Central European Time (Amsterdam)',
        },
        { value: 'Europe/Athens', label: 'Eastern European Time (Athens)' },
        { value: 'Europe/Moscow', label: 'Moscow Time' },
      ],
    },
    {
      group: 'Asia & Middle East',
      zones: [
        { value: 'Asia/Jerusalem', label: 'Israel Time (Jerusalem)' },
        { value: 'Asia/Dubai', label: 'Gulf Standard Time (Dubai)' },
        { value: 'Asia/Kuwait', label: 'Arabia Standard Time (Kuwait)' },
        { value: 'Asia/Tehran', label: 'Iran Standard Time (Tehran)' },
        { value: 'Asia/Kolkata', label: 'India Standard Time (Mumbai, Delhi)' },
        { value: 'Asia/Dhaka', label: 'Bangladesh Standard Time (Dhaka)' },
        { value: 'Asia/Bangkok', label: 'Indochina Time (Bangkok)' },
        { value: 'Asia/Singapore', label: 'Singapore Time' },
        { value: 'Asia/Hong_Kong', label: 'Hong Kong Time' },
        { value: 'Asia/Shanghai', label: 'China Standard Time (Shanghai)' },
        { value: 'Asia/Tokyo', label: 'Japan Standard Time (Tokyo)' },
        { value: 'Asia/Seoul', label: 'Korea Standard Time (Seoul)' },
      ],
    },
    {
      group: 'Oceania',
      zones: [
        {
          value: 'Australia/Sydney',
          label: 'Australian Eastern Time (Sydney)',
        },
        {
          value: 'Australia/Melbourne',
          label: 'Australian Eastern Time (Melbourne)',
        },
        {
          value: 'Australia/Brisbane',
          label: 'Australian Eastern Time (Brisbane)',
        },
        { value: 'Australia/Perth', label: 'Australian Western Time (Perth)' },
        { value: 'Pacific/Auckland', label: 'New Zealand Time (Auckland)' },
      ],
    },
    {
      group: 'Africa',
      zones: [
        { value: 'Africa/Cairo', label: 'Eastern European Time (Cairo)' },
        {
          value: 'Africa/Johannesburg',
          label: 'South Africa Standard Time (Johannesburg)',
        },
        { value: 'Africa/Lagos', label: 'West Africa Time (Lagos)' },
      ],
    },
  ];

  // Get venue UUID from URL params (only in browser to avoid prerendering errors)
  $: venueUuidFromUrl =
    browser && $page.url.searchParams
      ? $page.url.searchParams.get('venue_uuid')
      : null;

  /** @type {any} */
  let currentOwner = null;
  /** @type {any} */
  let venue = null;
  let isNewVenue = false;

  // Loading and error states
  let isLoading = false;
  let isSaving = false;
  let loadError = '';
  let saveError = '';

  // Form state
  let venueName = '';
  let venueAddress = '';
  let venueGeolocation = '';
  let venueComment = '';
  let venueBannerImage = '';
  let venueTimezone = '';

  // Event lists state (array of event list objects being edited)
  /** @type {any[]} */
  let eventListsData = [];

  // Undo stack
  /** @type {any[]} */
  let undoStack = [];
  const MAX_UNDO_STEPS = 50;

  // Preview pane state
  /** @type {string | null} */
  let previewEventListId = null;
  /** @type {string | null} */
  let copiedShareLinkEventListUuid = null;

  // Track if venue data has been loaded
  let dataLoaded = false;

  // Map state for geolocation picker
  /** @type {HTMLElement | null} */
  let geolocationMapContainer = null;
  /** @type {any} */
  let geolocationMap = null;
  /** @type {any} */
  let geolocationMarker = null;
  let leafletLoaded = false;
  let isGeocoding = false;
  /** @type {any[]} */
  let geocodingResults = [];
  let showGeocodingResults = false;

  // Timezone help popup state
  let showTimezoneHelp = false;
  /** @type {HTMLElement | null} */
  let timezoneHelpButton = null;
  /** @type {HTMLElement | null} */
  let timezoneHelpPopup = null;

  // Field-level validation (show on blur and on save failure)
  let venueNameError = '';
  /** @type {Record<string, string>} */
  let eventNameErrors = {};

  // Subscribe to owner store only (for auth check)
  currentOwnerStore.subscribe((val) => {
    currentOwner = val;
  });

  /** Validate venue name on blur and set field error */
  function validateVenueNameBlur() {
    venueNameError = !venueName.trim() ? 'Venue name is required' : '';
  }

  /** Validate event name on blur and set field error for that event
   * @param {string} eventUuid
   */
  function validateEventNameBlur(eventUuid) {
    const list = eventListsData.find((el) =>
      (el.events || []).some(
        (/** @type {any} */ e) => e.event_uuid === eventUuid,
      ),
    );
    const event = list?.events?.find(
      (/** @type {any} */ e) => e.event_uuid === eventUuid,
    );
    const msg = !event?.event_name?.trim() ? 'Event name is required' : '';
    eventNameErrors = { ...eventNameErrors, [eventUuid]: msg };
  }

  /** Set all required-field errors (e.g. after failed save) so user sees which fields are missing */
  function setRequiredFieldErrors() {
    venueNameError = !venueName.trim() ? 'Venue name is required' : '';
    const next = { ...eventNameErrors };
    for (const list of eventListsData) {
      for (const event of list.events || []) {
        const msg = !event.event_name?.trim() ? 'Event name is required' : '';
        if (msg) next[event.event_uuid] = msg;
        else delete next[event.event_uuid];
      }
    }
    eventNameErrors = next;
  }

  /**
   * Save current state to undo stack
   */
  function saveUndoState() {
    const state = {
      venue: venue ? JSON.parse(JSON.stringify(venue)) : null,
      eventListsData: JSON.parse(JSON.stringify(eventListsData)),
    };
    undoStack.push(state);
    if (undoStack.length > MAX_UNDO_STEPS) {
      undoStack.shift();
    }
  }

  /**
   * Undo last change
   */
  function handleUndo() {
    if (undoStack.length === 0) return;
    const previousState = undoStack.pop();
    if (previousState.venue) {
      venue = previousState.venue;
      loadVenueFormFields();
      // Reload event lists and events from API if venue exists
      if (venue.venue_uuid && !venue.venue_uuid.startsWith('temp-')) {
        loadVenueDataFromAPI();
      } else {
        eventListsData = previousState.eventListsData || [];
      }
    } else {
      eventListsData = previousState.eventListsData || [];
    }
  }

  /**
   * Validate UUID format
   * @param {string} uuid
   */
  function isValidUUID(uuid) {
    const uuidRegex =
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    return uuidRegex.test(uuid);
  }

  /**
   * Sanitize input to prevent XSS
   * @param {any} input
   */
  function sanitizeInput(input) {
    if (typeof input !== 'string') return input;
    const div = document.createElement('div');
    div.textContent = input;
    return div.innerHTML;
  }

  /**
   * Load venue data from API into form
   */
  async function loadVenueDataFromAPI() {
    if (!venue || !venue.venue_uuid) return;

    isLoading = true;
    loadError = '';

    try {
      // Load event lists for this venue
      const venueEventLists = await listEventListsForVenue(venue.venue_uuid);

      // Load events for each event list in parallel
      const eventListsWithEvents = await Promise.all(
        venueEventLists.map(async (el) => {
          try {
            const events = await listEventsForEventList(el.event_list_uuid);
            return { eventList: el, events };
          } catch (err) {
            console.error(
              'Failed to load events for event list',
              el.event_list_uuid,
              err,
            );
            return { eventList: el, events: [] };
          }
        }),
      );

      // Transform to eventListsData format
      eventListsData = eventListsWithEvents.map(({ eventList: el, events }) => {
        const eventListDate =
          el.date !== undefined && el.date !== null ? el.date : '';
        const eventsWithTime = events.map((e) => ({
          ...e,
          time: extractTimeFromRFC3339(e.datetime),
          duration_minutes:
            e.duration_minutes && e.duration_minutes > 0
              ? e.duration_minutes
              : '',
        }));
        return {
          ...el,
          date: eventListDate,
          visibility: el.visibility || 'private',
          event_uuids: events.map((e) => e.event_uuid),
          events: eventsWithTime,
        };
      });

      // Set preview to first event list
      if (eventListsData.length > 0) {
        previewEventListId = eventListsData[0].event_list_uuid;
      }
    } catch (err) {
      console.error('Failed to load venue data', err);
      loadError = 'Failed to load venue data. Please try again.';
      if (err instanceof ApiError && err.status === 404) {
        loadError = 'Venue not found.';
        goto('/venue-owner');
      }
    } finally {
      isLoading = false;
    }
  }

  /**
   * Load venue form fields from venue object (after API load)
   */
  function loadVenueFormFields() {
    if (!venue) return;
    venueName = venue.name || '';
    venueAddress = venue.address || '';
    venueGeolocation = venue.geolocation || '';
    venueComment = venue.comment || '';
    venueBannerImage = venue.banner_image || '';
    venueTimezone =
      venue.timezone !== undefined && venue.timezone !== null
        ? venue.timezone
        : '';
    venueNameError = '';
    eventNameErrors = {};
  }

  /**
   * Extract time (HH:MM) from RFC3339 datetime string
   * @param {string} rfc3339
   */
  function extractTimeFromRFC3339(rfc3339) {
    try {
      const date = new Date(rfc3339);
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      return `${hours}:${minutes}`;
    } catch (e) {
      return '00:00';
    }
  }

  /**
   * Convert time string (HH:MM) and event list date to RFC3339 datetime
   * If venueTimezone is provided, interprets the time as being in that timezone
   * Otherwise, interprets the time as being in the browser's local timezone
   * @param {string} timeStr
   * @param {string} dateStr
   * @param {string} [venueTimezone] - Optional venue timezone (IANA timezone string)
   */
  function combineTimeAndDate(timeStr, dateStr, venueTimezone) {
    try {
      const [hours, minutes] = timeStr.split(':').map(Number);
      const dateTimeStr = `${dateStr}T${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:00`;

      if (venueTimezone && venueTimezone.trim()) {
        // If venue timezone is set, we need to create a UTC timestamp that,
        // when displayed in the venue timezone, shows the desired wall-clock time
        // We do this by iteratively adjusting until we get the right display time

        // Start with a guess: create date as if it's in UTC
        let testDate = new Date(dateTimeStr + 'Z');
        const maxIterations = 10;
        let iterations = 0;

        while (iterations < maxIterations) {
          // Format this UTC time in the venue timezone
          const formatter = new Intl.DateTimeFormat('en-US', {
            timeZone: venueTimezone.trim(),
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            hour12: false,
          });

          const parts = formatter.formatToParts(testDate);
          const hourPart = parts.find((p) => p.type === 'hour');
          const minutePart = parts.find((p) => p.type === 'minute');
          if (!hourPart || !minutePart) {
            // If we can't parse the parts, break the loop
            break;
          }
          const venueHour = parseInt(hourPart?.value || '0');
          const venueMinute = parseInt(minutePart?.value || '0');

          // Check if we have the right time
          if (venueHour === hours && venueMinute === minutes) {
            return testDate.toISOString();
          }

          // Calculate the difference in minutes
          const desiredMinutes = hours * 60 + minutes;
          const currentMinutes = venueHour * 60 + venueMinute;
          const diffMinutes = desiredMinutes - currentMinutes;

          // Adjust the UTC time
          testDate = new Date(testDate.getTime() + diffMinutes * 60 * 1000);
          iterations++;
        }

        // If we couldn't converge, return the best guess
        return testDate.toISOString();
      } else {
        // No venue timezone - interpret as local time (browser timezone)
        const date = new Date(dateTimeStr);
        return date.toISOString();
      }
    } catch (e) {
      return new Date().toISOString();
    }
  }

  /**
   * Add new event list
   */
  function addEventList() {
    if (!currentOwner || !venue?.venue_uuid) return;
    saveUndoState();
    // For new venues, we'll create a temporary local-only event list
    // UUID will be assigned by backend when venue is saved
    const tempId = `temp-${Date.now()}-${Math.random()}`;
    const newList = {
      event_list_uuid: tempId,
      venue_uuid: venue.venue_uuid,
      name: '',
      date: '',
      comment: '',
      visibility: 'public', // Default to public so venue appears in main page dropdown
      event_uuids: [],
      events: [],
      sort_order: eventListsData.length,
      isNew: true, // Mark as new so we know to create via API on save
    };
    eventListsData = [...eventListsData, newList];
    if (!previewEventListId) {
      previewEventListId = newList.event_list_uuid;
    }
  }

  /**
   * Delete event list
   * @param {string} listUuid
   */
  function deleteEventList(listUuid) {
    const list = eventListsData.find((el) => el.event_list_uuid === listUuid);
    if (!list) return;
    if (
      !confirm(
        `Are you sure you want to delete the event list "${list.name || 'Untitled Event List'}"? This will also delete all events in this list.`,
      )
    ) {
      return;
    }
    saveUndoState();
    eventListsData = eventListsData.filter(
      (el) => el.event_list_uuid !== listUuid,
    );
    if (previewEventListId === listUuid) {
      previewEventListId =
        eventListsData.length > 0 ? eventListsData[0].event_list_uuid : null;
    }
  }

  /**
   * Move event list up
   * @param {number} index
   */
  function moveEventListUp(index) {
    if (index === 0) return;
    saveUndoState();
    const newData = [...eventListsData];
    const temp = newData[index];
    newData[index] = { ...newData[index - 1], sort_order: index };
    newData[index - 1] = { ...temp, sort_order: index - 1 };
    eventListsData = newData;
  }

  /**
   * Move event list down
   * @param {number} index
   */
  function moveEventListDown(index) {
    if (index === eventListsData.length - 1) return;
    saveUndoState();
    const newData = [...eventListsData];
    const temp = newData[index];
    newData[index] = { ...newData[index + 1], sort_order: index };
    newData[index + 1] = { ...temp, sort_order: index + 1 };
    eventListsData = newData;
  }

  /**
   * Clear date for an event list
   * @param {string} listUuid
   */
  function clearEventListDate(listUuid) {
    saveUndoState();
    const listIndex = eventListsData.findIndex(
      (el) => el.event_list_uuid === listUuid,
    );
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];

    // Clear the date
    const updatedList = {
      ...list,
      date: '',
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1),
    ];
  }

  /**
   * Clear duration for an event
   * @param {string} listUuid
   * @param {string} eventUuid
   */
  function clearEventDuration(listUuid, eventUuid) {
    saveUndoState();
    const listIndex = eventListsData.findIndex(
      (el) => el.event_list_uuid === listUuid,
    );
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];
    const eventIndex = list.events.findIndex(
      (/** @type {any} */ e) => e.event_uuid === eventUuid,
    );
    if (eventIndex === -1) return;

    // Create new events array with cleared duration
    const newEvents = [...list.events];
    newEvents[eventIndex] = {
      ...newEvents[eventIndex],
      duration_minutes: '',
    };

    // Create new list object with updated events
    const updatedList = {
      ...list,
      events: newEvents,
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1),
    ];
  }

  /**
   * Add event to event list
   * @param {string} listUuid
   */
  function addEvent(listUuid) {
    saveUndoState();
    const listIndex = eventListsData.findIndex(
      (el) => el.event_list_uuid === listUuid,
    );
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];

    // For new events without a date, use today's date as placeholder for datetime
    // The time field will be preserved separately
    const placeholderDate =
      list.date && list.date.trim()
        ? list.date
        : new Date().toISOString().split('T')[0]; // Today's date in YYYY-MM-DD format

    const tempEventId = `temp-event-${Date.now()}-${Math.random()}`;
    const newEvent = {
      event_uuid: tempEventId,
      event_list_uuid: listUuid,
      event_name: '',
      datetime: combineTimeAndDate('12:00', placeholderDate, venueTimezone),
      time: '12:00',
      comment: '',
      duration_minutes: '',
      sort_order: list.events.length,
      isNew: true, // Mark as new so we know to create via API on save
    };

    // Create new arrays with the new event added
    const newEvents = [...list.events, newEvent];
    const newEventUuids = [...(list.event_uuids || []), newEvent.event_uuid];

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids,
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1),
    ];
  }

  /**
   * Delete event
   * @param {string} listUuid
   * @param {string} eventUuid
   */
  function deleteEvent(listUuid, eventUuid) {
    saveUndoState();
    const listIndex = eventListsData.findIndex(
      (el) => el.event_list_uuid === listUuid,
    );
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];

    // Create new arrays with the event removed
    const newEvents = list.events.filter(
      (/** @type {any} */ e) => e.event_uuid !== eventUuid,
    );
    const newEventUuids = (list.event_uuids || []).filter(
      (/** @type {any} */ uuid) => uuid !== eventUuid,
    );

    // Clear validation error for deleted event
    eventNameErrors = { ...eventNameErrors };
    delete eventNameErrors[eventUuid];
    eventNameErrors = eventNameErrors;

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids,
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1),
    ];
  }

  /**
   * Duplicate event
   * @param {string} listUuid
   * @param {string} eventUuid
   */
  function duplicateEvent(listUuid, eventUuid) {
    saveUndoState();
    const listIndex = eventListsData.findIndex(
      (el) => el.event_list_uuid === listUuid,
    );
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];
    const originalEvent = list.events.find(
      (/** @type {any} */ e) => e.event_uuid === eventUuid,
    );
    if (!originalEvent) return;

    const tempEventId = `temp-event-${Date.now()}-${Math.random()}`;
    const newEvent = {
      ...originalEvent,
      event_uuid: tempEventId,
      event_name: '',
      sort_order:
        originalEvent.sort_order !== undefined
          ? originalEvent.sort_order + 1
          : list.events.length,
      isNew: true, // Mark as new so we know to create via API on save
    };

    const index = list.events.findIndex(
      (/** @type {any} */ e) => e.event_uuid === eventUuid,
    );

    // Create new arrays with the duplicated event inserted
    const newEvents = [...list.events];
    newEvents.splice(index + 1, 0, newEvent);
    const newEventUuids = [...(list.event_uuids || [])];
    newEventUuids.splice(index + 1, 0, newEvent.event_uuid);

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids,
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1),
    ];
  }

  /**
   * Move event up
   * @param {string} listUuid
   * @param {number} eventIndex
   */
  function moveEventUp(listUuid, eventIndex) {
    if (eventIndex === 0) return;
    saveUndoState();
    const listIndex = eventListsData.findIndex(
      (el) => el.event_list_uuid === listUuid,
    );
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];

    // Create new arrays with swapped elements
    const newEvents = [...list.events];
    const temp = newEvents[eventIndex];
    newEvents[eventIndex] = {
      ...newEvents[eventIndex - 1],
      sort_order: eventIndex,
    };
    newEvents[eventIndex - 1] = { ...temp, sort_order: eventIndex - 1 };

    const newEventUuids = [...(list.event_uuids || [])];
    const tempUuid = newEventUuids[eventIndex];
    newEventUuids[eventIndex] = newEventUuids[eventIndex - 1];
    newEventUuids[eventIndex - 1] = tempUuid;

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids,
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1),
    ];
  }

  /**
   * Move event down
   * @param {string} listUuid
   * @param {number} eventIndex
   */
  function moveEventDown(listUuid, eventIndex) {
    const listIndex = eventListsData.findIndex(
      (el) => el.event_list_uuid === listUuid,
    );
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];
    if (eventIndex === list.events.length - 1) return;
    saveUndoState();

    // Create new arrays with swapped elements
    const newEvents = [...list.events];
    const temp = newEvents[eventIndex];
    newEvents[eventIndex] = {
      ...newEvents[eventIndex + 1],
      sort_order: eventIndex,
    };
    newEvents[eventIndex + 1] = { ...temp, sort_order: eventIndex + 1 };

    const newEventUuids = [...(list.event_uuids || [])];
    const tempUuid = newEventUuids[eventIndex];
    newEventUuids[eventIndex] = newEventUuids[eventIndex + 1];
    newEventUuids[eventIndex + 1] = tempUuid;

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids,
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1),
    ];
  }

  /**
   * Handle image upload
   * @param {Event} event
   */
  function handleImageUpload(event) {
    const target = /** @type {HTMLInputElement} */ (event.target);
    const file = target.files?.[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      alert('Image size must be less than 5MB');
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      const result = e.target?.result;
      if (typeof result === 'string') {
        venueBannerImage = result;
      }
    };
    reader.readAsDataURL(file);
  }

  /**
   * Handle event list visibility change
   * @param {string} listUuid
   * @param {string} newVisibility
   */
  function handleEventListVisibilityChange(listUuid, newVisibility) {
    saveUndoState();
    const listIndex = eventListsData.findIndex(
      (el) => el.event_list_uuid === listUuid,
    );
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];

    // Preserve existing token - backend will generate if needed when saving
    // Don't set to undefined when switching to public - keep the token for future use
    const updatedList = {
      ...list,
      visibility: newVisibility,
      private_link_token: list.private_link_token, // Preserve token, backend generates if missing on save
    };

    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1),
    ];
  }

  /**
   * Copy share link for an event list to clipboard (no need to select text).
   * @param {{ event_list_uuid: string, private_link_token?: string | null }} listData
   */
  function copyShareLink(listData) {
    if (!listData?.private_link_token || typeof window === 'undefined') return;
    const url = `${window.location.origin}/?token=${listData.private_link_token}`;
    navigator.clipboard
      .writeText(url)
      .then(() => {
        copiedShareLinkEventListUuid = listData.event_list_uuid;
        setTimeout(() => {
          copiedShareLinkEventListUuid = null;
        }, 2000);
      })
      .catch(() => {});
  }

  /**
   * Save venue and all related data via API
   */
  async function saveVenue() {
    if (!currentOwner) {
      saveError = 'You must be logged in to save a venue';
      return;
    }

    // Validate required fields and show which fields are missing
    if (!venueName.trim()) {
      setRequiredFieldErrors();
      saveError = 'Please fix the required fields below.';
      return;
    }

    // Validate event names (required)
    for (const list of eventListsData) {
      for (const event of list.events) {
        if (!event.event_name || !event.event_name.trim()) {
          setRequiredFieldErrors();
          saveError = 'Please fix the required fields below.';
          return;
        }
      }
    }

    // Validate event list dates (optional, but if provided must be valid)
    for (const list of eventListsData) {
      const dateValue = list.date ? String(list.date).trim() : '';
      if (dateValue && !/^\d{4}-\d{2}-\d{2}$/.test(dateValue)) {
        saveError = `Invalid date format for event list "${list.name}". Please use ISO 8601 format (YYYY-MM-DD).`;
        return;
      }
      if (dateValue) {
        const parsedDate = parseISODate(dateValue);
        if (!parsedDate) {
          saveError = `Invalid date for event list "${list.name}". Please use a valid date in ISO 8601 format (YYYY-MM-DD).`;
          return;
        }
      }
    }

    // Validate event times and update datetimes
    for (const list of eventListsData) {
      for (const event of list.events) {
        const timeRegex = /^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/;
        if (!timeRegex.test(event.time)) {
          saveError = `Invalid time format for event "${event.event_name?.trim() || 'this event'}". Please use HH:MM format.`;
          return;
        }
        const dateToUse =
          list.date && list.date.trim()
            ? list.date
            : new Date().toISOString().split('T')[0];
        event.datetime = combineTimeAndDate(
          event.time,
          dateToUse,
          venueTimezone,
        );
      }
    }

    isSaving = true;
    saveError = '';

    try {
      // Step 1: Create or update venue
      if (isNewVenue) {
        venue = await createVenue({
          name: sanitizeInput(venueName.trim()),
          banner_image: venueBannerImage,
          address: sanitizeInput(venueAddress.trim()),
          geolocation: sanitizeInput(venueGeolocation.trim()),
          comment: sanitizeInput(venueComment.trim()),
          timezone: venueTimezone,
        });
      } else if (venue) {
        venue = await updateVenue(venue.venue_uuid, {
          name: sanitizeInput(venueName.trim()),
          banner_image: venueBannerImage,
          address: sanitizeInput(venueAddress.trim()),
          geolocation: sanitizeInput(venueGeolocation.trim()),
          comment: sanitizeInput(venueComment.trim()),
          timezone: venueTimezone,
        });
      }

      if (!venue || !venue.venue_uuid) {
        saveError = 'Failed to save venue';
        return;
      }

      // Step 2: Get existing event lists from API to determine what to delete
      const existingEventLists = await listEventListsForVenue(venue.venue_uuid);
      const existingListUuids = new Set(
        existingEventLists.map((el) => el.event_list_uuid),
      );
      const currentListUuids = new Set(
        eventListsData
          .map((el) => el.event_list_uuid)
          .filter((uuid) => !uuid.startsWith('temp-')),
      );

      // Delete event lists that were removed
      for (const existingList of existingEventLists) {
        if (!currentListUuids.has(existingList.event_list_uuid)) {
          try {
            await deleteEventListApi(existingList.event_list_uuid);
          } catch (err) {
            console.error(
              'Failed to delete event list',
              existingList.event_list_uuid,
              err,
            );
            // Continue with other operations even if delete fails
          }
        }
      }

      // Step 3: Create/update event lists and their events
      for (let listIndex = 0; listIndex < eventListsData.length; listIndex++) {
        const listData = eventListsData[listIndex];
        let eventListUuid = listData.event_list_uuid;
        let eventList;

        if (listData.isNew || eventListUuid.startsWith('temp-')) {
          // Create new event list
          eventList = await createEventList(venue.venue_uuid, {
            name: sanitizeInput(listData.name.trim()),
            date: listData.date || '',
            comment: sanitizeInput(listData.comment || ''),
            visibility: listData.visibility || 'private',
            sort_order: listIndex,
          });
          eventListUuid = eventList.event_list_uuid;

          // Update eventListsData with real UUID
          eventListsData[listIndex] = {
            ...listData,
            event_list_uuid: eventListUuid,
            event_uuids: eventList.event_uuids || [],
          };
        } else {
          // Update existing event list
          eventList = await updateEventList(eventListUuid, {
            name: sanitizeInput(listData.name.trim()),
            date: listData.date || '',
            comment: sanitizeInput(listData.comment || ''),
            visibility: listData.visibility || 'private',
            sort_order: listIndex,
          });
        }

        // Step 4: Get existing events for this list
        const existingEvents = await listEventsForEventList(eventListUuid);
        const existingEventUuids = new Set(
          existingEvents.map((/** @type {any} */ e) => e.event_uuid),
        );
        const currentEventUuids = new Set(
          listData.events
            .map((/** @type {any} */ e) => e.event_uuid)
            .filter(
              (/** @type {string} */ uuid) => !uuid.startsWith('temp-event-'),
            ),
        );

        // Delete events that were removed
        for (const existingEvent of existingEvents) {
          if (!currentEventUuids.has(existingEvent.event_uuid)) {
            try {
              await deleteEventApi(existingEvent.event_uuid);
            } catch (err) {
              console.error(
                'Failed to delete event',
                existingEvent.event_uuid,
                err,
              );
            }
          }
        }

        // Step 5: Create/update events
        for (
          let eventIndex = 0;
          eventIndex < listData.events.length;
          eventIndex++
        ) {
          const eventData = listData.events[eventIndex];
          const duration =
            eventData.duration_minutes &&
            eventData.duration_minutes !== '' &&
            Number(eventData.duration_minutes) > 0
              ? Number(eventData.duration_minutes)
              : null;

          if (
            eventData.isNew ||
            eventData.event_uuid.startsWith('temp-event-')
          ) {
            // Create new event
            await createEvent(eventListUuid, {
              event_name: sanitizeInput(eventData.event_name.trim()),
              datetime: eventData.datetime,
              comment: sanitizeInput(eventData.comment || ''),
              duration_minutes: duration,
              sort_order: eventIndex,
            });
          } else if (existingEventUuids.has(eventData.event_uuid)) {
            // Update existing event
            await updateEvent(eventData.event_uuid, {
              event_name: sanitizeInput(eventData.event_name.trim()),
              datetime: eventData.datetime,
              comment: sanitizeInput(eventData.comment || ''),
              duration_minutes: duration,
              sort_order: eventIndex,
            });
          }
        }
      }

      // Success - navigate back to venue owner page
      goto('/venue-owner');
    } catch (err) {
      console.error('Failed to save venue', err);
      if (err instanceof ApiError) {
        if (err.status === 401) {
          saveError = 'Your session has expired. Please log in again.';
          goto('/login');
        } else if (err.status === 404) {
          saveError = 'Venue not found. It may have been deleted.';
        } else {
          saveError = err.message || 'Failed to save venue. Please try again.';
        }
      } else {
        saveError = 'An unexpected error occurred. Please try again.';
      }
    } finally {
      isSaving = false;
    }
  }

  /**
   * Cancel editing
   */
  function cancelEdit() {
    if (
      confirm('Are you sure you want to cancel? Unsaved changes will be lost.')
    ) {
      goto('/venue-owner');
    }
  }

  /**
   * Parse geolocation string (format: "latitude,longitude")
   * @param {string} geolocation
   * @returns {{lat: number, lng: number} | null}
   */
  function parseGeolocation(geolocation) {
    if (!geolocation) return null;
    const parts = geolocation.split(',');
    if (parts.length !== 2) return null;
    const lat = parseFloat(parts[0].trim());
    const lng = parseFloat(parts[1].trim());
    if (isNaN(lat) || isNaN(lng)) return null;
    return { lat, lng };
  }

  /**
   * Geocode address to coordinates using OpenStreetMap Nominatim
   * @param {string} address
   * @returns {Promise<{lat: number, lng: number, display_name: string}[] | null>}
   */
  async function geocodeAddress(address) {
    if (!address || !address.trim()) return null;
    isGeocoding = true;
    try {
      const encodedAddress = encodeURIComponent(address.trim());
      // Improved query with better parameters for partial addresses:
      // - format=jsonv2: Better response format
      // - limit=5: Get multiple results to choose from
      // - addressdetails=1: Include detailed address components
      // - dedupe=1: Remove duplicates
      const response = await fetch(
        `https://nominatim.openstreetmap.org/search?format=jsonv2&q=${encodedAddress}&limit=5&addressdetails=1&dedupe=1`,
        {
          headers: {
            'User-Agent': 'times.place/1.0',
          },
        },
      );
      const data = await response.json();
      if (data && data.length > 0) {
        return data.map((/** @type {any} */ result) => ({
          lat: parseFloat(result.lat),
          lng: parseFloat(result.lon),
          display_name:
            result.display_name || result.name || 'Unknown location',
        }));
      }
      return null;
    } catch (error) {
      console.error('Geocoding error:', error);
      return null;
    } finally {
      isGeocoding = false;
    }
  }

  /**
   * Reverse geocode coordinates to address using OpenStreetMap Nominatim
   * @param {number} lat
   * @param {number} lng
   * @returns {Promise<string | null>}
   */
  async function reverseGeocode(lat, lng) {
    try {
      const response = await fetch(
        `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}&zoom=18&addressdetails=1`,
        {
          headers: {
            'User-Agent': 'times.place/1.0',
          },
        },
      );
      const data = await response.json();
      if (data && data.display_name) {
        return data.display_name;
      }
      return null;
    } catch (error) {
      console.error('Reverse geocoding error:', error);
      return null;
    }
  }

  /**
   * Initialize the geolocation map
   */
  function initGeolocationMap() {
    /** @type {any} */
    const win = window;
    const L = win.L;
    if (!geolocationMapContainer || !L || geolocationMap) return;

    // Default center (Jerusalem) or use existing geolocation
    let center = [31.7683, 35.2137]; // Jerusalem default
    let zoom = 13;

    const coords = parseGeolocation(venueGeolocation);
    if (coords) {
      center = [coords.lat, coords.lng];
      zoom = 15;
    }

    geolocationMap = L.map(geolocationMapContainer).setView(center, zoom);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      maxZoom: 19,
    }).addTo(geolocationMap);

    // Add marker if coordinates exist
    if (coords) {
      geolocationMarker = L.marker([coords.lat, coords.lng], {
        draggable: true,
      })
        .addTo(geolocationMap)
        .bindPopup('Drag to adjust location');

      geolocationMarker.on('dragend', async () => {
        const position = geolocationMarker.getLatLng();
        venueGeolocation = `${position.lat.toFixed(6)},${position.lng.toFixed(6)}`;
        // Optionally update address via reverse geocoding
        const address = await reverseGeocode(position.lat, position.lng);
        if (address) {
          venueAddress = address;
        }
      });
    }

    // Handle map clicks
    geolocationMap.on('click', async (/** @type {any} */ e) => {
      const { lat, lng } = e.latlng;
      venueGeolocation = `${lat.toFixed(6)},${lng.toFixed(6)}`;

      // Remove existing marker
      if (geolocationMarker) {
        geolocationMap.removeLayer(geolocationMarker);
      }

      // Add new marker
      geolocationMarker = L.marker([lat, lng], { draggable: true })
        .addTo(geolocationMap)
        .bindPopup('Drag to adjust location')
        .openPopup();

      // Update marker drag handler
      geolocationMarker.on('dragend', async () => {
        const position = geolocationMarker.getLatLng();
        venueGeolocation = `${position.lat.toFixed(6)},${position.lng.toFixed(6)}`;
        const address = await reverseGeocode(position.lat, position.lng);
        if (address) {
          venueAddress = address;
        }
      });

      // Optionally update address via reverse geocoding
      const address = await reverseGeocode(lat, lng);
      if (address) {
        venueAddress = address;
      }
    });
  }

  /**
   * Update geolocation map when coordinates change
   */
  function updateGeolocationMap() {
    /** @type {any} */
    const win = window;
    const L = win.L;
    if (!geolocationMap || !L) return;
    const coords = parseGeolocation(venueGeolocation);
    if (!coords) return;

    if (geolocationMarker) {
      geolocationMarker.setLatLng([coords.lat, coords.lng]);
      geolocationMap.setView([coords.lat, coords.lng], 15);
    } else {
      geolocationMarker = L.marker([coords.lat, coords.lng], {
        draggable: true,
      })
        .addTo(geolocationMap)
        .bindPopup('Drag to adjust location');

      geolocationMarker.on('dragend', async () => {
        const position = geolocationMarker.getLatLng();
        venueGeolocation = `${position.lat.toFixed(6)},${position.lng.toFixed(6)}`;
        const address = await reverseGeocode(position.lat, position.lng);
        if (address) {
          venueAddress = address;
        }
      });

      geolocationMap.setView([coords.lat, coords.lng], 15);
    }
  }

  /**
   * Handle "Find on Map" button click
   */
  async function handleFindOnMap() {
    if (!venueAddress.trim()) {
      alert('Please enter an address');
      return;
    }

    const results = await geocodeAddress(venueAddress);
    if (!results || results.length === 0) {
      alert(
        'Address not found. Please try a different address or click on the map to set location.',
      );
      return;
    }

    // If only one result, use it directly
    if (results.length === 1) {
      const result = results[0];
      venueGeolocation = `${result.lat.toFixed(6)},${result.lng.toFixed(6)}`;
      venueAddress = result.display_name;
      updateGeolocationMap();
      showGeocodingResults = false;
    } else {
      // Multiple results - show them for user to choose
      geocodingResults = results;
      showGeocodingResults = true;
    }
  }

  /**
   * Select a geocoding result
   * @param {{lat: number, lng: number, display_name: string}} result
   */
  function selectGeocodingResult(result) {
    venueGeolocation = `${result.lat.toFixed(6)},${result.lng.toFixed(6)}`;
    venueAddress = result.display_name;
    updateGeolocationMap();
    showGeocodingResults = false;
    geocodingResults = [];
  }

  /**
   * Handle click outside timezone help popup
   * @param {Event} event
   */
  function handleTimezoneHelpClickOutside(event) {
    if (
      showTimezoneHelp &&
      timezoneHelpButton &&
      timezoneHelpPopup &&
      !timezoneHelpButton.contains(/** @type {Node} */ (event.target)) &&
      !timezoneHelpPopup.contains(/** @type {Node} */ (event.target))
    ) {
      showTimezoneHelp = false;
    }
  }

  onMount(async () => {
    currentOwner = get(currentOwnerStore);
    if (!currentOwner) {
      goto('/login');
      return;
    }

    // Add click outside handler for timezone help popup
    if (browser) {
      document.addEventListener('click', handleTimezoneHelpClickOutside);
    }

    // Load Leaflet.js dynamically
    /** @type {any} */
    const win = window;
    if (typeof window !== 'undefined' && !win.L) {
      const link = document.createElement('link');
      link.rel = 'stylesheet';
      link.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';
      link.integrity = 'sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=';
      link.crossOrigin = '';
      document.head.appendChild(link);

      const script = document.createElement('script');
      script.src = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.js';
      script.integrity = 'sha256-20nQCchB9co0qIjJZRGuk2/Z9VM+kNiyxNV1lvTlZBo=';
      script.crossOrigin = '';
      script.onload = () => {
        leafletLoaded = true;
        if (geolocationMapContainer) {
          initGeolocationMap();
        }
      };
      document.head.appendChild(script);
    } else if (typeof window !== 'undefined' && win.L) {
      leafletLoaded = true;
    }

    if (venueUuidFromUrl) {
      // Editing existing venue - load from API
      isNewVenue = false;
      isLoading = true;
      loadError = '';

      try {
        venue = await getVenue(venueUuidFromUrl);
        // Check ownership (backend should handle this, but double-check)
        if (venue.owner_uuid !== currentOwner.owner_uuid) {
          loadError = 'You do not have permission to edit this venue';
          goto('/venue-owner');
          return;
        }

        loadVenueFormFields();
        await loadVenueDataFromAPI();
        dataLoaded = true;
      } catch (err) {
        console.error('Failed to load venue', err);
        if (err instanceof ApiError && err.status === 404) {
          loadError = 'Venue not found';
        } else if (err instanceof ApiError && err.status === 401) {
          loadError = 'Your session has expired. Please log in again.';
          goto('/login');
        } else {
          loadError = 'Failed to load venue. Please try again.';
        }
      } finally {
        isLoading = false;
      }
    } else {
      // Creating new venue
      isNewVenue = true;
      venue = null;
      eventListsData = [];
      previewEventListId = null;
      dataLoaded = true;
    }
  });

  // Data loading is now handled in onMount via API calls

  // Initialize geolocation map when container is ready
  afterUpdate(() => {
    if (leafletLoaded && geolocationMapContainer && !geolocationMap) {
      initGeolocationMap();
    } else if (geolocationMap && venueGeolocation) {
      updateGeolocationMap();
    }
  });

  onDestroy(() => {
    if (geolocationMap) {
      geolocationMap.remove();
      geolocationMap = null;
      geolocationMarker = null;
    }
    if (browser) {
      document.removeEventListener('click', handleTimezoneHelpClickOutside);
    }
  });

  // Get preview event list
  $: previewEventList = previewEventListId
    ? eventListsData.find((el) => el.event_list_uuid === previewEventListId)
    : null;

  // Get preview events
  $: previewEvents = previewEventList ? previewEventList.events : [];

  // Get venue owner for preview
  $: previewVenueOwner = currentOwner;
</script>

<svelte:head>
  <title>{isNewVenue ? 'Create' : 'Edit'} Venue - times.place</title>
</svelte:head>

<div
  class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-4 pb-4 md:pt-6 md:pb-8 lg:pt-6 lg:pb-8 w-full overflow-x-hidden"
>
  {#if isLoading}
    <div
      class="mb-4 md:mb-6 flex flex-col gap-2 md:gap-4 px-4 md:px-6 lg:sticky lg:top-20 lg:z-10 lg:bg-white lg:pb-4 lg:border-b lg:border-gray-200 lg:shadow-sm"
    >
      <div>
        <h1 class="text-[24px] md:text-3xl font-bold text-gray-900">
          {isNewVenue ? 'Create' : 'Edit'} Venue
        </h1>
      </div>
      <div class="flex gap-2 justify-start">
        {#if undoStack.length > 0}
          <button
            on:click={handleUndo}
            class="py-1.5 px-3 text-sm rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50 transition-colors md:py-2 md:px-4 md:text-base"
            title="Undo last change"
          >
            Undo
          </button>
        {/if}
        <button
          on:click={cancelEdit}
          class="py-1.5 px-3 text-sm rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium transition-colors duration-200 md:py-2 md:px-4 md:text-base"
        >
          Cancel
        </button>
        <button
          on:click={saveVenue}
          disabled={isSaving || isLoading}
          class="py-1.5 px-3 text-sm rounded-lg bg-blue-600 hover:bg-blue-700 text-white font-medium transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed md:py-2 md:px-4 md:text-base"
        >
          {isSaving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  {/if}

  {#if loadError}
    <div
      class="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
    >
      {loadError}
    </div>
  {/if}

  {#if saveError}
    <div
      class="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
    >
      {saveError}
    </div>
  {/if}

  {#if isLoading}
    <div class="mb-4 text-center text-gray-600">Loading venue data...</div>
  {:else if dataLoaded}
    <div
      class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6 w-full overflow-x-hidden max-w-full lg:h-[calc(100vh-6rem)] lg:min-h-0"
    >
      <!-- Left column on desktop: single flex container so header stays at top and form scrolls beneath (same as mobile) -->
      <div
        class="flex flex-col gap-4 lg:col-start-1 lg:row-start-1 lg:row-span-2 lg:min-h-0 lg:overflow-hidden"
      >
        <div
          class="mb-4 md:mb-0 flex flex-col gap-2 md:gap-4 px-4 md:px-6 flex-shrink-0 lg:pt-0 lg:gap-2 lg:sticky lg:top-20 lg:z-10 lg:bg-white lg:pb-4 lg:border-b lg:border-gray-200 lg:shadow-sm"
        >
          <div>
            <h1 class="text-[24px] md:text-3xl font-bold text-gray-900">
              {isNewVenue ? 'Create' : 'Edit'} Venue
            </h1>
          </div>
          <div class="flex gap-2 justify-start">
            {#if undoStack.length > 0}
              <button
                on:click={handleUndo}
                class="py-1.5 px-3 text-sm rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50 transition-colors md:py-2 md:px-4 md:text-base"
                title="Undo last change"
              >
                Undo
              </button>
            {/if}
            <button
              on:click={cancelEdit}
              class="py-1.5 px-3 text-sm rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium transition-colors duration-200 md:py-2 md:px-4 md:text-base"
            >
              Cancel
            </button>
            <button
              on:click={saveVenue}
              disabled={isSaving || isLoading}
              class="py-1.5 px-3 text-sm rounded-lg bg-blue-600 hover:bg-blue-700 text-white font-medium transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed md:py-2 md:px-4 md:text-base"
            >
              {isSaving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </div>
        <!-- Editing Pane (scrolls beneath header on desktop, same as mobile) -->
        <div
          class="bg-white rounded-xl shadow-lg p-4 md:p-6 space-y-4 md:space-y-6 overflow-y-auto overflow-x-hidden flex-1 min-h-0 w-full max-h-[calc(100vh-200px)] lg:max-h-none"
        >
          <!-- Basic Venue Information -->
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-800">Basic Information</h3>

            <div>
              <label
                for="venue-name"
                class="block text-sm font-medium text-gray-700 mb-1"
                >Venue Name *</label
              >
              <div class="relative">
                <input
                  type="text"
                  id="venue-name"
                  bind:value={venueName}
                  on:blur={validateVenueNameBlur}
                  class="w-full px-3 py-2 pr-9 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 {venueNameError
                    ? 'border-red-500'
                    : 'border-gray-300'}"
                  placeholder="Enter venue name"
                />
                <button
                  type="button"
                  on:click={() => (venueName = '')}
                  class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                  title="Clear"
                  aria-label="Clear"
                >
                  <svg
                    class="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>
              {#if venueNameError}
                <p class="mt-1 text-sm text-red-600">{venueNameError}</p>
              {/if}
            </div>

            <div>
              <label
                for="venue-address"
                class="block text-sm font-medium text-gray-700 mb-1"
                >Address (optional)</label
              >
              <div class="flex gap-2">
                <div class="relative flex-1 min-w-0">
                  <input
                    type="text"
                    id="venue-address"
                    bind:value={venueAddress}
                    class="w-full px-3 py-2 pr-9 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Enter address"
                    on:keydown={(/** @type {KeyboardEvent} */ e) =>
                      e.key === 'Enter' && handleFindOnMap()}
                  />
                  <button
                    type="button"
                    on:click={() => (venueAddress = '')}
                    class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                    title="Clear"
                    aria-label="Clear"
                  >
                    <svg
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      stroke-width="2"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </div>
                <button
                  type="button"
                  on:click={handleFindOnMap}
                  disabled={isGeocoding}
                  class="px-3 sm:px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm sm:text-base rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed shrink-0"
                  title="Find this address on the map below"
                >
                  <span class="hidden sm:inline"
                    >{isGeocoding ? 'Searching...' : 'Find on Map'}</span
                  >
                  <span class="sm:hidden"
                    >{isGeocoding ? 'Searching...' : 'Find'}</span
                  >
                </button>
              </div>
              {#if showGeocodingResults && geocodingResults.length > 0}
                <div
                  class="mt-2 border border-gray-300 rounded-lg bg-white shadow-lg max-h-48 overflow-auto z-10 relative"
                >
                  <div
                    class="p-2 text-xs text-gray-600 font-medium border-b bg-gray-50"
                  >
                    Multiple locations found. Select one:
                  </div>
                  {#each geocodingResults as result}
                    <button
                      type="button"
                      on:click={() => selectGeocodingResult(result)}
                      class="w-full text-left px-3 py-2 hover:bg-blue-50 border-b border-gray-100 last:border-b-0 text-sm transition-colors"
                    >
                      {result.display_name}
                    </button>
                  {/each}
                </div>
              {/if}
            </div>

            <div>
              <label
                for="venue-geolocation"
                class="block text-sm font-medium text-gray-700 mb-1"
                >Geolocation (optional)</label
              >
              <p class="text-xs text-gray-500 mb-2">
                Click on the map to set location, or use "Find on Map" button
                above to search by address. Drag the marker to adjust.
              </p>

              <div class="mb-2">
                <div
                  bind:this={geolocationMapContainer}
                  class="w-full h-64 rounded-lg overflow-hidden border border-gray-300"
                ></div>
              </div>

              <div class="mb-2">
                <label
                  for="venue-geolocation"
                  class="block text-xs font-medium text-gray-700 mb-1"
                  >Coordinates (latitude,longitude)</label
                >
                <div class="relative">
                  <input
                    type="text"
                    id="venue-geolocation"
                    bind:value={venueGeolocation}
                    class="w-full px-3 py-2 pr-9 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="latitude,longitude"
                    on:input={() => {
                      if (venueGeolocation) {
                        updateGeolocationMap();
                      }
                    }}
                  />
                  <button
                    type="button"
                    on:click={() => (venueGeolocation = '')}
                    class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                    title="Clear"
                    aria-label="Clear"
                  >
                    <svg
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      stroke-width="2"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </div>
              </div>
            </div>

            <div>
              <label
                for="venue-comment"
                class="block text-sm font-medium text-gray-700 mb-1"
                >Comment (optional)</label
              >
              <div class="relative">
                <textarea
                  id="venue-comment"
                  bind:value={venueComment}
                  rows="3"
                  class="w-full px-3 py-2 pr-9 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Optional comment about the venue"
                ></textarea>
                <button
                  type="button"
                  on:click={() => (venueComment = '')}
                  class="absolute right-2 top-2 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                  title="Clear"
                  aria-label="Clear"
                >
                  <svg
                    class="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <div>
              <div class="flex items-center gap-2 mb-1">
                <label
                  for="venue-timezone"
                  class="block text-sm font-medium text-gray-700"
                  >Timezone (optional, leave empty for your current timezone)</label
                >
                <div class="relative">
                  <button
                    type="button"
                    bind:this={timezoneHelpButton}
                    on:click={() => (showTimezoneHelp = !showTimezoneHelp)}
                    on:blur={() => {
                      // Close on blur, but delay to allow click on popup
                      setTimeout(() => {
                        if (
                          !timezoneHelpPopup?.contains(document.activeElement)
                        ) {
                          showTimezoneHelp = false;
                        }
                      }, 200);
                    }}
                    class="text-gray-400 hover:text-gray-600 focus:text-gray-600 focus:outline-none transition-colors"
                    aria-label="Timezone help"
                    aria-expanded={showTimezoneHelp}
                    aria-haspopup="true"
                  >
                    <svg
                      class="w-5 h-5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <circle
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        stroke-width="2"
                      />
                      <path
                        d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      />
                      <line
                        x1="12"
                        y1="17"
                        x2="12.01"
                        y2="17"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                      />
                    </svg>
                  </button>
                  {#if showTimezoneHelp}
                    <!-- Mobile: fixed positioning centered in viewport, Desktop: absolute positioning relative to button -->
                    <div
                      bind:this={timezoneHelpPopup}
                      class="fixed md:absolute z-50 w-[calc(100vw-2rem)] md:w-80 max-w-sm p-4 bg-white border border-gray-300 rounded-lg shadow-lg text-sm text-gray-700 left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 md:top-auto md:translate-y-0 md:left-auto md:translate-x-0 md:right-0 md:mt-2"
                      role="tooltip"
                    >
                      <div class="space-y-3">
                        <div
                          class="p-3 bg-blue-50 border-l-4 border-blue-500 rounded"
                        >
                          <p class="font-semibold text-blue-900 mb-1">
                            Timezone Set:
                          </p>
                          <p class="text-blue-800 text-xs">
                            All visitors see event times in the venue's
                            timezone, regardless of their location. For example,
                            if you set 14:00 in Jerusalem time, everyone sees
                            14:00 (Jerusalem time).
                          </p>
                        </div>
                        <div
                          class="p-3 bg-green-50 border-l-4 border-green-500 rounded"
                        >
                          <p class="font-semibold text-green-900 mb-1">
                            No Timezone:
                          </p>
                          <p class="text-green-800 text-xs">
                            Each visitor sees event times converted to their own
                            browser timezone. For example, 14:00 might show as
                            09:00 in New York or 14:00 in London, depending on
                            the visitor's location.
                          </p>
                        </div>
                      </div>
                      <button
                        type="button"
                        on:click={() => (showTimezoneHelp = false)}
                        class="mt-3 text-xs text-blue-600 hover:text-blue-800 underline"
                      >
                        Close
                      </button>
                    </div>
                  {/if}
                </div>
              </div>
              <div class="relative flex gap-2 items-center">
                <select
                  id="venue-timezone"
                  bind:value={venueTimezone}
                  class="flex-1 min-w-0 pl-3 pr-10 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3Csvg%20xmlns=%22http://www.w3.org/2000/svg%22%20viewBox=%220%200%2024%2024%22%20fill=%22none%22%20stroke=%22%23666%22%20stroke-width=%222%22%20stroke-linecap=%22round%22%20stroke-linejoin=%22round%22%3E%3Cpolyline%20points=%226%209%2012%2015%2018%209%22%3E%3C/polyline%3E%3C/svg%3E')] bg-no-repeat bg-[length:1.25em] bg-[position:right_0.75rem_center]"
                >
                  {#each timezones as tz}
                    {#if tz.group}
                      <optgroup label={tz.group}>
                        {#each tz.zones as zone}
                          <option value={zone.value}>{zone.label}</option>
                        {/each}
                      </optgroup>
                    {:else}
                      <option value={tz.value}>{tz.label}</option>
                    {/if}
                  {/each}
                </select>
                <button
                  type="button"
                  on:click={() => (venueTimezone = '')}
                  class="shrink-0 p-2 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                  title="Clear timezone"
                  aria-label="Clear timezone"
                >
                  <svg
                    class="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <div>
              <label
                for="venue-banner"
                class="block text-sm font-medium text-gray-700 mb-1"
                >Banner Image (optional) — preferred ratio 16:9</label
              >
              <input
                type="file"
                id="venue-banner"
                accept="image/*"
                on:change={handleImageUpload}
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
              {#if venueBannerImage}
                <BannerImage
                  src={venueBannerImage}
                  alt="Banner preview"
                  size="sm"
                  wrapperClass="mt-2"
                />
              {/if}
            </div>
          </div>

          <!-- Event Lists -->
          <div class="space-y-4">
            <div>
              <h3 class="text-lg font-medium text-gray-800">Event Lists</h3>
            </div>

            {#if eventListsData.length === 0}
              <p class="text-sm text-gray-500 py-4 text-center">
                No event lists yet. Click "+ Add Event List" to create one.
              </p>
            {/if}

            {#each eventListsData as listData, listIndex (listData.event_list_uuid)}
              <div
                class="border-2 border-gray-300 rounded-lg p-4 space-y-3 w-full min-w-0 overflow-x-hidden"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <button
                      on:click={() => moveEventListUp(listIndex)}
                      disabled={listIndex === 0}
                      class="p-1 text-gray-500 hover:text-gray-700 disabled:opacity-50"
                      title="Move up"
                    >
                      ↑
                    </button>
                    <button
                      on:click={() => moveEventListDown(listIndex)}
                      disabled={listIndex === eventListsData.length - 1}
                      class="p-1 text-gray-500 hover:text-gray-700 disabled:opacity-50"
                      title="Move down"
                    >
                      ↓
                    </button>
                  </div>
                  <button
                    on:click={() => deleteEventList(listData.event_list_uuid)}
                    class="px-2 py-1 bg-red-600 hover:bg-red-700 text-white text-sm rounded transition-colors"
                  >
                    Delete Event List
                  </button>
                </div>

                <div>
                  <label
                    for="event-list-name-{listData.event_list_uuid}"
                    class="block text-sm font-medium text-gray-700 mb-1"
                    >Event List Name (optional)</label
                  >
                  <div class="relative">
                    <input
                      type="text"
                      id="event-list-name-{listData.event_list_uuid}"
                      bind:value={listData.name}
                      class="w-full px-3 py-2 pr-9 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="e.g. Week of Dec 2"
                    />
                    <button
                      type="button"
                      on:click={() => (listData.name = '')}
                      class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                      title="Clear"
                      aria-label="Clear"
                    >
                      <svg
                        class="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        stroke-width="2"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          d="M6 18L18 6M6 6l12 12"
                        />
                      </svg>
                    </button>
                  </div>
                </div>

                <div>
                  <label
                    for="event-list-date-{listData.event_list_uuid}"
                    class="block text-sm font-medium text-gray-700 mb-1"
                    >Date (optional)</label
                  >
                  <div class="flex gap-2">
                    <input
                      type="date"
                      id="event-list-date-{listData.event_list_uuid}"
                      bind:value={listData.date}
                      on:input={() => {
                        // Update datetime for all events in this list when date changes
                        // If no date is provided, use today's date as placeholder
                        const dateToUse =
                          listData.date && listData.date.trim()
                            ? listData.date
                            : new Date().toISOString().split('T')[0];
                        listData.events.forEach((/** @type {any} */ event) => {
                          event.datetime = combineTimeAndDate(
                            event.time,
                            dateToUse,
                            venueTimezone,
                          );
                        });
                      }}
                      class="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    {#if listData.date && listData.date.trim()}
                      <button
                        type="button"
                        on:click={() =>
                          clearEventListDate(listData.event_list_uuid)}
                        class="px-3 py-2 bg-gray-200 hover:bg-gray-300 text-gray-700 text-sm font-medium rounded-lg transition-colors duration-200"
                        title="Clear date"
                      >
                        Clear
                      </button>
                    {/if}
                  </div>
                  <p class="mt-1 text-xs text-gray-500">
                    Date is stored in ISO 8601 format (YYYY-MM-DD)
                  </p>
                </div>

                <div>
                  <label
                    for="event-list-comment-{listData.event_list_uuid}"
                    class="block text-sm font-medium text-gray-700 mb-1"
                    >Comment (optional)</label
                  >
                  <div class="relative">
                    <textarea
                      id="event-list-comment-{listData.event_list_uuid}"
                      bind:value={listData.comment}
                      rows="2"
                      class="w-full px-3 py-2 pr-9 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="Optional comment"
                    ></textarea>
                    <button
                      type="button"
                      on:click={() => (listData.comment = '')}
                      class="absolute right-2 top-2 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                      title="Clear"
                      aria-label="Clear"
                    >
                      <svg
                        class="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        stroke-width="2"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          d="M6 18L18 6M6 6l12 12"
                        />
                      </svg>
                    </button>
                  </div>
                </div>

                <div>
                  <div class="block text-sm font-medium text-gray-700 mb-2">
                    Visibility
                  </div>
                  <div class="flex gap-4">
                    <label class="flex items-center">
                      <input
                        type="radio"
                        value="public"
                        checked={listData.visibility === 'public'}
                        on:change={() =>
                          handleEventListVisibilityChange(
                            listData.event_list_uuid,
                            'public',
                          )}
                        class="mr-2"
                      />
                      <span>Public</span>
                    </label>
                    <label class="flex items-center">
                      <input
                        type="radio"
                        value="private"
                        checked={listData.visibility === 'private'}
                        on:change={() =>
                          handleEventListVisibilityChange(
                            listData.event_list_uuid,
                            'private',
                          )}
                        class="mr-2"
                      />
                      <span>Private</span>
                    </label>
                  </div>
                  {#if listData.private_link_token}
                    <div class="mt-2 p-2 bg-gray-50 rounded text-xs">
                      <p class="text-gray-600 mb-1">
                        Share link (works for public and private):
                      </p>
                      <div class="flex items-center gap-2 flex-wrap">
                        <code class="text-blue-600 break-all flex-1 min-w-0">
                          {typeof window !== 'undefined'
                            ? `${window.location.origin}/?token=${listData.private_link_token}`
                            : ''}
                        </code>
                        <button
                          type="button"
                          on:click={() => copyShareLink(listData)}
                          class="shrink-0 p-1.5 rounded text-gray-600 hover:bg-gray-200 hover:text-gray-800 transition-colors"
                          title="Copy share link"
                          aria-label="Copy share link"
                        >
                          {#if copiedShareLinkEventListUuid === listData.event_list_uuid}
                            <svg
                              class="w-4 h-4 text-green-600"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M5 13l4 4L19 7"
                              />
                            </svg>
                          {:else}
                            <svg
                              class="w-4 h-4"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                              />
                            </svg>
                          {/if}
                        </button>
                      </div>
                    </div>
                  {:else}
                    <p class="mt-2 text-xs text-gray-500">
                      Save the venue to get a shareable direct link for this
                      event list.
                    </p>
                  {/if}
                </div>

                <div>
                  <div class="block text-sm font-medium text-gray-700 mb-2">
                    Events
                  </div>

                  {#if !listData.events || listData.events.length === 0}
                    <p class="text-sm text-gray-500 py-2">
                      No events yet. Click "+ Add Event" to add one.
                    </p>
                  {/if}
                  {#each listData.events || [] as event, eventIndex (event.event_uuid)}
                    <div
                      class="border border-gray-200 rounded p-3 mb-2 space-y-2 w-full overflow-x-hidden"
                    >
                      <div class="flex items-center justify-between">
                        <div class="flex items-center gap-1">
                          <button
                            on:click={() =>
                              moveEventUp(listData.event_list_uuid, eventIndex)}
                            disabled={eventIndex === 0}
                            class="p-1 text-gray-500 hover:text-gray-700 disabled:opacity-50 text-xs"
                            title="Move up"
                          >
                            ↑
                          </button>
                          <button
                            on:click={() =>
                              moveEventDown(
                                listData.event_list_uuid,
                                eventIndex,
                              )}
                            disabled={eventIndex ===
                              (listData.events || []).length - 1}
                            class="p-1 text-gray-500 hover:text-gray-700 disabled:opacity-50 text-xs"
                            title="Move down"
                          >
                            ↓
                          </button>
                        </div>
                        <div class="flex gap-1">
                          <button
                            on:click={() =>
                              duplicateEvent(
                                listData.event_list_uuid,
                                event.event_uuid,
                              )}
                            class="px-2 py-1 bg-blue-500 hover:bg-blue-600 text-white text-xs rounded transition-colors"
                          >
                            Duplicate
                          </button>
                          <button
                            on:click={() =>
                              deleteEvent(
                                listData.event_list_uuid,
                                event.event_uuid,
                              )}
                            class="px-2 py-1 bg-red-600 hover:bg-red-700 text-white text-xs rounded transition-colors"
                          >
                            Delete Event
                          </button>
                        </div>
                      </div>

                      <div>
                        <label
                          for="event-name-{event.event_uuid}"
                          class="block text-xs font-medium text-gray-700 mb-1"
                          >Event Name</label
                        >
                        <div class="relative">
                          <input
                            type="text"
                            id="event-name-{event.event_uuid}"
                            bind:value={event.event_name}
                            on:blur={() =>
                              validateEventNameBlur(event.event_uuid)}
                            class="w-full px-2 py-1 pr-8 text-sm border rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500 {eventNameErrors[
                              event.event_uuid
                            ]
                              ? 'border-red-500'
                              : 'border-gray-300'}"
                            placeholder="e.g. Shacharit, Mincha"
                          />
                          <button
                            type="button"
                            on:click={() => (event.event_name = '')}
                            class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                            title="Clear"
                            aria-label="Clear"
                          >
                            <svg
                              class="w-3.5 h-3.5"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M6 18L18 6M6 6l12 12"
                              />
                            </svg>
                          </button>
                        </div>
                        {#if eventNameErrors[event.event_uuid]}
                          <p class="mt-1 text-xs text-red-600">
                            {eventNameErrors[event.event_uuid]}
                          </p>
                        {/if}
                      </div>

                      <div
                        class="w-full min-w-0"
                        style="max-width: 100%; overflow: hidden; box-sizing: border-box;"
                      >
                        <label
                          for="event-time-{event.event_uuid}"
                          class="block text-xs font-medium text-gray-700 mb-1"
                          >Time (HH:MM)</label
                        >
                        <div class="relative flex gap-1">
                          <input
                            type="time"
                            id="event-time-{event.event_uuid}"
                            bind:value={event.time}
                            on:input={() => {
                              const dateToUse =
                                listData.date && listData.date.trim()
                                  ? listData.date
                                  : new Date().toISOString().split('T')[0];
                              event.datetime = combineTimeAndDate(
                                event.time,
                                dateToUse,
                                venueTimezone,
                              );
                            }}
                            class="flex-1 min-w-0 px-2 py-1 pr-8 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            style="box-sizing: border-box; max-width: 100%; min-width: 0; -webkit-appearance: none; appearance: none;"
                          />
                          <button
                            type="button"
                            on:click={() => {
                              event.time = '00:00';
                              const dateToUse =
                                listData.date && listData.date.trim()
                                  ? listData.date
                                  : new Date().toISOString().split('T')[0];
                              event.datetime = combineTimeAndDate(
                                '00:00',
                                dateToUse,
                                venueTimezone,
                              );
                            }}
                            class="shrink-0 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors self-center"
                            title="Clear time"
                            aria-label="Clear time"
                          >
                            <svg
                              class="w-3.5 h-3.5"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M6 18L18 6M6 6l12 12"
                              />
                            </svg>
                          </button>
                        </div>
                      </div>

                      <div>
                        <label
                          for="event-duration-{event.event_uuid}"
                          class="block text-xs font-medium text-gray-700 mb-1"
                          >Duration (minutes) (optional)</label
                        >
                        <div class="flex gap-2">
                          <input
                            type="number"
                            id="event-duration-{event.event_uuid}"
                            bind:value={event.duration_minutes}
                            min="1"
                            class="flex-1 min-w-0 px-2 py-1 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            placeholder="No duration"
                          />
                          {#if event.duration_minutes && event.duration_minutes !== '' && event.duration_minutes > 0}
                            <button
                              type="button"
                              on:click={() =>
                                clearEventDuration(
                                  listData.event_list_uuid,
                                  event.event_uuid,
                                )}
                              class="px-2 py-1 bg-gray-200 hover:bg-gray-300 text-gray-700 text-sm font-medium rounded transition-colors duration-200 shrink-0"
                              title="Clear duration"
                            >
                              <svg
                                class="w-4 h-4"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path
                                  stroke-linecap="round"
                                  stroke-linejoin="round"
                                  stroke-width="2"
                                  d="M6 18L18 6M6 6l12 12"
                                />
                              </svg>
                            </button>
                          {/if}
                        </div>
                      </div>

                      <div>
                        <label
                          for="event-comment-{event.event_uuid}"
                          class="block text-xs font-medium text-gray-700 mb-1"
                          >Comment (optional)</label
                        >
                        <div class="relative">
                          <textarea
                            id="event-comment-{event.event_uuid}"
                            bind:value={event.comment}
                            rows="2"
                            class="w-full px-2 py-1 pr-8 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            placeholder="Optional comment"
                          ></textarea>
                          <button
                            type="button"
                            on:click={() => (event.comment = '')}
                            class="absolute right-2 top-2 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                            title="Clear"
                            aria-label="Clear"
                          >
                            <svg
                              class="w-3.5 h-3.5"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M6 18L18 6M6 6l12 12"
                              />
                            </svg>
                          </button>
                        </div>
                      </div>
                    </div>
                  {/each}
                  <button
                    on:click={() => addEvent(listData.event_list_uuid)}
                    class="mt-2 px-2 py-1 bg-green-600 hover:bg-green-700 text-white text-sm rounded transition-colors"
                  >
                    + Add Event
                  </button>
                </div>
              </div>
            {/each}
            <button
              on:click={addEventList}
              class="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-lg transition-colors"
            >
              + Add Event List
            </button>
          </div>
        </div>
      </div>

      <!-- Preview Pane (constrained to grid height on desktop so page doesn't scroll) -->
      <div
        class="hidden lg:block bg-white rounded-xl shadow-lg p-6 space-y-4 overflow-y-auto lg:col-start-2 lg:row-start-1 lg:row-span-2 lg:min-h-0"
      >
        <h2 class="text-xl font-semibold text-gray-900 border-b pb-2">
          Live Preview
        </h2>

        {#if previewEventList && eventListsData.length > 1}
          <div>
            <label
              for="preview-event-list"
              class="block text-sm font-medium text-gray-700 mb-1"
              >Select Event List</label
            >
            <select
              id="preview-event-list"
              bind:value={previewEventListId}
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              {#each eventListsData as listData}
                <option value={listData.event_list_uuid}
                  >{listData.name || 'Untitled Event List'}</option
                >
              {/each}
            </select>
          </div>
        {/if}

        {#if venueName || previewEventList}
          <div>
            {#if venueBannerImage}
              <BannerImage
                src={venueBannerImage}
                alt={venueName || 'Venue'}
                size="sm"
                wrapperClass="mb-4"
              />
            {/if}

            <h2 class="text-3xl font-bold mb-3 text-gray-900">
              {venueName || 'Venue Name'}
            </h2>

            {#if venueAddress || venueGeolocation || venueTimezone || previewVenueOwner || venueComment}
              <div class="mb-3 grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="flex flex-col justify-start">
                  {#if venueAddress}
                    <p class="text-sm text-gray-600 mb-2">
                      <span class="font-medium text-gray-700">Address:</span>
                      {venueAddress}
                    </p>
                  {/if}

                  {#if venueGeolocation}
                    <p class="text-sm text-gray-600 mb-2">
                      <span class="font-medium text-gray-700">Geolocation:</span
                      >
                      {venueGeolocation}
                    </p>
                  {/if}

                  {#if venueTimezone}
                    <p class="text-sm text-gray-600 mb-2">
                      <span class="font-medium text-gray-700">Timezone:</span>
                      {venueTimezone}
                    </p>
                  {/if}

                  {#if previewVenueOwner}
                    <p class="text-sm text-gray-600 mb-2">
                      <span class="font-medium text-gray-700">Contact:</span>
                      {#if previewVenueOwner.name}
                        <span class="ml-2">{previewVenueOwner.name}</span>
                      {/if}
                      {#if previewVenueOwner.email}
                        <a
                          href="mailto:{previewVenueOwner.email}"
                          class="ml-2 text-blue-600 hover:text-blue-800 hover:underline"
                        >
                          {previewVenueOwner.email}
                        </a>
                      {/if}
                      {#if previewVenueOwner.mobile}
                        <a
                          href="tel:{previewVenueOwner.mobile}"
                          class="ml-2 text-blue-600 hover:text-blue-800 hover:underline"
                        >
                          {previewVenueOwner.mobile}
                        </a>
                      {/if}
                    </p>
                  {/if}

                  {#if venueComment}
                    <p class="text-sm text-gray-600 whitespace-pre-line">
                      {venueComment}
                    </p>
                  {/if}
                </div>
              </div>
            {/if}

            {#if previewEventList}
              <div class="mt-6">
                <h3 class="text-2xl font-semibold mb-1 text-gray-900">
                  {previewEventList.name}
                </h3>
                {#if previewEventList.comment}
                  <p class="text-xs text-gray-600 mb-4 whitespace-pre-line">
                    {previewEventList.comment}
                  </p>
                {/if}

                {#if previewEvents.length === 0}
                  <p class="text-gray-500">
                    No events scheduled for this list.
                  </p>
                {:else}
                  <div class="space-y-1">
                    {#each previewEvents as event}
                      <div
                        class="flex items-center justify-between py-1 px-3 bg-gray-50 rounded-lg"
                      >
                        <div>
                          <p class="font-medium text-gray-900 text-sm">
                            {event.event_name || 'Untitled Event'}
                          </p>
                          {#if event.comment}
                            <p
                              class="text-sm md:text-xs text-gray-600 whitespace-pre-line mt-0.5"
                            >
                              {event.comment}
                            </p>
                          {/if}
                        </div>
                        <div class="text-right">
                          {#if previewEventList.date && previewEventList.date.trim()}
                            <p class="text-base font-semibold text-blue-600">
                              {formatEventTime(
                                Math.floor(
                                  new Date(
                                    combineTimeAndDate(
                                      event.time,
                                      previewEventList.date,
                                      venueTimezone,
                                    ),
                                  ).getTime() / 1000,
                                ),
                                venueTimezone
                                  ? { timeZone: venueTimezone }
                                  : {},
                              )}
                            </p>
                          {:else}
                            <p class="text-base font-semibold text-gray-400">
                              {event.time}
                            </p>
                          {/if}
                          {#if event.duration_minutes}
                            <p class="text-xs text-gray-500 mt-0.5">
                              {event.duration_minutes} min
                            </p>
                          {/if}
                        </div>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {:else}
              <div class="mt-6 p-4 bg-gray-50 rounded-lg">
                <p class="text-gray-500 text-center">
                  Add an event list to see the preview.
                </p>
              </div>
            {/if}
          </div>
        {:else}
          <p class="text-gray-500 text-center">
            Start editing to see the preview.
          </p>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  /* Prevent horizontal scrolling on mobile for this page */
  :global(body) {
    overflow-x: hidden;
    position: relative;
  }

  /* Ensure all inputs and form elements don't overflow */
  :global(input),
  :global(textarea),
  :global(select) {
    max-width: 100%;
    box-sizing: border-box;
  }

  /* Fix time input width on mobile devices - use more specific selector */
  :global(input[type='time']) {
    min-width: 0 !important;
    width: 100% !important;
    max-width: 100% !important;
    box-sizing: border-box !important;
    -webkit-appearance: none !important;
    appearance: none !important;
  }

  /* Ensure event containers don't overflow */
  :global(.border.border-gray-200.rounded.p-3) {
    max-width: 100% !important;
    overflow-x: hidden !important;
    box-sizing: border-box !important;
  }

  /* Force all inputs in event containers to respect width */
  :global(.border.border-gray-200.rounded.p-3 input) {
    max-width: 100% !important;
    box-sizing: border-box !important;
    width: 100% !important;
  }

  /* Additional constraint for editing pane */
  :global(.bg-white.rounded-xl.shadow-lg.p-6) {
    max-width: 100% !important;
    overflow-x: hidden !important;
  }

  /* Mobile-specific fix for time inputs */
  @media (max-width: 768px) {
    :global(input[type='time']) {
      width: 100% !important;
      max-width: 100% !important;
      min-width: 0 !important;
      box-sizing: border-box !important;
      padding-left: 0.5rem !important;
      padding-right: 0.5rem !important;
    }
  }
</style>
