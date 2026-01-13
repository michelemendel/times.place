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
