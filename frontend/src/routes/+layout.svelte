<script>
  import '../app.css';
</script>

<div class="flex flex-col min-h-screen">
  <header class="bg-gray-100 shadow-sm">
    <nav class="container mx-auto h-20 flex items-center justify-between">
      <div class="flex items-center">
        <a
          href="/"
          class="flex items-center gap-4 hover:opacity-80 transition-opacity"
        >
          <img
            src="/house_clock.png"
            alt="time.place logo"
            class="h-14 w-auto object-contain"
            style="aspect-ratio: 886/762;"
          />
          <span class="text-2xl font-bold text-gray-900">time.place</span>
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

  <main class="flex-1 bg-white w-full">
    <div class="container mx-auto py-12">
      <slot />
    </div>
  </main>

  <footer class="bg-gray-100 border-t border-gray-200">
    <div class="container mx-auto py-12">
      <div class="flex flex-col sm:flex-row justify-between items-center gap-4">
        <p class="text-gray-600 text-sm">
          &copy; 2024 time.place. All rights reserved.
        </p>
        <p class="text-gray-600 text-sm">
          Contact: <a
            href="mailto:timeplaceadmin@atomicmail.io"
            target="_blank"
            rel="noopener noreferrer"
            class="text-blue-600 hover:text-blue-800 font-medium transition-colors"
            >timeplaceadmin@atomicmail.io</a
          >
        </p>
      </div>
    </div>
  </footer>
</div>
