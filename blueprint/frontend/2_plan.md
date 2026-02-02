# Technical Plan

## Tech Stack

- Frontend: Svelte and SvelteKit (for routing, reactivity and maintainability)
- CSS: Use standard UI from Tailwind CSS. We do not have the energy to tweak CSS.
- Data: Currently using localStorage (prototype). Migrating to backend API (see "Backend API Integration" section).
- Backend: Golang with PostgreSQL (complete and ready for integration - see `blueprint/backend/2_plan.md`)

## Data Model Structure

**Reference**: See the complete data model specification in `1_spec.md` → Structure → Data Model section.

### Implementation Details

- **Storage Format**: Data will be stored as nested JavaScript objects/JSON in localStorage
- **UUID Generation**: Use `crypto.randomUUID()` (available in modern browsers) to generate all entity UUIDs
- **Date/Time Implementation**: See utility functions in `frontend/src/utils/datetime.js`
  - Event List dates: Store and validate as ISO 8601 date strings (YYYY-MM-DD)
  - Event DateTime: Store as Unix epoch timestamp (seconds since January 1, 1970 UTC) representing the full date and time
  - Event Display: Extract only the time portion from the stored datetime for display to users
  - Timezone and locale handling: Automatically uses browser settings for timezone and locale

## UI Structure

### Visitor page

