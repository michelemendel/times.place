<script>
</script>

<svelte:head>
  <title>Demo - times.place</title>
</svelte:head>

<div class="bg-white rounded-xl shadow-lg p-8 md:p-12">
  <h1 class="text-4xl font-bold mb-8 text-gray-900 text-center">
    Demo
  </h1>

  <div class="space-y-6 text-gray-700 leading-relaxed max-w-3xl mx-auto">
    <div class="bg-amber-50 border-l-4 border-amber-400 p-4 mb-6">
      <p class="text-lg font-semibold text-amber-800 mb-2">
        Fully functional — currently in test mode
      </p>
      <p class="text-amber-700">
        times.place is fully functional. All features work end-to-end: registration, login, venue management, event lists, and public browsing. The site is currently running in test mode. Data is stored in a backend database and persists across sessions.
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
        These accounts are created when demo data is seeded in the backend database.
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

  </div>
</div>
