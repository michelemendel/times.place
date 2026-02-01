<script>
  import { get } from 'svelte/store';
  import { goto } from '$app/navigation';
  import { currentOwnerStore } from '$lib/stores';
  import { getAuthMe, deleteAccount } from '$lib/api/auth.js';
  import { onMount } from 'svelte';

  /** @type {number | null} */
  let venueCount = null;
  /** @type {number | null} */
  let venueLimit = null;
  let loading = true;
  /** @type {string | null} */
  let deleteError = null;
  let deleting = false;

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

<div class="bg-white rounded-xl shadow-lg pt-4 px-8 pb-8 md:pt-12 md:px-12 md:pb-12 max-w-2xl mx-auto text-[12px] md:text-base">
  <h1 class="text-[28px] md:text-4xl font-bold mb-4 text-gray-900 text-center">Account</h1>

  {#if loading}
    <p class="text-gray-600 text-center py-8">Loading…</p>
  {:else if !isLoggedIn}
    <p class="text-[14px] md:text-lg text-gray-600 mb-6 text-center">
      Sign in to view and manage your account.
    </p>
    <div class="flex flex-col sm:flex-row gap-4 justify-center">
      <a
        href="/login"
        class="inline-flex justify-center items-center px-6 py-3 bg-red-600 text-white font-medium rounded-lg hover:bg-red-700 transition-colors text-[12px] md:text-base"
      >
        Log in
      </a>
      <a
        href="/registration"
        class="inline-flex justify-center items-center px-6 py-3 border border-gray-300 text-gray-700 font-medium rounded-lg hover:bg-gray-50 transition-colors text-[12px] md:text-base"
      >
        Register
      </a>
    </div>
  {:else if owner}
    <div class="space-y-6">
      <section class="bg-gray-50 rounded-lg p-6">
        <h2 class="text-[14px] md:text-lg font-semibold text-gray-900 mb-3">Profile</h2>
        <dl class="space-y-2 text-gray-700">
          <div>
            <dt class="text-[10px] md:text-sm text-gray-500">Name</dt>
            <dd>{owner.name}</dd>
          </div>
          <div>
            <dt class="text-[10px] md:text-sm text-gray-500">Email</dt>
            <dd>{owner.email}</dd>
          </div>
          {#if owner.mobile}
            <div>
              <dt class="text-[10px] md:text-sm text-gray-500">Mobile</dt>
              <dd>{owner.mobile}</dd>
            </div>
          {/if}
        </dl>
      </section>

      <section class="bg-gray-50 rounded-lg p-6">
        <h2 class="text-[14px] md:text-lg font-semibold text-gray-900 mb-3">Account</h2>
        <ul class="space-y-2">
          <li>
            <a
              href="/venue-owner"
              class="text-red-600 hover:text-red-700 font-medium transition-colors"
            >
              My Venues
            </a>
            {#if venueCount != null && venueLimit != null}
              <span class="text-gray-500 text-[10px] md:text-sm ml-2">({venueCount} / {venueLimit})</span>
            {/if}
          </li>
          <!-- Placeholder for future billing / plan -->
          <li class="text-gray-500 text-[10px] md:text-sm">
            Billing and plan options will appear here when available.
          </li>
        </ul>
      </section>

      <section class="bg-gray-50 rounded-lg p-6 border border-red-200">
        <h2 class="text-[14px] md:text-lg font-semibold text-gray-900 mb-3">Danger zone</h2>
        <p class="text-gray-600 text-[10px] md:text-sm mb-3">
          Permanently delete your account and all your venues. This cannot be undone.
        </p>
        {#if deleteError}
          <p class="text-red-600 text-[10px] md:text-sm mb-3">{deleteError}</p>
        {/if}
        <button
          type="button"
          class="inline-flex items-center px-4 py-2 bg-red-600 text-white font-medium rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-[12px] md:text-base"
          disabled={deleting}
          onclick={async () => {
            if (!confirm('Permanently delete your account and all your venues? This cannot be undone.')) return;
            deleteError = null;
            deleting = true;
            try {
              await deleteAccount();
              goto('/');
            } catch (e) {
              deleteError = e instanceof Error ? e.message : 'Failed to delete account. Please try again.';
            } finally {
              deleting = false;
            }
          }}
        >
          {deleting ? 'Deleting…' : 'Delete account'}
        </button>
      </section>
    </div>
  {/if}
</div>
