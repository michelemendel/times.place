<script>
  import { seedDemoData } from '../../lib/demo_data';
  import { currentOwnerStore, ownersStore } from '../../lib/stores';
  import { get } from 'svelte/store';
  import { goto } from '$app/navigation';

  function handleResetData() {
    if (confirm('Reset all demo data? This will restore all deleted venues and clear any changes you\'ve made. You may need to log in again.')) {
      const currentOwner = get(currentOwnerStore);
      const currentEmail = currentOwner?.email;

      // Reset the data
      seedDemoData(true);

      // If the user was logged in as a demo owner, restore their login with the new owner object
      if (currentEmail) {
        const owners = get(ownersStore);
        const matchingOwner = owners.find(o => o.email === currentEmail);
        if (matchingOwner) {
          currentOwnerStore.set(matchingOwner);
        } else {
          // User was logged in as a non-demo owner, log them out
          currentOwnerStore.set(null);
          goto('/');
        }
      }
    }
  }
</script>

<svelte:head>
  <title>Prototype - times.place</title>
</svelte:head>

<div class="bg-white rounded-xl shadow-lg p-8 md:p-12">
  <h1 class="text-4xl font-bold mb-8 text-gray-900 text-center">
    Prototype Information
  </h1>

  <div class="space-y-6 text-gray-700 leading-relaxed max-w-3xl mx-auto">
    <div class="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-6">
      <p class="text-lg font-semibold text-yellow-800 mb-2">
        This is a prototype application
      </p>
      <p class="text-yellow-700">
        This version of times.place is a working prototype. All data is stored locally in your browser and will not persist across different devices or browsers. Data is stored using browser localStorage, which means it's only available on the device and browser where it was created.
      </p>
    </div>

    <div class="border-t border-gray-200 pt-6">
      <h2 class="text-2xl font-semibold mb-4 text-gray-900">
        Test User Accounts
      </h2>
      <p class="mb-4">
        Use these demo accounts to test the application:
      </p>
      <div class="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-6">
        <div class="space-y-3">
          <div>
            <p class="font-semibold text-gray-900">User 1:</p>
            <p class="text-gray-700">Email: <code class="bg-white px-2 py-1 rounded text-sm">abe@demo.org</code></p>
            <p class="text-gray-700">Password: <code class="bg-white px-2 py-1 rounded text-sm">demo</code></p>
          </div>
          <div>
            <p class="font-semibold text-gray-900">User 2:</p>
            <p class="text-gray-700">Email: <code class="bg-white px-2 py-1 rounded text-sm">ben@demo.org</code></p>
            <p class="text-gray-700">Password: <code class="bg-white px-2 py-1 rounded text-sm">demo</code></p>
          </div>
        </div>
      </div>
      <p class="text-sm text-gray-600 mb-4">
        These accounts are created when you use the "Reset All Demo Data" button below.
      </p>
    </div>

    <div class="border-t border-gray-200 pt-6">
      <h2 class="text-2xl font-semibold mb-4 text-gray-900">
        Report Bugs & Suggestions
      </h2>
      <p class="mb-4">
        Found a bug or have a suggestion for improvement? Please send your feedback to the developer:
      </p>
      <div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
        <p class="text-gray-700">
          <a
            href="mailto:timesandplaceadmin@atomicmail.io"
            target="_blank"
            rel="noopener noreferrer"
            class="text-blue-600 hover:text-blue-800 font-medium transition-colors"
          >
            timesandplaceadmin@atomicmail.io
          </a>
        </p>
      </div>
      <p class="text-sm text-gray-600">
        Your feedback helps improve the application. Please include details about any bugs you encounter or suggestions you have.
      </p>
    </div>

    <div class="border-t border-gray-200 pt-6">
      <h2 class="text-2xl font-semibold mb-4 text-gray-900">
        Reset Demo Data
      </h2>
      <p class="mb-4">
        If you want to reset all demo data and restore the original sample venues and event lists, you can use the button below. This will clear all your changes and restore the initial demo data.
      </p>
      <button
        type="button"
        class="px-6 py-3 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition-colors duration-200"
        on:click={handleResetData}
      >
        Reset All Demo Data
      </button>
      <p class="mt-4 text-sm text-gray-600">
        <strong>Note:</strong> This action cannot be undone. All your custom venues, event lists, and events will be replaced with the original demo data.
      </p>
    </div>
  </div>
</div>
