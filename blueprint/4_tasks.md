# Task List

## Installation and setup

- [ ] Install Svelte and SvelteKit
- [ ] Set up Svelte and SvelteKit project structure
- [ ] Set up Tailwind CSS
- [ ] Makefile: Add commands for Svelte and Tailwaind
- [ ] Start the app with a demo page to see that everything is set up correctly

## Data model

- [ ] Create venue data model
- [ ] Add ownerId field to venue data model for multi-owner support
- [ ] Add demo data
- [ ] Set up multiple venue owner demo accounts (at least two for testing isolation)

## Visitor Page

- [ ] Create header with navigation (Times, About, Venue Owner)
- [ ] Create About page with prototype information
- [ ] Create footer (empty for now)
- [ ] Create dropdown for list of venues using demo data
- [ ] Display selected venue details: banner, contacts, times, etc...

## Venue Owner Page

- [ ] Add venue owner button to main UI
- [ ] Build venue owner registration/account creation flow (for demo: simple form with hardcoded validation)
- [ ] Build venue owner login flow with hardcoded password for prototype
- [ ] Create venue owner dashboard/list view showing all venues owned by logged-in owner
- [ ] Implement client-side filtering to show only venues owned by current venue owner
- [ ] Add functionality to add a venue
- [ ] Add functionality to delete a venue
- [ ] Add functionality to edit a venue (see Venue Form below)
- [ ] Test on mobile + desktop

## Venue Form

- [ ] Form UI with two panes: editing pane (dynamic fields) and live preview pane
- [ ] Dynamic fields: functionality to add, delete, duplicate, and move fields up and down
- [ ] Input validation: date, time, XSS/SQL injection
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
