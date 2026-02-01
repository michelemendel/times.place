<script>
  import { onMount, afterUpdate, onDestroy, tick } from 'svelte';
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/stores';
  import { dev, browser } from '$app/environment';
  import { formatEventTime, formatModifiedAt } from '$lib/utils/datetime.js';
  import BannerImage from '$lib/BannerImage.svelte';
  import {
    listPublicVenues,
    getPublicEventListsForVenue,
    getPrivateVenueByToken,
    getPrivateEventListByToken,
    getPublicEventsForEventList,
  } from '$lib/api/public.js';

  /** @type {string | null} */
  let selectedVenueId = null;
  /** @type {string | null} */
  let selectedEventListId = null;

  /** @type {import('$lib/types').Venue[]} */
  let venues = [];
  /** @type {Record<string, import('$lib/types').EventList[]>} */
  let venueEventListsMap = {};
  /** @type {Record<string, import('$lib/types').Event[]>} */
  let eventListEventsMap = {};

  /** @type {import('$lib/types').Venue | null} */
  let tokenVenue = null;
  /** @type {import('$lib/types').EventList | null} */
  let tokenEventList = null;
  /** @type {import('$lib/types').Event[] | null} */
  let tokenEvents = null;

  /** @type {HTMLElement | null} */
  let mapContainer = null;
  /** @type {any} */
  let map = null;
  /** @type {any} */
  let marker = null;
  let leafletLoaded = false;

  // Searchable dropdown state
  let venueSearchQuery = '';
  let showVenueDropdown = false;
  let highlightedVenueIndex = -1;
  /** @type {HTMLElement | null} */
  let venueDropdownRef = null;
  /** @type {HTMLInputElement | null} */
  let searchInputRef = null;
  /** @type {{ top: string; left: string; width: string } | null} */
  let dropdownPosition = null;
  /** @type {(() => void) | null} */
  let dropdownScrollCleanup = null;
  /** @type {number} */
  let dropdownOpenedAt = 0;

  // Loading & error state
  let isLoadingVenues = false;
  let isLoadingListsAndEvents = false;
  let loadError = '';

  /**
   * Get timestamp string from event list (API may use snake_case or camelCase).
   * @param {import('$lib/types').EventList & { modifiedAt?: string; createdAt?: string }} el
   * @returns {string}
   */
  function getEventListTimestamp(el) {
    return el?.modified_at ?? el?.modifiedAt ?? el?.created_at ?? el?.createdAt ?? '';
  }

  // Private token from URL (only in browser to avoid prerender issues)
  $: privateLinkToken =
    browser && $page.url.searchParams
      ? $page.url.searchParams.get('token') || null
      : null;

  /**
   * Load public venues list (optionally with search query).
   * This drives the main dropdown.
   * @param {string} [query]
   */
  async function loadVenues(query) {
    isLoadingVenues = true;
    loadError = '';
    try {
      venues = await listPublicVenues(query ?? '');
    } catch (err) {
      console.error('Failed to load venues', err);
      loadError = 'Failed to load venues. Please try again.';
    } finally {
      isLoadingVenues = false;
    }
  }

  /**
   * When we have a token, resolve either a private venue or private event list.
   */
  async function resolveTokenIfPresent() {
    if (!privateLinkToken) return;

    loadError = '';
    isLoadingListsAndEvents = true;
    tokenVenue = null;
    tokenEventList = null;
    tokenEvents = null;

    try {
      // Try venue-by-token first
      try {
        const venueResult = await getPrivateVenueByToken(privateLinkToken);
        tokenVenue = venueResult.venue;
        const venue = venueResult.venue;
        // Merge into venues list if not already there
        const exists = venues.find((v) => v.venue_uuid === venue.venue_uuid);
        if (!exists) {
          venues = [...venues, venue];
        }
        // Populate event lists map for this venue
        venueEventListsMap[venue.venue_uuid] = venueResult.event_lists || [];
        selectedVenueId = venue.venue_uuid;

        // Auto-select first event list, if any
        if (venueResult.event_lists && venueResult.event_lists.length > 0) {
          selectedEventListId = venueResult.event_lists[0].event_list_uuid;
        }
        return;
      } catch (e) {
        // If venue token failed with 404 or similar, fall through to event-list token
      }

      // Try event-list-by-token
      const listResult = await getPrivateEventListByToken(privateLinkToken);
      tokenVenue = listResult.venue;
      tokenEventList = listResult.event_list;
      tokenEvents = listResult.events || [];

      // Merge into venues & maps
      if (tokenVenue) {
        const venue = tokenVenue;
        const exists = venues.find((v) => v.venue_uuid === venue.venue_uuid);
        if (!exists) {
          venues = [...venues, venue];
        }
        venueEventListsMap[venue.venue_uuid] = [
          ...(venueEventListsMap[venue.venue_uuid] || []),
          tokenEventList,
        ];
        eventListEventsMap[tokenEventList.event_list_uuid] = tokenEvents;
        selectedVenueId = venue.venue_uuid;
        selectedEventListId = tokenEventList.event_list_uuid;
      }
    } catch (err) {
      console.error('Failed to resolve private link token', err);
      loadError = 'Private link is invalid or expired.';
    } finally {
      isLoadingListsAndEvents = false;
    }
  }

  /**
   * Ensure we have event lists for the currently selected venue.
   * Uses the public venue event-lists endpoint.
   * @param {string} venueUuid
   */
  async function ensureEventListsForVenue(venueUuid) {
    if (venueEventListsMap[venueUuid]) return;
    isLoadingListsAndEvents = true;
    loadError = '';
    try {
      const lists = await getPublicEventListsForVenue(venueUuid);
      venueEventListsMap = {
        ...venueEventListsMap,
        [venueUuid]: lists,
      };
    } catch (err) {
      console.error('Failed to load event lists', err);
      loadError = 'Failed to load event lists. Please try again.';
    } finally {
      isLoadingListsAndEvents = false;
    }
  }

  /**
   * Ensure we have events for a given event list.
   * Fetches events for public event lists via the public API endpoint.
   * @param {string} eventListUuid
   */
  async function ensureEventsForEventList(eventListUuid) {
    // If we already have events for this list, return them
    if (eventListEventsMap[eventListUuid]) {
      return eventListEventsMap[eventListUuid];
    }

    // Try to fetch events for this public event list
    try {
      const events = await getPublicEventsForEventList(eventListUuid);
      eventListEventsMap = {
        ...eventListEventsMap,
        [eventListUuid]: events,
      };
      return events;
    } catch (err) {
      console.error('Failed to load events for event list', eventListUuid, err);
      // Return empty array if fetch fails (event list might be private or not found)
      return [];
    }
  }

  /** Refetch venues when page becomes visible (e.g. user returns from adding a venue). */
  function handleVisibilityChange() {
    if (
      typeof document !== 'undefined' &&
      document.visibilityState === 'visible'
    ) {
      loadVenues(venueSearchQuery || undefined);
    }
  }

  // Refetch venues whenever user navigates to home (e.g. after adding a venue and clicking Home)
  afterNavigate(({ to }) => {
    const path = to?.url?.pathname ?? '';
    if (browser && path === '/') {
      loadVenues(venueSearchQuery || undefined);
    }
  });

  onMount(async () => {
    // Initial venues load (no search)
    await loadVenues();

    // If a token is present, resolve it and update selection.
    if (privateLinkToken) {
      await resolveTokenIfPresent();
    }

    // Refetch venues when user returns to this tab so new venues show in dropdown
    if (browser && typeof document !== 'undefined') {
      document.addEventListener('visibilitychange', handleVisibilityChange);
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
        if (mapContainer && selectedVenue) {
          initMap();
        }
      };
      document.head.appendChild(script);
    } else if (typeof window !== 'undefined' && win.L) {
      leafletLoaded = true;
    }

    // Handle click outside for dropdown
    if (browser) {
      document.addEventListener('click', handleClickOutside);
    }

    // Close dropdown on scroll only after a short delay so opening on mobile does not close it immediately
    if (browser && typeof window !== 'undefined') {
      const closeDropdownOnScroll = () => {
        if (
          showVenueDropdown &&
          dropdownOpenedAt > 0 &&
          Date.now() - dropdownOpenedAt > 200
        ) {
          showVenueDropdown = false;
          highlightedVenueIndex = -1;
        }
      };
      window.addEventListener('scroll', closeDropdownOnScroll, true);
      dropdownScrollCleanup = () => {
        window.removeEventListener('scroll', closeDropdownOnScroll, true);
      };
    }
  });

  onDestroy(() => {
    if (browser && typeof document !== 'undefined') {
      document.removeEventListener('click', handleClickOutside);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    }
    dropdownScrollCleanup?.();
  });

  // Initialize or update map when venue changes
  afterUpdate(() => {
    if (leafletLoaded && mapContainer && selectedVenue && !map) {
      initMap();
    } else if (leafletLoaded && map && selectedVenue) {
      updateMap();
    } else if (map && !selectedVenue) {
      // Clean up map when no venue is selected
      if (marker) {
        map.removeLayer(marker);
        marker = null;
      }
      if (map) {
        map.remove();
        map = null;
      }
    }
    if (showVenueDropdown && searchInputRef) {
      updateDropdownPosition();
    }
  });

  // Visible venues (public list plus any token-derived venue)
  $: visibleVenues = venues;

  // Sorted venues
  $: sortedVenues = [...visibleVenues].sort((a, b) =>
    a.name.localeCompare(b.name),
  );

  // When user searches, the backend returns venues matching query across venue name, address, comment, event list names, and event names. We use that list as-is.
  $: filteredVenues = sortedVenues;

  // Position fixed dropdown when it opens so it is not clipped by overflow
  $: if (showVenueDropdown) {
    tick().then(updateDropdownPosition);
  } else {
    dropdownPosition = null;
  }

  // Get selected venue
  $: selectedVenue = selectedVenueId
    ? venues.find((v) => v.venue_uuid === selectedVenueId)
    : null;

  // Event lists for selected venue (public lists or token-derived lists)
  $: selectedVenueEventLists =
    selectedVenue && venueEventListsMap[selectedVenue.venue_uuid]
      ? venueEventListsMap[selectedVenue.venue_uuid]
      : [];

  // Auto-select first event list if none selected yet
  $: {
    if (selectedVenue && selectedVenueEventLists.length > 0) {
      if (!selectedEventListId) {
        selectedEventListId = selectedVenueEventLists[0].event_list_uuid;
      }
    } else if (!selectedVenue || selectedVenueEventLists.length === 0) {
      selectedEventListId = null;
    }
  }

  // Get selected event list
  $: selectedEventList = selectedEventListId
    ? selectedVenueEventLists.find(
        (el) => el.event_list_uuid === selectedEventListId,
      ) || null
    : null;

  // Events for selected list
  /** @type {import('$lib/types').Event[]} */
  let listEvents = /** @type {import('$lib/types').Event[]} */ ([]);
  let isLoadingEvents = false;

  // Reactive statement to load events when event list changes
  $: {
    if (selectedEventList) {
      isLoadingEvents = true;
      ensureEventsForEventList(selectedEventList.event_list_uuid)
        .then((events) => {
          listEvents = /** @type {import('$lib/types').Event[]} */ (events);
          isLoadingEvents = false;
        })
        .catch((err) => {
          console.error('Failed to load events', err);
          listEvents = /** @type {import('$lib/types').Event[]} */ ([]);
          isLoadingEvents = false;
        });
    } else {
      listEvents = /** @type {import('$lib/types').Event[]} */ ([]);
      isLoadingEvents = false;
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
   * Generate Google Maps directions URL
   * Prefers coordinates if available (more accurate), otherwise uses address
   * @param {import('$lib/types').Venue | null} venue
   * @returns {string | null}
   */
  function getDirectionsUrl(venue) {
    if (!venue) return null;

    // Prefer coordinates if available
    const coords = parseGeolocation(venue.geolocation);
    if (coords) {
      return `https://www.google.com/maps/dir/?api=1&destination=${coords.lat},${coords.lng}`;
    }

    // Fall back to address if available
    if (venue.address) {
      const encodedAddress = encodeURIComponent(venue.address);
      return `https://www.google.com/maps/search/?api=1&query=${encodedAddress}`;
    }

    return null;
  }

  /**
   * Initialize the map
   */
  function initMap() {
    /** @type {any} */
    const win = window;
    const L = win.L;
    if (!mapContainer || !L || !selectedVenue) return;

    const coords = parseGeolocation(selectedVenue.geolocation);
    if (!coords) return;

    map = L.map(mapContainer).setView([coords.lat, coords.lng], 15);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      maxZoom: 19,
    }).addTo(map);

    marker = L.marker([coords.lat, coords.lng])
      .addTo(map)
      .bindPopup(selectedVenue.name || 'Venue Location')
      .openPopup();
  }

  /**
   * Update map when venue changes
   */
  function updateMap() {
    /** @type {any} */
    const win = window;
    const L = win.L;
    if (!map || !selectedVenue || !L) return;

    const coords = parseGeolocation(selectedVenue.geolocation);
    if (!coords) return;

    map.setView([coords.lat, coords.lng], 15);

    if (marker) {
      map.removeLayer(marker);
    }

    marker = L.marker([coords.lat, coords.lng])
      .addTo(map)
      .bindPopup(selectedVenue.name || 'Venue Location')
      .openPopup();
  }

  // Convert RFC3339 datetime string to Unix timestamp (seconds)
  /**
   * @param {string} rfc3339
   * @returns {number}
   */
  function rfc3339ToUnixTimestamp(rfc3339) {
    return Math.floor(new Date(rfc3339).getTime() / 1000);
  }

  // Format event time from RFC3339 string
  /**
   * @param {string} rfc3339
   * @param {string} [venueTimezone] - Optional venue timezone to use for display
   * @returns {string}
   */
  function formatEventTimeFromRFC3339(rfc3339, venueTimezone) {
    const unixTimestamp = rfc3339ToUnixTimestamp(rfc3339);
    const timezoneToUse =
      venueTimezone && typeof venueTimezone === 'string' && venueTimezone.trim()
        ? venueTimezone.trim()
        : undefined;

    return formatEventTime(
      unixTimestamp,
      timezoneToUse ? { timeZone: timezoneToUse } : {},
    );
  }

  /**
   * @param {string | null} venueId
   */
  async function selectVenue(venueId) {
    selectedVenueId = venueId || null;
    selectedEventListId = null;
    const venue = venueId ? venues.find((v) => v.venue_uuid === venueId) : null;
    venueSearchQuery = venue ? venue.name : '';
    showVenueDropdown = false;
    highlightedVenueIndex = -1;

    if (selectedVenueId) {
      await ensureEventListsForVenue(selectedVenueId);
    }
  }

  /**
   * @param {Event} event
   */
  async function handleVenueSearchInput(event) {
    const target = /** @type {HTMLInputElement} */ (event.target);
    const newValue = target.value;
    venueSearchQuery = newValue;

    // If the search field is cleared, reset the selection
    if (!newValue || newValue.trim() === '') {
      selectedVenueId = null;
      selectedEventListId = null;
      await loadVenues();
    } else {
      await loadVenues(newValue);
    }

    showVenueDropdown = true;
    dropdownOpenedAt = Date.now();
    highlightedVenueIndex = -1;
    if (browser && typeof requestAnimationFrame !== 'undefined') {
      requestAnimationFrame(() => updateDropdownPosition());
    }
  }

  function handleVenueSearchFocus() {
    showVenueDropdown = true;
    dropdownOpenedAt = Date.now();
    highlightedVenueIndex = -1;
    if (browser && typeof requestAnimationFrame !== 'undefined') {
      requestAnimationFrame(() => updateDropdownPosition());
    }
  }

  /**
   * @param {KeyboardEvent} event
   */
  async function handleVenueSearchKeydown(event) {
    if (!showVenueDropdown || filteredVenues.length === 0) {
      if (event.key === 'ArrowDown' && filteredVenues.length > 0) {
        showVenueDropdown = true;
        dropdownOpenedAt = Date.now();
        highlightedVenueIndex = 0;
        event.preventDefault();
      }
      return;
    }

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        highlightedVenueIndex =
          (highlightedVenueIndex + 1) % filteredVenues.length;
        break;
      case 'ArrowUp':
        event.preventDefault();
        if (highlightedVenueIndex <= 0) {
          highlightedVenueIndex = filteredVenues.length - 1;
        } else {
          highlightedVenueIndex--;
        }
        break;
      case 'Enter':
        event.preventDefault();
        if (
          highlightedVenueIndex >= 0 &&
          highlightedVenueIndex < filteredVenues.length
        ) {
          await selectVenue(filteredVenues[highlightedVenueIndex].venue_uuid);
        }
        break;
      case 'Escape':
        event.preventDefault();
        showVenueDropdown = false;
        highlightedVenueIndex = -1;
        break;
    }
  }

  async function clearVenueSearch() {
    venueSearchQuery = '';
    selectedVenueId = null;
    selectedEventListId = null;
    showVenueDropdown = false;
    await loadVenues();
  }

  /**
   * @param {Event} event
   */
  function handleClickOutside(event) {
    if (
      venueDropdownRef &&
      !venueDropdownRef.contains(/** @type {Node} */ (event.target))
    ) {
      showVenueDropdown = false;
      highlightedVenueIndex = -1;
    }
  }

  /** Update fixed dropdown position from search input rect (so dropdown is not clipped by overflow). */
  function updateDropdownPosition() {
    if (!browser || !searchInputRef || !showVenueDropdown) {
      dropdownPosition = null;
      return;
    }
    const rect = searchInputRef.getBoundingClientRect();
    dropdownPosition = {
      top: `${rect.bottom + 4}px`,
      left: `${rect.left}px`,
      width: `${rect.width}px`,
    };
  }

  /**
   * @param {Event} event
   */
  function handleEventListChange(event) {
    const target = /** @type {HTMLSelectElement} */ (event.target);
    selectedEventListId = target.value || null;
  }
