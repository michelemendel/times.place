<script>
  import { goto } from '$app/navigation';
  import { register } from '$lib/api/auth.js';
  import { ApiError } from '$lib/api/client.js';

  let name = '';
  let mobile = '';
  let email = '';
  let password = '';
  let confirmPassword = '';

  let error = '';
  let success = '';
  let isLoading = false;

  // Field-level validation (show on blur and on submit failure)
  let nameError = '';
  let emailError = '';
  let passwordError = '';
  let confirmPasswordError = '';

  /** @param {string} value */
  function normalizeEmail(value) {
    return value.trim().toLowerCase();
  }

  /** @param {string} value */
  function normalizeMobile(value) {
    return value.trim();
  }

  /** @param {string} value */
  function isValidMobile(value) {
    // Very light validation for demo purposes: allow digits, spaces, +, -, ()
    const v = value.trim();
    if (!v) return true; // empty is valid (optional field)
    return /^[0-9+\-()\s]{7,}$/.test(v);
  }

  function validateNameBlur() {
    nameError = !name.trim() ? 'Name is required' : '';
  }

  function validateEmailBlur() {
    const em = normalizeEmail(email);
    emailError = !em ? 'Email is required' : !em.includes('@') ? 'Please enter a valid email address' : '';
  }

  function validatePasswordBlur() {
    passwordError = !password ? 'Password is required' : password.length < 6 ? 'Password must be at least 6 characters' : '';
  }

  function validateConfirmPasswordBlur() {
    confirmPasswordError = !confirmPassword ? 'Confirm password is required' : password !== confirmPassword ? 'Passwords do not match' : '';
  }

  /** Set all required-field errors (e.g. after failed submit) so user sees which fields are missing */
  function setRequiredFieldErrors() {
    nameError = !name.trim() ? 'Name is required' : '';
    const em = normalizeEmail(email);
    emailError = !em ? 'Email is required' : !em.includes('@') ? 'Please enter a valid email address' : '';
    passwordError = !password ? 'Password is required' : password.length < 6 ? 'Password must be at least 6 characters' : '';
    confirmPasswordError = !confirmPassword ? 'Confirm password is required' : password !== confirmPassword ? 'Passwords do not match' : '';
  }

  /** @param {SubmitEvent} e */
  async function handleSubmit(e) {
    e.preventDefault();
    error = '';
    success = '';
    isLoading = true;

    const n = name.trim();
    const m = normalizeMobile(mobile);
    const em = normalizeEmail(email);

    if (!n) {
      setRequiredFieldErrors();
      error = 'Please fix the required fields below.';
      isLoading = false;
      return;
    }
    if (!em || !em.includes('@')) {
      setRequiredFieldErrors();
      error = 'Please fix the required fields below.';
      isLoading = false;
      return;
    }
    if (!isValidMobile(m)) {
      error = 'Please enter a valid mobile number (or leave blank).';
      isLoading = false;
      return;
    }
    if (!password || password.length < 6) {
      setRequiredFieldErrors();
      error = 'Please fix the required fields below.';
      isLoading = false;
      return;
    }
    if (password !== confirmPassword) {
      setRequiredFieldErrors();
      error = 'Please fix the required fields below.';
      isLoading = false;
      return;
    }

    try {
      const response = await register({
        name: n,
        email: em,
        mobile: m,
        password,
      });
      success = `Account created. You are now logged in as ${response.owner.name}.`;
      goto('/venue-owner');
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 409) {
          error = 'An account with this email already exists. Please log in instead.';
        } else if (err.status === 400) {
          error = err.message || 'Please check your input and try again.';
        } else if (err.status === 0) {
          error = 'Network error. Please check your connection.';
        } else {
          error = err.message || 'Registration failed. Please try again.';
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
  <title>Registration - times.place</title>
</svelte:head>

<div class="max-w-5xl mx-auto px-6 sm:px-8 lg:px-12">
  <div class="bg-white rounded-xl shadow-lg p-8 md:p-12">
    <h1 class="text-4xl font-bold mb-4 text-gray-900 text-center">
      Register as Venue Owner
    </h1>
    <p class="text-lg text-gray-600 mb-8 text-center">
      Create an account to manage your venues and event schedules.
    </p>

    <div class="bg-gray-50 p-6 md:p-8 rounded-lg">
      <div class="max-w-xl mx-auto">
        <form class="space-y-5" on:submit={handleSubmit}>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1" for="name">Name *</label>
            <input
              id="name"
              class="w-full rounded-md border px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 {nameError
                ? 'border-red-500'
                : 'border-gray-300'}"
              bind:value={name}
              on:blur={validateNameBlur}
              autocomplete="name"
              placeholder="Your name"
            />
            {#if nameError}
              <p class="mt-1 text-sm text-red-600">{nameError}</p>
            {/if}
          </div>

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
            <label class="block text-sm font-medium text-gray-700 mb-1" for="mobile"
              >Mobile (optional)</label
            >
            <input
              id="mobile"
              class="w-full rounded-md border border-gray-300 px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              bind:value={mobile}
              autocomplete="tel"
              placeholder="+1 555-555-5555"
            />
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
              autocomplete="new-password"
              placeholder="Choose a password (at least 6 characters)"
            />
            {#if passwordError}
              <p class="mt-1 text-sm text-red-600">{passwordError}</p>
            {:else}
              <p class="mt-2 text-sm text-gray-600">
                Password must be at least 6 characters long.
              </p>
            {/if}
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1" for="confirmPassword"
              >Confirm password *</label
            >
            <input
              id="confirmPassword"
              type="password"
              class="w-full rounded-md border px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 {confirmPasswordError
                ? 'border-red-500'
                : 'border-gray-300'}"
              bind:value={confirmPassword}
              on:blur={validateConfirmPasswordBlur}
              autocomplete="new-password"
              placeholder="Re-enter password"
            />
            {#if confirmPasswordError}
              <p class="mt-1 text-sm text-red-600">{confirmPasswordError}</p>
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
              {isLoading ? 'Creating account...' : 'Create account'}
            </button>
            <a class="text-sm text-blue-600 hover:text-blue-800 font-medium" href="/login">
              Already have an account? Log in
            </a>
          </div>
        </form>
      </div>
    </div>
  </div>
</div>
