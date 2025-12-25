import { userStore, venueStore, eventListStore, eventStore } from './stores';
import type { VenueOwner, Venue, EventList, Event } from './types';
import { get } from 'svelte/store';

// Helper to generate IDs if not available in environment (e.g. older browsers)
const generateUUID = () => {
    if (typeof crypto !== 'undefined' && crypto.randomUUID) {
        return crypto.randomUUID();
    }
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
};

const now = new Date().toISOString();

export function seedDemoData() {
    // Only seed if no user exists
    if (get(userStore)) {
        console.log('Data already exists, skipping seed.');
        return;
    }

    console.log('Seeding demo data...');

    const ownerId = generateUUID();
    const owner: VenueOwner = {
        owner_uuid: ownerId,
        name: "Demo Rabbi",
        mobile: "+1-555-0199",
        email: "demo@synagogue.org",
        created_at: now,
        modified_at: now
    };

    const venueId = generateUUID();
    const venue: Venue = {
        venue_uuid: venueId,
        name: "Beth El Synagogue",
        banner_image: "https://placehold.co/600x200?text=Beth+El",
        address: "123 Shalom Street, Jerusalem",
        geolocation: "31.7767,35.2345",
        description: "A welcoming community in the heart of the city.",
        owner_uuid: ownerId,
        event_list_uuids: [], // Will link later
        timezone: "Asia/Jerusalem",
        created_at: now,
        modified_at: now
    };

    const listId = generateUUID();
    const eventList: EventList = {
        event_list_uuid: listId,
        venue_uuid: venueId,
        name: "Daily Minyan",
        description: "Morning and afternoon prayers",
        event_uuids: [], // Will link later
        created_at: now,
        modified_at: now
    };

    const eventId1 = generateUUID();
    const event1: Event = {
        event_uuid: eventId1,
        event_list_uuid: listId,
        event_name: "Shacharis",
        datetime: "2025-12-25T07:00:00+02:00",
        duration_minutes: 45,
        created_at: now,
        modified_at: now
    };

    const eventId2 = generateUUID();
    const event2: Event = {
        event_uuid: eventId2,
        event_list_uuid: listId,
        event_name: "Mincha",
        datetime: "2025-12-25T16:30:00+02:00",
        duration_minutes: 30,
        created_at: now,
        modified_at: now
    };

    // Link relations
    venue.event_list_uuids.push(listId);
    eventList.event_uuids.push(eventId1, eventId2);

    // Update stores
    userStore.set(owner);
    venueStore.set([venue]);
    eventListStore.set([eventList]);
    eventStore.set([event1, event2]);

    console.log('Demo data seeded!');
}
