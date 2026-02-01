<script>
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import '../app.css';
  import { dev } from '$app/environment';
  import { currentOwnerStore } from '$lib/stores';
  import { goto } from '$app/navigation';
  import { getCurrentOwner } from '$lib/api/auth.js';
  import { browser } from '$app/environment';

  $: isVenueForm = $page.url.pathname === '/venue-form';
  $: isVenueOwner = $page.url.pathname === '/venue-owner';

  let mobileMenuOpen = false;
  let userMenuOpen = false;
  /** @type {HTMLElement | undefined} */
  let userMenuEl;
  let isOnline = true;

  /** @type {() => void} */
  let handleOnline;
  /** @type {() => void} */
  let handleOffline;

  function toggleMobileMenu() {
    mobileMenuOpen = !mobileMenuOpen;
  }

  function closeMobileMenu() {
    mobileMenuOpen = false;
  }

  function toggleUserMenu() {
    userMenuOpen = !userMenuOpen;
  }

  function closeUserMenu() {
    userMenuOpen = false;
  }

  /** @param {MouseEvent} e */
  function handleClickOutside(e) {
    if (userMenuEl && !userMenuEl.contains(/** @type {Node} */ (e.target))) {
      closeUserMenu();
    }
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
      handleOnline = () => {
        isOnline = true;
      };
      handleOffline = () => {
        isOnline = false;
      };
      window.addEventListener('online', handleOnline);
      window.addEventListener('offline', handleOffline);
      document.addEventListener('click', handleClickOutside);
    }
  });

  onDestroy(() => {
    if (browser && handleOnline && handleOffline) {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    }
    if (browser) {
      document.removeEventListener('click', handleClickOutside);
    }
  });

  async function logout() {
    const { logout: logoutApi } = await import('$lib/api/auth.js');
    await logoutApi();
    closeMobileMenu();
    closeUserMenu();
    goto('/');
  }
</script>

