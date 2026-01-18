<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { get } from 'svelte/store';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';
  import { currentOwnerStore, venueStore, eventListStore, eventStore } from '$lib/stores';
  import { formatEventListDate, formatEventTime } from '$lib/utils/datetime.js';

  /**
   * Format event time from RFC3339 string
   * @param {string} rfc3339
   * @param {string} [venueTimezone] - Optional venue timezone to use for display
   * @returns {string}
   */
  function formatEventTimeFromRFC3339(rfc3339, venueTimezone) {
    const unixTimestamp = Math.floor(new Date(rfc3339).getTime() / 1000);
    return formatEventTime(unixTimestamp, venueTimezone ? { timeZone: venueTimezone } : {});
  }

  /** @type {import('$lib/types').VenueOwner | null} */
  let owner = null;
  /** @type {import('$lib/types').Venue | null} */
  let venue = null;
  /** @type {import('$lib/types').EventList | null} */
  let eventList = null;
  /** @type {import('$lib/types').Event[]} */
  let listEvents = [];

  // Get venue UUID and event list UUID from route params
  $: venueUuid = browser && $page.params ? $page.params.venue_uuid : null;
  $: eventListUuid = browser && $page.params ? $page.params.event_list_uuid : null;

  // Load data when UUIDs are available
  $: {
    if (venueUuid && eventListUuid) {
      const currentOwner = $currentOwnerStore;
      if (currentOwner) {
        owner = currentOwner;
        const allVenues = $venueStore;
        const foundVenue = allVenues.find(v => v.venue_uuid === venueUuid);

        // Verify ownership - only set venue if authorized
        if (foundVenue && foundVenue.owner_uuid === currentOwner.owner_uuid) {
          venue = foundVenue;
          const allEventLists = $eventListStore;
          const foundEventList = allEventLists.find(el =>
            el.event_list_uuid === eventListUuid && el.venue_uuid === foundVenue.venue_uuid
          );

          if (foundEventList) {
            eventList = foundEventList;
            const allEvents = $eventStore;
            listEvents = allEvents
              .filter(e => foundEventList.event_uuids.includes(e.event_uuid))
              .sort((a, b) => {
                // Sort by datetime
                return new Date(a.datetime).getTime() - new Date(b.datetime).getTime();
              });
          }
        } else if (foundVenue) {
          // Not authorized - venue will be null, handled in template
          venue = null;
        }
      }
    }
  }

  onMount(() => {
    owner = get(currentOwnerStore);
    if (!owner) {
      goto('/login');
      return;
    }

    // Check authorization after stores are loaded
    if (venueUuid && owner) {
      const allVenues = get(venueStore);
      const foundVenue = allVenues.find(v => v.venue_uuid === venueUuid);
      if (foundVenue && foundVenue.owner_uuid !== owner.owner_uuid) {
        // Not authorized - redirect
        goto('/venue-owner');
      }
    }
  });

  /**
   * Print the event list
   */
  function printEventList() {
    window.print();
  }

  /**
   * Go back to My Venues page
   */
  function goBack() {
    goto('/venue-owner');
  }
</script>

<svelte:head>
  <title>{eventList?.name || 'Event List'} - {venue?.name || 'Venue'} - times.place</title>
</svelte:head>

<div class="max-w-5xl mx-auto px-6 sm:px-8 lg:px-12 py-8">
  {#if !owner}
    <div class="text-center py-8">
      <p class="text-gray-600">Please log in to view event lists.</p>
    </div>
  {:else if !venue || !eventList}
    <div class="text-center py-8">
      <p class="text-gray-600">Event list not found.</p>
      <button
        on:click={goBack}
        class="mt-4 bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200"
      >
        Back to Venues
      </button>
    </div>
  {:else}
    <!-- Action Buttons (hidden when printing) -->
    <div class="mb-6 flex gap-4 no-print">
      <button
        on:click={goBack}
        class="flex items-center gap-2 px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition-colors duration-200"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        <span>Back</span>
      </button>

      <button
        on:click={printEventList}
        class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors duration-200"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
        </svg>
        <span>Print</span>
      </button>
    </div>

    <!-- Event List Content -->
    <div class="bg-white rounded-xl shadow-lg p-6 md:p-12">
      <!-- Banner Image -->
      {#if venue.banner_image}
        <div class="mb-6">
          <img
            src={venue.banner_image}
            alt={venue.name}
            class="w-full h-48 object-cover rounded-lg"
          />
        </div>
      {/if}

      <!-- Venue Header -->
      <div class="mb-6 pb-4 border-b border-gray-200">
        <h1 class="text-3xl font-bold mb-2 text-gray-900">{venue.name}</h1>
        {#if venue.address}
          <p class="text-lg text-gray-600">{venue.address}</p>
        {/if}
      </div>

      <!-- Event List Header -->
      <div class="mb-6">
        <h2 class="text-2xl font-semibold mb-2 text-gray-900">{eventList.name || 'Untitled Event List'}</h2>
        {#if eventList.date}
          <p class="text-lg text-gray-600 mb-2">
            <svg class="w-5 h-5 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            {formatEventListDate(eventList.date)}
          </p>
        {/if}
        {#if eventList.comment}
          <p class="text-gray-600 whitespace-pre-line">{eventList.comment}</p>
        {/if}
      </div>

      <!-- Events -->
      {#if listEvents.length === 0}
        <div class="py-8 text-center">
          <p class="text-gray-500">No events scheduled for this list.</p>
        </div>
      {:else}
        <div class="space-y-4">
          {#each listEvents as event}
            <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
              <div class="flex-1">
                <p class="font-medium text-lg text-gray-900">{event.event_name}</p>
                {#if event.comment}
                  <p class="text-sm text-gray-600 mt-1 whitespace-pre-line">{event.comment}</p>
                {/if}
              </div>
              <div class="text-right ml-4">
                <p class="text-xl font-semibold text-blue-600">
                  {formatEventTimeFromRFC3339(event.datetime, venue?.timezone)}
                </p>
                {#if event.duration_minutes}
                  <p class="text-sm text-gray-500 mt-1">
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

<style>
  @media print {
    .no-print {
      display: none !important;
    }

    .bg-white {
      background: white !important;
      box-shadow: none !important;
    }

    .bg-gray-50 {
      background: #f9fafb !important;
    }
  }
</style>
