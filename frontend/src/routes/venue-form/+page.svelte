<script>
  import { onMount, afterUpdate, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { get } from 'svelte/store';
  import { goto } from '$app/navigation';
  import { currentOwnerStore, venueStore, eventListStore, eventStore, ownersStore } from '$lib/stores';
  import { generateUUID } from '$lib/utils/uuid.js';
  import { getCurrentTimestamp, updateModifiedTimestamp, parseISODate, createEventTimestamp } from '$lib/utils/datetime.js';
  import { formatEventTime } from '$lib/utils/datetime.js';

  // Common timezones organized by region
  /** @type {any} */
  const timezones = [
    { value: '', label: 'No timezone' },
    { group: 'Americas', zones: [
      { value: 'America/New_York', label: 'Eastern Time (New York)' },
      { value: 'America/Chicago', label: 'Central Time (Chicago)' },
      { value: 'America/Denver', label: 'Mountain Time (Denver)' },
      { value: 'America/Los_Angeles', label: 'Pacific Time (Los Angeles)' },
      { value: 'America/Toronto', label: 'Eastern Time (Toronto)' },
      { value: 'America/Vancouver', label: 'Pacific Time (Vancouver)' },
      { value: 'America/Mexico_City', label: 'Central Time (Mexico City)' },
      { value: 'America/Sao_Paulo', label: 'Brasília Time (São Paulo)' },
      { value: 'America/Buenos_Aires', label: 'Argentina Time (Buenos Aires)' },
    ]},
    { group: 'Europe', zones: [
      { value: 'Europe/London', label: 'Greenwich Mean Time (London)' },
      { value: 'Europe/Paris', label: 'Central European Time (Paris)' },
      { value: 'Europe/Berlin', label: 'Central European Time (Berlin)' },
      { value: 'Europe/Rome', label: 'Central European Time (Rome)' },
      { value: 'Europe/Madrid', label: 'Central European Time (Madrid)' },
      { value: 'Europe/Amsterdam', label: 'Central European Time (Amsterdam)' },
      { value: 'Europe/Athens', label: 'Eastern European Time (Athens)' },
      { value: 'Europe/Moscow', label: 'Moscow Time' },
    ]},
    { group: 'Asia & Middle East', zones: [
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
    ]},
    { group: 'Oceania', zones: [
      { value: 'Australia/Sydney', label: 'Australian Eastern Time (Sydney)' },
      { value: 'Australia/Melbourne', label: 'Australian Eastern Time (Melbourne)' },
      { value: 'Australia/Brisbane', label: 'Australian Eastern Time (Brisbane)' },
      { value: 'Australia/Perth', label: 'Australian Western Time (Perth)' },
      { value: 'Pacific/Auckland', label: 'New Zealand Time (Auckland)' },
    ]},
    { group: 'Africa', zones: [
      { value: 'Africa/Cairo', label: 'Eastern European Time (Cairo)' },
      { value: 'Africa/Johannesburg', label: 'South Africa Standard Time (Johannesburg)' },
      { value: 'Africa/Lagos', label: 'West Africa Time (Lagos)' },
    ]},
  ];

  // Get venue UUID from URL params (only in browser to avoid prerendering errors)
  $: venueUuidFromUrl = browser && $page.url.searchParams ? $page.url.searchParams.get('venue_uuid') : null;

  /** @type {any} */
  let currentOwner = null;
  /** @type {any} */
  let venue = null;
  let isNewVenue = false;
  /** @type {any[]} */
  let allEventLists = [];
  /** @type {any[]} */
  let allEvents = [];
  /** @type {any[]} */
  let allOwners = [];

  // Form state
  let venueName = '';
  let venueAddress = '';
  let venueGeolocation = '';
  let venueComment = '';
  let venueBannerImage = '';
  let venueTimezone = 'Asia/Jerusalem';
  let venueVisibility = 'private';
  let privateLinkToken = '';

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

  // Subscribe to stores
  currentOwnerStore.subscribe((val) => {
    currentOwner = val;
  });

  eventListStore.subscribe((val) => {
    allEventLists = val;
  });

  eventStore.subscribe((val) => {
    allEvents = val;
  });

  ownersStore.subscribe((val) => {
    allOwners = val;
  });

  /**
   * Save current state to undo stack
   */
  function saveUndoState() {
    const state = {
      venue: venue ? JSON.parse(JSON.stringify(venue)) : null,
      eventListsData: JSON.parse(JSON.stringify(eventListsData))
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
      loadVenueData();
    }
    eventListsData = previousState.eventListsData;
  }

  /**
   * Validate UUID format
   * @param {string} uuid
   */
  function isValidUUID(uuid) {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
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
   * Load venue data into form
   */
  function loadVenueData() {
    if (!venue) return;
    venueName = venue.name || '';
    venueAddress = venue.address || '';
    venueGeolocation = venue.geolocation || '';
    venueComment = venue.comment || '';
    venueBannerImage = venue.banner_image || '';
    // Preserve empty timezone if it exists, only default for new venues
    venueTimezone = venue.timezone !== undefined && venue.timezone !== null ? venue.timezone : 'Asia/Jerusalem';
    venueVisibility = venue.visibility || 'private';
    privateLinkToken = venue.private_link_token || '';

    // Get current values from stores (in case stores haven't updated yet)
    const currentEventLists = get(eventListStore);
    const currentEvents = get(eventStore);

    // Load event lists for this venue
    const venueEventLists = currentEventLists.filter((el) => el.venue_uuid === venue.venue_uuid);
    // Always initialize as an array, even if empty
    // Deduplicate events by event_uuid within each event list to prevent duplicates when loadVenueData() is called multiple times
    eventListsData = venueEventLists.map((el) => {
      const listEvents = currentEvents.filter((e) => el.event_uuids.includes(e.event_uuid));
      // Ensure date is set - use existing date or default to today
      const eventListDate = el.date || new Date().toISOString().split('T')[0];
      // Deduplicate events by event_uuid within this event list
      const seenEventUuids = new Set();
      const uniqueEvents = [];
      for (const e of listEvents) {
        if (!seenEventUuids.has(e.event_uuid)) {
          seenEventUuids.add(e.event_uuid);
          uniqueEvents.push(e);
        }
      }
      const eventsWithTime = uniqueEvents.map((e) => ({
        ...e,
        // Extract time from datetime for editing
        time: extractTimeFromRFC3339(e.datetime)
      }));
      return {
        ...el,
        date: eventListDate,
        events: eventsWithTime || []
      };
    });
    // Ensure eventListsData is always an array
    if (!Array.isArray(eventListsData)) {
      eventListsData = [];
    }

    // Set preview to first event list
    if (eventListsData.length > 0) {
      previewEventListId = eventListsData[0].event_list_uuid;
    }
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
   * @param {string} timeStr
   * @param {string} dateStr
   */
  function combineTimeAndDate(timeStr, dateStr) {
    try {
      const [hours, minutes] = timeStr.split(':').map(Number);
      const date = new Date(`${dateStr}T${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:00`);
      return date.toISOString();
    } catch (e) {
      return new Date().toISOString();
    }
  }

  /**
   * Add new event list
   */
  function addEventList() {
    if (!currentOwner) return;
    // For new venues, create a temporary venue UUID if needed
    let venueUuid = venue?.venue_uuid;
    if (!venueUuid) {
      venueUuid = generateUUID();
      if (!venue) {
        venue = {
          venue_uuid: venueUuid,
          owner_uuid: currentOwner.owner_uuid
        };
      } else {
        venue.venue_uuid = venueUuid;
      }
    }
    saveUndoState();
    const newList = {
      event_list_uuid: generateUUID(),
      venue_uuid: venueUuid,
      name: 'New Event List',
      date: new Date().toISOString().split('T')[0], // Today's date in ISO format
      comment: '',
      private_link_token: generateUUID(),
      event_uuids: [],
      events: [],
      created_at: getCurrentTimestamp(),
      modified_at: getCurrentTimestamp()
    };
    // Reassign to trigger Svelte reactivity
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
    if (!confirm(`Are you sure you want to delete the event list "${list.name}"? This will also delete all events in this list.`)) {
      return;
    }
    saveUndoState();
    eventListsData = eventListsData.filter((el) => el.event_list_uuid !== listUuid);
    if (previewEventListId === listUuid) {
      previewEventListId = eventListsData.length > 0 ? eventListsData[0].event_list_uuid : null;
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
    newData[index] = newData[index - 1];
    newData[index - 1] = temp;
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
    newData[index] = newData[index + 1];
    newData[index + 1] = temp;
    eventListsData = newData;
  }

  /**
   * Add event to event list
   * @param {string} listUuid
   */
  function addEvent(listUuid) {
    saveUndoState();
    const listIndex = eventListsData.findIndex((el) => el.event_list_uuid === listUuid);
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];

    const newEvent = {
      event_uuid: generateUUID(),
      event_list_uuid: listUuid,
      event_name: 'New Event',
      datetime: combineTimeAndDate('12:00', list.date),
      time: '12:00',
      comment: '',
      duration_minutes: 0,
      created_at: getCurrentTimestamp(),
      modified_at: getCurrentTimestamp()
    };

    // Create new arrays with the new event added
    const newEvents = [...list.events, newEvent];
    const newEventUuids = [...list.event_uuids, newEvent.event_uuid];

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1)
    ];
  }

  /**
   * Delete event
   * @param {string} listUuid
   * @param {string} eventUuid
   */
  function deleteEvent(listUuid, eventUuid) {
    saveUndoState();
    const listIndex = eventListsData.findIndex((el) => el.event_list_uuid === listUuid);
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];

    // Create new arrays with the event removed
    const newEvents = list.events.filter((/** @type {any} */ e) => e.event_uuid !== eventUuid);
    const newEventUuids = list.event_uuids.filter((/** @type {any} */ uuid) => uuid !== eventUuid);

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1)
    ];
  }

  /**
   * Duplicate event
   * @param {string} listUuid
   * @param {string} eventUuid
   */
  function duplicateEvent(listUuid, eventUuid) {
    saveUndoState();
    const listIndex = eventListsData.findIndex((el) => el.event_list_uuid === listUuid);
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];
    const originalEvent = list.events.find((/** @type {any} */ e) => e.event_uuid === eventUuid);
    if (!originalEvent) return;

    const newEvent = {
      ...originalEvent,
      event_uuid: generateUUID(),
      event_name: `${originalEvent.event_name} (Copy)`,
      created_at: getCurrentTimestamp(),
      modified_at: getCurrentTimestamp()
    };

    const index = list.events.findIndex((/** @type {any} */ e) => e.event_uuid === eventUuid);

    // Create new arrays with the duplicated event inserted
    const newEvents = [...list.events];
    newEvents.splice(index + 1, 0, newEvent);
    const newEventUuids = [...list.event_uuids];
    newEventUuids.splice(index + 1, 0, newEvent.event_uuid);

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1)
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
    const listIndex = eventListsData.findIndex((el) => el.event_list_uuid === listUuid);
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];

    // Create new arrays with swapped elements
    const newEvents = [...list.events];
    const temp = newEvents[eventIndex];
    newEvents[eventIndex] = newEvents[eventIndex - 1];
    newEvents[eventIndex - 1] = temp;

    const newEventUuids = [...list.event_uuids];
    const tempUuid = newEventUuids[eventIndex];
    newEventUuids[eventIndex] = newEventUuids[eventIndex - 1];
    newEventUuids[eventIndex - 1] = tempUuid;

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1)
    ];
  }

  /**
   * Move event down
   * @param {string} listUuid
   * @param {number} eventIndex
   */
  function moveEventDown(listUuid, eventIndex) {
    const listIndex = eventListsData.findIndex((el) => el.event_list_uuid === listUuid);
    if (listIndex === -1) return;
    const list = eventListsData[listIndex];
    if (eventIndex === list.events.length - 1) return;
    saveUndoState();

    // Create new arrays with swapped elements
    const newEvents = [...list.events];
    const temp = newEvents[eventIndex];
    newEvents[eventIndex] = newEvents[eventIndex + 1];
    newEvents[eventIndex + 1] = temp;

    const newEventUuids = [...list.event_uuids];
    const tempUuid = newEventUuids[eventIndex];
    newEventUuids[eventIndex] = newEventUuids[eventIndex + 1];
    newEventUuids[eventIndex + 1] = tempUuid;

    // Create new list object with new arrays
    const updatedList = {
      ...list,
      events: newEvents,
      event_uuids: newEventUuids
    };

    // Create new eventListsData array with updated list
    eventListsData = [
      ...eventListsData.slice(0, listIndex),
      updatedList,
      ...eventListsData.slice(listIndex + 1)
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
   * Handle visibility change
   * @param {string} newVisibility
   */
  function handleVisibilityChange(newVisibility) {
    saveUndoState();
    venueVisibility = newVisibility;
    if (newVisibility === 'private' && !privateLinkToken) {
      privateLinkToken = generateUUID();
    } else if (newVisibility === 'public') {
      privateLinkToken = '';
    }
  }

  /**
   * Save venue and all related data
   */
  function saveVenue() {
    if (!currentOwner) {
      alert('You must be logged in to save a venue');
      return;
    }

    // Validate required fields
    if (!venueName.trim()) {
      alert('Venue name is required');
      return;
    }

    // Validate event list dates
    for (const list of eventListsData) {
      const dateValue = list.date ? String(list.date).trim() : '';
      if (!dateValue) {
        alert(`Date is required for event list "${list.name}".`);
        return;
      }
      // type="date" inputs should return YYYY-MM-DD format
      // Validate the format
      if (!/^\d{4}-\d{2}-\d{2}$/.test(dateValue)) {
        alert(`Invalid date format for event list "${list.name}". Please use ISO 8601 format (YYYY-MM-DD). Got: "${dateValue}"`);
        return;
      }
      // Validate that it's a valid date
      const parsedDate = parseISODate(dateValue);
      if (!parsedDate) {
        alert(`Invalid date for event list "${list.name}". Please use a valid date in ISO 8601 format (YYYY-MM-DD). Got: "${dateValue}"`);
        return;
      }
    }

    // Validate event times and update datetimes
    for (const list of eventListsData) {
      for (const event of list.events) {
        const timeRegex = /^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/;
        if (!timeRegex.test(event.time)) {
          alert(`Invalid time format for event "${event.event_name}". Please use HH:MM format.`);
          return;
        }
        event.datetime = combineTimeAndDate(event.time, list.date);
      }
    }

    const now = getCurrentTimestamp();

    // Create or update venue
    if (isNewVenue && currentOwner) {
      venue = {
        venue_uuid: generateUUID(),
        owner_uuid: currentOwner.owner_uuid,
        name: sanitizeInput(venueName.trim()),
        banner_image: venueBannerImage,
        address: sanitizeInput(venueAddress.trim()),
        geolocation: sanitizeInput(venueGeolocation.trim()),
        comment: sanitizeInput(venueComment.trim()),
        event_list_uuids: eventListsData.map((el) => el.event_list_uuid),
        timezone: venueTimezone,
        visibility: venueVisibility,
        private_link_token: venueVisibility === 'private' ? privateLinkToken : undefined,
        created_at: now,
        modified_at: now
      };
    } else if (venue) {
      venue = updateModifiedTimestamp({
        ...venue,
        name: sanitizeInput(venueName.trim()),
        banner_image: venueBannerImage,
        address: sanitizeInput(venueAddress.trim()),
        geolocation: sanitizeInput(venueGeolocation.trim()),
        comment: sanitizeInput(venueComment.trim()),
        event_list_uuids: eventListsData.map((el) => el.event_list_uuid),
        timezone: venueTimezone,
        visibility: venueVisibility,
        private_link_token: venueVisibility === 'private' ? privateLinkToken : undefined
      });
    }

    // Update venue store
    if (!venue) return;
    const currentVenues = get(venueStore);
    if (isNewVenue) {
      venueStore.set([...currentVenues, venue]);
    } else {
      venueStore.set(currentVenues.map((v) => (v.venue_uuid === venue.venue_uuid ? venue : v)));
    }

    // Update event lists
    if (!venue) return;
    const currentEventLists = get(eventListStore);
    /** @type {import('$lib/types').EventList[]} */
    const updatedEventLists = eventListsData.map((listData) => {
      const existingList = currentEventLists.find((el) => el.event_list_uuid === listData.event_list_uuid);
      if (existingList) {
        // Preserve private_link_token if it exists, otherwise generate one
        const token = existingList.private_link_token || listData.private_link_token || generateUUID();
        const updated = /** @type {import('$lib/types').EventList} */ (updateModifiedTimestamp({
          ...existingList,
          name: sanitizeInput(listData.name.trim()),
          date: listData.date,
          comment: sanitizeInput(listData.comment || ''),
          event_uuids: listData.event_uuids,
          private_link_token: token
        }));
        return updated;
      } else {
        /** @type {import('$lib/types').EventList} */
        const newList = {
          event_list_uuid: listData.event_list_uuid,
          venue_uuid: venue.venue_uuid,
          name: sanitizeInput(listData.name.trim()),
          date: listData.date,
          comment: sanitizeInput(listData.comment || ''),
          private_link_token: listData.private_link_token || generateUUID(),
          event_uuids: listData.event_uuids,
          created_at: listData.created_at || now,
          modified_at: now
        };
        return newList;
      }
    });

    // Remove deleted event lists
    const existingListUuids = eventListsData.map((/** @type {any} */ el) => el.event_list_uuid);
    const listsToKeep = currentEventLists.filter(
      (el) => el.venue_uuid !== venue.venue_uuid || existingListUuids.includes(el.event_list_uuid)
    );
    /** @type {import('$lib/types').EventList[]} */
    const allEventLists = [...listsToKeep.filter((el) => el.venue_uuid !== venue.venue_uuid), ...updatedEventLists];
    eventListStore.set(allEventLists);

    // Update events
    if (!venue) return;
    const currentEvents = get(eventStore);
    const allEventUuids = eventListsData.flatMap((el) => el.event_uuids);
    const updatedEvents = eventListsData.flatMap((listData) =>
      listData.events.map((/** @type {any} */ eventData) => {
        const existingEvent = currentEvents.find((e) => e.event_uuid === eventData.event_uuid);
        if (existingEvent) {
          return updateModifiedTimestamp({
            ...existingEvent,
            event_name: sanitizeInput(eventData.event_name.trim()),
            datetime: eventData.datetime,
            comment: sanitizeInput(eventData.comment || ''),
            duration_minutes: eventData.duration_minutes || 0
          });
        } else {
          return {
            event_uuid: eventData.event_uuid,
            event_list_uuid: listData.event_list_uuid,
            event_name: sanitizeInput(eventData.event_name.trim()),
            datetime: eventData.datetime,
            comment: sanitizeInput(eventData.comment || ''),
            duration_minutes: eventData.duration_minutes || 0,
            created_at: eventData.created_at || now,
            modified_at: now
          };
        }
      })
    );

    // Deduplicate updatedEvents by event_uuid to prevent duplicates
    const uniqueUpdatedEvents = [];
    const seenEventUuids = new Set();
    for (const event of updatedEvents) {
      if (!seenEventUuids.has(event.event_uuid)) {
        seenEventUuids.add(event.event_uuid);
        uniqueUpdatedEvents.push(event);
      }
    }

    // Remove all events belonging to this venue's event lists (we'll replace them with updatedEvents)
    // Keep only events that don't belong to this venue's event lists
    const venueEventListUuids = eventListsData.map((el) => el.event_list_uuid);
    const eventsToKeep = currentEvents.filter(
      (e) => !venueEventListUuids.includes(e.event_list_uuid)
    );
    // Add all updated events for this venue's event lists
    eventStore.set([...eventsToKeep, ...uniqueUpdatedEvents]);

    // Navigate back to venue owner page
    goto('/venue-owner');
  }

  /**
   * Cancel editing
   */
  function cancelEdit() {
    if (confirm('Are you sure you want to cancel? Unsaved changes will be lost.')) {
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
            'User-Agent': 'time.place/1.0'
          }
        }
      );
      const data = await response.json();
      if (data && data.length > 0) {
        return data.map((/** @type {any} */ result) => ({
          lat: parseFloat(result.lat),
          lng: parseFloat(result.lon),
          display_name: result.display_name || result.name || 'Unknown location'
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
            'User-Agent': 'time.place/1.0'
          }
        }
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
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      maxZoom: 19
    }).addTo(geolocationMap);

    // Add marker if coordinates exist
    if (coords) {
      geolocationMarker = L.marker([coords.lat, coords.lng], { draggable: true })
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
      geolocationMarker = L.marker([coords.lat, coords.lng], { draggable: true })
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
      alert('Address not found. Please try a different address or click on the map to set location.');
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

  onMount(() => {
    currentOwner = get(currentOwnerStore);
    if (!currentOwner) {
      goto('/login');
      return;
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
      // Editing existing venue
      const allVenues = get(venueStore);
      venue = allVenues.find((v) => v.venue_uuid === venueUuidFromUrl);
      if (!venue) {
        alert('Venue not found');
        goto('/venue-owner');
        return;
      }
      // Check ownership
      if (venue.owner_uuid !== currentOwner.owner_uuid) {
        alert('You do not have permission to edit this venue');
        goto('/venue-owner');
        return;
      }
      isNewVenue = false;
      dataLoaded = false;
      // Try to load data - always call loadVenueData to initialize eventListsData
      loadVenueData();
      // Check if we successfully loaded data for this venue
      const currentEventLists = get(eventListStore);
      const venueEventLists = currentEventLists.filter((el) => el.venue_uuid === venue.venue_uuid);
      // Mark as loaded if stores have data (even if this venue has no event lists yet)
      if (currentEventLists.length > 0) {
        dataLoaded = true;
      }
    } else {
      // Creating new venue
      isNewVenue = true;
      venue = null;
      eventListsData = [];
      previewEventListId = null;
      // Default to private visibility
      venueVisibility = 'private';
      privateLinkToken = generateUUID();
    }
  });

  // Reload venue data when stores update (if we're editing an existing venue and data hasn't been loaded yet)
  // Watch allEventLists to trigger when stores are populated
  // Check if THIS venue has event lists, or if stores are populated (to attempt loading)
  $: if (venue && !isNewVenue && !dataLoaded) {
    const currentEventLists = get(eventListStore);
    const venueEventLists = currentEventLists.filter((el) => el.venue_uuid === venue.venue_uuid);
    // Trigger loading if stores are populated (even if this venue has no event lists yet)
    // This ensures we load the data when stores become available
    if (currentEventLists.length > 0 || allEventLists.length > 0) {
      loadVenueData();
      // Mark as loaded after calling loadVenueData (it always sets eventListsData, even if empty)
      dataLoaded = true;
    }
  }

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
  <title>{isNewVenue ? 'Create' : 'Edit'} Venue - time.place</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
  <div class="mb-6 flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold text-gray-900">{isNewVenue ? 'Create' : 'Edit'} Venue</h1>
      <p class="text-sm text-gray-600 mt-1">Manage your venue details, event lists, and schedules.</p>
    </div>
    <div class="flex gap-2">
      {#if undoStack.length > 0}
        <button
          on:click={handleUndo}
          class="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
          title="Undo last change"
        >
          Undo
        </button>
      {/if}
      <button
        on:click={cancelEdit}
        class="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
      >
        Cancel
      </button>
      <button
        on:click={saveVenue}
        class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
      >
        Save Venue
      </button>
    </div>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- Editing Pane -->
    <div class="bg-white rounded-xl shadow-lg p-6 space-y-6 overflow-y-auto max-h-[calc(100vh-200px)]">
      <h2 class="text-xl font-semibold text-gray-900 border-b pb-2">Edit Venue</h2>

      <!-- Basic Venue Information -->
      <div class="space-y-4">
        <h3 class="text-lg font-medium text-gray-800">Basic Information</h3>

        <div>
          <label for="venue-name" class="block text-sm font-medium text-gray-700 mb-1">Venue Name *</label>
          <input
            type="text"
            id="venue-name"
            bind:value={venueName}
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            placeholder="Enter venue name"
          />
        </div>

        <div>
          <label for="venue-address" class="block text-sm font-medium text-gray-700 mb-1">Address (optional)</label>
          <div class="flex gap-2">
            <input
              type="text"
              id="venue-address"
              bind:value={venueAddress}
              class="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Enter address"
              on:keydown={(/** @type {KeyboardEvent} */ e) => e.key === 'Enter' && handleFindOnMap()}
            />
            <button
              type="button"
              on:click={handleFindOnMap}
              disabled={isGeocoding}
              class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
              title="Find this address on the map below"
            >
              {isGeocoding ? 'Searching...' : 'Find on Map'}
            </button>
          </div>
          {#if showGeocodingResults && geocodingResults.length > 0}
            <div class="mt-2 border border-gray-300 rounded-lg bg-white shadow-lg max-h-48 overflow-auto z-10 relative">
              <div class="p-2 text-xs text-gray-600 font-medium border-b bg-gray-50">Multiple locations found. Select one:</div>
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
          <label for="venue-geolocation" class="block text-sm font-medium text-gray-700 mb-1">Geolocation (optional)</label>
          <p class="text-xs text-gray-500 mb-2">Click on the map to set location, or use "Find on Map" button above to search by address. Drag the marker to adjust.</p>

          <div class="mb-2">
            <div
              bind:this={geolocationMapContainer}
              class="w-full h-64 rounded-lg overflow-hidden border border-gray-300"
            ></div>
          </div>

          <div class="mb-2">
            <label for="venue-geolocation" class="block text-xs font-medium text-gray-700 mb-1">Coordinates (latitude,longitude)</label>
            <input
              type="text"
              id="venue-geolocation"
              bind:value={venueGeolocation}
              class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="latitude,longitude"
              on:input={() => {
                if (venueGeolocation) {
                  updateGeolocationMap();
                }
              }}
            />
          </div>
        </div>

        <div>
          <label for="venue-comment" class="block text-sm font-medium text-gray-700 mb-1">Comment (optional)</label>
          <textarea
            id="venue-comment"
            bind:value={venueComment}
            rows="3"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            placeholder="Optional comment about the venue"
          ></textarea>
        </div>

        <div>
          <label for="venue-timezone" class="block text-sm font-medium text-gray-700 mb-1">Timezone (optional, leave empty for your current timezone)</label>
          <select
            id="venue-timezone"
            bind:value={venueTimezone}
            class="w-full pl-3 pr-10 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3Csvg%20xmlns=%22http://www.w3.org/2000/svg%22%20viewBox=%220%200%2024%2024%22%20fill=%22none%22%20stroke=%22%23666%22%20stroke-width=%222%22%20stroke-linecap=%22round%22%20stroke-linejoin=%22round%22%3E%3Cpolyline%20points=%226%209%2012%2015%2018%209%22%3E%3C/polyline%3E%3C/svg%3E')] bg-no-repeat bg-[length:1.25em] bg-[position:right_0.75rem_center]"
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
        </div>

        <div>
          <div class="block text-sm font-medium text-gray-700 mb-2">Visibility</div>
          <div class="flex gap-4">
            <label class="flex items-center">
              <input
                type="radio"
                value="public"
                bind:group={venueVisibility}
                on:change={() => handleVisibilityChange('public')}
                class="mr-2"
              />
              <span>Public</span>
            </label>
            <label class="flex items-center">
              <input
                type="radio"
                value="private"
                bind:group={venueVisibility}
                on:change={() => handleVisibilityChange('private')}
                class="mr-2"
              />
              <span>Private</span>
            </label>
          </div>
          {#if venueVisibility === 'private' && privateLinkToken}
            <div class="mt-2 p-2 bg-gray-50 rounded text-xs">
              <p class="text-gray-600 mb-1">Private Link:</p>
              <code class="text-blue-600 break-all">
                {typeof window !== 'undefined' ? `${window.location.origin}/?token=${privateLinkToken}` : ''}
              </code>
            </div>
          {/if}
        </div>

        <div>
          <label for="venue-banner" class="block text-sm font-medium text-gray-700 mb-1">Banner Image (optional)</label>
          <input
            type="file"
            id="venue-banner"
            accept="image/*"
            on:change={handleImageUpload}
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
          {#if venueBannerImage}
            <div class="mt-2">
              <img src={venueBannerImage} alt="Banner preview" class="max-w-full h-32 object-cover rounded" />
            </div>
          {/if}
        </div>
      </div>

      <!-- Event Lists -->
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-medium text-gray-800">Event Lists</h3>
          <button
            on:click={addEventList}
            class="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-lg transition-colors"
          >
            + Add Event List
          </button>
        </div>

        {#if eventListsData.length === 0}
          <p class="text-sm text-gray-500 italic py-4 text-center">No event lists yet. Click "+ Add Event List" to create one.</p>
        {/if}

        {#each eventListsData as listData, listIndex (listData.event_list_uuid)}
          <div class="border border-gray-200 rounded-lg p-4 space-y-3">
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
                Delete
              </button>
            </div>

            <div>
              <label for="event-list-name-{listData.event_list_uuid}" class="block text-sm font-medium text-gray-700 mb-1">Event List Name (optional)</label>
              <input
                type="text"
                id="event-list-name-{listData.event_list_uuid}"
                bind:value={listData.name}
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="Event list name"
              />
            </div>

            <div>
              <label for="event-list-date-{listData.event_list_uuid}" class="block text-sm font-medium text-gray-700 mb-1">Date (ISO 8601: YYYY-MM-DD)</label>
              <input
                type="date"
                id="event-list-date-{listData.event_list_uuid}"
                bind:value={listData.date}
                on:input={() => {
                  // Update datetime for all events in this list when date changes
                  listData.events.forEach((/** @type {any} */ event) => {
                    event.datetime = combineTimeAndDate(event.time, listData.date);
                  });
                }}
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
              <p class="mt-1 text-xs text-gray-500">Date is stored in ISO 8601 format (YYYY-MM-DD)</p>
            </div>

            <div>
              <label for="event-list-comment-{listData.event_list_uuid}" class="block text-sm font-medium text-gray-700 mb-1">Comment (optional)</label>
              <textarea
                id="event-list-comment-{listData.event_list_uuid}"
                bind:value={listData.comment}
                rows="2"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="Optional comment"
              ></textarea>
            </div>

            <div>
              <div class="flex items-center justify-between mb-2">
                <div class="block text-sm font-medium text-gray-700">Events</div>
                <button
                  on:click={() => addEvent(listData.event_list_uuid)}
                  class="px-2 py-1 bg-green-600 hover:bg-green-700 text-white text-sm rounded transition-colors"
                >
                  + Add Event
                </button>
              </div>

              {#if !listData.events || listData.events.length === 0}
                <p class="text-sm text-gray-500 italic py-2">No events yet. Click "+ Add Event" to add one.</p>
              {/if}
              {#each (listData.events || []) as event, eventIndex (event.event_uuid)}
                <div class="border border-gray-200 rounded p-3 mb-2 space-y-2">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-1">
                      <button
                        on:click={() => moveEventUp(listData.event_list_uuid, eventIndex)}
                        disabled={eventIndex === 0}
                        class="p-1 text-gray-500 hover:text-gray-700 disabled:opacity-50 text-xs"
                        title="Move up"
                      >
                        ↑
                      </button>
                      <button
                        on:click={() => moveEventDown(listData.event_list_uuid, eventIndex)}
                        disabled={eventIndex === (listData.events || []).length - 1}
                        class="p-1 text-gray-500 hover:text-gray-700 disabled:opacity-50 text-xs"
                        title="Move down"
                      >
                        ↓
                      </button>
                    </div>
                    <div class="flex gap-1">
                      <button
                        on:click={() => duplicateEvent(listData.event_list_uuid, event.event_uuid)}
                        class="px-2 py-1 bg-blue-500 hover:bg-blue-600 text-white text-xs rounded transition-colors"
                      >
                        Duplicate
                      </button>
                      <button
                        on:click={() => deleteEvent(listData.event_list_uuid, event.event_uuid)}
                        class="px-2 py-1 bg-red-600 hover:bg-red-700 text-white text-xs rounded transition-colors"
                      >
                        Delete
                      </button>
                    </div>
                  </div>

                  <div>
                    <label for="event-name-{event.event_uuid}" class="block text-xs font-medium text-gray-700 mb-1">Event Name (optional)</label>
                    <input
                      type="text"
                      id="event-name-{event.event_uuid}"
                      bind:value={event.event_name}
                      class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="Event name"
                    />
                  </div>

                  <div>
                    <label for="event-time-{event.event_uuid}" class="block text-xs font-medium text-gray-700 mb-1">Time (HH:MM)</label>
                    <input
                      type="time"
                      id="event-time-{event.event_uuid}"
                      bind:value={event.time}
                      on:input={() => {
                        // Update datetime immediately when time changes
                        event.datetime = combineTimeAndDate(event.time, listData.date);
                      }}
                      class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div>
                    <label for="event-duration-{event.event_uuid}" class="block text-xs font-medium text-gray-700 mb-1">Duration (minutes) (optional)</label>
                    <input
                      type="number"
                      id="event-duration-{event.event_uuid}"
                      bind:value={event.duration_minutes}
                      min="0"
                      class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div>
                    <label for="event-comment-{event.event_uuid}" class="block text-xs font-medium text-gray-700 mb-1">Comment (optional)</label>
                    <textarea
                      id="event-comment-{event.event_uuid}"
                      bind:value={event.comment}
                      rows="2"
                      class="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="Optional comment"
                    ></textarea>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Preview Pane -->
    <div class="bg-white rounded-xl shadow-lg p-6 space-y-4 overflow-y-auto max-h-[calc(100vh-200px)]">
      <h2 class="text-xl font-semibold text-gray-900 border-b pb-2">Live Preview</h2>

      {#if previewEventList && eventListsData.length > 1}
        <div>
          <label for="preview-event-list" class="block text-sm font-medium text-gray-700 mb-1">Select Event List</label>
          <select
            id="preview-event-list"
            bind:value={previewEventListId}
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            {#each eventListsData as listData}
              <option value={listData.event_list_uuid}>{listData.name}</option>
            {/each}
          </select>
        </div>
      {/if}

      {#if venueName || previewEventList}
        <div class="space-y-4">
          {#if venueBannerImage}
            <div>
              <img src={venueBannerImage} alt={venueName || 'Venue'} class="w-full h-48 object-cover rounded-lg" />
            </div>
          {/if}

          <h2 class="text-3xl font-bold text-gray-900">{venueName || 'Venue Name'}</h2>

          {#if venueAddress}
            <p class="text-sm text-gray-600">
              <span class="font-medium">Address:</span> {venueAddress}
            </p>
          {/if}

          {#if venueGeolocation}
            <p class="text-sm text-gray-600">
              <span class="font-medium">Geolocation:</span> {venueGeolocation}
            </p>
          {/if}

          {#if venueTimezone}
            <p class="text-sm text-gray-600">
              <span class="font-medium">Timezone:</span> {venueTimezone}
            </p>
          {/if}

          {#if previewVenueOwner}
            <div class="text-sm text-gray-600">
              <span class="font-medium">Contact:</span>
              {#if previewVenueOwner.name}
                <span class="ml-2">{previewVenueOwner.name}</span>
              {/if}
              {#if previewVenueOwner.email}
                <a href="mailto:{previewVenueOwner.email}" class="ml-2 text-blue-600 hover:text-blue-800">
                  {previewVenueOwner.email}
                </a>
              {/if}
              {#if previewVenueOwner.mobile}
                <a href="tel:{previewVenueOwner.mobile}" class="ml-2 text-blue-600 hover:text-blue-800">
                  {previewVenueOwner.mobile}
                </a>
              {/if}
            </div>
          {/if}

          {#if venueComment}
            <p class="text-sm text-gray-600 italic">{venueComment}</p>
          {/if}

          {#if previewEventList}
            <div class="mt-6">
              <h3 class="text-2xl font-semibold mb-4 text-gray-900">{previewEventList.name}</h3>
              {#if previewEventList.comment}
                <p class="text-gray-600 mb-4 italic">{previewEventList.comment}</p>
              {/if}

              {#if previewEvents.length === 0}
                <p class="text-gray-500 italic">No events scheduled for this list.</p>
              {:else}
                <div class="space-y-3">
                  {#each previewEvents as event}
                    <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                      <div>
                        <p class="font-medium text-gray-900">{event.event_name}</p>
                        {#if event.comment}
                          <p class="text-sm text-gray-600 italic">{event.comment}</p>
                        {/if}
                      </div>
                      <div class="text-right">
                        <p class="text-lg font-semibold text-blue-600">
                          {formatEventTime(Math.floor(new Date(combineTimeAndDate(event.time, previewEventList.date)).getTime() / 1000), venueTimezone ? { timeZone: venueTimezone } : {})}
                        </p>
                        {#if event.duration_minutes}
                          <p class="text-xs text-gray-500">{event.duration_minutes} min</p>
                        {/if}
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {:else}
            <div class="mt-6 p-4 bg-gray-50 rounded-lg">
              <p class="text-gray-500 text-center italic">Add an event list to see the preview.</p>
            </div>
          {/if}
        </div>
      {:else}
        <p class="text-gray-500 italic text-center">Start editing to see the preview.</p>
      {/if}
    </div>
  </div>
</div>
