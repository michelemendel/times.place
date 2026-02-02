<script>
  import { currentOwnerStore } from '$lib/stores';
</script>

<div class="space-y-6">
  <h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
  <p class="text-gray-600">Welcome, {$currentOwnerStore?.name}.</p>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <a
      href="/backoffice/owners"
      class="block p-6 bg-white border border-gray-200 rounded-lg shadow-sm hover:shadow-md transition-shadow"
    >
      <h3 class="text-lg font-semibold text-gray-900 mb-2">Manage Owners</h3>
      <p class="text-gray-600">
        View registered owners, manage admins, and delete accounts.
      </p>
    </a>
    <a
      href="/backoffice/venues"
      class="block p-6 bg-white border border-gray-200 rounded-lg shadow-sm hover:shadow-md transition-shadow"
    >
      <h3 class="text-lg font-semibold text-gray-900 mb-2">Manage Venues</h3>
      <p class="text-gray-600">View all venues across the platform.</p>
    </a>
  </div>
</div>
