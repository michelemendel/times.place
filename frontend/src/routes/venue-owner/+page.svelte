<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { getAuthMe, resendVerificationEmail } from '$lib/api/auth.js';
  import { listVenues, deleteVenue } from '$lib/api/venues.js';
  import BannerImage from '$lib/BannerImage.svelte';

  /** @type {import('$lib/types').VenueOwner | null} */
  let owner = null;
  /** @type {import('$lib/types').Venue[]} */
  let ownerVenues = [];
  /** @type {Record<string, import('$lib/types').EventList[]>} */
  let eventListsByVenue = {};
  /** @type {import('$lib/types').Venue | null} */
  let deleteConfirmVenue = null;
  /** @type {string | null} */
  let copiedLinkToken = null;
  let isLoading = false;
  let loadError = '';
  /** Max venues allowed (from API). Default 2. */
  let venueLimit = 2;
  let showLimitPopup = false;
  let resendVerificationLoading = false;
  let resendVerificationMessage = '';

  $: venueCount = ownerVenues.length;
  $: atVenueLimit = venueCount >= venueLimit;
  $: addVenueDisabled = atVenueLimit || owner?.email_verified === false;

  async function loadVenuesAndLists() {
    isLoading = true;
    loadError = '';
    try {
      // GET /api/venues returns venues with event_lists embedded (one call, no per-venue requests)
      const venues = await listVenues();
      ownerVenues = venues.sort((a, b) => a.name.localeCompare(b.name));

      /** @type {Record<string, import('$lib/types').EventList[]>} */
      const map = {};
      for (const venue of ownerVenues) {
        map[venue.venue_uuid] = Array.isArray(venue.event_lists) ? venue.event_lists : [];
      }
      eventListsByVenue = map;
    } catch (err) {
      console.error('Failed to load venues', err);
      loadError = 'Failed to load venues. Please try again.';
    } finally {
      isLoading = false;
    }
  }

  onMount(async () => {
    try {
      const me = await getAuthMe();
      owner = me.owner;
      venueLimit = me.venue_limit ?? 2;
    } catch {
      goto('/login');
      return;
    }
    if (!owner) {
      goto('/login');
      return;
    }

    await loadVenuesAndLists();
  });

  /**
   * @param {import('$lib/types').Venue} venue
   * @returns {number}
   */
  function getEventListCount(venue) {
    const lists = eventListsByVenue[venue.venue_uuid] || [];
    return lists.length;
  }

  function handleAddVenue() {
    if (atVenueLimit) {
      showLimitPopup = true;
      return;
    }
    goto('/venue-form');
  }

  function closeLimitPopup() {
    showLimitPopup = false;
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

  async function handleDeleteConfirm() {
    if (!deleteConfirmVenue || !owner) return;

    const venue = deleteConfirmVenue;

    try {
      await deleteVenue(venue.venue_uuid);
      ownerVenues = ownerVenues.filter((v) => v.venue_uuid !== venue.venue_uuid);
      const { [venue.venue_uuid]: _removed, ...rest } = eventListsByVenue;
      eventListsByVenue = rest;
    } catch (err) {
      console.error('Failed to delete venue', err);
      loadError = 'Failed to delete venue. Please try again.';
    } finally {
      deleteConfirmVenue = null;
    }
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
    const lists = eventListsByVenue[venue.venue_uuid] || [];
    return lists
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
    // In the API-backed version, private_link_token is managed by the backend.
    // We simply rely on what the server returns; if it's missing, we don't fabricate one.
    return eventList;
  }

  /**
   * Get private link for an event list
   * @param {import('$lib/types').EventList} eventList
   * @returns {string}
   */
  function getPrivateLink(eventList) {
    const listWithToken = ensureEventListToken(eventList);
    if (!listWithToken.private_link_token) return '';
    if (typeof window === 'undefined') return '';
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

  async function handleResendVerification() {
    if (resendVerificationLoading) return;
    resendVerificationLoading = true;
    resendVerificationMessage = '';
    try {
      await resendVerificationEmail();
      resendVerificationMessage = 'Verification email sent. Check your inbox.';
      const me = await getAuthMe();
      owner = me.owner;
    } catch (err) {
      resendVerificationMessage = err instanceof Error ? err.message : 'Failed to send. Try again later.';
    } finally {
      resendVerificationLoading = false;
    }
  }
</script>

<svelte:head>
  <title>My Venues - times.place</title>
</svelte:head>

<div class="max-w-5xl mx-auto px-6 sm:px-8 lg:px-12 pt-1 pb-2 md:pt-6 md:pb-8">
  <div class="mb-1 md:mb-0 text-center">
    <h1 class="text-2xl md:text-4xl font-bold mb-1 md:mb-0.5 text-gray-900">My Venues</h1>
    <p class="text-base text-gray-600 hidden md:block">Manage your venues and event schedules.</p>
  </div>

  {#if owner}
    <div class="mb-3 md:mb-3 text-center">
      <div class="text-sm text-gray-700">
        Signed in as <span class="font-semibold">{owner.name}</span>
        <span class="text-gray-500">({owner.email})</span>
      </div>
    </div>

    {#if owner.email_verified === false}
      <div class="mb-4 p-4 rounded-lg bg-amber-50 border border-amber-200 text-amber-900">
        <p class="font-medium">Verify your email to add or edit venues and events.</p>
        <p class="text-sm mt-1">Check your inbox for a verification link, or request a new one. If you don't see it, check your spam or junk folder.</p>
        <div class="mt-3 flex flex-wrap items-center gap-2">
          <button
            type="button"
            on:click={handleResendVerification}
            disabled={resendVerificationLoading}
            class="text-sm font-medium py-1.5 px-3 rounded-md bg-amber-600 hover:bg-amber-700 text-white disabled:opacity-50"
          >
            {resendVerificationLoading ? 'Sending…' : 'Resend verification email'}
          </button>
          {#if resendVerificationMessage}
            <span class="text-sm {resendVerificationMessage.startsWith('Verification') ? 'text-green-700' : 'text-red-600'}">{resendVerificationMessage}</span>
          {/if}
        </div>
      </div>
    {/if}

    <div class="mb-6 flex justify-center md:justify-end">
      <button
        on:click={handleAddVenue}
        disabled={addVenueDisabled}
        class="font-semibold py-1.5 px-3 text-sm rounded-lg shadow-md transition-colors duration-200 flex items-center gap-1.5 md:py-3 md:px-6 md:text-base md:gap-2 {addVenueDisabled
          ? 'bg-gray-400 text-gray-200 cursor-not-allowed'
          : 'bg-blue-600 hover:bg-blue-700 text-white'}"
      >
        <svg class="w-4 h-4 md:w-5 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        <span>Add Venue</span>
      </button>
    </div>

    <!-- Venue limit reached popup -->
    {#if showLimitPopup}
      <!-- svelte-ignore a11y_no_static_element_interactions - backdrop click to close -->
      <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
        aria-modal="true"
        on:click={closeLimitPopup}
        on:keydown={(e) => e.key === 'Escape' && closeLimitPopup()}
      >
        <div
          class="bg-white rounded-xl shadow-xl max-w-sm w-full p-6"
          role="dialog"
          aria-labelledby="limit-popup-title"
          tabindex="-1"
          on:click|stopPropagation
          on:keydown|stopPropagation
        >
          <h2 id="limit-popup-title" class="text-lg font-semibold text-gray-900 mb-3">
            Venue limit reached
          </h2>
          <p class="text-gray-600 mb-5">
            Free tier allows at most {venueLimit} venues. Upgrade to add more.
          </p>
          <button
            type="button"
            on:click={closeLimitPopup}
            class="w-full py-2 px-4 rounded-lg bg-gray-200 hover:bg-gray-300 text-gray-800 font-medium transition-colors"
          >
            OK
          </button>
        </div>
      </div>
    {/if}

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
          disabled={owner.email_verified === false}
          class="font-semibold py-2 px-6 rounded-lg shadow-md transition-colors duration-200 {owner.email_verified === false
            ? 'bg-gray-400 text-gray-200 cursor-not-allowed'
            : 'bg-blue-600 hover:bg-blue-700 text-white'}"
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
              <BannerImage
                src={venue.banner_image}
                alt={venue.name}
                size="sm"
                wrapperClass="rounded-t-xl rounded-b-none bg-gray-200"
              />
            {/if}
            <div class="p-4 sm:p-6 flex flex-col flex-grow">
              <h3 class="text-xl font-bold text-gray-900 mb-2">{venue.name}</h3>
              {#if owner?.name}
                <p class="text-sm text-gray-600 mb-1">Owner: {owner.name}</p>
              {/if}
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
                  <p class="text-sm font-medium text-gray-700 mb-2 md:mb-3">Event Lists:</p>
                  <div class="space-y-1 md:space-y-2">
                    {#each venueEventLists as eventList (eventList.event_list_uuid)}
                      <div class="flex items-center justify-between gap-2 py-1.5 px-2 md:p-2 bg-gray-50 rounded-lg">
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
                            title="Copy shareable direct link"
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

              <div class="flex flex-row gap-2 mt-auto">
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
