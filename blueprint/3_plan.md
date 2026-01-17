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

- Default/demo data pre-loaded on first use
- On change, serialize data and save to localStorage

## Localization

- Support for both English and Hebrew via label toggles or i18n helper
- Consider direction swap for Hebrew (RTL support)

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

## Deployment & Hosting

- App will be hosted on Render.com for ease of deployment and free-tier prototyping
- No custom domain to start; app will be accessed via Render’s default subdomain
- When (or if) the project continues past the demo phase, revisit domain registration and production hosting
- Initial rollout: only the frontend (Svelte/SvelteKit) will be deployed; backend integration planned for future (Golang/PostgreSQL)
