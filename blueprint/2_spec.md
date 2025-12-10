# Specification

## Purpose

To prototype a simple, user-friendly web app (frontend-only) for listing venues and their event times. The app should make it easy for the public to access up-to-date info and allow venue owners to register and manage their own venues.

## Definitions

- **Venue**: A location (synagogue, community center, etc.) that hosts events with scheduled times.
- **Visitor**: A user visiting the application searching for venues and events.
- **Venue Owner**: A registered user who can create and manage their own venues. Similar to Facebook pages, each person registers and manages venues related to them. A venue owner can have multiple venues but can only see and edit venues they own. A venue owner has:
  - ownerUUID (unique identifier, UUID format)

## User Stories

- As a visitor, I want to select a venue from a dropdown and see its upcoming event times, banner image, and contact details.
- As a visitor, I want information presented clearly in both English and Hebrew.
- As a venue owner, I want to register/log in and see a list of my venues
- As a venue owner, I want to add a new venue that I manage
- As a venue owner, I want to delete one of my venues
- As a venue owner, I want to select one of my venues to edit
- As a venue owner, I want to update any detail (name, banner/image, contact, events, times, description) for my selected venue

## Structure

### Data Model

The application manages four main entities with the following relationships:

- **Venue Owner**: A registered user who can create and manage venues. A venue owner has:

  - ownerUUID (unique identifier, UUID format)

- **Venue**: A location that can have zero or more event lists. A venue has:

  - venueUUID (unique identifier, UUID format)
  - Name
  - Banner/image
  - Contact details (mobile, email, address)
  - Description (optional)
  - ownerUUID (identifier linking venue to venue owner for multi-owner support, UUID format)
  - Zero or more event lists

- **Event List**: A named collection of events belonging to a venue. An event list has:

  - eventListUUID (unique identifier, UUID format)
  - venueUUID (identifier linking event list to venue, UUID format)
  - Name/title
  - Date (ISO 8601 date format, e.g., "2024-12-25")
  - Description (optional)
  - Zero or more events

- **Event**: A scheduled occurrence with a specific time. An event has:
  - eventUUID (unique identifier, UUID format)
  - eventListUUID (identifier linking event to event list, UUID format)
  - Event name
  - DateTime (stored as full date and time in Unix epoch timestamp format; only the time portion is displayed to users)
  - Description (optional)

### Data Format Specifications

- **UUIDs**: All entity identifiers use UUID format (e.g., `550e8400-e29b-41d4-a716-446655440000`)
- **Date Format**: Event list dates are stored and displayed in ISO 8601 date format (YYYY-MM-DD, e.g., "2024-12-25")
- **Time Format**:
  - Event DateTime is stored internally as full date and time in Unix epoch timestamp format (seconds since January 1, 1970 UTC)
  - When displaying events, only the time portion is shown to users (e.g., "14:30" or "2:30 PM")
  - The date is not displayed for events since it comes from the event list they belong to
  - For internal operations (sorting, filtering), the full datetime stored in the event is used

### User Interface Behavior

- Event lists are selectable on both the visitor's page and the edit page when a venue has been selected
- When a venue is selected, users can choose which event list to view (if the venue has multiple event lists)
- If a venue has no event lists, no events are displayed
- Venue owners can reorder event lists when editing a venue
- Events are displayed with only the time (the date comes from the event list they belong to)

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

## References

- [Spec-driven agentic coding conversation](https://www.perplexity.ai/search/you-are-a-specialist-in-spec-d-SPJpkcwqQ9KIRazcPMfT1g)
- [github_spec-kit](https://github.com/github/spec-kit?tab=readme-ov-file)
- [A Practical Guide to Spec-Driven Development](https://docs.zencoder.ai/user-guides/tutorials/spec-driven-development-guide#api-design)
