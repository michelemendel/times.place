<script>
  import { get } from 'svelte/store';
  import { currentOwnerStore } from '$lib/stores';
  import { getAuthMe } from '$lib/api/auth.js';
  import { onMount } from 'svelte';

  /** @type {number | null} */
  let venueCount = null;
  /** @type {number | null} */
  let venueLimit = null;
  let loading = true;

  $: owner = $currentOwnerStore;
  $: isLoggedIn = !!owner;

  onMount(async () => {
    if (!get(currentOwnerStore)) {
      loading = false;
      return;
    }
    try {
      const me = await getAuthMe();
      venueCount = me.venue_count ?? null;
      venueLimit = me.venue_limit ?? null;
    } catch {
      // leave venue counts null
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Account - times.place</title>
</svelte:head>

<div class="bg-white rounded-xl shadow-lg p-8 md:p-12 max-w-2xl mx-auto">
  <h1 class="text-4xl font-bold mb-4 text-gray-900 text-center">Account</h1>

  {#if loading}
    <p class="text-gray-600 text-center py-8">Loading…</p>
  {:else if !isLoggedIn}
    <p class="text-lg text-gray-600 mb-6 text-center">
      Sign in to view and manage your account.
    </p>
    <div class="flex flex-col sm:flex-row gap-4 justify-center">
      <a
        href="/login"
        class="inline-flex justify-center items-center px-6 py-3 bg-red-600 text-white font-medium rounded-lg hover:bg-red-700 transition-colors"
      >
        Log in
      </a>
      <a
        href="/registration"
        class="inline-flex justify-center items-center px-6 py-3 border border-gray-300 text-gray-700 font-medium rounded-lg hover:bg-gray-50 transition-colors"
      >
        Register
      </a>
    </div>
  {:else if owner}
    <div class="space-y-6">
      <section class="bg-gray-50 rounded-lg p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-3">Profile</h2>
        <dl class="space-y-2 text-gray-700">
          <div>
            <dt class="text-sm text-gray-500">Name</dt>
            <dd>{owner.name}</dd>
          </div>
          <div>
            <dt class="text-sm text-gray-500">Email</dt>
            <dd>{owner.email}</dd>
          </div>
          {#if owner.mobile}
            <div>
              <dt class="text-sm text-gray-500">Mobile</dt>
              <dd>{owner.mobile}</dd>
            </div>
          {/if}
        </dl>
      </section>

      <section class="bg-gray-50 rounded-lg p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-3">Account</h2>
        <ul class="space-y-2">
          <li>
            <a
              href="/venue-owner"
              class="text-red-600 hover:text-red-700 font-medium transition-colors"
            >
              My Venues
            </a>
            {#if venueCount != null && venueLimit != null}
              <span class="text-gray-500 text-sm ml-2">({venueCount} / {venueLimit})</span>
            {/if}
          </li>
          <!-- Placeholder for future billing / plan -->
          <li class="text-gray-500 text-sm">
            Billing and plan options will appear here when available.
          </li>
        </ul>
      </section>
    </div>
  {/if}
</div>
