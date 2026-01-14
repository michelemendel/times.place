# Implementation Log

## 2025-12-09

### Project Structure Setup

- Created monorepo structure with `frontend/` and `backend/` directories
- Moved all Go-related files to `backend/`:
  - `cmd/`, `domain/`, `utils/`, `go.mod`
- Decided on monorepo approach for easier code sharing and coordinated development

### Frontend Setup (SvelteKit)

- Initialized SvelteKit project in `frontend/` using JavaScript (not TypeScript)
- Configured for static site export using `@sveltejs/adapter-static`
- Set up prerendering with `export const prerender = true` in `+layout.js`
- Created basic project structure:
  - `src/routes/` for pages
  - `src/lib/` for reusable components
  - `static/` for static assets

### Tailwind CSS Configuration

- Installed Tailwind CSS v4
- Configured PostCSS with `@tailwindcss/postcss` plugin (v4 requires separate package)
- Updated CSS to use v4 syntax: `@import 'tailwindcss'` instead of `@tailwind` directives
- Created `tailwind.config.js` and `postcss.config.js`

### Makefile Setup

- Created Makefile with clear naming convention:
  - Frontend targets: `f-dev`, `f-build`, `f-preview`, `f-install`
  - Backend targets: `b-build`, `b-run`, `b-install` (placeholders)
  - Convenience shortcuts: `dev`, `build`, `preview` (default to frontend)
- Organized help output by component

### Assets and Logo

- Copied `house_clock.png` logo from `assets/` to `frontend/static/`
- Created favicon from logo:
  - Generated `favicon.ico` (32x32) using `sharp` and `to-ico` packages
  - Also kept `favicon.png` as fallback
  - Updated `app.html` to reference favicon
- Established convention: all static assets go in `frontend/static/` and are served from root path

### Demo Page

- Created initial demo page at `src/routes/+page.svelte` to verify setup
- Used Tailwind classes to confirm CSS is working
- Verified dev server runs successfully on `http://localhost:5173/`

### Issues Resolved

- Fixed Tailwind CSS v4 PostCSS configuration (required `@tailwindcss/postcss` package)
- Fixed static adapter prerendering requirement
- Fixed missing favicon 404 error during build

## 2025-12-25

### Data Model Implementation

- **Type Definitions (`src/lib/types.ts`)**:

  - Defined TypeScript interfaces for `VenueOwner`, `Venue`, `EventList`, and `Event`.
  - Established relationships using UUIDs (`owner_uuid`, `venue_uuid`, etc.).

- **State Management (`src/lib/stores.ts`)**:

  - Implemented Svelte `writable` stores for all data entities.
  - Added automatic persistence to `localStorage` (with safety check for SSR).
  - Used generic `load` helper to initialize stores from storage or default values.

- **Date/Time Handling (`src/lib/utils/datetime.js`)**:

  - Created utility functions for timezone-aware formatting:
    - `formatEventTime`: formats Unix timestamp to time string based on user locale/timezone.
    - `formatEventListDate`: formats ISO date string to full date string.
    - `formatFullDateTime`: useful for debugging/admin views.
  - Implemented JSDoc type safety (refining optional parameters).

- **Demo Data (`src/lib/demo_data.ts`)**:
  - Implemented `seedDemoData` function to generate initial valid data.
  - Includes a Venue Owner, Venue ("Beth El Synagogue"), Event List ("Daily Minyan"), and Events.
  - Uses a UUID generator fallback for wider compatibility.

## 2026-01-12

### Store Naming Refinement

- **Renamed `userStore` to `currentOwnerStore`** (`src/lib/stores.ts`):
  - Clarified that this store tracks the currently logged-in venue owner (not anonymous public users).
  - Updated localStorage key from `times_place_user` to `times_place_current_owner`.
  - Updated all imports and references in `demo_data.ts`.
  - Public visitors are anonymous and don't require any stored user data.

### Multiple Venue Owner Demo Accounts

- **Extended demo data** (`src/lib/demo_data.ts`):
  - Added `ownersStore` to store all venue owner accounts (for login functionality).
  - Created two venue owner accounts for testing isolation:
    - **Owner 1: "Demo Rabbi"** with 2 venues:
      - "Beth El Synagogue" (has 2 event lists: "Daily Minyan" and "Shabbat Services")
      - "Community Center" (has no event lists)
    - **Owner 2: "Sarah Cohen"** with 2 venues:
      - "Beit Midrash" (has 1 event list: "Weekly Schedule")
      - "Chabad House" (has no event lists)
  - Updated seed function to check for existing data in `ownersStore` or `venueStore` instead of just checking for logged-in user.
  - Added `seedDemoData()` call to layout (`src/routes/+layout.svelte`) to automatically seed on app load.

