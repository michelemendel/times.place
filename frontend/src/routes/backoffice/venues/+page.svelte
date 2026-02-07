<script>
  import { onMount } from 'svelte';
  import { listVenues } from '$lib/api/admin';

  /** @type {import('$lib/types').Venue[]} */
  let venues = [];
  let loading = true;
  /** @type {string | null} */
  let error = null;
  let searchQuery = '';

  onMount(async () => {
    try {
      // @ts-ignore - listVenues in admin might not return exactly Venue[] or is missing typedef
      venues = await listVenues();
    } catch (err) {
      error = err instanceof Error ? err.message : 'An error occurred';
    } finally {
      loading = false;
    }
  });

  $: filteredVenues = venues.filter(
    (v) =>
      v.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (v.owner_name &&
        v.owner_name.toLowerCase().includes(searchQuery.toLowerCase())) ||
      (v.owner_email &&
        v.owner_email.toLowerCase().includes(searchQuery.toLowerCase())),
  );
</script>

<div class="space-y-4">
  <div class="space-y-2">
    <h1 class="text-xl font-bold text-gray-900">Manage Venues</h1>
    <div class="relative max-w-xs">
      <input
        type="text"
        placeholder="Search venues..."
        bind:value={searchQuery}
        class="w-full pl-3 pr-8 py-1.5 text-xs border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
      />
      <span class="absolute right-2 top-2 text-gray-400">
        <svg
          class="h-4 w-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      </span>
    </div>
  </div>

  {#if loading}
    <div class="flex justify-center py-10">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"
      ></div>
    </div>
  {:else if error}
    <div class="bg-red-50 text-red-600 p-4 rounded-md">
      Error: {error}
    </div>
  {:else}
    <div
      class="overflow-x-auto bg-white border border-gray-200 rounded-lg shadow-sm"
    >
      <table class="min-w-[40rem] w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th
              class="px-4 py-2 text-left text-[10px] font-medium text-gray-500 uppercase tracking-wider"
              >Venue Name</th
            >
            <th
              class="px-4 py-2 text-left text-[10px] font-medium text-gray-500 uppercase tracking-wider"
              >Owner</th
            >
            <th
              class="px-4 py-2 text-left text-[10px] font-medium text-gray-500 uppercase tracking-wider"
              >Address</th
            >
            <th
              class="px-4 py-2 text-left text-[10px] font-medium text-gray-500 uppercase tracking-wider"
              >Public</th
            >
            <th
              class="px-4 py-2 text-left text-[10px] font-medium text-gray-500 uppercase tracking-wider"
              >Private</th
            >
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          {#each filteredVenues as venue (venue.venue_uuid)}
            <tr class="hover:bg-gray-50">
              <td
                class="px-4 py-2 text-xs font-medium text-gray-900 max-w-[150px] truncate"
                title={venue.name}
              >
                <a
                  href="/?venue={venue.venue_uuid}"
                  target="_blank"
                  class="hover:underline text-blue-600"
                >
                  {venue.name}
                </a>
              </td>
              <td class="px-4 py-2 text-xs text-gray-500 max-w-[200px]">
                <div
                  class="text-gray-900 truncate"
                  title={venue.owner_name || ''}
                >
                  {venue.owner_name || 'Unknown'}
                </div>
                <div
                  class="text-[10px] text-gray-400 truncate"
                  title={venue.owner_email || ''}
                >
                  {venue.owner_email || ''}
                </div>
              </td>
              <td
                class="px-4 py-2 text-xs text-gray-500 max-w-[200px] truncate"
                title={venue.address}>{venue.address}</td
              >
              <td class="px-4 py-2 whitespace-nowrap text-xs text-gray-500">
                {venue.public_events_count ?? 0}
              </td>
              <td class="px-4 py-2 whitespace-nowrap text-xs text-gray-500">
                {venue.private_events_count ?? 0}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
      {#if filteredVenues.length === 0}
        <div class="px-6 py-4 text-center text-gray-500">No venues found.</div>
      {/if}
    </div>
  {/if}
</div>
