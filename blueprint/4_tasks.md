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

## Venue Owner Page Registration, Authentication, and Authorization

- [x] Build venue owner registration/account creation flow (for demo: simple form with hardcoded validation)
- [x] Build venue owner login flow with hardcoded password for prototype

## Venue Owner Page

- [ ] Create venue owner dashboard/list view showing all venues owned by logged-in owner
- [ ] Implement client-side filtering to show only venues owned by current venue owner
- [ ] Add functionality to add a venue
- [ ] Add functionality to delete a venue
- [ ] Add functionality to edit a venue (see Venue Form below)
- [ ] Venues are sorted alphabetically
- [ ] Test on mobile + desktop

## Venue Form

- [ ] Form UI with two panes: editing pane (dynamic fields) and live preview pane
- [ ] Event list management: functionality to add, delete, and reorder event lists
- [ ] For each event list: manage name, date (ISO 8601 format), comment, and events
- [ ] Event management within event lists: functionality to add, delete, duplicate, and move events up and down
- [ ] Event DateTime input: allow venue owners to input time, combine with event list date to create full datetime, convert to Unix epoch timestamp for storage
- [ ] Display event times (time portion only) in both edit and preview panes
- [ ] Event list selector in preview pane to test how different event lists appear to visitors
- [ ] Input validation: date (ISO 8601 format), time (validate and combine with event list date to create full datetime, convert to Unix epoch), XSS/SQL injection
- [ ] Validate UUID format for all entity identifiers
- [ ] Undo functionality
- [ ] Prototype: Save to local storage for all edits
- [ ] Add image upload for banners (prototype: local storage as base64/blob)

## Localization

- [ ] Implement English/Hebrew content support (spec requires both languages)
- [ ] Evaluate and implement RTL support for Hebrew content if needed
- [ ] Test bilingual content display

## Decision/Exploration Tasks

- [ ] Evaluate multilingual switching UX: Do we need an English/Hebrew switch in the application, or will it work using the computer's keyboard input source?

## Deployment

- [ ] Configure SvelteKit for static site export
- [ ] Build static site for deployment
- [ ] Deploy to Render.com
- [ ] Verify deployment and test on Render's default subdomain

## Documentation Tasks

- [ ] Document demo data and editing process
- [ ] Record agentic coding decisions and workflow notes in implement.md
