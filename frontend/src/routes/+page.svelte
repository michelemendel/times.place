<script>
  import { onMount, afterUpdate, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { venueStore, eventListStore, eventStore, ownersStore } from '$lib/stores';
  import { seedDemoData } from '$lib/demo_data';
  import { formatEventTime } from '$lib/utils/datetime.js';
  import { dev, browser } from '$app/environment';

  /** @type {string | null} */
  let selectedVenueId = null;
  /** @type {string | null} */
  let selectedEventListId = null;
  /** @type {any[]} */
  let venues = [];
  /** @type {any[]} */
  let eventLists = [];
  /** @type {any[]} */
  let events = [];
  /** @type {any[]} */
  let owners = [];

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

  // Subscribe to stores
  venueStore.subscribe((v) => {
    venues = v;
    // Reset selection if selected venue no longer exists
    if (selectedVenueId && !venues.find((v) => v.venue_uuid === selectedVenueId)) {
      selectedVenueId = null;
      selectedEventListId = null;
    }
  });

  eventListStore.subscribe((v) => {
    eventLists = v;
    // Reset event list selection if it no longer exists
    if (selectedEventListId && !eventLists.find((el) => el.event_list_uuid === selectedEventListId)) {
      selectedEventListId = null;
    }
  });

  eventStore.subscribe((v) => {
    events = v;
  });

  ownersStore.subscribe((v) => {
    owners = v;
  });

  onMount(() => {
    // Seed demo data if needed
    // Seed demo data only when storage is empty.
    // Note: forcing a re-seed clears localStorage and will wipe newly-registered accounts.
    seedDemoData(false);

    // If there's a private link token, try to find and select the venue or event list
    if (privateLinkToken) {
      // First check if it's a venue token
      const privateVenue = venues.find(
        (v) => v.visibility === 'private' && v.private_link_token === privateLinkToken
      );
      if (privateVenue) {
        selectedVenueId = privateVenue.venue_uuid;
      } else {
        // Check if it's an event list token
        const eventList = eventLists.find((el) => el.private_link_token === privateLinkToken);
        if (eventList) {
          const venue = venues.find((v) => v.venue_uuid === eventList.venue_uuid);
          if (venue) {
            selectedVenueId = venue.venue_uuid;
            selectedEventListId = eventList.event_list_uuid;
          }
        }
      }
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
  });

  onDestroy(() => {
    if (browser) {
      document.removeEventListener('click', handleClickOutside);
    }
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
  });

  // Get private link token from URL (only in browser to avoid prerendering errors)
  $: privateLinkToken = browser && $page.url.searchParams ? $page.url.searchParams.get('token') || null : null;

  // Check if token matches an event list
  $: eventListFromToken = privateLinkToken
    ? eventLists.find((el) => el.private_link_token === privateLinkToken)
    : null;

  // If token matches an event list, find its venue
  $: venueFromEventListToken = eventListFromToken
    ? venues.find((v) => v.venue_uuid === eventListFromToken.venue_uuid)
    : null;

  // Filter venues: show venues that have at least one public event list OR venues with event lists accessed via token
  // If all event lists are private (and not accessed via token), the venue should not appear
  $: visibleVenues = venues.filter((venue) => {
    const venueEventLists = eventLists.filter((el) =>
      venue.event_list_uuids.includes(el.event_list_uuid)
    );

    // If venue has no event lists, don't show it
    if (venueEventLists.length === 0) return false;

    // Show venue if it contains an event list accessed via token
    if (venueFromEventListToken && venue.venue_uuid === venueFromEventListToken.venue_uuid) {
      return true;
    }

    // Show venue if it has at least one public event list
    const hasPublicEventList = venueEventLists.some((el) => el.visibility === 'public');
    if (hasPublicEventList) return true;

    return false;
  });

  // Sorted venues (only visible ones)
  $: sortedVenues = [...visibleVenues].sort((a, b) => a.name.localeCompare(b.name));

  /**
   * Check if a venue matches the search query by searching across all searchable fields:
   * - Venue: name, address, comment
   * - Venue Owner: name, mobile
   * - Event List: name, comment
   * - Event: event_name, comment
   * @param {any} venue
   * @param {string} query
   * @returns {boolean}
   */
  function venueMatchesSearch(venue, query) {
    if (!query) return true;
    const lowerQuery = query.toLowerCase();

    // Check venue fields
    if (venue.name?.toLowerCase().includes(lowerQuery)) return true;
    if (venue.address?.toLowerCase().includes(lowerQuery)) return true;
    if (venue.comment?.toLowerCase().includes(lowerQuery)) return true;

    // Check venue owner fields
    const owner = owners.find((o) => o.owner_uuid === venue.owner_uuid);
    if (owner) {
      if (owner.name?.toLowerCase().includes(lowerQuery)) return true;
      if (owner.mobile?.toLowerCase().includes(lowerQuery)) return true;
    }

    // Check event list fields
    const venueEventLists = eventLists.filter((el) =>
      venue.event_list_uuids.includes(el.event_list_uuid)
    );
    for (const eventList of venueEventLists) {
      if (eventList.name?.toLowerCase().includes(lowerQuery)) return true;
      if (eventList.comment?.toLowerCase().includes(lowerQuery)) return true;
    }

    // Check event fields
    for (const eventList of venueEventLists) {
      const listEvents = events.filter((e) =>
        eventList.event_uuids.includes(e.event_uuid)
      );
      for (const event of listEvents) {
        if (event.event_name?.toLowerCase().includes(lowerQuery)) return true;
        if (event.comment?.toLowerCase().includes(lowerQuery)) return true;
      }
    }

    return false;
  }

  // Filtered venues based on search
  $: filteredVenues = venueSearchQuery
    ? sortedVenues.filter((v) => venueMatchesSearch(v, venueSearchQuery))
    : sortedVenues;

  // Get selected venue
  $: selectedVenue = selectedVenueId
    ? venues.find((v) => v.venue_uuid === selectedVenueId)
    : null;

  // Get event lists for selected venue, filtered by visibility
  // Show public event lists OR private event lists accessed via token
  $: venueEventLists = selectedVenue
    ? eventLists
        .filter((el) => selectedVenue.event_list_uuids.includes(el.event_list_uuid))
        .filter((el) => {
          // Show public event lists
          if (el.visibility === 'public') return true;
          // Show private event lists if accessed via token
          if (el.visibility === 'private' && privateLinkToken && el.private_link_token === privateLinkToken) {
            return true;
          }
          return false;
        })
    : [];

  // Auto-select first event list or event list from token
  $: {
    if (selectedVenue && venueEventLists.length > 0) {
      // If we have an event list from token, select it
      if (eventListFromToken && venueEventLists.find(el => el.event_list_uuid === eventListFromToken.event_list_uuid)) {
        selectedEventListId = eventListFromToken.event_list_uuid;
      } else if (!selectedEventListId) {
        // Otherwise, select first event list
        selectedEventListId = venueEventLists[0].event_list_uuid;
      }
    } else if (!selectedVenue || venueEventLists.length === 0) {
      selectedEventListId = null;
    }
  }

  // Get selected event list
  $: selectedEventList = selectedEventListId
    ? eventLists.find((el) => el.event_list_uuid === selectedEventListId)
    : null;

  // Get events for selected event list
  $: listEvents = selectedEventList
    ? events.filter((e) => selectedEventList.event_uuids.includes(e.event_uuid))
    : [];

  // Get venue owner for selected venue
  $: venueOwner = selectedVenue
    ? owners.find((o) => o.owner_uuid === selectedVenue.owner_uuid)
    : null;

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
   * @param {any} venue
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
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      maxZoom: 19
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
    return formatEventTime(unixTimestamp, venueTimezone ? { timeZone: venueTimezone } : {});
  }

  /**
   * @param {Event} event
   */
  function handleVenueChange(event) {
    const target = /** @type {HTMLSelectElement} */ (event.target);
    selectedVenueId = target.value || null;
    selectedEventListId = null; // Reset event list selection
  }

  /**
   * @param {string | null} venueId
   */
  function selectVenue(venueId) {
    selectedVenueId = venueId || null;
    selectedEventListId = null;
    // Look up the venue directly instead of relying on reactive selectedVenue
    const venue = venueId ? venues.find((v) => v.venue_uuid === venueId) : null;
    venueSearchQuery = venue ? venue.name : '';
    showVenueDropdown = false;
    highlightedVenueIndex = -1;
  }

  /**
   * @param {Event} event
   */
  function handleVenueSearchInput(event) {
    const target = /** @type {HTMLInputElement} */ (event.target);
    const newValue = target.value;
    venueSearchQuery = newValue;

    // If the search field is cleared, reset the selection
    if (!newValue || newValue.trim() === '') {
      selectedVenueId = null;
      selectedEventListId = null;
    }

    showVenueDropdown = true;
    highlightedVenueIndex = -1;
  }

  function handleVenueSearchFocus() {
    showVenueDropdown = true;
    highlightedVenueIndex = -1;
  }

  /**
   * @param {KeyboardEvent} event
   */
  function handleVenueSearchKeydown(event) {
    if (!showVenueDropdown || filteredVenues.length === 0) {
      if (event.key === 'ArrowDown' && filteredVenues.length > 0) {
        showVenueDropdown = true;
        highlightedVenueIndex = 0;
        event.preventDefault();
      }
      return;
    }

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        highlightedVenueIndex = (highlightedVenueIndex + 1) % filteredVenues.length;
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
        if (highlightedVenueIndex >= 0 && highlightedVenueIndex < filteredVenues.length) {
          selectVenue(filteredVenues[highlightedVenueIndex].venue_uuid);
        }
        break;
      case 'Escape':
        event.preventDefault();
        showVenueDropdown = false;
        highlightedVenueIndex = -1;
        break;
    }
  }

  function clearVenueSearch() {
    venueSearchQuery = '';
    selectedVenueId = null;
    selectedEventListId = null;
    showVenueDropdown = false;
  }

  /**
   * @param {Event} event
   */
  function handleClickOutside(event) {
    if (venueDropdownRef && !venueDropdownRef.contains(/** @type {Node} */ (event.target))) {
      showVenueDropdown = false;
      highlightedVenueIndex = -1;
    }
  }

  // Update search query when venue is selected via click/keyboard
  // This is handled in selectVenue(), so we don't need a reactive statement here
  // to avoid conflicts when user is typing to clear/reset

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

