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
- As a venue owner, I want to update any detail (name, banner/image, contact, events, times, comment) for my selected venue

## Structure

### Data Model

The application manages four main entities with the following relationships:

- **Venue Owner**: A registered user who can create and manage venues. A venue owner has:

  - `owner_uuid` (UUID, unique identifier)
  - `name` (STRING, venue owner's name)
  - `mobile` (STRING, contact mobile phone number)
  - `email` (STRING, contact email address)
  - `created_at` (STRING, RFC3339 datetime with timezone offset, e.g., "2024-12-25T10:00:00-05:00")
  - `modified_at` (STRING, RFC3339 datetime with timezone offset, e.g., "2024-12-25T10:00:00-05:00")

- **Venue**: A location that can have zero or more event lists. A venue has:

  - `venue_uuid` (UUID, unique identifier)
  - `name` (STRING, venue name)
  - `banner_image` (STRING, banner/image URL or base64 encoded data)
  - `address` (STRING, physical address of the venue)
  - `geolocation` (STRING, geolocation coordinates, e.g., "latitude,longitude" or JSON format)
  - `comment` (STRING, optional comment)
  - `owner_uuid` (UUID, foreign key linking venue to venue owner for multi-owner support)
  - `timezone` (STRING, timezone identifier, e.g., "Asia/Jerusalem")
  - `event_list_uuids` (ARRAY[UUID], array of event list UUIDs belonging to this venue; cardinality: 0..\*)
  - `created_at` (STRING, RFC3339 datetime with timezone offset, e.g., "2024-12-25T10:00:00-05:00")
  - `modified_at` (STRING, RFC3339 datetime with timezone offset, e.g., "2024-12-25T10:00:00-05:00")

- **Event List**: A named collection of events belonging to a venue. An event list has:

  - `event_list_uuid` (UUID, unique identifier)
  - `venue_uuid` (UUID, foreign key linking event list to venue)
  - `name` (STRING, event list name/title)
  - `comment` (STRING, optional comment)
  - `event_uuids` (ARRAY[UUID], array of event UUIDs belonging to this event list; cardinality: 0..\*)
  - `created_at` (STRING, RFC3339 datetime with timezone offset, e.g., "2024-12-25T10:00:00-05:00")
  - `modified_at` (STRING, RFC3339 datetime with timezone offset, e.g., "2024-12-25T10:00:00-05:00")

- **Event**: A scheduled occurrence with a specific time. An event has:
  - `event_uuid` (UUID, unique identifier)
  - `event_list_uuid` (UUID, foreign key linking event to event list)
  - `event_name` (STRING, event name)
  - `datetime` (STRING, RFC3339 datetime with timezone offset, e.g., "2025-12-25T14:30:00-05:00")
  - `duration_minutes` (INTEGER, duration of the event in minutes)
  - `comment` (STRING, optional comment)
  - `created_at` (STRING, RFC3339 datetime with timezone offset, e.g., "2024-12-25T10:00:00-05:00")
  - `modified_at` (STRING, RFC3339 datetime with timezone offset, e.g., "2024-12-25T10:00:00-05:00")

### Searchable Fields

The search functionality allows users to find venues by matching search queries against fields that are visible on the home page. The following fields are searchable:

- **Venue Owner**:

  - `name` (STRING, venue owner's name)
  - `mobile` (STRING, contact mobile phone number)

- **Venue**:

  - `name` (STRING, venue name)
  - `address` (STRING, physical address of the venue)
  - `comment` (STRING, optional comment)

- **Event List**:

  - `name` (STRING, event list name/title)
  - `comment` (STRING, optional comment)

- **Event**:
  - `event_name` (STRING, event name)
  - `comment` (STRING, optional comment)

When a user performs a search, the system should match the query against all of these fields across all entities. A venue should appear in search results if the query matches any of the searchable fields associated with that venue, its owner, its event lists, or its events.

### Data Format Specifications

- **UUIDs**: All entity identifiers use UUID format (e.g., `550e8400-e29b-41d4-a716-446655440000`)
- **DateTime Format**: All datetime fields use RFC3339 format with timezone offset (e.g., "2025-12-25T14:30:00-05:00")
  - Event `datetime` field stores the full date and time with timezone information
  - `created_at` and `modified_at` fields track when entities were created and last modified
  - The presentation layer determines what portion of the datetime to display (date only, time only, or full datetime) based on context

### User Interface Behavior

- Event lists are selectable on both the visitor's page and the edit page when a venue has been selected
- When a venue is selected, users can choose which event list to view (if the venue has multiple event lists)
- If a venue has no event lists, no events are displayed
- Venue owners can reorder event lists when editing a venue
- The presentation layer determines how to display event datetimes (e.g., time only, date and time, or grouped by date)

## Security & Safety Considerations

### Public Visibility & Physical Safety

- **Primary Concern**: Publicizing Jewish places of gathering could make them targets for attacks. The platform must balance accessibility for community members with physical safety considerations.
- A venue owner can choose if a venue is public (searchable on the public page) or private (only users with the link can view the venue).
- **Recommendation**: Consider defaulting venues to private/unlisted, requiring explicit opt-in for public visibility. This ensures venues are only publicized when owners consciously choose to do so.

### Privacy & Data Protection

- **Venue Owner Information**: Email addresses and mobile phone numbers should not be publicly visible. Consider:
  - Only displaying contact information to authenticated venue owners
  - Providing a contact form that forwards messages without exposing owner details
  - Allowing venue owners to choose what contact information (if any) is publicly displayed
- **Location Privacy**: Consider options for:
  - Showing approximate location (neighborhood/area) instead of exact addresses for public venues
  - Allowing venue owners to hide precise geolocation coordinates
  - Requiring explicit consent before displaying exact addresses publicly

### Access Control & Link Security

- **Private Link Security**: If venues can be accessed via private links:
  - Use cryptographically secure, unguessable tokens (not sequential IDs)
  - Consider expiration dates for shared links
  - Allow venue owners to revoke/regenerate private links
- **Venue Owner Authentication**: Ensure proper access control so venue owners can only edit their own venues (already mentioned in constraints, but critical for security)

### Content & Abuse Prevention

- **Verification**: Consider a verification process to ensure venue owners are legitimate (e.g., email verification, manual approval for sensitive venues)
- **Rate Limiting**: Implement measures to prevent automated scraping or enumeration of venues (especially important for private venues)
- **Content Moderation**: Define policies for handling malicious or inappropriate content

### Operational Security

- **Time-based Privacy**: Consider whether exact event times should be publicly visible, or if approximate times (e.g., "evening services") would be safer
- **Search Functionality**: For public venues, ensure search doesn't inadvertently expose private information or enable targeted discovery of all venues in a specific area

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