</script>

<svelte:head>
  <title>Find Venues and Event Times</title>
  <link
    rel="stylesheet"
    href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"
    integrity="sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY="
    crossorigin=""
  />
</svelte:head>

<div
  class="w-full min-w-0 max-w-[100vw] overflow-x-clip lg:max-w-[60%] lg:mx-auto"
>
  <div class="-mt-1 md:mt-0 mb-2 md:mb-2 text-center no-print-header">
    <h1 class="text-[16px] md:text-4xl font-bold mb-0.5 md:mb-2 text-gray-900">
      Find Venues and Events
    </h1>
    <p class="text-[10px] md:text-base text-gray-600 mb-1 md:mb-0">
      Select a venue to view its event schedules and contact information.
    </p>
  </div>

  <div
    class="bg-white rounded-xl shadow-lg p-2 md:p-12 md:pt-4 min-w-0 max-w-full w-full overflow-x-clip box-border"
  >
    <!-- Venue Searchable Dropdown (no overflow-hidden here so dropdown is not clipped) -->
    <div
      class="mb-2 md:mb-2 relative no-print-venue-dropdown"
      bind:this={venueDropdownRef}
    >
      <div class="relative">
        <input
          type="text"
          id="venue-search"
          bind:this={searchInputRef}
          value={venueSearchQuery}
          on:input={handleVenueSearchInput}
          on:focus={handleVenueSearchFocus}
          on:keydown={handleVenueSearchKeydown}
          placeholder="Search and select a venue..."
          class="w-full max-w-full px-3 md:px-4 py-1.5 md:py-2 pr-10 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white text-[14px] md:text-base text-gray-900 box-border"
        />
        {#if venueSearchQuery}
          <button
            type="button"
            on:click={clearVenueSearch}
            class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600"
            aria-label="Clear search"
          >
            <svg
              class="w-5 h-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
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
        <!-- Dropdown: fixed when position set (not clipped), else absolute below input so list shows immediately -->
        {#if showVenueDropdown}
          <div
            class="z-[9999] bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-auto text-[14px] md:text-base {dropdownPosition
              ? 'fixed'
              : 'absolute left-0 top-full w-full mt-1'}"
            style={dropdownPosition
              ? `top: ${dropdownPosition.top}; left: ${dropdownPosition.left}; width: ${dropdownPosition.width};`
              : ''}
            role="listbox"
          >
            {#if filteredVenues.length > 0}
              {#each filteredVenues as venue, index}
                <button
                  type="button"
                  on:click={() => selectVenue(venue.venue_uuid)}
                  on:mouseenter={() => (highlightedVenueIndex = index)}
                  class="w-full text-left px-4 py-2 focus:outline-none text-[14px] md:text-base {highlightedVenueIndex ===
                  index
                    ? 'bg-blue-100'
                    : selectedVenueId === venue.venue_uuid
                      ? 'bg-blue-50'
                      : 'hover:bg-gray-100'}"
                  role="option"
                  aria-selected={highlightedVenueIndex === index}
                >
                  {venue.name}
                </button>
              {/each}
            {:else if venueSearchQuery}
              <div class="px-4 py-2 text-gray-500 text-[14px] md:text-sm">
                No venues found
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </div>

    {#if selectedVenue}
      <!-- Venue Details (overflow-x-clip only here so dropdown above is not clipped) -->
      <div
        class="pt-0 md:pt-2 md:mt-0 min-w-0 max-w-[100vw] overflow-x-clip w-full"
      >
        <!-- Banner Image -->
        {#if selectedVenue.banner_image}
          <BannerImage
            src={selectedVenue.banner_image}
            alt={selectedVenue.name}
            size="md"
            wrapperClass="mb-4"
          />
        {/if}

        <!-- Venue Name -->
        <h2
          class="text-[14px] md:text-3xl font-bold mb-1 md:mb-3 text-gray-900 break-words min-w-0"
        >
          {selectedVenue.name}
        </h2>

        <!-- Address, Comment, and Map -->
        {#if selectedVenue.address || selectedVenue.geolocation || selectedVenue.comment}
          <div
            class="mb-1 md:mb-3 grid grid-cols-1 md:grid-cols-2 gap-1 md:gap-4 min-w-0 max-w-full overflow-hidden"
          >
            <div class="flex flex-col justify-start min-w-0 gap-0 md:gap-0">
              {#if selectedVenue.owner_name}
                <p class="text-[12px] md:text-sm text-gray-600 mb-0 md:mb-1">
                  <span class="font-medium text-gray-700">Admin:</span>
                  {selectedVenue.owner_name}
                </p>
              {/if}
              {#if selectedVenue.owner_email}
                <p
                  class="text-[12px] md:text-sm text-gray-600 mb-0 md:mb-1 break-all min-w-0"
                >
                  <a
                    href="mailto:{selectedVenue.owner_email}"
                    class="text-blue-600 hover:text-blue-800 hover:underline break-all"
                  >
                    {selectedVenue.owner_email}
                  </a>
                </p>
              {/if}
              {#if selectedVenue.address}
                <p
                  class="text-[12px] md:text-sm text-gray-600 mb-0 md:mb-2 break-words min-w-0"
                >
                  <span class="font-medium text-gray-700">Address:</span>
                  {#if getDirectionsUrl(selectedVenue)}
                    <a
                      href={getDirectionsUrl(selectedVenue)}
                      target="_blank"
                      rel="noopener noreferrer"
                      class="text-blue-600 hover:text-blue-800 hover:underline break-all"
                    >
                      {selectedVenue.address}
                    </a>
                  {:else}
                    {selectedVenue.address}
                  {/if}
                </p>
              {/if}
              {#if selectedVenue.geolocation}
                <p
                  class="hidden md:block text-[12px] md:text-sm text-gray-600 mb-0 md:mb-2 no-print-geolocation"
                >
                  <span class="font-medium text-gray-700">Geolocation:</span>
                  {selectedVenue.geolocation}
                </p>
              {/if}
              {#if selectedVenue.timezone}
                <p
                  class="text-[12px] md:text-sm text-gray-600 mb-0 md:mb-2 no-print-timezone"
                >
                  <span class="font-medium text-gray-700">Timezone:</span>
                  {selectedVenue.timezone}
                </p>
              {/if}
              {#if selectedVenue.comment}
                <p
                  class="text-[12px] md:text-sm text-gray-600 whitespace-pre-line break-words mt-0"
                >
                  {selectedVenue.comment}
                </p>
              {/if}
            </div>
            {#if selectedVenue.geolocation}
              <div
                class="hidden md:block w-full h-48 rounded-lg overflow-hidden border border-gray-300 no-print-map"
              >
                <div bind:this={mapContainer} class="w-full h-full"></div>
              </div>
            {/if}
          </div>
        {/if}

        <!-- Event Lists Section -->
        {#if selectedVenueEventLists.length === 0}
          <div class="mt-3 md:mt-6 p-4 bg-gray-50 rounded-lg">
            <p class="text-gray-500 text-center">
              This venue currently has no event schedules available.
            </p>
          </div>
        {:else if selectedVenueEventLists.length === 1}
          <!-- Single event list - no selector needed -->
          {#if selectedEventList}
            <div class="mt-3 md:mt-6 min-w-0 max-w-full overflow-hidden">
              <h3
                class="text-[14px] md:text-2xl font-semibold mb-0 md:mb-1 text-gray-900 break-words min-w-0"
              >
                {selectedEventList.name}
              </h3>
              {#if getEventListTimestamp(selectedEventList)}
                <p class="text-xs text-gray-600 mb-1 md:mb-2">
                  Modified: {formatModifiedAt(getEventListTimestamp(selectedEventList))}
                </p>
              {/if}
              {#if selectedEventList.comment}
                <p
                  class="text-xs text-gray-600 mb-2 md:mb-4 whitespace-pre-line break-words min-w-0 overflow-wrap-anywhere"
                >
                  {selectedEventList.comment}
                </p>
              {/if}

              {#if listEvents.length === 0}
                <p class="text-gray-500">No events scheduled for this list.</p>
              {:else}
                <div class="space-y-0.5 md:space-y-1 min-w-0 overflow-hidden">
                  {#each listEvents as event}
                    <div
                      class="flex items-center justify-between gap-2 py-0.5 md:py-1 px-2 md:px-3 bg-gray-50 rounded-lg min-w-0 overflow-hidden"
                    >
                      <div class="min-w-0 flex-1">
                        <p
                          class="font-medium text-gray-900 text-sm break-words"
                        >
                          {event.event_name}
                        </p>
                        {#if event.comment}
                          <p
                            class="text-[12px] md:text-xs text-gray-600 whitespace-pre-line mt-0 md:mt-0.5 break-words"
                          >
                            {event.comment}
                          </p>
                        {/if}
                      </div>
                      <div class="text-right flex-shrink-0 min-w-0">
                        <p
                          class="text-[14px] md:text-base font-semibold text-blue-600"
                        >
                          {formatEventTimeFromRFC3339(
                            event.datetime,
                            selectedVenue?.timezone,
                          )}
                        </p>
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
          {/if}
        {:else}
          <!-- Multiple event lists - show selector (same as venue-form preview) -->
          <div class="mt-3 md:mt-6 min-w-0 max-w-full overflow-hidden px-2 md:px-3">
            <div class="no-print-event-list-selector">
              <label
                for="event-list-select"
                class="block text-[10px] md:text-sm font-medium text-gray-700 mb-0.5 md:mb-1"
                >Select Event List</label
              >
              <select
                id="event-list-select"
                on:change={handleEventListChange}
                class="w-full px-2 md:px-3 py-1.5 md:py-2 text-[12px] md:text-base border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white text-gray-900 mb-3 md:mb-4 box-border"
              >
                {#each selectedVenueEventLists as eventList}
                  <option
                    value={eventList.event_list_uuid}
                    selected={selectedEventListId === eventList.event_list_uuid}
                  >
                    {eventList.name}
                  </option>
                {/each}
              </select>
            </div>

            {#if selectedEventList}
              <div class="min-w-0 max-w-full overflow-hidden">
                <h3
                  class="text-[14px] md:text-2xl font-semibold mb-0 md:mb-1 text-gray-900 break-words min-w-0"
                >
                  {selectedEventList.name}
                </h3>
                {#if getEventListTimestamp(selectedEventList)}
                  <p class="text-xs text-gray-600 mb-1 md:mb-2">
                    Modified: {formatModifiedAt(getEventListTimestamp(selectedEventList))}
                  </p>
                {/if}
                {#if selectedEventList.comment}
                  <p
                    class="text-xs text-gray-600 mb-2 md:mb-4 whitespace-pre-line break-words min-w-0 overflow-wrap-anywhere"
                  >
                    {selectedEventList.comment}
                  </p>
                {/if}

                {#if listEvents.length === 0}
                  <p class="text-gray-500">
                    No events scheduled for this list.
                  </p>
                {:else}
                  <div class="space-y-0.5 md:space-y-1 min-w-0 overflow-hidden">
                    {#each listEvents as event}
                      <div
                        class="flex items-center justify-between gap-2 py-0.5 md:py-1 px-2 md:px-3 bg-gray-50 rounded-lg min-w-0 overflow-hidden"
                      >
                        <div class="min-w-0 flex-1">
                          <p
                            class="font-medium text-gray-900 text-sm break-words"
                          >
                            {event.event_name}
                          </p>
                          {#if event.comment}
                            <p
                              class="text-[12px] md:text-xs text-gray-600 whitespace-pre-line mt-0 md:mt-0.5 break-words"
                            >
                              {event.comment}
                            </p>
                          {/if}
                        </div>
                        <div class="text-right flex-shrink-0 min-w-0">
                          <p
                            class="text-[14px] md:text-base font-semibold text-blue-600"
                          >
                            {formatEventTimeFromRFC3339(
                              event.datetime,
                              selectedVenue?.timezone,
                            )}
                          </p>
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
            {/if}
          </div>
        {/if}
      </div>
    {:else}
      <p class="text-gray-500 text-center text-[10px] md:text-sm">
        The search works across venue names, addresses, comments, owner
        information, event list names, and event names.
      </p>
    {/if}
  </div>
</div>
