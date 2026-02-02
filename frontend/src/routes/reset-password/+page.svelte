<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { resetPassword } from '$lib/api/auth.js';
  import { ApiError } from '$lib/api/client.js';
  import { goto } from '$app/navigation';

  let token = '';
  let password = '';
  let confirmPassword = '';
  let error = '';
  let success = '';
  let isLoading = false;
  let passwordError = '';
  let confirmError = '';

  onMount(() => {
    token = $page.url.searchParams.get('token') || '';
    if (!token) {
      error =
        'Reset token is missing. Please use the link provided in your email.';
    }
  });

  function validatePasswords() {
    passwordError = '';
    confirmError = '';

    if (password.length < 6) {
      passwordError = 'Password must be at least 6 characters';
      return false;
    }

    if (password !== confirmPassword) {
      confirmError = 'Passwords do not match';
      return false;
    }

    return true;
  }

  /** @param {SubmitEvent} e */
  async function handleSubmit(e) {
    e.preventDefault();
    if (!validatePasswords()) return;

    error = '';
    success = '';
    isLoading = true;

    try {
      const response = await resetPassword(token, password);
      success = response.message;
      // Redirect to login after 3 seconds
      setTimeout(() => {
        goto('/login');
      }, 3000);
    } catch (err) {
      if (err instanceof ApiError) {
        error = err.message || 'Reset failed. Your link may have expired.';
      } else {
        error = 'An unexpected error occurred. Please try again.';
      }
    } finally {
      isLoading = false;
    }
  }
</script>

<svelte:head>
  <title>Reset Password - times.place</title>
</svelte:head>

<div class="max-w-5xl mx-auto px-6 sm:px-8 lg:px-12 py-12">
  <div class="bg-white rounded-xl shadow-lg p-8 md:p-12 max-w-2xl mx-auto">
    <h1 class="text-3xl font-bold mb-4 text-gray-900 text-center">
      Reset Password
    </h1>

    {#if !token && !success}
      <div
        class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800 mb-6"
      >
        {error}
      </div>
      <div class="text-center">
        <a
          href="/forgot-password"
          class="text-sm text-blue-600 hover:text-blue-800 font-medium"
        >
          Request a new reset link
        </a>
      </div>
    {:else}
      <p class="text-gray-600 mb-8 text-center">
        Enter your new password below.
      </p>

      <div class="bg-gray-50 p-6 md:p-8 rounded-lg">
        <form class="space-y-5" on:submit={handleSubmit}>
          <div>
            <label
              class="block text-sm font-medium text-gray-700 mb-1"
              for="password">New Password</label
            >
            <input
              id="password"
              type="password"
              class="w-full rounded-md border px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 {passwordError
                ? 'border-red-500'
                : 'border-gray-300'}"
              bind:value={password}
              autocomplete="new-password"
              placeholder="At least 6 characters"
              required
              disabled={isLoading || !!success}
            />
            {#if passwordError}
              <p class="mt-1 text-sm text-red-600">{passwordError}</p>
            {/if}
          </div>

          <div>
            <label
              class="block text-sm font-medium text-gray-700 mb-1"
              for="confirmPassword">Confirm New Password</label
            >
            <input
              id="confirmPassword"
              type="password"
              class="w-full rounded-md border px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 {confirmError
                ? 'border-red-500'
                : 'border-gray-300'}"
              bind:value={confirmPassword}
              autocomplete="new-password"
              placeholder="Confirm your new password"
              required
              disabled={isLoading || !!success}
            />
            {#if confirmError}
              <p class="mt-1 text-sm text-red-600">{confirmError}</p>
            {/if}
          </div>

          {#if error}
            <div
              class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
            >
              {error}
            </div>
          {/if}

          {#if success}
            <div
              class="rounded-md border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800"
            >
              {success} Redirecting to login...
            </div>
          {:else}
            <button
              type="submit"
              disabled={isLoading}
              class="w-full inline-flex justify-center rounded-md bg-blue-600 px-4 py-2 text-white font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? 'Resetting...' : 'Reset Password'}
            </button>
          {/if}
        </form>
      </div>
    {/if}
  </div>
</div>
