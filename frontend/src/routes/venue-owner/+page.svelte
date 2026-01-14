<script>
  import { onMount } from 'svelte';
  import { get } from 'svelte/store';
  import { goto } from '$app/navigation';
  import { currentOwnerStore } from '$lib/stores';

  /** @type {import('$lib/types').VenueOwner | null} */
  let owner = null;

  onMount(() => {
    owner = get(currentOwnerStore);
    if (!owner) {
      goto('/login');
      return;
    }
  });
</script>

<svelte:head>
  <title>My Venues - time.place</title>
</svelte:head>

<div class="max-w-5xl mx-auto px-6 sm:px-8 lg:px-12">
  <div class="mb-8 text-center">
    <h1 class="text-4xl font-bold mb-4 text-gray-900">My Venues</h1>
    <p class="text-lg text-gray-600">Manage your venues and event schedules.</p>
  </div>

  <div class="bg-white rounded-xl shadow-lg p-8 md:p-12">
    {#if owner}
      <div class="mb-6 rounded-md border border-gray-200 bg-gray-50 px-4 py-3">
        <div class="text-sm text-gray-700">
          Signed in as <span class="font-semibold">{owner.name}</span>
          <span class="text-gray-500">({owner.email})</span>
        </div>
      </div>
    {/if}

    <p class="text-gray-500 italic text-center">
      Venue owner dashboard will be implemented next (this page is now protected and tied to the logged-in owner).
    </p>
  </div>
</div>