- Header with an image and navigation: Times, About, Venue Owner
- Times:
  - Landing page: Dropdown to select venue, details panel, venue owner button.
  - Detail panel: Banner image, venue name, contact info, event list selector (if venue has multiple event lists), events display, print button.
  - Event list selector: Dropdown or tabs to switch between event lists when a venue has multiple event lists
  - Events display: Shows events from the selected event list (or the only event list if there's just one)
  - Print functionality: Print button that triggers browser print dialog with print-optimized CSS styles
- About: info about the prototype
- Venue Owner:
  - Button shows login (simple password for demo), then edit page/modal.
  - Edit page: Divided in two panes, one for editing, and the other that shows the rendered result as it will be on the Times/Detail page
- Footer: Empty for now

### Venue owner page - form

- Dynamic form fields for managing event lists (add/remove event lists)
- For each event list: dynamic fields for managing events (input/add/remove for each event)
- Event list selector in preview pane to test how different event lists appear
- Undo functionality

## Venue Owner System & Authentication

- For demo: Use simple password-based authentication (hardcoded credentials)
- Support multiple venue owners: Each venue owner has a unique identifier and can only see/edit venues they created
- Demo setup: Two venue owner accounts will be created to verify multi-owner isolation
- Venue owner data isolation: Venues will have an `ownerUUID` field (UUID format); filtering happens client-side
- Future: Replace with proper authentication and authorization system

## Security & Safety Implementation

### Public/Private Venue Visibility

- **Data Model**: Add a `visibility` field to the Venue entity with values: `"public"` (searchable on public page) or `"private"` (only accessible via shareable link)
- **Default Behavior**: Default all new venues to `"private"` to ensure venues are only publicized when owners explicitly choose to do so
- **UI Implementation**:
  - Add visibility toggle/radio buttons in venue form (Public/Private)
  - For private venues, generate and display a shareable link (cryptographically secure token, not sequential ID)
  - On visitor page, filter venues by visibility: public venues appear in dropdown/search, private venues only accessible via direct link
- **Link Security**: Use `crypto.randomUUID()` or similar to generate unguessable tokens for private venue links (store as `private_link_token` field on venue)

### Privacy Protection

- **Contact Information**:
  - Venue owner email and mobile phone numbers should NOT be displayed on public visitor page
  - Only show contact info to authenticated venue owners (on their own venue edit pages)
  - Consider adding a contact form option for public venues that forwards messages without exposing owner details
- **Location Privacy**:
  - For public venues, consider showing approximate location (neighborhood/area) instead of exact address
  - Add venue owner option to hide precise geolocation coordinates
  - Store both `address` (full) and `address_public` (approximate) fields; use `address_public` for public display
- **Time Privacy**: Consider venue owner option to show approximate times (e.g., "evening services") instead of exact times for public venues

### Access Control

- **Venue Owner Authorization**: Ensure client-side filtering strictly enforces that venue owners can only see/edit venues where `ownerUUID` matches their own UUID
- **Private Link Access**: When accessing a venue via private link, verify the `private_link_token` matches before displaying venue details

### Demo Limitations

- For prototype: Focus on implementing public/private visibility toggle and basic link sharing
- Contact information hiding can be implemented by simply not displaying owner email/mobile on visitor page
- Advanced features (link expiration, revocation, approximate locations) can be deferred to future iterations

## Data Storage

- **Current (prototype)**: Default/demo data pre-loaded on first use, serialized to localStorage on change
- **Migration in progress**: Backend API is complete and ready. We are migrating from localStorage to API-only data storage (see "Backend API Integration" section below).

## Image Handling

- Use simple file upload for banners, store as base64 or blob URL in localStorage

## Print Functionality

- **Print Button**: Add a print button on the visitor page detail panel when an event list is displayed
- **Print Styles**: Implement CSS print media queries (`@media print`) to optimize the printed output:
  - Hide navigation, buttons, and non-essential UI elements
  - Show venue banner image, name, address, and contact info
  - Display event list name and date prominently
  - Format events in a clean, readable list with times
  - Ensure proper page breaks and avoid splitting event information across pages
  - Use appropriate font sizes and spacing for printed output
- **Print Trigger**: Use browser's native `window.print()` API when print button is clicked
- **Print Content**: Include venue details (name, address, timezone), event list information (name, date), and all events with their times formatted clearly

## Risks & Uncertainties

- Security is minimal for demo
- Local storage limitations for images

## Backend API Integration

### Backend Status

**The backend is now complete and ready for integration.** All API endpoints are implemented and tested. See `blueprint/backend/2_plan.md` for complete backend architecture and API details.

### API Structure

The backend provides the following endpoint groups:

- **Authentication** (`/api/auth/*`):
  - `POST /api/auth/register` - Register new venue owner
  - `POST /api/auth/login` - Login and receive JWT access token
  - `POST /api/auth/refresh` - Refresh access token using refresh token cookie
  - `POST /api/auth/logout` - Logout and revoke refresh token
  - `GET /api/auth/me` - Get current authenticated owner (protected)

- **Owner-scoped CRUD** (all require JWT authentication):
  - **Venues**: `GET /api/venues`, `POST /api/venues`, `GET /api/venues/:venue_uuid`, `PATCH /api/venues/:venue_uuid`, `DELETE /api/venues/:venue_uuid`
  - **Event Lists**: `GET /api/venues/:venue_uuid/event-lists`, `POST /api/venues/:venue_uuid/event-lists`, `GET /api/event-lists/:event_list_uuid`, `PATCH /api/event-lists/:event_list_uuid`, `DELETE /api/event-lists/:event_list_uuid`
  - **Events**: `GET /api/event-lists/:event_list_uuid/events`, `POST /api/event-lists/:event_list_uuid/events`, `GET /api/events/:event_uuid`, `PATCH /api/events/:event_uuid`, `DELETE /api/events/:event_uuid`

- **Public endpoints** (no authentication required):
  - `GET /api/public/venues?query=...` - Search public venues
  - `GET /api/public/venues/:venue_uuid/event-lists` - Get public event lists for a venue
  - `GET /api/public/venues/by-token/:token` - Access private venue via token
  - `GET /api/public/event-lists/by-token/:token` - Access private event list via token

### Authentication

The backend uses JWT access tokens + refresh tokens:

- **Access tokens**: Short-lived JWT tokens (15 minutes) returned in JSON response body
- **Refresh tokens**: Long-lived tokens (30 days) stored as HttpOnly cookies by the backend
- **Token refresh**: Frontend automatically refreshes access token when it expires (on 401 responses)
- **Storage**: Access tokens stored in memory (not localStorage) for security; refresh tokens handled automatically by browser

### Development Setup

Local development uses a two-process setup:

- **Frontend**: SvelteKit dev server on `http://localhost:5173` (with HMR)
- **Backend**: Go/Echo API server on `http://localhost:8080`

The SvelteKit dev server proxies `/api/*` requests to the backend, so:

- Browser makes requests to `http://localhost:5173/api/...`
- Vite proxies to `http://localhost:8080/api/...`
- No CORS issues (same-origin from browser perspective)
- Refresh token cookies work correctly (HttpOnly, same-origin)

### Production Deployment

In production (Render.com), a single Go binary serves both:

- **Frontend**: Static SvelteKit build served from `/` (with SPA fallback)
- **API**: All `/api/*` routes handled by Echo

The frontend uses relative URLs (`/api/...`) so the same code works in both dev (proxied) and production (served by Go).

### Migration Strategy: Remove localStorage Code (Cutover Approach)

We will **remove all localStorage-based data storage code** and switch to API-only data access.

**Rationale**:

- Simpler codebase: single source of truth (the API)
- No code duplication or divergence risk between localStorage and API implementations
- Forces proper API integration testing
- The local dev setup (SvelteKit dev server + Go API) already supports frontend development with a real backend

**Migration approach (cutover)**:

1. **Phase 1**: Implement API integration alongside existing localStorage code (temporary parallel implementation for testing).
2. **Phase 2**: Once API is stable and tested, remove all localStorage code in a single cutover:
   - Remove localStorage-based stores (`stores.ts` localStorage persistence)
   - Replace with API client calls (fetch wrapper)
   - Remove demo data seeding from localStorage
   - Update all components to use API-based stores
   - Remove localStorage-related utilities if no longer needed

**What gets removed**:

- localStorage persistence in `stores.ts` (the `subscribe` handlers that save to localStorage)
- `demo_data.ts` seeding function (or replace with API-based seeding if needed for development)
- Any localStorage-specific data loading/saving logic
- Client-side UUID generation (backend generates UUIDs)

**What stays**:

- UI components and routing (no changes needed)
- Date/time formatting utilities (still needed for display, but may need updates for RFC3339 format from API)
- Form validation logic
- All business logic and UI behavior

**API client implementation**:

- Use relative URLs (`/api/...`) so the same code works in dev (proxied) and production (served by Go)
- Store JWT access tokens in memory (not localStorage) for security
- Handle refresh token cookies automatically (HttpOnly, set by backend)
- Implement automatic token refresh on 401 responses
- Implement proper error handling and loading states
- Handle authentication state (logged-in owner)

## Deployment & Hosting

- App will be hosted on Render.com for ease of deployment and free-tier prototyping
- No custom domain to start; app will be accessed via Render’s default subdomain
- When (or if) the project continues past the demo phase, revisit domain registration and production hosting
- **Deployment**: Single Go binary serves both frontend (static SvelteKit build) and backend API
- **Next step**: Complete frontend migration to backend API, then deploy integrated application

## Security

### Email verification

Owners must verify their email before editing. The backend returns `email_verified` on the owner object from `GET /api/auth/me`; when false, the UI shows a persistent banner and blocks write actions (or shows a clear message on 403 `email_not_verified`). A dedicated `/verify-email?token=...` page calls the verify endpoint and redirects on success; a "Resend verification email" action calls the resend endpoint (rate-limited).

## Backoffice

### Manual handling of owners, their accounts and venues

The Backoffice interface provides administrators with a dedicated section for platform management.

- **Route**: Exposed under `/backoffice`, protected by client-side and server-side admin checks.
- **Interface**:
  - **Dashboard**: High-level statistics of users and venues.
  - **Owner Management**: Searchable list of all owners with actions to manually verify emails, view all their venues, or delete accounts.
  - **Venue Management**: Global list of all venues for oversight.
- **Navigation**: A "Backoffice" link appears in the main navigation for authenticated admin users.

### Viewing statistics (see intrumentation)

## Instrumentation and Observability

### Basic statistics (to save database cost)
