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

## 2026-01-16

### Venue Form Implementation

- **Two-Pane Layout** (`frontend/src/routes/venue-form/+page.svelte`):

  - Implemented split-screen layout with editing pane (left) and live preview pane (right).
  - Preview pane updates in real-time as user edits venue information, event lists, and events.
  - Both panes scroll independently with `max-h-[calc(100vh-200px)]` for better UX on long forms.

- **Venue Basic Information Form**:

  - Venue name (required), address, geolocation, comment, banner image, timezone, and visibility fields.
  - All optional fields labeled with "(optional)" for clarity.
  - Form validation ensures venue name is required before saving.

- **Event List Management**:

  - Add event list: Creates new event list with default name "New Event List" and today's date.
  - Delete event list: With confirmation dialog to prevent accidental deletion.
  - Reorder event lists: Move up/down buttons with proper disabled states at boundaries.
  - Each event list has: name (optional), date (required, ISO 8601 format), comment (optional), and events.

- **Event Management**:

  - Add event: Creates new event with default name "New Event" and time "12:00".
  - Delete event: Removes event from list and updates event_uuids array.
  - Duplicate event: Creates copy with "(Copy)" suffix in name.
  - Reorder events: Move up/down buttons with proper disabled states.
  - Each event has: name (optional), time (required, HH:MM format), duration (optional, minutes), comment (optional).

