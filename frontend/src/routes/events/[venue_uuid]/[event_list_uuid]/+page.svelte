<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { get } from 'svelte/store';
  import { goto } from '$app/navigation';
  import {
    listPublicVenues,
    getPublicEventListsForVenue,
    getPublicEventsForEventList,
  } from '$lib/api/public.js';
  import { formatEventListDate, formatEventTime } from '$lib/utils/datetime.js';
  import BannerImage from '$lib/BannerImage.svelte';

  /**
   * Format event time from RFC3339 string
   * @param {string} rfc3339
   * @param {string} [venueTimezone] - Optional venue timezone to use for display
   * @returns {string}
   */
  function formatEventTimeFromRFC3339(rfc3339, venueTimezone) {
    const unixTimestamp = Math.floor(new Date(rfc3339).getTime() / 1000);
    return formatEventTime(
      unixTimestamp,
      venueTimezone ? { timeZone: venueTimezone } : {},
    );
  }

  /** @type {import('$lib/types').Venue | null} */
  let venue = null;
  /** @type {import('$lib/types').EventList | null} */
  let eventList = null;
  /** @type {import('$lib/types').Event[]} */
  let listEvents = [];
  /** @type {boolean} */
  let loading = true;
  /** @type {string} */
  let loadError = '';

  onMount(async () => {
    const params = get(page).params;
    const venueUuid = params?.venue_uuid;
    const eventListUuid = params?.event_list_uuid;

    if (!venueUuid || !eventListUuid) {
      loading = false;
      loadError = 'Invalid link.';
      return;
    }

    loading = true;
    loadError = '';

    try {
      // Fetch all public venues and find the one we need
      // Note: Ideally we would have a getPublicVenue(uuid) endpoint
      const venues = await listPublicVenues();
      venue = venues.find((v) => v.venue_uuid === venueUuid) || null;

      if (!venue) {
        throw new Error('Venue not found.');
      }

      // Fetch event lists to find the specific one
      const eventLists = await getPublicEventListsForVenue(venueUuid);
      eventList =
        eventLists.find((el) => el.event_list_uuid === eventListUuid) || null;

      if (!eventList) {
        throw new Error('Event list not found.');
      }

      // Fetch events
      const events = await getPublicEventsForEventList(eventListUuid);
      listEvents = (events ?? []).slice().sort((a, b) => {
        return new Date(a.datetime).getTime() - new Date(b.datetime).getTime();
      });
    } catch (e) {
      venue = null;
      eventList = null;
      listEvents = [];
      loadError = e instanceof Error ? e.message : 'Failed to load event list.';
    } finally {
      loading = false;
    }
  });

  /**
   * Print the event list
   */
  function printEventList() {
    window.print();
  }

  /**
   * Go back to Main Page
   */
  function goBack() {
    if (venue && eventList) {
      goto(`/?venue=${venue.venue_uuid}&list=${eventList.event_list_uuid}`);
    } else {
      goto('/');
    }
  }
</script>

<svelte:head>
  <title
    >{eventList?.name || 'Event List'} - {venue?.name || 'Venue'} - times.place</title
  >
</svelte:head>

