<script>
  import { currentOwnerStore, authInitialized } from '$lib/stores';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';

  $: if (browser && $authInitialized) {
    if (!$currentOwnerStore) {
      goto('/login?redirect=/backoffice');
    } else if (!$currentOwnerStore.is_admin) {
      goto('/');
    }
  }
</script>

{#if $authInitialized && $currentOwnerStore && $currentOwnerStore.is_admin}
  <div class="flex flex-col md:flex-row min-h-[calc(100vh-200px)] gap-6">
    <aside
      class="w-full md:w-64 flex-shrink-0 bg-gray-50 border border-gray-200 rounded-lg p-4 h-fit"
    >
      <h2 class="text-lg font-bold text-gray-800 mb-4 border-b pb-2">
        Backoffice
      </h2>
      <nav class="space-y-1">
        <a
          href="/backoffice"
          class="block px-3 py-2 rounded-md text-gray-700 hover:bg-gray-200 hover:text-gray-900 transition-colors"
          >Dashboard</a
        >
        <a
          href="/backoffice/owners"
          class="block px-3 py-2 rounded-md text-gray-700 hover:bg-gray-200 hover:text-gray-900 transition-colors"
          >Owners</a
        >
        <a
          href="/backoffice/venues"
          class="block px-3 py-2 rounded-md text-gray-700 hover:bg-gray-200 hover:text-gray-900 transition-colors"
          >Venues</a
        >
      </nav>
    </aside>
    <main class="flex-1 bg-white border border-gray-200 rounded-lg p-6">
      <slot />
    </main>
  </div>
{:else}
  <div class="flex justify-center items-center py-20">
    <div
      class="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"
    ></div>
  </div>
{/if}
