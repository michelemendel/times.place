# Specification

## Purpose

To prototype a simple, user-friendly web app (frontend-only) for listing venues and their event times. The app should make it easy for the public to access up-to-date info and allow venue owners to register and manage their own venues.

## Definitions

- **Venue**: A location (synagogue, community center, etc.) that hosts events with scheduled times.
- **Visitor**: A user visiting the application searching for venues and events.
- **Venue Owner**: A registered user who can create and manage their own venues. Similar to Facebook pages, each person registers and manages venues related to them. A venue owner can have multiple venues but can only see and edit venues they own.

## User Stories

- As a visitor, I want to select a venue from a dropdown and see its upcoming event times, banner image, and contact details.
- As a visitor, I want information presented clearly in both English and Hebrew.
- As a venue owner, I want to register/log in and see a list of my venues
- As a venue owner, I want to add a new venue that I manage
- As a venue owner, I want to delete one of my venues
- As a venue owner, I want to select one of my venues to edit
- As a venue owner, I want to update any detail (name, banner/image, contact, events, times, extra info) for my selected venue

## Success Criteria

- App displays demo data with multiple venues and their times.
- Venue owners can edit all info directly in the frontend; changes reflect immediately and persist locally in the browser (no server required)
- Clean responsive design, working on mobile and desktop.
- Demo supports both English and Hebrew for content/labelling.

## Constraints & Open Decisions

- No backend for this prototype (frontend-only; local storage demo).
- Multiple venue owners must be supported, with each owner having access only to venues they created
- Venue owner authentication should be simple for demo purposes (no complex auth system)
- No user registration for the general public (visitors don't need accounts)
- UI should support Hebrew (RTL?) if demoed to native users
- Hosting: Postpone domain registration until demo feedback is obtained

## Learning Goals

- Learn Svelte and SvelteKit
- Apply agentic, specification-driven development with Svelte
- Compare dynamic forms vs markdown-parsing input for venue owner editing

## References

- [Spec-driven agentic coding conversation](https://www.perplexity.ai/search/you-are-a-specialist-in-spec-d-SPJpkcwqQ9KIRazcPMfT1g)
- [github_spec-kit](https://github.com/github/spec-kit?tab=readme-ov-file)
- [A Practical Guide to Spec-Driven Development](https://docs.zencoder.ai/user-guides/tutorials/spec-driven-development-guide#api-design)