<div class="w-full min-w-0 max-w-full lg:max-w-[60%] lg:mx-auto">
  {#if loading}
    <div class="text-center py-4 md:py-6">
      <p class="text-[12px] md:text-sm text-gray-600">Loading...</p>
    </div>
  {:else if loadError || !venue || !eventList}
    <div class="text-center py-4 md:py-6">
      <p class="text-[12px] md:text-sm text-gray-600">
        {loadError || 'Event list not found.'}
      </p>
      <button
        on:click={goBack}
        class="mt-2 md:mt-4 bg-blue-600 hover:bg-blue-700 text-white text-[12px] md:text-sm font-medium py-1 md:py-2 px-2 md:px-4 rounded-lg transition-colors duration-200"
      >
        Go Home
      </button>
    </div>
  {:else}
    <!-- Action Buttons (hidden when printing) -->
    <div class="mb-2 md:mb-4 flex gap-2 md:gap-4 no-print pl-2 md:pl-0">
      <button
        on:click={goBack}
        class="flex items-center gap-1 md:gap-2 px-2 md:px-3 py-1 md:py-1.5 text-[10px] md:text-sm bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition-colors duration-200"
      >
        <svg
          class="w-3 h-3 md:w-4 md:h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 19l-7-7 7-7"
          />
        </svg>
        <span>Back to App</span>
      </button>

      <button
        on:click={printEventList}
        class="flex items-center gap-1 md:gap-2 px-2 md:px-3 py-1 md:py-1.5 text-[10px] md:text-sm bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors duration-200"
      >
        <svg
          class="w-3 h-3 md:w-4 md:h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"
          />
        </svg>
        <span>Print</span>
      </button>
    </div>

    <!-- Event List Content -->
    <div class="bg-white rounded-xl shadow-lg p-2 md:p-6 mx-2 md:mx-0">
      <!-- Banner Image -->
      {#if venue.banner_image}
        <BannerImage
          src={venue.banner_image}
          alt={venue.name}
          size="md"
          wrapperClass="mb-2 md:mb-4"
        />
      {/if}

      <!-- Venue Header -->
      <div class="mb-2 md:mb-3 pb-1 md:pb-2 border-b border-gray-200">
        <h1
          class="text-[16px] md:text-3xl font-bold mb-1 md:mb-3 text-gray-900"
        >
          {venue.name}
        </h1>
        {#if venue.address}
          <p
            class="text-[10px] md:text-sm text-gray-600 mb-1 md:mb-2 text-left"
          >
            {venue.address}
          </p>
        {/if}
      </div>

      <!-- Event List Header -->
      <div class="mb-2 md:mb-4">
        <h2
          class="text-[14px] md:text-2xl font-semibold mb-0.5 md:mb-1 text-gray-900"
        >
          {eventList.name || 'Untitled Event List'}
        </h2>
        {#if eventList.date}
          <p class="text-[10px] md:text-sm text-gray-600 mb-1 md:mb-2">
            <svg
              class="w-3 h-3 md:w-4 md:h-4 inline mr-0.5 md:mr-1"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
            {formatEventListDate(eventList.date)}
          </p>
        {/if}
        {#if eventList.comment}
          <p
            class="text-[9px] md:text-xs text-gray-600 mb-2 md:mb-4 whitespace-pre-line text-left"
          >
            {eventList.comment}
          </p>
        {/if}
      </div>

      <!-- Events -->
      {#if listEvents.length === 0}
        <div class="py-2 md:py-4 text-center">
          <p class="text-[10px] md:text-sm text-gray-500">
            No events scheduled for this list.
          </p>
        </div>
      {:else}
        <div class="space-y-0.5 md:space-y-1">
          {#each listEvents as event}
            <div
              class="flex items-center justify-between py-0.5 md:py-1 px-2 md:px-3 bg-gray-50 rounded-lg"
            >
              <div class="flex-1 min-w-0 text-left">
                <p class="font-medium text-[12px] md:text-sm text-gray-900">
                  {event.event_name}
                </p>
                {#if event.comment}
                  <p
                    class="text-[10px] md:text-xs text-gray-600 mt-0.5 whitespace-pre-line"
                  >
                    {event.comment}
                  </p>
                {/if}
              </div>
              <div class="text-right ml-2 md:ml-4 flex-shrink-0">
                <p class="text-[14px] md:text-base font-semibold text-blue-600">
                  {formatEventTimeFromRFC3339(event.datetime, venue?.timezone)}
                </p>
                {#if event.duration_minutes}
                  <p class="text-[9px] md:text-xs text-gray-500 mt-0.5">
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

    /* Ensure content is full width on print */
    :global(body) {
      background: white !important;
    }
  }
</style>
