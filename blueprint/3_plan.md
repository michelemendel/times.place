# Technical Plan

## Tech Stack

- Frontend: Svelte and SvelteKit (for routing, reactivity and maintainability)
- CSS: Use standard UI from Tailwind CSS. We do not have the energy to tweak CSS.
- Data: Initially, static JS object/JSON, persisted to local storage.
- Backend (planned later): Golang with PostgreSQL

## UI Structure

### Visitor page

- Header with an image and navigation: Times, About, Venue Owner
- Times:
  - Landing page: Dropdown to select venue, details panel, venue owner button.
  - Detail panel: Banner image, venue name, contact info, times list.
- About: info about the prototype
- Venue Owner:
  - Button shows login (simple password for demo), then edit page/modal.
  - Edit page: Divided in two panes, one for editing, and the other that shows the rendered result as it will be on the Times/Detail page
- Footer: Empty for now

### Venue owner page - form

- Option 1: Dynamic form fields (input/add/remove for each event).
- Option 2: Markdown input field for all info, parsed/rendered live
- Will prototype both, and let user feedback guide choice.
- Undo functionality

## Venue Owner System & Authentication

- For demo: Use simple password-based authentication (hardcoded credentials)
- Support multiple venue owners: Each venue owner has a unique identifier and can only see/edit venues they created
- Demo setup: Two venue owner accounts will be created to verify multi-owner isolation
- Venue owner data isolation: Venues will have an `ownerId` field (or `venueOwnerId`); filtering happens client-side
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
- Markdown editor usability (especially for bi-lingual content)
- Local storage limitations for images

## Deployment & Hosting

- App will be hosted on Render.com for ease of deployment and free-tier prototyping
- No custom domain to start; app will be accessed via Render’s default subdomain
- When (or if) the project continues past the demo phase, revisit domain registration and production hosting
- Initial rollout: only the frontend (Svelte/SvelteKit) will be deployed; backend integration planned for future (Golang/PostgreSQL)
