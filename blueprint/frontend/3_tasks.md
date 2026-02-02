# Task List

## Installation and setup

- [x] Set up folder structure: frontend/, backend/ (move current go and backend related files and folders here)
- [x] Install SvelteKit with javascript, not typescript
- [x] Set up SvelteKit project structure in frontend/
- [x] Set up Tailwind CSS
- [x] Makefile: Add commands for Svelte and Tailwaind
- [x] Start the app with a demo page to see that everything is set up correctly

## Setup routes

- [x] Registration page
- [x] Login page
- [x] About page
- [x] Visitor page (merged into landing page)
- [x] Venue owner page
- [x] Venue form page

## Setup overall layout with some dummy text

- [x] Create header with navigation (About, Login). Use /frontend/static/house_clock.png as logo icon and Times & Place as logo title in header.
- [x] Create footer. Add standard copyright text with symbol. Add mailto contact email, `timeplaceadmin@atomicmail.io`.
- [x] Create About page with prototype information and contact info `timeplaceadmin@atomicmail.io`.
- [x] Add the other pages
- [x] Check navigation

## Data model

- [x] Create venue owner data model (see spec for details)
- [x] Create venue data model (see spec for details)
- [x] Create event list data model (see spec for details)
- [x] Create event data model (see spec for details)
- [x] Implement UUID generation for all entities (use crypto.randomUUID() or similar)
- [x] Implement date/time formatting: store event list dates as ISO 8601, store event DateTime as Unix epoch (full date+time), display only time portion from events
- [x] Add demo data with venues that have multiple event lists and venues with no event lists
- [x] Set up multiple venue owner demo accounts (at least two for testing isolation)

## Visitor Page

