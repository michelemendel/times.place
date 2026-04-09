<script>
  import { page } from '$app/stores';
  $: canonicalUrl = $page.url.origin + $page.url.pathname;
  // Disclaimer page
</script>

<svelte:head>
  <title>Disclaimer - times.place</title>
  <meta name="description" content="Disclaimer and terms for times.place." />
  <meta property="og:title" content="Disclaimer - times.place" />
  <meta property="og:description" content="Disclaimer and terms for times.place." />
  <meta property="og:url" content={canonicalUrl} />
  <link rel="canonical" href={canonicalUrl} />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="Disclaimer - times.place" />
  <meta name="twitter:description" content="Disclaimer and terms for times.place." />
  <meta name="twitter:image" content="{$page.url.origin}/favicon.png" />
</svelte:head>

<div class="bg-white rounded-xl shadow-lg p-8 md:p-12">
  <h1 class="text-4xl font-bold mb-8 text-gray-900 text-center">
    Disclaimer
  </h1>

  <div class="space-y-6 text-gray-700 leading-relaxed max-w-4xl mx-auto">
    <h2 class="text-2xl font-semibold mt-8 mb-4 text-gray-900">
      Accuracy of Information
    </h2>
    <p>
      times.place is a platform that allows venue owners to publish their event schedules and venue information. We are <strong>not responsible</strong> for the accuracy, completeness, or timeliness of the data displayed on this site. All venue details, schedules, event names, times, and other information are provided and maintained by the respective venue owners. It is the sole responsibility of venue owners to ensure that their published information is correct and up to date.
    </p>

    <h2 class="text-2xl font-semibold mt-8 mb-4 text-gray-900">
      No Warranty
    </h2>
    <p>
      This website and its content are provided &quot;as is&quot; without any warranties of any kind, either express or implied. We do not warrant that the site will be uninterrupted, error-free, or free of harmful components. Use of this service is at your own risk.
    </p>

    <h2 class="text-2xl font-semibold mt-8 mb-4 text-gray-900">
      Limitation of Liability
    </h2>
    <p>
      To the fullest extent permitted by applicable law, times.place and its operators shall not be liable for any direct, indirect, incidental, special, consequential, or punitive damages arising from your use of this site or reliance on any information published here—including but not limited to scheduling errors, cancelled or changed events, or inaccuracies in venue or event data.
    </p>

    <h2 class="text-2xl font-semibold mt-8 mb-4 text-gray-900">
      Third-Party Content and Links
    </h2>
    <p>
      This site may contain links to external websites (e.g. maps, directions) or content provided by third parties. We do not control and are not responsible for the content or practices of external sites. The inclusion of any link does not imply endorsement by times.place.
    </p>

    <h2 class="text-2xl font-semibold mt-8 mb-4 text-gray-900">
      Contact
    </h2>
    <p>
      If you have questions about this disclaimer or the use of data on this site, please contact us at <a href="mailto:timesandplaceadmin@atomicmail.io" class="text-blue-600 hover:text-blue-800 hover:underline">timesandplaceadmin@atomicmail.io</a>.
    </p>
  </div>
</div>
