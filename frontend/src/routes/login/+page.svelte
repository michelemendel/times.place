<script>
  import { get } from 'svelte/store';
  import { currentOwnerStore } from '$lib/stores';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { login } from '$lib/api/auth.js';
  import { ApiError } from '$lib/api/client.js';

  let email = '';
  let password = '';
  let error = '';
  let success = '';
  let isLoading = false;

  // If already logged in, take them to owner page
  onMount(() => {
    const current = get(currentOwnerStore);
    if (current) goto('/venue-owner');
  });

  /** @param {string} value */
  function normalizeEmail(value) {
    return value.trim().toLowerCase();
  }

  /** @param {SubmitEvent} e */
  async function handleSubmit(e) {
    e.preventDefault();
    error = '';
    success = '';
    isLoading = true;

    const normalizedEmail = normalizeEmail(email);
    if (!normalizedEmail || !normalizedEmail.includes('@')) {
      error = 'Please enter a valid email address.';
      isLoading = false;
      return;
    }

    if (!password) {
      error = 'Please enter your password.';
      isLoading = false;
      return;
    }

    try {
      const response = await login(normalizedEmail, password);
      success = `Welcome back, ${response.owner.name}.`;
      goto('/venue-owner');
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 401) {
          error = 'Invalid email or password.';
        } else if (err.status === 0) {
          error = 'Network error. Please check your connection.';
        } else {
          error = err.message || 'Login failed. Please try again.';
        }
      } else {
        error = 'An unexpected error occurred. Please try again.';
      }
    } finally {
      isLoading = false;
    }
  }
</script>

<svelte:head>
  <title>Login</title>
</svelte:head>

<div class="bg-white rounded-xl shadow-lg p-8 md:p-12">
  <h1 class="text-4xl font-bold mb-4 text-gray-900 text-center">Login</h1>
  <p class="text-lg text-gray-600 mb-8 text-center">
    Sign in to manage your venues and event schedules.
  </p>

  <div class="bg-gray-50 p-6 md:p-8 rounded-lg">
    <form class="max-w-xl mx-auto space-y-5" on:submit={handleSubmit}>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1" for="email">Email</label>
        <input
          id="email"
          type="email"
          class="w-full rounded-md border border-gray-300 px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          bind:value={email}
          autocomplete="email"
          placeholder="you@example.com"
          required
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1" for="password">Password</label>
        <input
          id="password"
          type="password"
          class="w-full rounded-md border border-gray-300 px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          bind:value={password}
          autocomplete="current-password"
          placeholder="Your password"
          required
        />
      </div>

      {#if error}
        <div class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {error}
        </div>
      {/if}

      {#if success}
        <div class="rounded-md border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
          {success}
        </div>
      {/if}

      <div class="flex flex-col sm:flex-row gap-3 sm:items-center sm:justify-between">
        <button
          type="submit"
          disabled={isLoading}
          class="inline-flex justify-center rounded-md bg-blue-600 px-4 py-2 text-white font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? 'Signing in...' : 'Sign in'}
        </button>
        <a class="text-sm text-blue-600 hover:text-blue-800 font-medium" href="/registration">
          Need an account? Register as a venue owner
        </a>
      </div>
    </form>
  </div>
</div>