- [x] Create dropdown for list of venues using demo data
- [x] Display selected venue details: banner, contacts, comment
- [x] Implement event list selector (dropdown/tabs) when venue has multiple event lists
- [x] Display events from selected event list (or only event list if there's just one)
- [x] Display only the time portion from each event (date comes from the event list)
- [x] Format event times for display (extract time from stored Unix epoch datetime, format for user's timezone)
- [x] Handle case when venue has no event lists (show appropriate message)
- [x] Add print button to event list display
- [x] Implement print-friendly CSS styles using @media print
- [x] Ensure print output includes venue details, event list name/date, and all events with times
- [x] Test print functionality across different browsers

## Venue Owner Page Registration, Authentication, and Authorization

- [x] Build venue owner registration/account creation flow (for demo: simple form with hardcoded validation)
- [x] Build venue owner login flow with hardcoded password for prototype

## Venue Owner Page

- [x] Create venue owner dashboard/list view showing all venues owned by logged-in owner
- [x] Implement client-side filtering to show only venues owned by current venue owner
- [x] Add functionality to add a venue
- [x] Add functionality to delete a venue
- [x] Add functionality to edit a venue (see Venue Form below)
- [x] Venues are sorted alphabetically
- [x] Test on mobile + desktop

## Venue Form

- [x] Form UI with two panes: editing pane (dynamic fields) and live preview pane
- [x] Event list management: functionality to add, delete, and reorder event lists
- [x] For each event list: manage name, date (ISO 8601 format), comment, and events
- [x] Event management within event lists: functionality to add, delete, duplicate, and move events up and down
- [x] Event DateTime input: allow venue owners to input time, combine with event list date to create full datetime, convert to Unix epoch timestamp for storage
- [x] Display event times (time portion only) in both edit and preview panes
- [x] Event list selector in preview pane to test how different event lists appear to visitors
- [x] Input validation: date (ISO 8601 format), time (validate and combine with event list date to create full datetime, convert to Unix epoch), XSS/SQL injection
- [x] Validate UUID format for all entity identifiers
- [x] Undo functionality
- [x] Prototype: Save to local storage for all edits
- [x] Add image upload for banners (prototype: local storage as base64/blob)
- [x] Add timezone field with dropdown selector (organized by region)
- [x] Display timezone on preview pane and public page
- [x] Implement venue timezone-aware time display (times shown in venue's timezone)
- [x] Add geolocation picker with interactive map (address geocoding, map click, draggable marker)

## Security & Safety Features

- [x] Add `visibility` field to Venue data model (values: "public" or "private")
- [x] Default all new venues to "private" visibility
- [x] Add visibility toggle/radio buttons in venue form (Public/Private option)
- [x] Implement private link generation: generate cryptographically secure token (using crypto.randomUUID() or similar) for private venues
- [x] Store `private_link_token` field on venue entity for private venues
- [x] Update visitor page to filter venues: only show public venues in dropdown/search
- [x] Implement private link access: allow visitors to access private venues via URL with token parameter (e.g., `/?token=...`)
- [x] Hide venue owner contact information (email, mobile) from public visitor page
- [x] Only display contact information to authenticated venue owners on their own venue edit pages
- [x] Update demo data to include mix of public and private venues for testing
- [x] Test that private venues are not searchable/visible in public dropdown
- [x] Test that private venue links work correctly and are unguessable

## Documentation Tasks

- [x] Document demo data and editing process
- [x] Record agentic coding decisions and workflow notes in implement.md

## Backend API Integration (migration from localStorage)

### API Client Implementation

- [x] Create `src/lib/api/client.js` with fetch wrapper:
  - [x] Implement base URL configuration (relative `/api/...` paths for dev proxy and production)
  - [x] Add request interceptors for adding JWT access token to Authorization header
  - [x] Add response interceptors for error handling (parse error responses, handle 401/403/404/500)
  - [x] Implement JWT access token storage (memory-based, not localStorage)
  - [x] Handle refresh token cookies (automatic via browser, HttpOnly)
  - [x] Implement automatic token refresh on 401 responses (call `/api/auth/refresh`, retry original request)
  - [x] Add loading state management utilities
  - [x] Handle network errors gracefully

### Authentication API Integration

- [x] Replace localStorage login with `POST /api/auth/login`:
  - [x] Update login page to call API endpoint
  - [x] Store access token in memory from response
  - [x] Store owner data from response in store
  - [x] Handle login errors (invalid credentials, network errors)
- [x] Replace localStorage registration with `POST /api/auth/register`:
  - [x] Update registration page to call API endpoint
  - [x] Store access token in memory from response
  - [x] Store owner data from response in store
  - [x] Handle registration errors (email already exists, validation errors)
- [x] Implement `POST /api/auth/refresh` for token refresh:
  - [x] Call endpoint when access token expires (on 401 response)
  - [x] Update stored access token with new token
  - [x] Retry original failed request with new token
- [x] Implement `POST /api/auth/logout` to revoke refresh token:
  - [x] Call endpoint on logout
  - [x] Clear access token from memory
  - [x] Clear owner data from store
  - [x] Redirect to home page
- [x] Implement `GET /api/auth/me` to get current owner:
  - [x] Call on app initialization to restore session
  - [x] Update owner store with response data
  - [x] Handle unauthenticated state (no valid token)
- [x] Update auth state management:
  - [x] Remove localStorage-based owner storage
  - [x] Use API responses for owner data
  - [x] Handle authentication state changes (login/logout)

### Venues API Integration

- [x] Replace `venueStore` localStorage with API calls:
  - [x] `GET /api/venues` - List all venues owned by authenticated owner
  - [x] `POST /api/venues` - Create new venue
  - [x] `GET /api/venues/:venue_uuid` - Get single venue by UUID
  - [x] `PATCH /api/venues/:venue_uuid` - Update venue
  - [x] `DELETE /api/venues/:venue_uuid` - Delete venue
- [x] Update venue form (`src/routes/venue-form/+page.svelte`) to use API endpoints:
  - [x] Load venue data from API on mount
  - [x] Save venue changes via PATCH endpoint
  - [x] Create new venues via POST endpoint
  - [x] Handle API errors and loading states
- [x] Update venue owner dashboard (`src/routes/venue-owner/+page.svelte`) to fetch from API:
  - [x] Load venues list from API on mount
  - [x] Handle delete via DELETE endpoint
  - [x] Handle API errors and loading states
- [x] Remove client-side filtering (backend handles owner scoping)

### Event Lists API Integration

- [x] Replace `eventListStore` localStorage with API calls:
  - [x] `GET /api/venues/:venue_uuid/event-lists` - List event lists for a venue
  - [x] `POST /api/venues/:venue_uuid/event-lists` - Create new event list
  - [x] `GET /api/event-lists/:event_list_uuid` - Get single event list by UUID
  - [x] `PATCH /api/event-lists/:event_list_uuid` - Update event list
  - [x] `DELETE /api/event-lists/:event_list_uuid` - Delete event list
- [x] Update venue form event list management to use API:
  - [x] Load event lists from API when venue is loaded
  - [x] Create event lists via POST endpoint
  - [x] Update event lists via PATCH endpoint
  - [x] Delete event lists via DELETE endpoint
  - [x] Handle sort_order updates via API
  - [x] Handle API errors and loading states

### Events API Integration

- [x] Replace `eventStore` localStorage with API calls:
  - [x] `GET /api/event-lists/:event_list_uuid/events` - List events for an event list
  - [x] `POST /api/event-lists/:event_list_uuid/events` - Create new event
  - [x] `GET /api/events/:event_uuid` - Get single event by UUID
  - [x] `PATCH /api/events/:event_uuid` - Update event
  - [x] `DELETE /api/events/:event_uuid` - Delete event
- [x] Update venue form event management to use API:
  - [x] Load events from API when event list is selected
  - [x] Create events via POST endpoint
  - [x] Update events via PATCH endpoint
  - [x] Delete events via DELETE endpoint
  - [x] Handle sort_order updates via API (move up/down)
  - [x] Handle API errors and loading states

### Public Endpoints Integration

- [x] Replace visitor page localStorage queries with public API:
  - [x] `GET /api/public/venues?query=...` - Search public venues (replace dropdown filtering)
  - [x] `GET /api/public/venues/:venue_uuid/event-lists` - Get public event lists for a venue
  - [x] `GET /api/public/venues/by-token/:token` - Access private venue via token (replace client-side token lookup)
  - [x] `GET /api/public/event-lists/by-token/:token` - Access private event list via token (replace client-side token lookup)
- [x] Update visitor page (`src/routes/+page.svelte`) to fetch from public API:
  - [x] Load public venues list from API on mount
  - [x] Implement search functionality using query parameter
  - [x] Load event lists from API when venue is selected
  - [x] Handle private venue/event list access via token URL parameter
  - [x] Remove client-side filtering (backend handles visibility)
  - [x] Handle API errors and loading states

### Remove localStorage Code (Cutover)

- [x] Remove localStorage persistence from `stores.ts`:
  - [x] Remove `subscribe` handlers that save to localStorage
  - [x] Remove `load()` helper function
  - [x] Remove storage key constants
  - [x] Update stores to initialize from API instead of localStorage
- [x] Remove or replace `demo_data.ts`:
  - [x] Remove localStorage seeding function
  - [x] Removed obsolete demo_data.ts file (backend seed.go handles demo data)
- [x] Remove client-side UUID generation:
  - [x] Remove `crypto.randomUUID()` calls
  - [x] Use UUIDs returned from API responses
- [x] Remove localStorage-specific utilities:
  - [x] Search for and remove any remaining localStorage references
  - [x] Clean up any localStorage-related helper functions
- [x] Update all components to work with API-based stores:
  - [x] Test venue owner dashboard
  - [x] Test venue form (create, edit, delete)
  - [x] Test visitor page (public venues, private token access)
  - [x] Test authentication flow (login, register, logout)
  - [x] Verify data consistency between frontend and backend

### Data Format Updates

- [x] Update data models to match backend API response formats:
  - [x] Review API response types from backend handlers
  - [x] Update TypeScript interfaces in `src/lib/types.ts` if needed
  - [x] Ensure field names match (snake_case from API vs camelCase in frontend)
- [x] Handle RFC3339 datetime strings from API:
  - [x] Update datetime utilities to parse RFC3339 format (backend uses RFC3339, not Unix epoch)
  - [x] Update event datetime handling (API returns RFC3339 strings, not Unix timestamps)
  - [x] Ensure timezone handling works with RFC3339 format
- [x] Update UUID handling:
  - [x] Ensure UUIDs are strings from API (not client-generated)
  - [x] Remove any UUID generation logic
- [x] Update date/time display:
  - [x] Ensure event list dates work with API format (ISO 8601 date strings)
  - [x] Ensure event times display correctly from RFC3339 datetime strings

### Error Handling & UX

- [x] Implement consistent error handling for API failures:
  - [x] Parse error responses from API (check for error message fields)
  - [x] Display user-friendly error messages
  - [x] Handle different error types (network, validation, auth, server errors)
- [x] Add loading indicators for API operations:
  - [x] Show loading state during API calls
  - [x] Disable forms/buttons during API operations
  - [x] Provide visual feedback for async operations
- [x] Handle authentication errors:
  - [x] Redirect to login on 401 responses (when refresh fails)
  - [x] Show appropriate error messages for auth failures
  - [x] Clear invalid auth state
- [x] Handle network errors gracefully:
  - [x] Show retry options for network failures (error messages with "try again" guidance)
  - [x] Provide offline state detection (navigator.onLine API with visual indicator)
  - [x] Handle timeout errors (detected and reported with clear messages)

### Testing & Validation

- [ ] Test all CRUD operations via API:
  - [ ] Create, read, update, delete venues
  - [ ] Create, read, update, delete event lists
  - [ ] Create, read, update, delete events
  - [ ] Verify data persists correctly in database
- [ ] Test authentication flow:
  - [ ] Register new owner
  - [ ] Login with credentials
  - [ ] Token refresh on expiration
  - [ ] Logout and token revocation
  - [ ] Session restoration on page reload
- [ ] Test public endpoints and private token access:
  - [ ] Browse public venues
  - [ ] Search public venues
  - [ ] Access private venue via token
  - [ ] Access private event list via token
  - [ ] Verify private venues don't appear in public search
- [ ] Test error scenarios:
  - [ ] Network failures
  - [ ] Authentication failures (invalid token, expired token)
  - [ ] Validation errors (invalid data)
  - [ ] Server errors (500, 503)
- [ ] Verify data consistency:
  - [ ] Ensure frontend displays match backend data
  - [ ] Test concurrent edits (if applicable)
  - [ ] Verify owner isolation (owners only see their venues)

## Misc

- [x] Rename page Prototype to Demo with information about the site being fully functional but currently in test modus
- [x] Add disclaimer
- [x] Remove Reload button
- [x] Edit mode: Move Cancel and Save buttons to the left (aligned with the actual edit column), and also add them to the bottom, so the user doesn't have to scroll to the top if he is editing something closer to the bottom of the page.
- [x] Search: It doesn't filter on event list names or event names
- [x] UI:Mobile:Edit mode: The UI has to take up less space. There is to much white space.
  - [x] "My Venues": smaller font. Move it up. Remove "Manage your venues and event schedules.".
  - [x] "Signed in as ...": Remove border and background. Move it closer to "My Venues" and center it.
  - [x] Make banner (the image) shorter (reduce height)
- [x] Fix image to fit in banner space
- [x] Change name of Demo page to something like "Test Phase", center it, and make font 22px, and put in brackets: [Test Phase]
- [ ] New Feature: Main Page: Add the possibility to search near me. The distance can be hardcoded. Set it to a radius of 500 meter.
- [x] Add updated time on each event list

## Deployment

See 3_tasks.md under backend

## Security

- [x] Delete account button in account page
- [x] Email verification GUI

## Backoffice

- [x] Client-side Admin Protection:
  - [x] Implement `isAdmin` flag in the owner store <!-- id: 200 -->
  - [x] Protect `/backoffice` routes with auth check <!-- id: 201 -->
- [x] UI Implementation:
  - [x] Create Backoffice Dashboard with platform stats <!-- id: 202 -->
  - [x] Implement Owner Management table: list, search, manual verification, deletion <!-- id: 203 -->
  - [x] Implement Global Venue Management list <!-- id: 204 -->
  - [x] Add "Backoffice" link to main navigation for admin users <!-- id: 205 -->

## Instrumentation and Observability

- [ ] Basic statistics (to save database cost)
