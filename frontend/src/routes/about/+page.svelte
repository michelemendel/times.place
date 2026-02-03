<script>
  // About page
</script>

<svelte:head>
  <title>About - times.place</title>
</svelte:head>

<div
  class="bg-white rounded-xl shadow-lg pt-4 px-8 pb-8 md:pt-12 md:px-12 md:pb-12 text-[12px] md:text-base"
>
  <h1 class="text-[28px] md:text-4xl font-bold mb-8 text-gray-900 text-center">
    About times.place
  </h1>

  <div class="space-y-6 text-gray-700 leading-relaxed max-w-4xl mx-auto">
    <p class="text-[14px] md:text-lg text-center">
      times.place is a venue management and event scheduling platform designed
      to help venues share their event schedules with the public. Whether you're
      looking for prayer times at a synagogue, community events, or other
      scheduled activities, times.place makes it easy to find up-to-date
      information.
    </p>

    <h2 class="text-[20px] md:text-2xl font-semibold mt-10 mb-4 text-gray-900">
      How It Works
    </h2>
    <p>times.place organizes information in a hierarchical structure:</p>

    <!-- Entity Relationship Diagram -->
    <div class="my-8 p-6 bg-gray-50 rounded-lg border border-gray-200">
      <div class="flex flex-col items-center space-y-4">
        <!-- Owner -->
        <div
          class="bg-blue-100 border-2 border-blue-300 rounded-lg px-6 py-4 text-center min-w-[200px]"
        >
          <div class="font-bold text-[14px] md:text-lg text-blue-900">
            Owner
          </div>
          <div class="text-[10px] md:text-sm text-blue-700 mt-1">
            Manages venues
          </div>
        </div>

        <!-- Arrow down -->
        <svg
          class="w-6 h-6 text-gray-600"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 14l-7 7m0 0l-7-7m7 7V3"
          />
        </svg>

        <!-- Venue -->
        <div
          class="bg-green-100 border-2 border-green-300 rounded-lg px-6 py-4 text-center min-w-[200px]"
        >
          <div class="font-bold text-[14px] md:text-lg text-green-900">
            Venue
          </div>
          <div class="text-[10px] md:text-sm text-green-700 mt-1">
            A physical location
          </div>
        </div>

        <!-- Arrow down -->
        <svg
          class="w-6 h-6 text-gray-600"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 14l-7 7m0 0l-7-7m7 7V3"
          />
        </svg>

        <!-- Event List -->
        <div
          class="bg-purple-100 border-2 border-purple-300 rounded-lg px-6 py-4 text-center min-w-[200px]"
        >
          <div class="font-bold text-[14px] md:text-lg text-purple-900">
            Event List
          </div>
          <div class="text-[10px] md:text-sm text-purple-700 mt-1">
            A collection of events
          </div>
        </div>

        <!-- Arrow down -->
        <svg
          class="w-6 h-6 text-gray-600"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 14l-7 7m0 0l-7-7m7 7V3"
          />
        </svg>

        <!-- Events -->
        <div
          class="bg-orange-100 border-2 border-orange-300 rounded-lg px-6 py-4 text-center min-w-[200px]"
        >
          <div class="font-bold text-[14px] md:text-lg text-orange-900">
            Events
          </div>
          <div class="text-[10px] md:text-sm text-orange-700 mt-1">
            Individual scheduled items
          </div>
        </div>
      </div>
    </div>

    <div class="mt-6 space-y-4">
      <h3 class="text-[16px] md:text-xl font-semibold text-gray-900">
        What is an Event?
      </h3>
      <p>
        An <strong>event</strong> is a single scheduled item with a specific time.
        For example, a prayer service, a meeting, or a class session. Each event
        has:
      </p>
      <ul class="list-disc list-inside space-y-2 ml-4">
        <li>A name (e.g., "Morning Prayer", "Evening Service")</li>
        <li>An optional date (e.g., "2024-12-25")</li>
        <li>A specific time (e.g., "8:00 AM", "6:30 PM")</li>
        <li>An optional duration</li>
        <li>An optional comment or description</li>
      </ul>
      <p>
        Events are grouped into <strong>event lists</strong>, which can
        represent a day's schedule, a weekly pattern, or any other logical
        grouping. Event lists belong to a <strong>venue</strong>, and venues are
        managed by an <strong>owner</strong>.
      </p>
      <div class="p-4 bg-red-50 border-l-4 border-red-500 rounded mt-4">
        <p class="font-semibold text-[12px] md:text-base text-red-900 mb-2">
          Note for Venue Owners:
        </p>
        <p class="text-[12px] md:text-base text-red-800">
          While both event lists and individual events can have dates, it's best
          to use one or the other (or neither). Using both in the same list can
          be confusing to your visitors.
        </p>
      </div>
    </div>

    <h2 class="text-[20px] md:text-2xl font-semibold mt-10 mb-4 text-gray-900">
      For Visitors
    </h2>
    <p>
      Browse venues and view their event schedules. Find the information you
      need quickly and easily.
    </p>

    <h2 class="text-[20px] md:text-2xl font-semibold mt-10 mb-4 text-gray-900">
      For Venue Owners
    </h2>
    <p>
      Register and manage your venues, create event lists, and keep your
      schedules up to date. You have full control over your venue information
      and event times.
    </p>

    <h2 class="text-[20px] md:text-2xl font-semibold mt-10 mb-4 text-gray-900">
      Timezone Settings
    </h2>
    <p>
      When creating events for your venue, you can optionally set a timezone.
      This affects how event times are displayed to visitors:
    </p>
    <div class="mt-4 space-y-4">
      <div class="p-4 bg-blue-50 border-l-4 border-blue-500 rounded">
        <p class="font-semibold text-[12px] md:text-base text-blue-900 mb-2">
          Timezone Set:
        </p>
        <p class="text-[12px] md:text-base text-blue-800">
          If you set a timezone for your venue (e.g., "Israel Time
          (Jerusalem)"), all visitors will see event times displayed in that
          timezone, regardless of where they are located. For example, if you
          set an event for 14:00 in Jerusalem time, a visitor in New York will
          also see 14:00 (Jerusalem time), not the converted time in their local
          timezone.
        </p>
      </div>
      <div class="p-4 bg-green-50 border-l-4 border-green-500 rounded">
        <p class="font-semibold text-[12px] md:text-base text-green-900 mb-2">
          No Timezone Set:
        </p>
        <p class="text-[12px] md:text-base text-green-800">
          If you leave the timezone empty, each visitor will see event times
          converted to their own browser's timezone. For example, if you set an
          event for 14:00 and a visitor is in New York, they might see 09:00
          (their local time), while a visitor in London might see 14:00 (their
          local time).
        </p>
      </div>
    </div>
    <p class="mt-4">
      <strong>When to set a timezone:</strong> Set a timezone when you want all visitors
      to see the same local time for your venue, regardless of where they are. This
      is especially useful for venues that serve a local community and want to display
      times in the venue's local time.
    </p>
    <p>
      <strong>When to leave it empty:</strong> Leave the timezone empty if you want
      each visitor to see times automatically converted to their own location. This
      is useful for events that need to be understood in the visitor's local context.
    </p>

    <h2 class="text-[20px] md:text-2xl font-semibold mt-10 mb-4 text-gray-900">
      Visibility & Security
    </h2>
    <p>
      times.place includes a visibility feature as a security measure to help
      you control who can see your event list and event information.
    </p>
    <p>
      <strong>Public visibility:</strong> Event lists marked as public are visible
      to everyone and appear in search results. A venue appears in the public venue
      list if it has at least one public event list. This is ideal for community
      events that you want to share broadly.
    </p>
    <p>
      <strong>Private visibility:</strong> Event lists marked as private are only
      accessible via a special shareable link. This link contains a secure token
      that you can share with specific people. Private event lists do not appear
      in public searches, providing an additional layer of privacy and security for
      your information.
    </p>
    <p>
      As a venue owner, you choose the visibility setting for each event list,
      giving you granular control over who can access your schedules.
    </p>

    <h2 class="text-[20px] md:text-2xl font-semibold mt-10 mb-4 text-gray-900">
      Contact
    </h2>
    <p>For questions, support, or feedback, please contact us at:</p>
    <p>
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
</div>
