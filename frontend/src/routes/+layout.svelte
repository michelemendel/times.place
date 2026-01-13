<script>
  import { onMount } from 'svelte';
  import { seedDemoData } from '../lib/demo_data';
  import '../app.css';
  import { dev } from '$app/environment';

  let mobileMenuOpen = false;

  function toggleMobileMenu() {
    mobileMenuOpen = !mobileMenuOpen;
  }

  function closeMobileMenu() {
    mobileMenuOpen = false;
  }

  onMount(() => {
    // Force re-seed in development mode so changes to demo_data.ts are reflected
    seedDemoData(dev);
  });
</script>

<div class="flex flex-col min-h-screen">
  <header class="bg-gray-100 shadow-sm">
    <nav class="container mx-auto h-20 flex items-center justify-between relative">
      <div class="flex items-center pl-4 md:pl-0">
        <a
          href="/"
          class="flex items-center gap-2 md:gap-4 hover:opacity-80 transition-opacity"
          on:click={closeMobileMenu}
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

      <!-- Desktop Navigation -->
      <div class="hidden md:flex items-center gap-8">
        <a
          href="/"
          class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
          >Home</a
        >
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

      <!-- Mobile Hamburger Button -->
      <button
        type="button"
        class="md:hidden p-2 rounded-md text-gray-700 hover:text-gray-900 hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
        on:click={toggleMobileMenu}
        aria-label="Toggle menu"
        aria-expanded={mobileMenuOpen}
      >
        {#if mobileMenuOpen}
          <!-- Close icon (X) -->
          <svg
            class="h-6 w-6"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        {:else}
          <!-- Hamburger icon -->
          <svg
            class="h-6 w-6"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 6h16M4 12h16M4 18h16"
            />
          </svg>
        {/if}
      </button>

      <!-- Mobile Menu -->
      {#if mobileMenuOpen}
        <div class="absolute top-20 left-0 right-0 bg-gray-100 border-t border-gray-200 shadow-lg md:hidden z-50">
          <div class="container mx-auto py-4 flex flex-col gap-4">
            <a
              href="/"
              class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-2 hover:bg-gray-200 rounded-md"
              on:click={closeMobileMenu}
              >Home</a
            >
            <a
              href="/about"
              class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-2 hover:bg-gray-200 rounded-md"
              on:click={closeMobileMenu}
              >About</a
            >
            <a
              href="/login"
              class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-2 hover:bg-gray-200 rounded-md"
              on:click={closeMobileMenu}
              >Login</a
            >
          </div>
        </div>
      {/if}
    </nav>
  </header>

  <main class="flex-1 bg-white w-full">
    <div class="container mx-auto py-4 md:py-12">
      <slot />
    </div>
  </main>

  <footer class="bg-gray-100 border-t border-gray-200">
    <div class="container mx-auto h-20 flex items-center">
      <div class="flex flex-col sm:flex-row justify-between items-center gap-4 w-full">
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