### Date/Time Utility Functions

- **Centralized date creation** (`src/lib/utils/datetime.js`):
  - Added `getCurrentTimestamp()`: Returns current timestamp as ISO 8601 string (RFC3339 format) for `created_at` and `modified_at` fields.
  - Added `updateModifiedTimestamp(entity)`: Helper function to update `modified_at` field when entities are modified (for future use).
  - Updated `demo_data.ts` to use `getCurrentTimestamp()` instead of direct `new Date().toISOString()` calls.
  - All date creation and modification operations are now centralized behind utility functions for easier maintenance.

## 2025-01-13

### Search Functionality Specification and Implementation

- **Searchable Fields Specification** (`blueprint/2_spec.md`):

  - Added "Searchable Fields" section after Data Model section.
  - Documented all fields that should be searchable:
    - Venue Owner: `name`, `mobile`
    - Venue: `name`, `address`, `comment`
    - Event List: `name`, `comment`
    - Event: `event_name`, `comment`
  - Specified that venues appear in search results if query matches any searchable field across venue, owner, event lists, or events.

- **Comprehensive Search Implementation** (`frontend/src/routes/+page.svelte`):
  - Implemented `venueMatchesSearch()` function that searches across all specified fields.
  - Replaced simple venue name filter with comprehensive multi-field search.
  - Search is case-insensitive and handles optional fields safely.
  - Users can now find venues by searching owner names, mobile numbers, addresses, comments, event list names, and event names.

### Clickable Address Links for Directions

- **Google Maps Integration** (`frontend/src/routes/+page.svelte`):
  - Added `getDirectionsUrl()` function that generates Google Maps directions URLs.
  - Prefers coordinates (from `geolocation` field) when available for accuracy.
  - Falls back to address string if coordinates are not available.
  - Made venue addresses clickable links that open Google Maps in a new tab.
  - Uses free Google Maps URL scheme (no API key required):
    - With coordinates: `https://www.google.com/maps/dir/?api=1&destination={lat},{lng}`
    - With address: `https://www.google.com/maps/search/?api=1&query={encoded_address}`
  - Links styled as blue clickable text with hover effects.

### Demo Data Updates

- **Real Jerusalem Addresses** (`frontend/src/lib/demo_data.ts`):
  - Updated all venue addresses to real Jerusalem street addresses for testing:
    - "Beth El Synagogue": `15 King George Street, Jerusalem` (31.7787, 35.2175)
    - "Community Center": `42 Ben Yehuda Street, Jerusalem` (31.7800, 35.2167)
    - "Beit Midrash": `28 Jaffa Road, Jerusalem` (31.7820, 35.2180)
    - "Chabad House": `12 Rechov Agron, Jerusalem` (31.7750, 35.2200)
  - Updated geolocation coordinates to match real Jerusalem locations.
  - All addresses are now testable with Google Maps links.

## 2026-01-14

### Venue Owner Registration, Authentication, and Authorization

- **Registration Flow** (`frontend/src/routes/registration/+page.svelte`):

  - Implemented complete registration form with fields: name, email, mobile, password, and confirm password.
  - Added form validation:
    - Name: required, non-empty
    - Email: required, must contain '@'
    - Mobile: required, validated with regex pattern (allows digits, spaces, +, -, parentheses, minimum 7 characters)
    - Password: required, minimum 4 characters
    - Confirm password: must match password
  - Creates new `VenueOwner` entity and stores in `ownersStore` (persisted to localStorage).
  - Automatically logs in the newly registered user by setting `currentOwnerStore`.
  - Redirects to `/venue-owner` after successful registration.

- **Login Flow** (`frontend/src/routes/login/+page.svelte`):

  - Implemented email + password authentication form.
  - Validates credentials by:
    - Looking up owner by email (case-insensitive, normalized)
    - Comparing provided password with stored password (or fallback to "demo" for legacy accounts)
  - Sets `currentOwnerStore` on successful login and redirects to `/venue-owner`.
  - Auto-redirects to `/venue-owner` if user is already logged in.

- **Authorization / Route Protection** (`frontend/src/routes/venue-owner/+page.svelte`):

  - Added client-side authorization guard: checks `currentOwnerStore` on mount.
  - Redirects to `/login` if no owner is logged in.
  - Displays logged-in owner information when authenticated.

- **Navigation Updates** (`frontend/src/routes/+layout.svelte`):

  - Updated header navigation to conditionally show:
    - **When logged out**: "Login" and "Register" links
    - **When logged in**: "My Venues" link and "Logout" button
  - Implemented logout functionality that clears `currentOwnerStore` and redirects to home.
  - Applied same conditional navigation to mobile menu.

