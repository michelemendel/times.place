<script>
  import { onMount, onDestroy } from 'svelte';
  import '../app.css';
  import { dev } from '$app/environment';
  import { currentOwnerStore } from '$lib/stores';
  import { goto } from '$app/navigation';
  import { getCurrentOwner } from '$lib/api/auth.js';
  import { browser } from '$app/environment';

  let mobileMenuOpen = false;
  let isOnline = true;
  
  let handleOnline;
  let handleOffline;

  function toggleMobileMenu() {
    mobileMenuOpen = !mobileMenuOpen;
  }

  function closeMobileMenu() {
    mobileMenuOpen = false;
  }

  onMount(async () => {
    // Restore session on app initialization
    if (browser) {
      try {
        await getCurrentOwner();
      } catch (error) {
        // If /api/auth/me fails, user is not authenticated
        // This is expected for unauthenticated users, so we silently ignore
        // The error could be 401 (no valid token) or network error
      }
      
      // Set up offline detection
      isOnline = navigator.onLine;
      handleOnline = () => { isOnline = true; };
      handleOffline = () => { isOnline = false; };
      window.addEventListener('online', handleOnline);
      window.addEventListener('offline', handleOffline);
    }
  });
  
  onDestroy(() => {
    if (browser && handleOnline && handleOffline) {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    }
  });

  async function logout() {
    const { logout: logoutApi } = await import('$lib/api/auth.js');
    await logoutApi();
    closeMobileMenu();
    goto('/');
  }

</script>

<div class="flex flex-col min-h-screen">
  {#if !isOnline}
    <div class="bg-red-600 text-white text-center py-2 px-4">
      <p class="text-sm font-medium">You are currently offline. Some features may not be available.</p>
    </div>
  {/if}
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
            alt="times.place logo"
            class="h-14 w-auto object-contain"
            style="aspect-ratio: 886/762;"
          />
          <span class="text-2xl font-bold text-gray-900">times.place</span>
        </a>
      </div>

      <!-- Desktop Navigation -->
      <div class="hidden md:flex items-center gap-2">
        <a
          href="/prototype"
          class="text-red-600 hover:text-red-700 font-medium text-base transition-colors"
          >Prototype</a
        >
        <span class="text-gray-400">|</span>
        <a
          href="/"
          class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
          >Home</a
        >
        <span class="text-gray-400">|</span>
        <a
          href="/about"
          class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
          >About</a
        >
        {#if $currentOwnerStore}
          <span class="text-gray-400">|</span>
          <a
            href="/venue-owner"
            class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
            >My Venues</a
          >
          <span class="text-gray-400">|</span>
          <button
            type="button"
            class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
            on:click={logout}
          >
            Logout
          </button>
        {:else}
          <span class="text-gray-400">|</span>
          <a
            href="/login"
            class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
            >Login</a
          >
          <span class="text-gray-400">|</span>
          <a
            href="/registration"
            class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
            >Register</a
          >
        {/if}
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
          <div class="container mx-auto py-4 flex flex-col gap-2">
            <a
              href="/prototype"
              class="text-red-600 hover:text-red-700 font-medium text-base transition-colors px-4 py-2 hover:bg-gray-200 rounded-md"
              on:click={closeMobileMenu}
              >Prototype</a
            >
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
            {#if $currentOwnerStore}
              <a
                href="/venue-owner"
                class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-2 hover:bg-gray-200 rounded-md"
                on:click={closeMobileMenu}
                >My Venues</a
              >
              <button
                type="button"
                class="text-left text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-2 hover:bg-gray-200 rounded-md"
                on:click={logout}
              >
                Logout
              </button>
            {:else}
              <a
                href="/login"
                class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-2 hover:bg-gray-200 rounded-md"
                on:click={closeMobileMenu}
                >Login</a
              >
              <a
                href="/registration"
                class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-2 hover:bg-gray-200 rounded-md"
                on:click={closeMobileMenu}
                >Register</a
              >
            {/if}
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
          &copy; 2024 times.place. All rights reserved.
        </p>
        <p class="text-gray-600 text-sm">
          Contact: <a
            href="mailto:timesandplaceadmin@atomicmail.io"
            target="_blank"
            rel="noopener noreferrer"
            class="text-blue-600 hover:text-blue-800 font-medium transition-colors"
            >timesandplaceadmin@atomicmail.io</a
          >
        </p>
      </div>
    </div>
  </footer>
</div>
