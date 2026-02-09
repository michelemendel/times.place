<script>
  import { page } from '$app/stores';
  $: canonicalUrl = $page.url.origin + $page.url.pathname;
  // Price page
</script>

<svelte:head>
  <title>Price - times.place</title>
  <meta name="description" content="Pricing for times.place. Free for now, with a limit of 2 venues per owner." />
  <meta property="og:title" content="Price - times.place" />
  <meta property="og:description" content="Pricing for times.place. Free for now, with a limit of 2 venues per owner." />
  <meta property="og:url" content={canonicalUrl} />
  <link rel="canonical" href={canonicalUrl} />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="Price - times.place" />
  <meta name="twitter:description" content="Pricing for times.place. Free for now, with a limit of 2 venues per owner." />
  <meta name="twitter:image" content="{$page.url.origin}/favicon.png" />
</svelte:head>

<div class="bg-white rounded-xl shadow-lg pt-4 px-8 pb-8 md:pt-12 md:px-12 md:pb-12 text-[12px] md:text-base">
  <h1 class="text-[28px] md:text-4xl font-bold mb-8 text-gray-900 text-center">
    Price
  </h1>

  <div class="space-y-6 text-gray-700 leading-relaxed max-w-2xl mx-auto text-center">
    <p>
      Everything is free for now, but with a limit of 2 venues per owner.
    </p>
    <p>
      Enjoy.
    </p>
  </div>
</div>
