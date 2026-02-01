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

  // Field-level validation (show on blur and on submit failure)
  let emailError = '';
  let passwordError = '';

  // If already logged in, take them to owner page
  onMount(() => {
    const current = get(currentOwnerStore);
    if (current) goto('/venue-owner');
  });

  /** @param {string} value */
  function normalizeEmail(value) {
    return value.trim().toLowerCase();
  }

  function validateEmailBlur() {
    const em = normalizeEmail(email);
    emailError = !em ? 'Email is required' : !em.includes('@') ? 'Please enter a valid email address' : '';
  }

  function validatePasswordBlur() {
    passwordError = !password ? 'Password is required' : '';
  }

  /** Set all required-field errors (e.g. after failed submit) so user sees which fields are missing */
  function setRequiredFieldErrors() {
    const em = normalizeEmail(email);
    emailError = !em ? 'Email is required' : !em.includes('@') ? 'Please enter a valid email address' : '';
    passwordError = !password ? 'Password is required' : '';
  }

  /** @param {SubmitEvent} e */
  async function handleSubmit(e) {
    e.preventDefault();
    error = '';
    success = '';
    isLoading = true;

    const normalizedEmail = normalizeEmail(email);
    if (!normalizedEmail || !normalizedEmail.includes('@')) {
      setRequiredFieldErrors();
      error = 'Please fix the required fields below.';
      isLoading = false;
      return;
    }

    if (!password) {
      setRequiredFieldErrors();
      error = 'Please fix the required fields below.';
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
  <title>Login - times.place</title>
</svelte:head>

<div class="max-w-5xl mx-auto px-6 sm:px-8 lg:px-12">
  <div class="bg-white rounded-xl shadow-lg p-8 md:p-12">
    <h1 class="text-4xl font-bold mb-4 text-gray-900 text-center">Login</h1>
    <p class="text-lg text-gray-600 mb-8 text-center">
      Sign in to manage your venues and event schedules.
    </p>

    <div class="bg-gray-50 p-6 md:p-8 rounded-lg">
      <div class="max-w-xl mx-auto">
        <form class="space-y-5" on:submit={handleSubmit}>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1" for="email">Email *</label>
            <input
              id="email"
              type="email"
              class="w-full rounded-md border px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 {emailError
                ? 'border-red-500'
                : 'border-gray-300'}"
              bind:value={email}
              on:blur={validateEmailBlur}
              autocomplete="email"
              placeholder="you@example.com"
            />
            {#if emailError}
              <p class="mt-1 text-sm text-red-600">{emailError}</p>
            {/if}
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1" for="password">Password *</label>
            <input
              id="password"
              type="password"
              class="w-full rounded-md border px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 {passwordError
                ? 'border-red-500'
                : 'border-gray-300'}"
              bind:value={password}
              on:blur={validatePasswordBlur}
              autocomplete="current-password"
              placeholder="Your password"
            />
            {#if passwordError}
              <p class="mt-1 text-sm text-red-600">{passwordError}</p>
            {/if}
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
  </div>
</div>
