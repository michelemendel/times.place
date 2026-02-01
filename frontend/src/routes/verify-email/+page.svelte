<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';

  let status = 'loading'; // 'loading' | 'success' | 'error'
  let message = '';

  onMount(async () => {
    const token = $page.url.searchParams.get('token');
    if (!token) {
      status = 'error';
      message = 'Verification link is missing. Please use the link from your email.';
      return;
    }

    try {
      const response = await fetch(`/api/auth/verify-email?token=${encodeURIComponent(token)}`, {
        method: 'GET',
        credentials: 'include',
      });
      const data = await response.json().catch(() => ({}));
      if (response.ok) {
        status = 'success';
        message = data.message || 'Email verified.';
        setTimeout(() => goto('/venue-owner'), 2000);
      } else {
        status = 'error';
        message = data?.error?.message || 'Invalid or expired verification link. Please request a new one.';
      }
    } catch {
      status = 'error';
      message = 'Something went wrong. Please try again.';
    }
  });
</script>

<svelte:head>
  <title>Verify email – Times.Place</title>
</svelte:head>

<div class="min-h-[60vh] flex flex-col items-center justify-center px-4">
  {#if status === 'loading'}
    <p class="text-gray-600">Verifying your email…</p>
  {:else if status === 'success'}
    <p class="text-green-700 font-medium">{message}</p>
    <p class="text-sm text-gray-500 mt-2">Redirecting to My Venues…</p>
  {:else}
    <p class="text-red-600 font-medium">{message}</p>
    <a href="/venue-owner" class="mt-4 text-blue-600 hover:underline">Back to My Venues</a>
  {/if}
</div>