- **Data Model Updates** (`frontend/src/lib/types.ts`):

  - Added optional `password?: string` field to `VenueOwner` interface for prototype password storage.
  - Documented that password storage is prototype-only and not secure.

- **Demo Data Updates** (`frontend/src/lib/demo_data.ts`):

  - Added `password: "demo"` to both demo venue owner accounts for testing.

- **UUID Utility Refactoring** (`frontend/src/lib/utils/uuid.js`):

  - Extracted `generateUUID()` function from inline implementations into shared utility module.
  - Updated `demo_data.ts` and `registration/+page.svelte` to import from `$lib/utils/uuid.js`.
  - Centralized UUID generation logic for consistency across the codebase.

- **Critical Bug Fixes**:
  - **Fixed data persistence issue**: Changed `seedDemoData(dev)` to `seedDemoData(false)` in both `+layout.svelte` and `+page.svelte`.
    - Previously, force re-seeding in dev mode was clearing localStorage and wiping newly registered accounts.
    - Now demo data only seeds when storage is empty, preserving user-registered accounts.
  - **Password field implementation**: Added password and confirm password fields to registration form.
    - Passwords are stored per-owner in localStorage (prototype-only, not secure).
    - Login validates against stored password with fallback to "demo" for legacy accounts without passwords.

### Venue Owner Dashboard Implementation

- **Venue Owner Dashboard** (`frontend/src/routes/venue-owner/+page.svelte`):

  - Implemented complete venue owner dashboard showing all venues owned by the logged-in owner.
  - **Client-side filtering**: Uses reactive Svelte store subscriptions (`$venueStore`) to automatically filter venues by `owner_uuid` matching the current logged-in owner.
  - **Alphabetical sorting**: Venues are sorted by name using `localeCompare()` for proper locale-aware sorting.
  - **Card-based layout**: Each venue displayed in a responsive card showing:
    - Banner image (if available)
    - Venue name, address, and comment
    - Event list count indicator
    - Edit and Delete action buttons
  - **Empty state**: Shows helpful message and "Add Your First Venue" button when no venues exist.
  - **Responsive design**:
    - Single column layout on mobile
    - Two-column grid layout on desktop (md breakpoint and above)
    - All buttons and text scale appropriately for different screen sizes

- **Add Venue Functionality**:

  - "Add Venue" button navigates to `/venue-form` (venue form implementation pending).

- **Edit Venue Functionality**:

  - "Edit" button on each venue card navigates to `/venue-form?venue_uuid={venue_uuid}` (venue form implementation pending).

- **Delete Venue Functionality**:

  - "Delete" button opens a confirmation modal with accessibility features:
    - Proper ARIA roles (`role="dialog"`, `aria-modal="true"`, `aria-labelledby`)
    - Keyboard support (Escape key to cancel)
    - Click outside modal to cancel (checks event target to prevent accidental closes)
  - On confirmation, deletes:
    - The venue from `venueStore`
    - All associated event lists from `eventListStore`
    - All associated events from `eventStore`
  - Uses proper cascade deletion to maintain data integrity.

- **Reactive Store Updates**:

  - Fixed initial implementation issue where venue list didn't update after deletion.
  - Changed from `get(venueStore)` to reactive store syntax `$venueStore` to ensure automatic updates when stores change.
  - All reactive statements properly subscribe to store changes for real-time UI updates.

- **Reset Demo Data Functionality** (`frontend/src/routes/+layout.svelte`):

  - Added "Reset Data" button to navigation menu (desktop and mobile) for development use.
  - Button appears between "My Venues" and "Logout" when user is logged in.
  - **Smart owner preservation**: When resetting demo data:
    - Saves current owner's email before reset
    - After reset, finds the owner with matching email in newly seeded data
    - If found (demo owner), updates `currentOwnerStore` with new owner object (preserves login and venue ownership)
    - If not found (custom registered user), logs them out since their account isn't in demo data
  - Prevents issue where venues would disappear after reset due to UUID mismatches.
  - Styled as red text to indicate it's a destructive action.

- **Technical Decisions**:
  - Used Svelte's reactive store syntax (`$store`) instead of `get(store)` in reactive statements for automatic subscription and updates.
  - Implemented modal backdrop click handling by checking `event.target === event.currentTarget` instead of using `stopPropagation` to avoid accessibility warnings.
  - Used JSDoc type annotations (`@type`, `@param`) instead of TypeScript syntax since project uses JavaScript.
  - All linter errors resolved, including accessibility warnings for modal interactions.
