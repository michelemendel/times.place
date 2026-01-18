<script>
  import { onMount } from 'svelte';
  import { get } from 'svelte/store';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';
  import { currentOwnerStore, venueStore, eventListStore, eventStore } from '$lib/stores';
  import { generateUUID } from '$lib/utils/uuid.js';
  import { updateModifiedTimestamp } from '$lib/utils/datetime.js';

  /** @type {import('$lib/types').VenueOwner | null} */
  let owner = null;
  /** @type {import('$lib/types').Venue[]} */
  let ownerVenues = [];
  /** @type {import('$lib/types').EventList[]} */
  let allEventLists = [];
  /** @type {import('$lib/types').Venue | null} */
  let deleteConfirmVenue = null;
  /** @type {string | null} */
  let copiedLinkToken = null;

  // Use reactive store syntax ($store) to automatically update when stores change
  $: {
    const currentOwner = $currentOwnerStore;
    if (currentOwner) {
      owner = currentOwner;
      const allVenues = $venueStore;
      // Filter venues by owner_uuid and sort alphabetically
      ownerVenues = allVenues
        .filter(v => v.owner_uuid === currentOwner.owner_uuid)
        .sort((a, b) => a.name.localeCompare(b.name));
    }
    allEventLists = $eventListStore;
  }

  onMount(() => {
    owner = get(currentOwnerStore);
    if (!owner) {
      goto('/login');
      return;
    }
  });

  /**
   * @param {import('$lib/types').Venue} venue
   * @returns {number}
   */
  function getEventListCount(venue) {
    return venue.event_list_uuids.length;
  }

  function handleAddVenue() {
    goto('/venue-form');
  }

  /**
   * @param {import('$lib/types').Venue} venue
   */
  function handleEditVenue(venue) {
    goto(`/venue-form?venue_uuid=${venue.venue_uuid}`);
  }

  /**
   * @param {import('$lib/types').Venue} venue
   */
  function handleDeleteClick(venue) {
    deleteConfirmVenue = venue;
  }

  function handleDeleteConfirm() {
    if (!deleteConfirmVenue || !owner) return;

    const venue = deleteConfirmVenue;

    // Get all event lists for this venue
    const venueEventLists = allEventLists.filter(el => el.venue_uuid === venue.venue_uuid);

    // Get all events for these event lists
    const allEvents = get(eventStore);
    const venueEventUuids = venueEventLists.flatMap(el => el.event_uuids);

    // Remove venue from store
    const currentVenues = get(venueStore);
    venueStore.set(currentVenues.filter(v => v.venue_uuid !== venue.venue_uuid));

    // Remove event lists from store
    const currentEventLists = get(eventListStore);
    eventListStore.set(currentEventLists.filter(el => el.venue_uuid !== venue.venue_uuid));

    // Remove events from store
    const currentEvents = get(eventStore);
    eventStore.set(currentEvents.filter(e => !venueEventUuids.includes(e.event_uuid)));

    deleteConfirmVenue = null;
  }

  function handleDeleteCancel() {
    deleteConfirmVenue = null;
  }

  /**
   * @param {MouseEvent} event
   */
  function handleModalBackdropClick(event) {
    // Only close if clicking directly on the backdrop, not on the modal content
    if (event.target === event.currentTarget) {
      handleDeleteCancel();
    }
  }

  /**
   * Get event lists for a venue
   * @param {import('$lib/types').Venue} venue
   * @returns {import('$lib/types').EventList[]}
   */
  function getVenueEventLists(venue) {
    return allEventLists
      .filter(el => el.venue_uuid === venue.venue_uuid)
      .sort((a, b) => {
        // Sort by date, then by name
        if (a.date !== b.date) {
          return a.date.localeCompare(b.date);
        }
        return a.name.localeCompare(b.name);
      });
  }

  /**
   * Ensure event list has a private link token, generate if missing
   * @param {import('$lib/types').EventList} eventList
   * @returns {import('$lib/types').EventList}
   */
  function ensureEventListToken(eventList) {
    if (eventList.private_link_token) {
      return eventList;
    }

    // Generate token and update store
    const token = generateUUID();
    const updatedList = /** @type {import('$lib/types').EventList} */ (updateModifiedTimestamp({
      ...eventList,
      private_link_token: token
    }));

    // Update in store
    const currentEventLists = get(eventListStore);
    const updatedEventLists = currentEventLists.map(el =>
      el.event_list_uuid === eventList.event_list_uuid ? updatedList : el
    );
    eventListStore.set(updatedEventLists);

    return updatedList;
  }

  /**
   * Get private link for an event list
   * @param {import('$lib/types').EventList} eventList
   * @returns {string}
   */
  function getPrivateLink(eventList) {
    const listWithToken = ensureEventListToken(eventList);
    if (!listWithToken.private_link_token) return '';
    if (!browser) return '';
    return `${window.location.origin}/?token=${listWithToken.private_link_token}`;
  }

  /**
   * Copy private link to clipboard
   * @param {import('$lib/types').EventList} eventList
   */
  async function copyPrivateLink(eventList) {
    const link = getPrivateLink(eventList);
    if (!link) return;

    try {
      await navigator.clipboard.writeText(link);
      copiedLinkToken = eventList.event_list_uuid;
      // Reset after 2 seconds
      setTimeout(() => {
        copiedLinkToken = null;
      }, 2000);
    } catch (err) {
      console.error('Failed to copy link:', err);
    }
  }

  /**
   * Navigate to view/print page for an event list
   * @param {string} venueUuid
   * @param {string} eventListUuid
   */
  function viewPrintEventList(venueUuid, eventListUuid) {
    goto(`/venue-owner/${venueUuid}/event-lists/${eventListUuid}`);
  }
