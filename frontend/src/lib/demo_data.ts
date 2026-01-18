import { currentOwnerStore, ownersStore, venueStore, eventListStore, eventStore } from './stores';
import type { VenueOwner, Venue, EventList, Event } from './types';
import { get } from 'svelte/store';
import { getCurrentTimestamp } from './utils/datetime.js';
import { generateUUID } from './utils/uuid.js';

export function seedDemoData(force: boolean = false) {
    // Only seed if no data exists, unless force is true
    if (!force && (get(ownersStore).length > 0 || get(venueStore).length > 0)) {
        console.log('Data already exists, skipping seed.');
        return;
    }

    // If forcing, clear existing data first
    if (force) {
        console.log('Force re-seeding: clearing existing data...');
        ownersStore.set([]);
        venueStore.set([]);
        eventListStore.set([]);
        eventStore.set([]);
    }

    console.log('Seeding demo data...');

    const owners: VenueOwner[] = [];
    const venues: Venue[] = [];
    const eventLists: EventList[] = [];
    const events: Event[] = [];

    const now = getCurrentTimestamp();

    // Owner 1: Demo Rabbi
    const owner1Id = generateUUID();
    const owner1: VenueOwner = {
        owner_uuid: owner1Id,
        name: "Demo Rabbi",
        mobile: "+1-555-0199",
        email: "demo@synagogue.org",
        password: "demo",
        created_at: now,
        modified_at: now
    };
    owners.push(owner1);

    // Owner 1 - Venue 1: Beth El Synagogue (has multiple event lists)
    const venue1Id = generateUUID();
    const venue1: Venue = {
        venue_uuid: venue1Id,
        name: "Beth El Synagogue",
        banner_image: "https://placehold.co/600x200?text=Beth+El",
        address: "15 King George Street, Jerusalem",
        geolocation: "31.7787,35.2175",
        comment: "A welcoming community in the heart of the city.",
        owner_uuid: owner1Id,
        event_list_uuids: [],
        timezone: "Asia/Jerusalem",
        created_at: now,
        modified_at: now
    };
    venues.push(venue1);

    // Event List 1 for Venue 1: Daily Minyan - PUBLIC
    const list1Id = generateUUID();
    const eventList1: EventList = {
        event_list_uuid: list1Id,
        venue_uuid: venue1Id,
        name: "Daily Minyan",
        date: "2025-12-25",
        comment: "Morning and afternoon prayers",
        visibility: "public",
        event_uuids: [],
        created_at: now,
        modified_at: now
    };
    eventLists.push(eventList1);
    venue1.event_list_uuids.push(list1Id);

    const event1Id = generateUUID();
    const event1: Event = {
        event_uuid: event1Id,
        event_list_uuid: list1Id,
        event_name: "Shacharis",
        datetime: "2025-12-25T06:00:00+02:00",
        comment: "we start on time",
        duration_minutes: 0,
        created_at: now,
        modified_at: now
    };
    events.push(event1);

    const event2Id = generateUUID();
    const event2: Event = {
        event_uuid: event2Id,
        event_list_uuid: list1Id,
        event_name: "Mincha",
        datetime: "2025-12-25T16:30:00+02:00",
        duration_minutes: 0,
        created_at: now,
        modified_at: now
    };
    events.push(event2);
    eventList1.event_uuids.push(event1Id, event2Id);

    // Event List 2 for Venue 1: Shabbat Services - PUBLIC
    const list2Id = generateUUID();
    const eventList2: EventList = {
        event_list_uuid: list2Id,
        venue_uuid: venue1Id,
        name: "Shabbat Services",
        date: "2025-12-26",
        comment: "Friday evening and Saturday morning services",
        visibility: "public",
        event_uuids: [],
        created_at: now,
        modified_at: now
    };
    eventLists.push(eventList2);
    venue1.event_list_uuids.push(list2Id);

    const event3Id = generateUUID();
    const event3: Event = {
        event_uuid: event3Id,
        event_list_uuid: list2Id,
        event_name: "Kabbalat Shabbat",
        datetime: "2025-12-26T17:30:00+02:00",
        duration_minutes: 60,
        created_at: now,
        modified_at: now
    };
    events.push(event3);

    const event4Id = generateUUID();
    const event4: Event = {
        event_uuid: event4Id,
        event_list_uuid: list2Id,
        event_name: "Shabbat Morning",
        datetime: "2025-12-27T09:00:00+02:00",
        duration_minutes: 120,
        created_at: now,
        modified_at: now
    };
    events.push(event4);
    eventList2.event_uuids.push(event3Id, event4Id);

    // Owner 1 - Venue 2: Community Center (has no event lists)
    const venue2Id = generateUUID();
    const venue2: Venue = {
        venue_uuid: venue2Id,
        name: "Community Center",
        banner_image: "https://placehold.co/600x200?text=Community+Center",
        address: "42 Ben Yehuda Street, Jerusalem",
        geolocation: "31.7800,35.2167",
        comment: "Multi-purpose community space.",
        owner_uuid: owner1Id,
        event_list_uuids: [],
        timezone: "Asia/Jerusalem",
        created_at: now,
        modified_at: now
    };
    venues.push(venue2);

    // Owner 2: Sarah Cohen
    const owner2Id = generateUUID();
    const owner2: VenueOwner = {
        owner_uuid: owner2Id,
        name: "Sarah Cohen",
        mobile: "+1-555-0200",
        email: "sarah@community.org",
        password: "demo",
        created_at: now,
        modified_at: now
    };
    owners.push(owner2);

    // Owner 2 - Venue 1: Beit Midrash (has one event list)
    const venue3Id = generateUUID();
    const venue3: Venue = {
        venue_uuid: venue3Id,
        name: "Beit Midrash",
        banner_image: "https://placehold.co/600x200?text=Beit+Midrash",
        address: "28 Jaffa Road, Jerusalem",
        geolocation: "31.7820,35.2180",
        comment: "Study hall and prayer space.",
        owner_uuid: owner2Id,
        event_list_uuids: [],
        timezone: "Asia/Jerusalem",
        created_at: now,
        modified_at: now
    };
    venues.push(venue3);

    // Event List 3 for Venue 3: Weekly Schedule - PRIVATE (to demonstrate private event lists)
    const list3Id = generateUUID();
    const eventList3Token = generateUUID();
    const eventList3: EventList = {
        event_list_uuid: list3Id,
        venue_uuid: venue3Id,
        name: "Weekly Schedule",
        date: "2025-12-25",
        comment: "Regular weekly learning sessions",
        visibility: "private",
        private_link_token: eventList3Token,
        event_uuids: [],
        created_at: now,
        modified_at: now
    };
    eventLists.push(eventList3);
    venue3.event_list_uuids.push(list3Id);

    const event5Id = generateUUID();
    const event5: Event = {
        event_uuid: event5Id,
        event_list_uuid: list3Id,
        event_name: "Morning Learning",
        datetime: "2025-12-25T08:00:00+02:00",
        duration_minutes: 90,
        created_at: now,
        modified_at: now
    };
    events.push(event5);

    const event6Id = generateUUID();
    const event6: Event = {
        event_uuid: event6Id,
        event_list_uuid: list3Id,
        event_name: "Evening Shiur",
        datetime: "2025-12-25T19:30:00+02:00",
        duration_minutes: 60,
        created_at: now,
        modified_at: now
    };
    events.push(event6);
    eventList3.event_uuids.push(event5Id, event6Id);

    // Owner 2 - Venue 2: Chabad House (has no event lists)
    const venue4Id = generateUUID();
    const venue4: Venue = {
        venue_uuid: venue4Id,
        name: "Chabad House",
        banner_image: "https://placehold.co/600x200?text=Chabad+House",
        address: "12 Rechov Agron, Jerusalem",
        geolocation: "31.7750,35.2200",
        comment: "Warm and welcoming Chabad center.",
        owner_uuid: owner2Id,
        event_list_uuids: [],
        timezone: "Asia/Jerusalem",
        created_at: now,
        modified_at: now
    };
    venues.push(venue4);

    // Update stores (don't set currentOwnerStore - owners will log in separately)
    ownersStore.set(owners);
    venueStore.set(venues);
    eventListStore.set(eventLists);
    eventStore.set(events);

    console.log(`Demo data seeded! Created ${owners.length} venue owners, ${venues.length} venues, ${eventLists.length} event lists, and ${events.length} events.`);
}
