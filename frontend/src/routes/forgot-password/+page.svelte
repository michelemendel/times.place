<script lang="ts">
  import { forgotPassword } from '$lib/api/auth.js';
  import { ApiError } from '$lib/api/client.js';

  let email = '';
  let error = '';
  let success = '';
  let isLoading = false;
  let emailError = '';

  function normalizeEmail(value: string) {
    return value.trim().toLowerCase();
  }

  function validateEmailBlur() {
    const em = normalizeEmail(email);
    emailError = !em
      ? 'Email is required'
      : !em.includes('@')
        ? 'Please enter a valid email address'
        : '';
  }

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    success = '';
    emailError = '';

    const normalizedEmail = normalizeEmail(email);
    if (!normalizedEmail || !normalizedEmail.includes('@')) {
      emailError = !normalizedEmail
        ? 'Email is required'
        : 'Please enter a valid email address';
      return;
    }

    isLoading = true;
    try {
      const response = await forgotPassword(normalizedEmail);
      success = response.message;
      email = '';
    } catch (err) {
      if (err instanceof ApiError) {
        error = err.message || 'Request failed. Please try again.';
      } else {
        error = 'An unexpected error occurred. Please try again.';
      }
    } finally {
      isLoading = false;
    }
  }
</script>

<svelte:head>
  <title>Forgot Password - times.place</title>
</svelte:head>

<div class="max-w-5xl mx-auto px-6 sm:px-8 lg:px-12 py-12">
  <div class="bg-white rounded-xl shadow-lg p-8 md:p-12 max-w-2xl mx-auto">
    <h1 class="text-3xl font-bold mb-4 text-gray-900 text-center">
      Forgot Password
    </h1>
    <p class="text-gray-600 mb-8 text-center">
      Enter your email address and we'll send you a link to reset your password.
    </p>

    <div class="bg-gray-50 p-6 md:p-8 rounded-lg">
      <form class="space-y-5" on:submit={handleSubmit}>
        <div>
          <label
            class="block text-sm font-medium text-gray-700 mb-1"
            for="email">Email Address</label
          >
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
            required
            disabled={isLoading}
          />
          {#if emailError}
            <p class="mt-1 text-sm text-red-600">{emailError}</p>
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
            {success}
          </div>
        {:else}
          <button
            type="submit"
            disabled={isLoading}
            class="w-full inline-flex justify-center rounded-md bg-blue-600 px-4 py-2 text-white font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Sending Link...' : 'Send Reset Link'}
          </button>
        {/if}

        <div class="text-center pt-4">
          <a
            href="/login"
            class="text-sm text-blue-600 hover:text-blue-800 font-medium"
          >
            Back to login
          </a>
        </div>
      </form>
    </div>
  </div>
</div>