<div class="flex flex-col min-h-screen w-full min-w-0 max-w-[100vw] overflow-x-clip">
  {#if !isOnline}
    <div class="bg-red-600 text-white text-center py-2 px-4">
      <p class="text-sm font-medium">
        You are currently offline. Some features may not be available.
      </p>
    </div>
  {/if}
  <header class="bg-gray-100 shadow-sm min-w-0 w-full max-w-[100vw] overflow-x-clip">
    <nav
      class="container mx-auto h-14 md:h-20 flex items-center justify-between relative min-w-0 max-w-full overflow-visible"
    >
      <div class="flex items-center pl-4 md:pl-0">
        <a
          href="/"
          class="flex items-center gap-1.5 md:gap-4 hover:opacity-80 transition-opacity"
          on:click={closeMobileMenu}
        >
          <img
            src="/house_clock.png"
            alt="times.place logo"
            class="h-10 md:h-14 w-auto object-contain"
            style="aspect-ratio: 886/762;"
          />
          <span class="text-xl md:text-2xl font-bold text-gray-900">times.place</span>
        </a>
      </div>

      <!-- Desktop: [Test Phase] centered in header -->
      <div class="hidden md:flex absolute left-1/2 -translate-x-1/2 h-full items-center pointer-events-none">
        <a
          href="/demo"
          class="pointer-events-auto text-red-600 hover:text-red-700 font-medium text-[14px] transition-colors"
          >[Test Phase]</a
        >
      </div>

      <!-- Desktop Navigation -->
      <div class="hidden md:flex items-center gap-2">
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
        <span class="text-gray-400">|</span>
        <a
          href="/price"
          class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors"
          >Price</a
        >
        <span class="text-gray-400">|</span>
        <!-- User menu: icon + dropdown (login, logout, account links) -->
        <div class="relative" bind:this={userMenuEl}>
          <button
            type="button"
            class="flex items-center justify-center w-10 h-10 rounded-full text-gray-600 hover:text-gray-900 hover:bg-gray-200 transition-colors focus:outline-none focus:ring-2 focus:ring-inset focus:ring-red-500"
            on:click|stopPropagation={toggleUserMenu}
            aria-label="Account menu"
            aria-expanded={userMenuOpen}
            aria-haspopup="true"
          >
            {#if $currentOwnerStore}
              <span
                class="flex items-center justify-center w-8 h-8 rounded-full bg-red-100 text-red-700 font-semibold text-sm"
                title={$currentOwnerStore.name}
              >
                {$currentOwnerStore.name?.charAt(0)?.toUpperCase() ?? '?'}
              </span>
            {:else}
              <svg
                class="w-6 h-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                />
              </svg>
            {/if}
          </button>
          {#if userMenuOpen}
            <div
              class="absolute right-0 mt-2 w-52 py-1 bg-white rounded-lg shadow-lg border border-gray-200 z-50"
              role="menu"
            >
              {#if $currentOwnerStore}
                <a
                  href="/venue-owner"
                  class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  role="menuitem"
                  on:click={closeUserMenu}>My Venues</a
                >
                <a
                  href="/my"
                  class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  role="menuitem"
                  on:click={closeUserMenu}>Account</a
                >
                <button
                  type="button"
                  class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  role="menuitem"
                  on:click={logout}>Logout</button
                >
              {:else}
                <a
                  href="/login"
                  class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  role="menuitem"
                  on:click={closeUserMenu}>Login</a
                >
                <a
                  href="/registration"
                  class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  role="menuitem"
                  on:click={closeUserMenu}>Register</a
                >
                <a
                  href="/my"
                  class="block px-4 py-2 text-sm text-gray-500 hover:bg-gray-100"
                  role="menuitem"
                  on:click={closeUserMenu}>Account</a
                >
              {/if}
            </div>
          {/if}
        </div>
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
        <div
          class="absolute top-14 md:top-20 left-0 right-0 bg-gray-100 border-t border-gray-200 shadow-lg md:hidden z-50"
        >
          <div class="container mx-auto py-2 flex flex-col gap-0 min-w-0 max-w-full">
            <a
              href="/demo"
              class="text-red-600 hover:text-red-700 font-medium text-[14px] transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
              on:click={closeMobileMenu}>[Test Phase]</a
            >
            <a
              href="/"
              class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
              on:click={closeMobileMenu}>Home</a
            >
            <a
              href="/about"
              class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
              on:click={closeMobileMenu}>About</a
            >
            <a
              href="/price"
              class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
              on:click={closeMobileMenu}>Price</a
            >
            {#if $currentOwnerStore}
              <a
                href="/venue-owner"
                class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
                on:click={closeMobileMenu}>My Venues</a
              >
              <a
                href="/my"
                class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
                on:click={closeMobileMenu}>Account</a
              >
              <button
                type="button"
                class="text-left text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
                on:click={logout}
              >
                Logout
              </button>
            {:else}
              <a
                href="/my"
                class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
                on:click={closeMobileMenu}>Account</a
              >
              <a
                href="/login"
                class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
                on:click={closeMobileMenu}>Login</a
              >
              <a
                href="/registration"
                class="text-gray-700 hover:text-gray-900 font-medium text-base transition-colors px-4 py-1.5 hover:bg-gray-200 rounded-md"
                on:click={closeMobileMenu}>Register</a
              >
            {/if}
          </div>
        </div>
      {/if}
    </nav>
  </header>

  <main class="flex-1 bg-white w-full min-w-0 max-w-[100vw] overflow-x-clip">
    <div
      class="container mx-auto py-4 w-full min-w-0 max-w-[100vw] box-border overflow-x-clip {isVenueForm || isVenueOwner
        ? 'md:pt-0 md:pb-12'
        : 'md:py-12'}"
    >
      <slot />
    </div>
  </main>

  <footer class="bg-gray-100 border-t border-gray-200 min-w-0 w-full max-w-[100vw] overflow-x-clip">
    <div class="container mx-auto py-2 sm:py-3 flex items-center min-h-0 min-w-0 max-w-full">
      <div
        class="flex flex-col sm:flex-row justify-between items-center gap-1 sm:gap-3 w-full"
      >
        <p class="text-gray-600 text-xs sm:text-sm">
          &copy; 2026 times.place. All rights reserved.
          <a
            href="/disclaimer"
            class="text-blue-600 hover:text-blue-800 font-medium transition-colors ml-2"
            >Disclaimer</a
          >
        </p>
        <p class="text-gray-600 text-xs sm:text-sm">
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