<div class="mb-2 md:mb-2 text-center no-print-header">
  <h1 class="text-xl md:text-4xl font-bold mb-1 md:mb-2 text-gray-900">Find Venues and Events</h1>
  <p class="text-xs md:text-lg text-gray-600 mb-1 md:mb-0">
    Select a venue to view its event schedules and contact information.
  </p>
</div>

<div class="bg-white rounded-xl shadow-lg p-2 md:p-12 md:pt-4">
  <!-- Venue Searchable Dropdown -->
  <div class="mb-2 md:mb-2 relative no-print-venue-dropdown" bind:this={venueDropdownRef}>
    <div class="relative">
      <input
        type="text"
        id="venue-search"
        value={venueSearchQuery}
        on:input={handleVenueSearchInput}
        on:focus={handleVenueSearchFocus}
        on:keydown={handleVenueSearchKeydown}
        placeholder="Search and select a venue..."
        class="w-full px-4 py-2 pr-10 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white text-gray-900"
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
    </div>
    {#if showVenueDropdown && filteredVenues.length > 0}
      <div class="absolute z-50 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-auto">
        {#each filteredVenues as venue, index}
          <button
            type="button"
            on:click={() => selectVenue(venue.venue_uuid)}
            on:mouseenter={() => highlightedVenueIndex = index}
            class="w-full text-left px-4 py-2 focus:outline-none {highlightedVenueIndex === index ? 'bg-blue-100' : selectedVenueId === venue.venue_uuid ? 'bg-blue-50' : 'hover:bg-gray-100'}"
          >
            {venue.name}
          </button>
        {/each}
      </div>
    {/if}
    {#if showVenueDropdown && filteredVenues.length === 0 && venueSearchQuery}
      <div class="absolute z-50 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg">
        <div class="px-4 py-2 text-gray-500 text-sm">No venues found</div>
      </div>
    {/if}
  </div>

  {#if selectedVenue}
    <!-- Venue Details -->
    <div class="pt-0 md:pt-2 md:mt-0">
      <!-- Banner Image -->
      {#if selectedVenue.banner_image}
        <div class="mb-4">
          <img
            src={selectedVenue.banner_image}
            alt={selectedVenue.name}
            class="w-full h-48 object-cover rounded-lg"
          />
        </div>
      {/if}

      <!-- Venue Name -->
      <h2 class="text-3xl font-bold mb-3 text-gray-900">{selectedVenue.name}</h2>

      <!-- Address, Contact, Comment, and Map -->
      {#if selectedVenue.address || selectedVenue.geolocation || venueOwner || selectedVenue.comment}
        <div class="mb-3 grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="flex flex-col justify-start">
            {#if selectedVenue.address}
              <p class="text-sm text-gray-600 mb-2">
                <span class="font-medium text-gray-700">Address:</span>
                {#if getDirectionsUrl(selectedVenue)}
                  <a
                    href={getDirectionsUrl(selectedVenue)}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-blue-600 hover:text-blue-800 hover:underline"
                  >
                    {selectedVenue.address}
                  </a>
                {:else}
                  {selectedVenue.address}
                {/if}
              </p>
            {/if}
            {#if selectedVenue.geolocation}
              <p class="text-sm text-gray-600 mb-2 no-print-geolocation">
                <span class="font-medium text-gray-700">Geolocation:</span> {selectedVenue.geolocation}
              </p>
            {/if}
            {#if selectedVenue.timezone}
              <p class="text-sm text-gray-600 mb-2 no-print-timezone">
                <span class="font-medium text-gray-700">Timezone:</span> {selectedVenue.timezone}
              </p>
            {/if}
            <!-- Contact information is hidden from public visitor page for security -->
            {#if selectedVenue.comment}
              <p class="text-sm text-gray-600 italic">{selectedVenue.comment}</p>
            {/if}
          </div>
          {#if selectedVenue.geolocation}
            <div class="w-full h-48 rounded-lg overflow-hidden border border-gray-300 no-print-map">
              <div bind:this={mapContainer} class="w-full h-full"></div>
            </div>
          {/if}
        </div>
      {/if}

      <!-- Event Lists Section -->
      {#if venueEventLists.length === 0}
        <div class="mt-6 p-4 bg-gray-50 rounded-lg">
          <p class="text-gray-500 text-center italic">
            This venue currently has no event schedules available.
          </p>
        </div>
      {:else if venueEventLists.length === 1}
        <!-- Single event list - no selector needed -->
        {#if selectedEventList}
          <div class="mt-6">
            <h3 class="text-2xl font-semibold mb-4 text-gray-900">{selectedEventList.name}</h3>
            {#if selectedEventList.comment}
              <p class="text-gray-600 mb-4 italic">{selectedEventList.comment}</p>
            {/if}

            {#if listEvents.length === 0}
              <p class="text-gray-500 italic">No events scheduled for this list.</p>
            {:else}
              <div class="space-y-3">
                {#each listEvents as event}
                  <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p class="font-medium text-gray-900">{event.event_name}</p>
                      {#if event.comment}
                        <p class="text-sm text-gray-600 italic">{event.comment}</p>
                      {/if}
                    </div>
                    <div class="text-right">
                      <p class="text-lg font-semibold text-blue-600">
                        {formatEventTimeFromRFC3339(event.datetime, selectedVenue?.timezone)}
                      </p>
                      {#if event.duration_minutes}
                        <p class="text-xs text-gray-500">
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
        <!-- Multiple event lists - show selector -->
        <div class="mt-6">
          <select
            id="event-list-select"
            on:change={handleEventListChange}
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white text-gray-900 mb-4 no-print-event-list-selector"
          >
            {#each venueEventLists as eventList}
              <option value={eventList.event_list_uuid} selected={selectedEventListId === eventList.event_list_uuid}>
                {eventList.name}
              </option>
            {/each}
          </select>

          {#if selectedEventList}
            <div>
              <h3 class="text-2xl font-semibold mb-4 text-gray-900">{selectedEventList.name}</h3>
              {#if selectedEventList.comment}
                <p class="text-gray-600 mb-4 italic">{selectedEventList.comment}</p>
              {/if}

              {#if listEvents.length === 0}
                <p class="text-gray-500 italic">No events scheduled for this list.</p>
              {:else}
                <div class="space-y-3">
                  {#each listEvents as event}
                    <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                      <div>
                        <p class="font-medium text-gray-900">{event.event_name}</p>
                        {#if event.comment}
                          <p class="text-sm text-gray-600 italic">{event.comment}</p>
                        {/if}
                      </div>
                      <div class="text-right">
                        <p class="text-lg font-semibold text-blue-600">
                          {formatEventTimeFromRFC3339(event.datetime)}
                        </p>
                        {#if event.duration_minutes}
                          <p class="text-xs text-gray-500">
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
    <p class="text-gray-500 italic text-center">
      Please select a venue from the dropdown above to view its details and event schedules.
    </p>
  {/if}
</div>
