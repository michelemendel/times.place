<script>
  import { onMount } from 'svelte';
  import { listOwners, deleteOwner } from '$lib/api/admin';
  import { currentOwnerStore } from '$lib/stores';

  let owners = [];
  let loading = true;
  let error = null;
  let searchQuery = '';

  onMount(async () => {
    try {
      owners = await listOwners();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  });

  async function handleDelete(uuid, name) {
    if (
      !confirm(
        `Are you sure you want to delete owner "${name}"? This action cannot be undone.`,
      )
    ) {
      return;
    }

    try {
      await deleteOwner(uuid);
      owners = owners.filter((o) => o.owner_uuid !== uuid);
    } catch (err) {
      alert(`Failed to delete owner: ${err.message}`);
    }
  }

  $: filteredOwners = owners.filter(
    (o) =>
      o.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      o.email.toLowerCase().includes(searchQuery.toLowerCase()),
  );
</script>

<div class="space-y-6">
  <div class="flex justify-between items-center">
    <h1 class="text-2xl font-bold text-gray-900">Manage Owners</h1>
    <div class="relative">
      <input
        type="text"
        placeholder="Search owners..."
        bind:value={searchQuery}
        class="pl-4 pr-10 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
      />
      <span class="absolute right-3 top-2.5 text-gray-400">
        <svg
          class="h-5 w-5"
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
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
              >Name</th
            >
            <th
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
              >Email</th
            >
            <th
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
              >Venues</th
            >
            <th
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
              >Role</th
            >
            <th
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
              >Created</th
            >
            <th
              class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
              >Actions</th
            >
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          {#each filteredOwners as owner (owner.owner_uuid)}
            <tr class="hover:bg-gray-50">
              <td
                class="px-6 py-4 text-sm font-medium text-gray-900 max-w-[200px] truncate"
                title={owner.name}>{owner.name}</td
              >
              <td
                class="px-6 py-4 text-sm text-gray-500 max-w-[250px] truncate"
                title={owner.email}>{owner.email}</td
              >
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500"
                >{owner.venue_count}</td
              >
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {#if owner.is_admin}
                  <span
                    class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800"
                  >
                    Admin
                  </span>
                {/if}
                {#if owner.is_demo}
                  <span
                    class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 ml-1"
                  >
                    Demo
                  </span>
                {/if}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {new Date(owner.created_at).toLocaleDateString()}
              </td>
              <td
                class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium"
              >
                {#if owner.owner_uuid !== $currentOwnerStore?.owner_uuid && !owner.is_demo}
                  <button
                    on:click={() => handleDelete(owner.owner_uuid, owner.name)}
                    class="text-red-600 hover:text-red-900 focus:outline-none"
                  >
                    Delete
                  </button>
                {:else if owner.is_demo}
                  <span
                    class="text-gray-400 cursor-not-allowed"
                    title="Demo accounts cannot be deleted">Delete</span
                  >
                {:else}
                  <span class="text-gray-400 cursor-not-allowed"
                    >Current User</span
                  >
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
      {#if filteredOwners.length === 0}
        <div class="px-6 py-4 text-center text-gray-500">No owners found.</div>
      {/if}
    </div>
  {/if}
</div>
