<script>
  import '../app.css';
</script>

<div class="flex flex-col min-h-screen">
  <header class="bg-white shadow-sm">
    <nav class="container mx-auto h-20 flex items-center justify-between">
      <div class="flex items-center">
        <a
          href="/"
          class="flex items-center gap-4 hover:opacity-80 transition-opacity"
        >
          <img
            src="/house_clock.png"
            alt="Times & Place logo"
            class="h-14 w-auto object-contain"
            style="aspect-ratio: 886/762;"
          />
          <span class="text-2xl font-bold text-gray-900">Times & Place</span>
        </a>
      </div>
      <div class="flex items-center gap-8">
        <a
          href="/about"
          class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
          >About</a
        >
        <a
          href="/login"
          class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
          >Login</a
        >
      </div>
    </nav>
  </header>

  <main class="flex-1 bg-gray-50 w-full">
    <div class="container mx-auto py-12">
      <slot />
    </div>
  </main>

  <footer class="bg-white border-t border-gray-100">
    <div class="container mx-auto py-12">
      <div class="flex flex-col sm:flex-row justify-between items-center gap-4">
        <p class="text-gray-600 text-sm">
          &copy; 2024 times.place. All rights reserved.
        </p>
        <p class="text-gray-600 text-sm">
          Contact: <a
            href="mailto:timesplaceadmin@atomicmail.io"
            target="_blank"
            rel="noopener noreferrer"
            class="text-blue-600 hover:text-blue-800 font-medium transition-colors"
            >timesplaceadmin@atomicmail.io</a
          >
        </p>
      </div>
    </div>
  </footer>
</div>