- **Event DateTime Handling**:

  - Time input uses HTML5 `type="time"` for native time picker.
  - Time is combined with event list date to create full RFC3339 datetime string.
  - Datetime is converted to Unix epoch timestamp (seconds) for storage.
  - Real-time datetime updates when time or event list date changes.
  - Preview pane shows formatted time using venue timezone (or user's timezone if venue timezone not set).

- **Input Validation**:

  - Date validation: Ensures ISO 8601 format (YYYY-MM-DD) and valid date values.
  - Time validation: Regex pattern ensures HH:MM format.
  - XSS prevention: `sanitizeInput()` function escapes HTML entities in all text inputs.
  - UUID validation: `isValidUUID()` function validates UUID format for all entity identifiers.
  - Clear error messages guide users to fix validation issues.

- **Undo Functionality**:

  - Implements undo stack with maximum 50 steps.
  - Saves state before each modification (add, delete, move, duplicate operations).
  - Undo button appears in header when undo stack has items.
  - Restores previous venue and event lists state.

- **Local Storage Persistence**:

  - All venue, event list, and event changes saved to localStorage via Svelte stores.
  - Automatic persistence on every save operation.
  - Data persists across page refreshes and browser sessions.

- **Image Upload**:

  - Banner image upload converts images to base64 data URLs.
  - 5MB file size limit with validation.
  - Preview shows uploaded image immediately.
  - Stored as base64 string in venue's `banner_image` field.

- **Reactivity Fixes**:

  - Fixed event duplication issues by ensuring proper array reassignment triggers Svelte reactivity.
  - Fixed event list visibility by improving reactive statement logic for data loading.
  - All array mutations now create new arrays/objects to trigger Svelte updates.
  - Fixed event moving functionality by properly reassigning arrays after swaps.

- **Accessibility Improvements**:

  - All form inputs have matching `id` and `label` `for` attributes.
  - Section headers changed from `<label>` to `<div>` where not associated with controls.
  - Proper ARIA attributes on interactive elements.
  - Keyboard navigation support throughout form.

- **Prerendering Fixes**:
  - Wrapped `$page.url.searchParams` access in `browser` checks to prevent SSR errors.
  - All client-only code properly guarded with `browser` checks.

### Security & Safety Features Implementation

- **Visibility Field** (`frontend/src/lib/types.ts`):

  - Added `visibility: 'public' | 'private'` field to `Venue` interface.
  - Added optional `private_link_token?: string` field for private venues.

- **Visibility Toggle** (`frontend/src/routes/venue-form/+page.svelte`):

  - Radio button group for Public/Private selection.
  - Private link token automatically generated when visibility set to private.
  - Private link displayed in form when venue is private.
  - Link format: `{origin}/?token={private_link_token}`.

- **Private Link Generation**:

  - Uses `crypto.randomUUID()` or fallback UUID generator for cryptographically secure tokens.
  - Token generated automatically when venue visibility changed to private.
  - Token cleared when venue changed to public.

- **Visitor Page Filtering** (`frontend/src/routes/+page.svelte`):

  - Filters venues to show only public venues OR private venues accessed via valid token.
  - Token extracted from URL query parameter: `/?token=...`
  - Private venues not accessible without valid token are hidden from search/dropdown.

- **Contact Information Privacy**:

  - Venue owner email and mobile hidden from public visitor page.
  - Contact information only visible to authenticated venue owners on their own venue edit pages.
  - Preview pane in venue form shows contact info for venue owner's reference.

- **Demo Data Updates** (`frontend/src/lib/demo_data.ts`):
  - Updated all venues to include `visibility` field (mix of public and private).
  - Added `private_link_token` to private venues.
  - Added `date` field to all event lists (ISO 8601 format).

### Timezone Implementation

- **Timezone Field** (`frontend/src/routes/venue-form/+page.svelte`):

  - Replaced text input with organized dropdown selector.
  - Timezones grouped by region: Americas, Europe, Asia & Middle East, Oceania, Africa.
  - Human-readable labels (e.g., "Israel Time (Jerusalem)" instead of "Asia/Jerusalem").
  - Default option: "No timezone" (uses visitor's browser timezone).
  - Label: "Timezone (optional, leave empty for your current timezone)".

- **Timezone-Aware Time Display**:

  - Updated `formatEventTime()` function to accept optional `timeZone` parameter.
  - Preview pane displays event times in venue's timezone when set.
  - Public page displays event times in venue's timezone when set.
  - Falls back to visitor's browser timezone if venue timezone not set.
  - Times update immediately when venue timezone changes.

- **Timezone Display**:
  - Timezone shown on preview pane and public page when set.
  - Displayed after address field with consistent styling.

### Geolocation Picker Implementation

- **Interactive Map** (`frontend/src/routes/venue-form/+page.svelte`):

  - Integrated Leaflet.js map for visual location selection.
  - Map shows existing geolocation when editing venue.
  - Default center: Jerusalem (31.7683, 35.2137) if no location set.

- **Address Geocoding**:

  - "Find on Map" button next to address field.
  - Uses OpenStreetMap Nominatim API (free, no API key required).
  - Improved query parameters for better partial address matching:
    - `format=jsonv2`: Better response format
    - `limit=5`: Get multiple results
    - `addressdetails=1`: Include detailed address components
    - `dedupe=1`: Remove duplicates
  - Handles partial addresses (e.g., "Jerusalem", "Main Street").
  - Shows dropdown with multiple results when multiple matches found.
  - Auto-selects single result when only one match.

- **Map Interaction**:

  - Click anywhere on map to set location.
  - Draggable marker for fine-tuning location.
  - Marker updates coordinates and triggers reverse geocoding to update address field.
  - Two-way sync: address → map, map → address.

- **Reverse Geocoding**:

  - When marker moved or map clicked, reverse geocodes coordinates to address.
  - Updates address field automatically with full address string.
  - Uses OpenStreetMap Nominatim reverse geocoding API.

- **Coordinates Display**:

  - Shows latitude,longitude in text field below map.
  - Can be edited manually.
  - Updates map when coordinates changed manually.
  - Format: `{lat.toFixed(6)},{lng.toFixed(6)}` for precision.

- **User Experience**:

  - Helper text explains how to use the picker.
  - Loading state on "Find on Map" button during geocoding.
  - Results dropdown styled with hover effects.
  - Map height: 256px (h-64) for good visibility without taking too much space.

### Bug Fixes and Refinements

- **Event Duplication Fix**:

  - Fixed issue where events were duplicated when loading/saving.
  - Improved deduplication logic in `loadVenueData()` and `saveVenue()`.
  - Ensures events appear only once per event list.

- **Event List Visibility Fix**:

  - Fixed reactive statement logic to properly load event lists when stores become available.
  - Improved `dataLoaded` flag management to prevent infinite retries.
  - Event lists now display correctly in both edit and preview panes.

- **Reactivity Improvements**:

  - All array/object mutations now create new objects to trigger Svelte reactivity.
  - Fixed event moving, adding, deleting, and duplicating to update UI immediately.
  - Fixed event list moving, adding, and deleting to update UI immediately.

- **Date Input Improvements**:

  - Kept HTML5 date picker (`type="date"`) for better UX.
  - Added helper text explaining ISO 8601 storage format.
  - Improved date validation with clearer error messages.
  - Defaults to today's date when creating new event lists.

- **Field Labeling**:

  - All optional fields labeled with "(optional)" for clarity.
  - Required fields marked with asterisk (*).
  - Consistent labeling across all form sections.

- **Timezone Field Fix**:
  - Fixed issue where empty timezone field disappeared.
  - Preserves empty strings when loading existing venues.
  - Only defaults to 'Asia/Jerusalem' for new venues or when value is truly missing.

### Technical Decisions

- **Svelte Reactivity**: Used array/object reassignment instead of mutations to ensure Svelte detects changes and updates UI.
- **Geocoding Service**: Chose OpenStreetMap Nominatim over Google Maps Geocoding API for free usage without API keys.
- **Map Library**: Used Leaflet.js (already in use on public page) for consistency and no additional dependencies.
- **Timezone Storage**: Stored as IANA timezone strings (e.g., "Asia/Jerusalem") for compatibility with JavaScript's Intl API.
- **Coordinate Precision**: Stored with 6 decimal places (approximately 0.1 meter precision) for accuracy without excessive storage.
- **Error Handling**: Graceful fallbacks when geocoding fails (user can still click on map or enter coordinates manually).

## 2025-01-17

### Event List Private Links and Print Functionality

- **Data Model Updates** (`frontend/src/lib/types.ts`):

  - Added `private_link_token?: string` field to `EventList` interface.
  - Each event list can now have its own private link token, independent of venue visibility.

- **Event Lists View Page** (`frontend/src/routes/venue-owner/[venue_uuid]/event-lists/+page.svelte`):

  - Created new route `/venue-owner/[venue_uuid]/event-lists` to display all event lists for a specific venue.
  - Shows venue name and basic info at the top.
  - Lists all event lists sorted by date, then by name.
  - For each event list, displays:
    - Event list name and date
    - "Get Private Link" button (copies link to clipboard with visual feedback)
    - "Print" button (opens print dialog using hidden iframe)
    - Preview of events (first 3 events with times)
  - Includes "Back to Venues" navigation button.
  - Handles authorization - redirects if venue owner doesn't own the venue.
  - Created `+page.js` file to mark route as non-prerenderable (requires authentication).

- **Venue Owner Page Updates** (`frontend/src/routes/venue-owner/+page.svelte`):

  - Added "View Event Lists" button to each venue card.
  - Button positioned alongside "Edit" and "Delete" buttons.
  - Updated card layout to use flexbox for consistent button alignment at bottom of cards.

- **Private Link Token Generation** (`frontend/src/routes/venue-form/+page.svelte`):

  - Automatically generates `private_link_token` when creating new event lists.
  - Preserves existing tokens when updating event lists.
  - Uses `crypto.randomUUID()` or fallback UUID generator (same pattern as venue tokens).
  - Tokens are generated on-the-fly if missing when accessing event lists page.

- **Visitor Page Updates** (`frontend/src/routes/+page.svelte`):

  - Supports accessing event lists via private link token in URL: `/?token={event_list_token}`.
  - When token matches an event list's `private_link_token`:
    - Automatically selects the venue that owns that event list.
    - Automatically selects that specific event list.
    - Displays the event list (same view as public access).
  - Works independently of venue visibility (private event lists can be shared even if venue is public).

- **Print Functionality**:

  - Print button uses hidden iframe to load public view without navigating away from event lists page.
  - Iframe loads the public view URL with the event list's private link token.
  - Triggers browser print dialog for iframe content.
  - User stays on event lists page throughout the process.
  - Falls back to opening new window if iframe approach fails.

- **Print Styles** (`frontend/src/app.css`):

  - Added comprehensive `@media print` CSS rules.
  - Hides navigation, buttons, and non-essential UI elements when printing.
  - Specifically hides:
    - Header "Find Venues and Events" and venue dropdown
    - Map display
    - Geolocation information
    - Timezone information
    - Event list dropdown selector (when multiple lists exist)
  - Optimizes layout for print output:
    - Removes background colors and shadows
    - Ensures proper page breaks
    - Prevents event information from splitting across pages
    - Uses appropriate font sizes for printed output

- **Demo Data Updates** (`frontend/src/lib/demo_data.ts`):

  - Added `private_link_token` to all event lists in demo data.
  - Generated tokens for all existing event lists using UUID generation pattern.

- **Build Configuration** (`frontend/svelte.config.js`):
  - Added `fallback: 'index.html'` to adapter-static configuration.
  - Enables SPA mode to support dynamic routes that require authentication.
  - Allows dynamic route `/venue-owner/[venue_uuid]/event-lists` to work in static build.

### Implementation Decisions

- **Event List Private Links**: Each event list gets its own `private_link_token` (stored on EventList), independent of venue-level private links. Link format: `/?token={event_list_token}`.
- **Print Implementation**: Used hidden iframe approach to print public view content while keeping user on event lists page. This avoids navigation and provides seamless printing experience.
- **Token Generation**: Tokens are generated automatically when event lists are created, and on-the-fly if missing when accessing the event lists page. This ensures all event lists have tokens available.
- **Print Styles**: Used CSS classes (`no-print-*`) to selectively hide UI elements during printing, providing clean print output with only essential information.
- **Route Configuration**: Dynamic route marked as non-prerenderable since it requires authentication and venue ownership verification.

## 2026-01-18

### My Venues Page Event List Management Redesign

- **Removed "View Event Lists" Button** (`frontend/src/routes/venue-owner/+page.svelte`):
  - Removed the separate "View Event Lists" button from each venue card.
  - Event lists are now displayed inline within each venue card for better accessibility and workflow.

- **Inline Event Lists Display** (`frontend/src/routes/venue-owner/+page.svelte`):
  - Event lists are now shown directly in each venue card below the venue comment.
  - Each event list displays:
    - Visibility icon (lock icon for private, globe icon for public) with tooltip
    - Event list name (truncated if too long)
    - "View/Print" button (green) - navigates to dedicated view/print page
    - "Get Private Link" button (blue) - copies link to clipboard with visual feedback
  - Event lists are sorted by date, then by name.
  - Only displays event lists section if venue has at least one event list.
  - Styled with gray background and proper spacing for clear visual separation.

- **View/Print Page for Single Event List** (`frontend/src/routes/venue-owner/[venue_uuid]/event-lists/[event_list_uuid]/+page.svelte`):
  - Created new route `/venue-owner/[venue_uuid]/event-lists/[event_list_uuid]` for viewing and printing individual event lists.
  - Displays full event list with all events sorted by datetime.
  - Shows venue name, address, event list name, date, and comment.
  - Includes "Back" button to return to My Venues page.
  - Includes "Print" button that triggers browser print dialog.
  - Print styles hide action buttons when printing (using `no-print` class).
  - Handles authorization - redirects if venue owner doesn't own the venue.
  - Created `+page.js` file to mark route as non-prerenderable (requires authentication).

- **Get Private Link Functionality** (`frontend/src/routes/venue-owner/+page.svelte`):
  - "Get Private Link" button copies the private link to clipboard without navigating.
  - Shows "Copied!" feedback with checkmark icon for 2 seconds after copying.
  - Automatically generates `private_link_token` if missing when copying link.
  - Uses `ensureEventListToken()` helper function to guarantee token exists.
  - Link format: `{origin}/?token={event_list_token}`.

- **Helper Functions** (`frontend/src/routes/venue-owner/+page.svelte`):
  - `getVenueEventLists(venue)`: Gets and sorts event lists for a specific venue.
  - `ensureEventListToken(eventList)`: Ensures event list has a private link token, generates if missing.
  - `getPrivateLink(eventList)`: Generates the full private link URL for an event list.
  - `copyPrivateLink(eventList)`: Copies private link to clipboard with visual feedback.
  - `viewPrintEventList(venueUuid, eventListUuid)`: Navigates to the view/print page for an event list.

- **Visibility Icons** (`frontend/src/routes/venue-owner/+page.svelte`):
  - Added lock icon (🔒) for private event lists with "Private" tooltip.
  - Added globe icon (🌐) for public event lists with "Public" tooltip.
  - Icons displayed next to event list name for quick visibility status identification.
  - Icons are properly sized (w-4 h-4) and colored (text-gray-600) for consistency.

- **TypeScript Type Fixes**:
  - Fixed type errors related to `updateModifiedTimestamp()` returning generic `object` type.
  - Added type cast using JSDoc syntax: `/** @type {import('$lib/types').EventList} */` to ensure TypeScript recognizes the result as `EventList`.
  - Fixed `map()` operation type inference by ensuring `updatedList` is properly typed before use in array operations.
  - Used `{@const}` directive at `{#each}` block level to compute event lists once per venue iteration.

- **Build Configuration**:
  - All routes properly configured with `prerender = false` for dynamic authentication-based routes.
  - Build completes successfully with no type errors or warnings.

### Implementation Decisions

- **Inline Display**: Chose to display event lists inline in venue cards instead of separate page for better workflow - users can quickly access event list actions without navigation.
- **Dedicated View/Print Page**: Created separate route for viewing/printing single event lists to provide focused, print-optimized view with back navigation.
- **Copy vs Navigate**: "Get Private Link" copies to clipboard without navigation for quick sharing, while "View/Print" navigates to dedicated page for focused viewing and printing.
- **Token Generation**: Tokens are generated on-demand when needed (when copying link) rather than requiring pre-generation, ensuring all event lists can be shared even if token was missing.
- **Visual Feedback**: Used temporary state (`copiedLinkToken`) to show "Copied!" feedback for 2 seconds, providing clear user confirmation of successful copy operation.
- **Type Safety**: Used JSDoc type annotations with type casts to ensure TypeScript properly infers types throughout the event list management flow.


### UI/UX Refinements and Bug Fixes

- **Event Time Field Fix** (`frontend/src/routes/venue-form/+page.svelte`):
  - Fixed issue where new events' time field was stuck at 11:07 and couldn't be changed.
  - Root cause: When creating new events without a date, `datetime` was set to current timestamp, and time was extracted from that datetime, overwriting user input.
  - Solution: Use today's date as placeholder for `datetime` when event list has no date, and preserve the `time` field separately when loading events.
  - Time field now properly editable for all events, regardless of whether event list has a date.

- **Event List Date Field Labeling** (`frontend/src/routes/venue-form/+page.svelte`):
  - Updated event list date label to show "Date (optional)" to clarify that date is optional for event lists.

- **Event List Comment Styling** (`frontend/src/routes/venue-form/+page.svelte`, `frontend/src/routes/+page.svelte`):
  - Reduced font size to `text-xs` (from default size).
  - Removed italic styling.
  - Reduced spacing between event list name and comment from `mb-4` to `mb-1` on heading.
  - Maintained `mb-4` spacing between comment and events list for proper visual separation.
  - Applied consistent styling to both preview pane and public page.

- **Preview Pane Styling Alignment** (`frontend/src/routes/venue-form/+page.svelte`):
  - Made preview pane match public page styling exactly:
    - Banner image: Added `mb-4` wrapper to match spacing.
    - Venue info: Added `grid grid-cols-1 md:grid-cols-2 gap-4` structure with `flex flex-col justify-start` inner div.
    - Contact links: Added `hover:underline` to email and mobile links.
    - Events: Changed spacing from `space-y-3` to `space-y-1`, padding from `p-3` to `py-2 px-3`.
    - Event name: Added `text-sm` class.
    - Event time: Changed from `text-lg` to `text-base`.
    - Event comment: Changed to `text-sm md:text-xs` with `mt-0.5`.
    - Event duration: Added `mt-0.5` spacing.
  - Preview now provides accurate representation of how venue will appear on public page.

- **Typography Improvements** (`frontend/src/routes/+page.svelte`):
  - Updated "Select a venue to view its event schedules and contact information." text:
    - Desktop: 16px (`text-base`).
    - Mobile: 14px (`text-sm`).
    - Uses responsive Tailwind classes: `text-sm md:text-base`.
  - Updated search description text to 14px using inline style.

- **Navigation Menu Improvements** (`frontend/src/routes/+layout.svelte`):
  - Reduced spacing between menu items:
    - Desktop: Changed from `gap-8` to `gap-2`.
    - Mobile: Changed from `gap-4` to `gap-2`.
  - Added `|` separators between menu items on desktop navigation.
  - Separators styled with `text-gray-400` for subtle appearance.
  - Mobile menu remains vertical without separators (vertical separators not needed).

- **Removed All Italic Styling**:
  - Removed `italic` class from all text elements across the application:
    - Venue comments, event comments, event list comments.
    - Empty state messages.
    - Preview messages.
  - Applied to: `venue-form/+page.svelte`, `+page.svelte`, `venue-owner` pages.
  - Consistent non-italic styling throughout the application.

- **Event Padding Refinement**:
  - User adjusted event padding from `py-2` to `py-1` for tighter vertical spacing between events.
  - Applied to both public page and preview pane for consistency.

### Implementation Decisions

- **Time Field Handling**: When event list has no date, use today's date as placeholder for `datetime` (required field) while preserving separate `time` field for editing. This ensures datetime is always valid while allowing flexible time editing.
- **Preview Accuracy**: Preview pane now exactly matches public page styling to provide accurate preview of how content will appear to visitors.
- **Menu Spacing**: Reduced menu spacing and added separators for more compact, professional appearance while maintaining readability.
- **Typography Consistency**: Removed all italic styling for cleaner, more modern appearance. Used responsive font sizing for better mobile experience.

## 2026-01-19

### Timezone Functionality Fix

- **Root Cause Identified** (`frontend/src/routes/venue-form/+page.svelte`):
  - Fixed critical bug where event datetimes were being created in the venue owner's browser timezone instead of the venue's specified timezone.
  - When creating `new Date('2024-01-15T14:00:00')` without a timezone, JavaScript interprets it as local time (owner's browser timezone), not the venue's timezone.
  - This caused visitors to see times converted to their own timezone even when a venue timezone was set.

- **Solution Implemented**:
  - Updated `combineTimeAndDate()` function to accept optional `venueTimezone` parameter.
  - When venue timezone is set, uses iterative adjustment algorithm to create a UTC timestamp that, when displayed in the venue timezone, shows the desired wall-clock time.
  - When no venue timezone is set, continues to interpret times as the owner's local timezone (existing behavior).
  - Updated all calls to `combineTimeAndDate()` throughout the venue form to pass the venue timezone.

- **Timezone Display Logic** (`frontend/src/routes/+page.svelte`, `frontend/src/lib/utils/datetime.js`):
  - Enhanced `formatEventTimeFromRFC3339()` with explicit validation for non-empty timezone strings (trims whitespace).
  - Added validation in `formatEventTime()` to ensure timezone values are valid strings before use.
  - Both single and multiple event list display paths now correctly pass venue timezone to formatting functions.

- **Behavior After Fix**:
  - **Timezone Set**: All visitors see event times in the venue's timezone (e.g., everyone sees "14:00" in Jerusalem time, regardless of their location).
  - **No Timezone**: Each visitor sees event times converted to their own browser timezone (e.g., 14:00 might show as 09:00 in New York or 14:00 in London).

### Timezone Documentation and User Education

- **About Page Updates** (`frontend/src/routes/about/+page.svelte`):
  - Added comprehensive "Timezone Settings" section explaining how timezone functionality works.
  - Included color-coded examples:
    - Blue box: Explains behavior when timezone is set (fixed display for all visitors).
    - Green box: Explains behavior when no timezone is set (relative to each visitor's location).
  - Added guidance on when to set a timezone vs. leaving it empty.

- **Venue Form Help Popup** (`frontend/src/routes/venue-form/+page.svelte`):
  - Added question mark icon (circle with "?") next to timezone label for inline help.
  - Implemented click-to-toggle popup with same explanation as About page (condensed format).
  - Popup features:
    - Responsive positioning: Fixed and centered in viewport on mobile, absolute positioned relative to button on desktop.
    - Click-outside-to-close functionality.
    - Accessible with proper ARIA attributes (`aria-label`, `aria-expanded`, `aria-haspopup`, `role="tooltip"`).
    - Color-coded content matching About page format for consistency.
  - Mobile-optimized: Uses `fixed` positioning with viewport centering to prevent overflow on small screens.

### Implementation Decisions

- **Timezone Storage**: Timezone is stored as IANA timezone string (e.g., "Asia/Jerusalem"). Actual time offsets are calculated dynamically by JavaScript's `Intl.DateTimeFormat` API based on the timezone, date, and DST rules. This is the correct approach as offsets change with DST and historical rules.

- **Datetime Creation**: When venue timezone is set, event datetimes are created using an iterative adjustment algorithm that ensures the stored UTC timestamp, when displayed in the venue timezone, shows the exact wall-clock time entered by the owner. This ensures consistency across all visitors.

- **User Education**: Added documentation in two places (About page and inline help) to help users understand the timezone functionality, as it was identified as confusing for regular users. The help popup provides quick access without navigating away from the form.

## 2026-01-22

### Summary

- Documented migration strategy from localStorage to backend API integration.
- Decided to **remove all localStorage code** once API is stable (cutover approach, no dual-mode support).

### Notes

- **Migration approach**:
  - Remove localStorage code entirely once backend API is implemented and stable.
  - Rationale: simpler codebase, single source of truth, no divergence risk, forces proper API testing.
  - Strategy: cutover approach (implement API integration, then remove localStorage in one step).
- **What will be removed**:
  - localStorage persistence in `stores.ts` (subscribe handlers that save to localStorage).
  - `demo_data.ts` seeding function (or replace with API-based seeding if needed for dev).
  - Client-side UUID generation (backend generates UUIDs).
  - Any localStorage-specific utilities.
- **What stays**:
  - All UI components and routing (no changes needed).
  - Date/time formatting utilities.
  - Form validation logic.
  - All business logic and UI behavior.
- **API integration plan**:
  - Use relative URLs (`/api/...`) so same code works in dev (proxied) and production (served by Go).
  - Handle JWT access tokens and refresh token cookies (HttpOnly, set by backend).
  - Implement proper error handling and loading states.
- **Tasks added**: Created comprehensive task list for API client implementation, store migration, and localStorage removal.

## 2026-01-28

### Summary

- **API Client Implementation**: Created comprehensive fetch wrapper with JWT token management, automatic token refresh, error handling, and loading state management.
- **Authentication API Integration**: Migrated login and registration from localStorage to backend API endpoints with proper token handling and session restoration.

### Notes

- **API Client** (`frontend/src/lib/api/client.js`):
  - Created fetch wrapper with base URL configuration using relative `/api/...` paths (works in both dev proxy and production).
  - Implemented memory-based JWT access token storage (not localStorage for security).
  - Added automatic token refresh on 401 responses: calls `/api/auth/refresh`, updates token, and retries original request.
  - Implemented request interceptors to automatically add `Authorization: Bearer {token}` header when token is available.
  - Added response interceptors for error handling: parses error responses from API (checks for `error.message` and `error.code` fields).
  - Handles refresh token cookies automatically via `credentials: 'include'` (HttpOnly cookies set by backend).
  - Implemented loading state management utilities (`onLoadingChange()` callback system).
  - Added network error handling with user-friendly error messages.
  - Created custom `ApiError` class with status code and error code for better error handling.
  - Provides convenience methods: `get()`, `post()`, `patch()`, `delete()`, `getJSON()`, `postJSON()`, `patchJSON()`.

- **Authentication API** (`frontend/src/lib/api/auth.js`):
  - Created authentication API functions wrapping backend endpoints:
    - `register()`: POST `/api/auth/register` - Creates new venue owner account.
    - `login()`: POST `/api/auth/login` - Authenticates user and receives access token.
    - `logout()`: POST `/api/auth/logout` - Revokes refresh token and clears session.
    - `getCurrentOwner()`: GET `/api/auth/me` - Gets current authenticated owner, with automatic token refresh if no access token but refresh token cookie exists.
  - All functions automatically store access token in memory and update `currentOwnerStore` with owner data from API responses.
  - Handles authentication errors gracefully with appropriate error messages.

- **Login Page** (`frontend/src/routes/login/+page.svelte`):
  - Replaced localStorage-based authentication with API call to `POST /api/auth/login`.
  - Removed dependency on `ownersStore` (no longer queries localStorage for owner lookup).
  - Added loading state (`isLoading`) with disabled button during API call.
  - Improved error handling: shows specific error messages for invalid credentials (401), network errors, and other API errors.
  - Automatically stores access token and owner data from API response.
  - Redirects to `/venue-owner` on successful login.

- **Registration Page** (`frontend/src/routes/registration/+page.svelte`):
  - Replaced localStorage-based registration with API call to `POST /api/auth/register`.
  - Removed client-side UUID generation (backend generates UUIDs).
  - Removed dependency on `ownersStore` and `getCurrentTimestamp()` (no longer stores owner in localStorage).
  - Updated password validation: minimum 6 characters (matches backend requirement).
  - Added loading state with disabled button during API call.
  - Improved error handling: shows specific error messages for email conflicts (409), validation errors (400), network errors, and other API errors.
  - Automatically stores access token and owner data from API response.
  - Redirects to `/venue-owner` on successful registration.

- **Session Restoration** (`frontend/src/routes/+layout.svelte`):
  - Added `getCurrentOwner()` call on app initialization to restore session from refresh token cookie.
  - If no access token exists but refresh token cookie is present, automatically refreshes token before calling `/api/auth/me`.
  - Handles unauthenticated state gracefully (silently ignores errors for unauthenticated users).
  - Updated logout function to call API `logout()` endpoint before clearing local state.

- **State Management** (`frontend/src/lib/stores.ts`):
  - Removed localStorage persistence for `currentOwnerStore` (now managed entirely by API/auth).
  - `currentOwnerStore` initialized to `null` instead of loading from localStorage.
  - Added comments explaining that `currentOwnerStore` is API-managed while other stores still use localStorage (pending future migration).
  - Owner data now comes exclusively from API responses, not localStorage.

- **Development Configuration** (`frontend/vite.config.js`):
  - Added API proxy configuration to dev server:
    - Proxies `/api/*` requests to `http://localhost:8080` (backend server).
    - Enables seamless API calls during development without CORS issues.
    - Uses `changeOrigin: true` for proper proxy behavior.
  - Frontend uses relative URLs (`/api/...`) so same code works in dev (proxied) and production (served by Go backend).

### Implementation Decisions

- **Memory-based Token Storage**: Access tokens stored in memory (not localStorage) for better security - tokens are cleared on page refresh and must be refreshed via HttpOnly cookie.
- **Automatic Token Refresh**: API client automatically refreshes expired tokens on 401 responses, retrying the original request transparently to the caller.
- **Session Restoration**: On app initialization, attempts to restore session by refreshing token if refresh token cookie exists, then calling `/api/auth/me` to get current owner.
- **Error Handling**: Consistent error parsing from API responses (checks for `error.message` and `error.code` fields matching backend `ErrorResponse` format).
- **Cutover Approach**: Removed localStorage persistence for owner/authentication data immediately (not dual-mode) - simpler codebase, single source of truth.
- **Proxy Configuration**: Dev server proxies API requests to backend, allowing same-origin requests from browser perspective (refresh token cookies work correctly).

### Summary

- **Venues & public endpoints**: Wired the venue owner dashboard and public visitor page to the new backend API, including owner-scoped venues and public/token-based venue access.
- **LocalStorage cutover**: Removed remaining localStorage-backed stores and demo seeding from the runtime, so all data now flows through the backend API.

### Notes

- **API Client** (`frontend/src/lib/api/venues.js`, `frontend/src/lib/api/eventLists.js`, `frontend/src/lib/api/events.js`, `frontend/src/lib/api/public.js`):
  - Added dedicated clients for owner-scoped venues, event lists, and events (e.g., `/api/venues`, `/api/venues/:venue_uuid/event-lists`, `/api/event-lists/:event_list_uuid/events`, and related GET/PATCH/DELETE endpoints).
  - Added public client functions for unauthenticated access: `listPublicVenues`, `getPublicEventListsForVenue`, `getPrivateVenueByToken`, and `getPrivateEventListByToken` bound to `/api/public/...` endpoints.
  - All new clients use the shared `api` wrapper so they inherit JWT handling, automatic refresh, error parsing, and loading-state callbacks.
- **Routes/Pages**:
  - `frontend/src/routes/venue-owner/+page.svelte`: Replaced `venueStore` usage with `listVenues()`; loads the authenticated owner’s venues from `/api/venues`, loads per-venue event lists via `listEventListsForVenue()`, deletes venues via `deleteVenue()`, and surfaces loading + error states in the UI instead of mutating Svelte stores directly.
  - `frontend/src/routes/+page.svelte`: Replaced all localStorage/store-based visitor logic with API calls:
    - Loads public venues via `listPublicVenues()` (search backed by `?query=` parameter).
    - Loads public event lists per venue via `getPublicEventListsForVenue()`.
    - Resolves `?token=` URLs using `getPrivateVenueByToken()` and `getPrivateEventListByToken()` to support private venue/event-list sharing.
    - Keeps time display RFC3339-aware using existing `formatEventTime` helpers and venue timezone, and adds explicit loading/error UX around API calls.
  - `frontend/src/routes/+layout.svelte`: Removed `seedDemoData(false)` from app initialization so the frontend no longer seeds local demo data on load; session restoration is now purely API-based via `getCurrentOwner()`.
- **State Management / Cutover** (`frontend/src/lib/stores.ts`):
  - Removed all entity stores (`ownersStore`, `venueStore`, `eventListStore`, `eventStore`) and their localStorage persistence helpers (`load()`, subscribe-handlers, storage key constants).
  - Simplified `stores.ts` to a single `currentOwnerStore` that is populated exclusively from authentication API responses, aligning with the cutover plan away from localStorage.
- **Venue Form API Integration** (`frontend/src/routes/venue-form/+page.svelte`):
  - Completely refactored venue form to use API endpoints instead of localStorage stores:
    - **Loading**: Replaced store subscriptions with `getVenue()`, `listEventListsForVenue()`, and `listEventsForEventList()` API calls in `onMount`. Added `loadVenueDataFromAPI()` async function that loads venue, event lists, and events in parallel.
    - **Saving**: Replaced `saveVenue()` store mutations with comprehensive API operations:
      - Creates/updates venue via `createVenue()`/`updateVenue()`.
      - For each event list: creates new lists via `createEventList()`, updates existing via `updateEventList()`, deletes removed lists via `deleteEventListApi()`.
      - For each event: creates new events via `createEvent()`, updates existing via `updateEvent()`, deletes removed events via `deleteEventApi()`.
      - Handles `sort_order` updates for both event lists and events when reordering.
    - **UUID Generation**: Removed all `generateUUID()` calls and `crypto.randomUUID()` usage. Uses temporary IDs (`temp-*`) for new entities during editing, then replaces with backend-assigned UUIDs on save.
    - **Error Handling**: Added `isLoading`, `isSaving`, `loadError`, and `saveError` state variables with UI feedback (error messages, loading indicators, disabled save button during operations).
    - **Store Dependencies**: Removed all `venueStore`, `eventListStore`, `eventStore`, `ownersStore` imports and subscriptions. Only retains `currentOwnerStore` for authentication checks.
    - **Undo Functionality**: Updated `handleUndo()` to work with API-loaded data structure.
    - **Form State**: Updated `addEventList()`, `addEvent()`, `duplicateEvent()` to use temporary IDs and mark entities with `isNew` flag for API create/update logic.
    - **Sort Order**: Updated `moveEventListUp()`, `moveEventListDown()`, `moveEventUp()`, `moveEventDown()` to track `sort_order` changes that are persisted via API on save.

### Summary

- **Lazy event loading**: Implemented on-demand loading of events for public event lists to improve initial page load performance.
- **Public events API client**: Added function to retrieve events for public event lists separately from event list data.

### Notes

- **API Client** (`frontend/src/lib/api/public.js`):
  - Added `getPublicEventsForEventList(eventListUuid)`: Calls `GET /api/public/event-lists/:event_list_uuid/events` to retrieve events for a public event list.
  - Enables lazy loading of events separately from event list metadata for better performance.
- **Routes/Pages** (`frontend/src/routes/+page.svelte`):
  - Implemented `ensureEventsForEventList()` function that lazily loads events for event lists when needed.
  - Events are cached in `eventListEventsMap` to avoid redundant API calls.
  - Events are loaded on-demand when a user selects an event list, rather than loading all events upfront.
  - Improves initial page load time by deferring event data fetching until it's actually needed.

## 2026-01-29

### Summary

- **Data consistency with backend API**: Aligned frontend TypeScript types and venue-form usage with backend response shapes. Made `event_list_uuids` and `event_uuids` optional (API does not return them; they are derived client-side). Added `sort_order` to EventList and Event; clarified `duration_minutes` as `number | null`. Guarded all reads of `event_uuids` in venue-form with `(list.event_uuids || [])`.

### Notes

- **Types** (`frontend/src/lib/types.ts`):
  - **Venue**: `event_list_uuids` is now optional (`event_list_uuids?: string[]`) with a comment that it is client-only and derived from listing event-lists.
  - **EventList**: `event_uuids` is now optional (`event_uuids?: string[]`) with a comment that it is client-only and derived from listing events; added required `sort_order: number` to match backend `EventListResponse`.
  - **Event**: Added required `sort_order: number`; changed `duration_minutes` to `duration_minutes?: number | null` to match backend nullable response.
- **Routes/Pages** (`frontend/src/routes/venue-form/+page.svelte`):
  - All reads of `list.event_uuids` now use `(list.event_uuids || [])` so optional field is safe when API returns event lists without the array (e.g. after load from API).
  - Updated when appending a new event, when removing an event, and when copying for reorder/update (multiple call sites).

### Summary

- **Demo page**: Renamed route "Prototype" to "Demo" and updated copy to state that the site is fully functional and currently in test mode.

### Notes

- **Routes/Pages**:
  - Added `frontend/src/routes/demo/+page.svelte`: New Demo page with title "Demo - times.place", heading "Demo", and an info block stating the site is fully functional (registration, login, venue management, event lists, public browsing) and currently running in test mode; kept Test User Accounts and Report Bugs sections.
  - Removed `frontend/src/routes/prototype/+page.svelte` and the empty `prototype/` route directory.
- **Layout** (`frontend/src/routes/+layout.svelte`):
  - Desktop and mobile nav links updated from `/prototype` and label "Prototype" to `/demo` and label "Demo".

### Summary

- **Venue visibility removed**: Removed venue-level visibility from types, API client, and venue-form so visibility is controlled only at event-list level (the only level exposed in the GUI). Public listing is determined by "at least one public event list" per venue.

### Notes

- **Types** (`frontend/src/lib/types.ts`):
  - Removed `visibility: 'public' | 'private'` from `Venue` interface; venues no longer have a visibility field from the API.
- **API Client** (`frontend/src/lib/api/venues.js`):
  - Removed `visibility` from `createVenue()` payload and JSDoc; removed `visibility` from `updateVenue()` payload and JSDoc.
- **Routes/Pages** (`frontend/src/routes/venue-form/+page.svelte`):
  - Removed `visibility: 'public'` from the `createVenue()` call when saving a new venue (no longer sent to backend).
- **Routes/Pages** (`frontend/src/routes/about/+page.svelte`):
  - Updated Visibility & Security copy to state that only event lists have visibility; a venue appears in the public list if it has at least one public event list.

## 2026-01-30

### Summary

- **Disclaimer and UX polish**: Added Disclaimer page and footer link; removed Reload button from main page; fixed search to use backend results as-is so event list and event name matches show correctly; banner image constrained with object-cover/object-center.
- **Edit Mode (venue-form)**: Title and Cancel/Save left-aligned with form (matching padding); desktop layout uses single flex column so header stays at top and form scrolls beneath; grid fixed height on desktop so page doesn't scroll; bottom duplicate Cancel/Save removed; "Save Venue" renamed to "Save"; event list/event delete buttons labeled "Delete Event List" and "Delete Event"; duplicate "Edit Venue" h2 removed; layout uses reduced top padding on /venue-form and /venue-owner so titles sit higher.
- **My Venues (venue-owner)**: Desktop/mobile spacing tightened (title further up, less space to subtitle and "Signed in as"); Add Venue right on desktop, centered and smaller on mobile; event list items tighter spacing; Edit/Delete side by side.
- **Layout**: Footer link to Disclaimer; main content uses `md:pt-0 md:pb-12` when path is `/venue-form` or `/venue-owner` for tighter top spacing on those pages.

### Notes

- **Routes/Pages** (`frontend/src/routes/disclaimer/+page.svelte`):
  - New Disclaimer page with sections: Accuracy of Information (venue owners responsible for data), No Warranty, Limitation of Liability, Third-Party Content and Links, Contact.
- **Routes/Pages** (`frontend/src/routes/+layout.svelte`):
  - Footer: added "Disclaimer" link next to copyright.
  - Main content container: `isVenueForm` and `isVenueOwner` derived from `$page.url.pathname`; when true, use `md:pt-0 md:pb-12` instead of `md:py-12` so Edit Venue and My Venues content sits higher on desktop.
- **Routes/Pages** (`frontend/src/routes/+page.svelte`):
  - Removed Reload (refresh) button next to venue search so the main page no longer exposes an unclear control.
  - Search: `filteredVenues` now equals `sortedVenues`; backend already filters by venue, event list, and event names, so client-side `venueMatchesSearch` filter was removed to avoid dropping venues that matched only on event list/event name.
  - Banner: fixed-height container with `object-cover object-center` so images fit the banner area.
- **Routes/Pages** (`frontend/src/routes/venue-form/+page.svelte`):
  - Header block (title + Cancel/Save) given same horizontal padding as form (`px-4 md:px-6`) for left alignment; loading-state header uses same padding.
  - Desktop: left column is single flex container (`lg:col-start-1 lg:row-span-2`) with header (flex-shrink-0) and editing pane (flex-1 min-h-0 overflow-y-auto); grid has `lg:h-[calc(100vh-6rem)]` so the page doesn't scroll and only the form scrolls; preview pane has `lg:min-h-0` and scrolls within the grid.
  - Bottom duplicate Cancel/Save/Undo block removed (header is fixed at top on desktop, single set on mobile).
  - "Save Venue" button label changed to "Save" (both action bars).
  - Event list delete button label: "Delete Event List"; event delete button label: "Delete Event".
  - Duplicate "Edit Venue" h2 removed from editing pane; page title remains in header.
  - Desktop: outer page container uses `lg:pt-0`, `lg:-mt-6` for title/buttons further up.
- **Routes/Pages** (`frontend/src/routes/venue-owner/+page.svelte`):
  - Desktop: page container `md:pt-0 md:pb-8`, `md:-mt-4`; header block and h1 margins reduced (e.g. `md:mb-0`, `md:mb-0.5`) for tighter spacing between title, subtitle, and "Signed in as".
  - Add Venue: `justify-end` on desktop; on mobile `justify-center`, smaller button (`py-1.5 px-3 text-sm`, icon `w-4 h-4`).
  - Event list items: `space-y-1` on mobile, smaller item padding; Edit and Delete buttons always in a row (`flex-row`).
- **Makefile**:
  - Comment above `fbuild` explaining rollup optional-deps issue and suggesting `make finstall-clean && make fbuild`.

### Summary (banner images)

- **Banner images fixed**: Same image now fits consistently in every banner; images scale to fit without cropping; letterboxing is white; edit form shows preferred ratio 16:9.

### Notes (banner images)

- **Components** (`frontend/src/lib/BannerImage.svelte`): New shared component with size variants (sm/md/lg), `object-contain`, white background for areas outside the image.
- **Routes/Pages**: All banner usages (venue-owner grid, home selected venue, venue-form upload preview and right panel, event-list print view) use `<BannerImage>`.
- **Styling** (`frontend/src/app.css`): Print rule for `.banner-img` so banners scale to fit when printing.
- **Routes/Pages** (`frontend/src/routes/venue-form/+page.svelte`): Label "Banner Image (optional)" now includes "— preferred ratio 16:9".

### Summary (My page and user menu)

- **My page**: New account page at `/my` for future account and billing; shows profile (name, email, mobile) and account links (My Venues, placeholder for billing) when logged in; when anonymous, prompts to log in or register.
- **User menu in nav**: Replaced inline Login/Register and My Venues/Logout with a single user icon and dropdown: anonymous users see user silhouette icon and dropdown (Login, Register, My account); logged-in users see initial-in-circle and dropdown (Account, My Venues, Logout). Dropdown closes on outside click; same account links available in mobile hamburger menu.

### Notes (My page and user menu)

- **Routes/Pages** (`frontend/src/routes/my/+page.svelte`):
  - New route `/my`; uses `getAuthMe()` for venue_count/venue_limit when logged in; shows profile section and account section with My Venues link and billing placeholder.
- **Routes/Pages** (`frontend/src/routes/+layout.svelte`):
  - Desktop nav: user icon (anon = person icon, logged in = initial in circle) with dropdown; dropdown items: Account (/my), My Venues, Logout when logged in; Login, Register, My account when anon. Click-outside closes dropdown.
  - Mobile nav: added "My account" link; kept Login/Register or My Venues/Logout as before.
  - `userMenuOpen`, `userMenuEl`, `toggleUserMenu`, `closeUserMenu`, `handleClickOutside` for dropdown behavior.

## 2026-02-01

### Summary

- **Demo page renamed to Test Phase**: Renamed the Demo page and nav label to "Test Phase" to avoid confusion; page title is "Test Phase" (no brackets), centered with 22px font; copy updated from "demo" to "test" wording (test accounts, test data).
- **Nav [Test Phase] placement**: On desktop, [Test Phase] is centered in the header using absolute positioning (`left-1/2 -translate-x-1/2`); nav link uses 14px font and brackets in the label. On mobile (hamburger menu), [Test Phase] is left-aligned like other items (no centering).

### Notes

- **Routes/Pages** (`frontend/src/routes/demo/+page.svelte`):
  - `<title>` set to "Test Phase - times.place"; main heading "Test Phase" (no brackets), `text-[22px]`, `text-center`, `w-full`.
  - Copy: "Use these test accounts to try the application"; "when test data is seeded" (replacing "demo" wording).
- **Routes/Pages** (`frontend/src/routes/+layout.svelte`):
  - Desktop: [Test Phase] in a centered block (`hidden md:flex absolute left-1/2 -translate-x-1/2 h-full items-center pointer-events-none`); link has `pointer-events-auto`, `text-[14px]`, label "[Test Phase]".
  - Right nav block no longer includes [Test Phase] or leading pipe; starts with Home | About | Price | user menu.
  - Mobile: [Test Phase] link left-aligned (no flex/justify-center), same styling as other menu items.

### Summary (account deletion)

- **Account deletion**: Added permanent account deletion on the My page: calls `DELETE /api/auth/me`, clears session and redirects; includes confirmation and error handling.

### Notes (account deletion)

- **API Client** (`frontend/src/lib/api/auth.js`):
  - New `deleteAccount()`: sends `DELETE /api/auth/me`, then clears access token and `currentOwnerStore` so the user is logged out after deletion.
- **Routes/Pages** (`frontend/src/routes/my/+page.svelte`):
  - New "Delete account" section: warning copy ("Permanently delete your account and all your venues. This cannot be undone."), inline error display (`deleteError`), confirm dialog ("Permanently delete your account and all your venues? This cannot be undone."), loading state (`deleting`), calls `deleteAccount()` then redirects (e.g. home); surfaces API errors for failed deletion.
