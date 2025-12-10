# Technical Plan

## Tech Stack

- Frontend: Svelte and SvelteKit (for routing, reactivity and maintainability)
- CSS: Use standard UI from Tailwind CSS. We do not have the energy to tweak CSS.
- Data: Initially, static JS object/JSON, persisted to local storage.
- Backend (planned later): Golang with PostgreSQL

## Data Model Structure

**Reference**: See the complete data model specification in `2_spec.md` → Structure → Data Model section.

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
  - Detail panel: Banner image, venue name, contact info, event list selector (if venue has multiple event lists), events display.
  - Event list selector: Dropdown or tabs to switch between event lists when a venue has multiple event lists
  - Events display: Shows events from the selected event list (or the only event list if there's just one)
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

## Data Storage

- Default/demo data pre-loaded on first use
- On change, serialize data and save to localStorage

## Localization

- Support for both English and Hebrew via label toggles or i18n helper
- Consider direction swap for Hebrew (RTL support)

## Image Handling

- Use simple file upload for banners, store as base64 or blob URL in localStorage

## Risks & Uncertainties

- Security is minimal for demo
- Local storage limitations for images

## Deployment & Hosting

- App will be hosted on Render.com for ease of deployment and free-tier prototyping
- No custom domain to start; app will be accessed via Render’s default subdomain
- When (or if) the project continues past the demo phase, revisit domain registration and production hosting
- Initial rollout: only the frontend (Svelte/SvelteKit) will be deployed; backend integration planned for future (Golang/PostgreSQL)