</script>

<svelte:head>
  <title>My Venues - time.place</title>
</svelte:head>

<div class="max-w-5xl mx-auto px-6 sm:px-8 lg:px-12 py-8">
  <div class="mb-8 text-center">
    <h1 class="text-4xl font-bold mb-4 text-gray-900">My Venues</h1>
    <p class="text-lg text-gray-600">Manage your venues and event schedules.</p>
  </div>

  {#if owner}
    <div class="mb-6 rounded-md border border-gray-200 bg-gray-50 px-4 py-3">
      <div class="text-sm text-gray-700">
        Signed in as <span class="font-semibold">{owner.name}</span>
        <span class="text-gray-500">({owner.email})</span>
      </div>
    </div>

    <div class="mb-6 flex justify-end">
      <button
        on:click={handleAddVenue}
        class="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-4 sm:py-3 sm:px-6 rounded-lg shadow-md transition-colors duration-200 flex items-center gap-2"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        <span>Add Venue</span>
      </button>
    </div>

    {#if ownerVenues.length === 0}
      <div class="bg-white rounded-xl shadow-lg p-8 md:p-12 text-center">
        <div class="mb-4">
          <svg class="w-16 h-16 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
        </div>
        <h2 class="text-2xl font-semibold text-gray-900 mb-2">No venues yet</h2>
        <p class="text-gray-600 mb-6">Get started by adding your first venue.</p>
        <button
          on:click={handleAddVenue}
          class="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-6 rounded-lg shadow-md transition-colors duration-200"
        >
          Add Your First Venue
        </button>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6">
        {#each ownerVenues as venue (venue.venue_uuid)}
          {@const venueEventLists = getVenueEventLists(venue)}
          <div class="bg-white rounded-xl shadow-md hover:shadow-lg transition-shadow duration-200 overflow-hidden flex flex-col h-full">
            {#if venue.banner_image}
              <div class="w-full h-32 sm:h-40 bg-gray-200 overflow-hidden">
                <img
                  src={venue.banner_image}
                  alt={venue.name}
                  class="w-full h-full object-cover"
                />
              </div>
            {/if}
            <div class="p-4 sm:p-6 flex flex-col flex-grow">
              <h3 class="text-xl font-bold text-gray-900 mb-2">{venue.name}</h3>
              {#if venue.address}
                <p class="text-sm text-gray-600 mb-2 flex items-start gap-2">
                  <svg class="w-4 h-4 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                  <span>{venue.address}</span>
                </p>
              {/if}
              <div class="flex items-center gap-2 text-sm text-gray-600 mb-3">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <span>{getEventListCount(venue)} event list{getEventListCount(venue) !== 1 ? 's' : ''}</span>
              </div>
              {#if venue.comment}
                <p class="text-sm text-gray-500 mb-4 line-clamp-2">{venue.comment}</p>
              {/if}

              <!-- Event Lists -->
              {#if venueEventLists.length > 0}
                <div class="mb-4 pt-4 border-t border-gray-200">
                  <p class="text-sm font-medium text-gray-700 mb-3">Event Lists:</p>
                  <div class="space-y-2">
                    {#each venueEventLists as eventList (eventList.event_list_uuid)}
                      <div class="flex items-center justify-between gap-2 p-2 bg-gray-50 rounded-lg">
                        <div class="flex items-center gap-2 flex-1 min-w-0">
                          {#if eventList.visibility === 'private'}
                            <div title="Private" class="shrink-0">
                              <svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                              </svg>
                            </div>
                          {:else}
                            <div title="Public" class="shrink-0">
                              <svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                              </svg>
                            </div>
                          {/if}
                          <span class="text-sm text-gray-900 truncate">{eventList.name}</span>
                        </div>
                        <div class="flex gap-2 shrink-0">
                          <button
                            on:click={() => viewPrintEventList(venue.venue_uuid, eventList.event_list_uuid)}
                            class="px-3 py-1 bg-green-600 hover:bg-green-700 text-white text-xs font-medium rounded transition-colors duration-200"
                            title="View/Print"
                          >
                            View/Print
                          </button>
                          <button
                            on:click={() => copyPrivateLink(eventList)}
                            class="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium rounded transition-colors duration-200"
                            title="Get Private Link"
                          >
                            {#if copiedLinkToken === eventList.event_list_uuid}
                              <span class="flex items-center gap-1">
                                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                                </svg>
                                Copied
                              </span>
                            {:else}
                              Get Link
                            {/if}
                          </button>
                        </div>
                      </div>
                    {/each}
                  </div>
                </div>
              {/if}

              <div class="flex flex-col sm:flex-row gap-2 mt-auto">
                <button
                  on:click={() => handleEditVenue(venue)}
                  class="flex-1 bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200 text-sm sm:text-base"
                >
                  Edit
                </button>
                <button
                  on:click={() => handleDeleteClick(venue)}
                  class="flex-1 bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200 text-sm sm:text-base"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}

  <!-- Delete Confirmation Modal -->
  {#if deleteConfirmVenue}
    <div
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="delete-modal-title"
      tabindex="-1"
      on:click={handleModalBackdropClick}
      on:keydown={(e) => e.key === 'Escape' && handleDeleteCancel()}
    >
      <article class="bg-white rounded-xl shadow-xl max-w-md w-full p-6">
        <h3 id="delete-modal-title" class="text-xl font-bold text-gray-900 mb-4">Delete Venue</h3>
        <p class="text-gray-700 mb-6">
          Are you sure you want to delete <span class="font-semibold">"{deleteConfirmVenue.name}"</span>?
          This will also delete all associated event lists and events. This action cannot be undone.
        </p>
        <div class="flex flex-col sm:flex-row gap-3 justify-end">
          <button
            on:click={handleDeleteCancel}
            class="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors duration-200 font-medium"
          >
            Cancel
          </button>
          <button
            on:click={handleDeleteConfirm}
            class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors duration-200 font-medium"
          >
            Delete
          </button>
        </div>
      </article>
    </div>
  {/if}
</div>
