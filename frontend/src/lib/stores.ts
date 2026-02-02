import { writable } from 'svelte/store';
import type { VenueOwner } from './types';

// Authentication state store:
// The current authenticated venue owner, populated from /api/auth/me
// and cleared on logout. No localStorage persistence is used; the
// backend manages refresh tokens via HttpOnly cookies.
export const currentOwnerStore = writable<VenueOwner | null>(null);

// Flag to indicate if the initial auth check has completed
export const authInitialized = writable<boolean>(false);

