package testdata

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Execer is an interface that both *sql.DB and *sql.Tx satisfy
// This allows SeedTestData to work with both database connections and transactions
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// TestData holds references to seeded test data for use in tests
type TestData struct {
	Owner1UUID uuid.UUID
	Owner2UUID uuid.UUID
	Venue1UUID uuid.UUID
	Venue2UUID uuid.UUID
	Venue3UUID uuid.UUID
	Venue4UUID uuid.UUID
	EventList1UUID uuid.UUID
	EventList2UUID uuid.UUID
	EventList3UUID uuid.UUID
	EventList4UUID uuid.UUID // הדגמה: אוהל אברהם - זמני תפילות חול, public so venue appears in public list
	Event1UUID uuid.UUID
	Event2UUID uuid.UUID
	Event3UUID uuid.UUID
	Event4UUID uuid.UUID
	Event5UUID uuid.UUID
	Event6UUID uuid.UUID
	Event7UUID uuid.UUID // שחרית - Ohel Avraham
	Event8UUID uuid.UUID // מנחה - Ohel Avraham
	Event9UUID uuid.UUID // מעריב - Ohel Avraham
}

// SeedTestData inserts test data into the database and returns references to the created records.
// This is designed for use in tests and development, NOT for production.
//
// The seeded data matches the structure from frontend/src/lib/demo_data.ts:
// - Owner 1 (Abe): 2 venues (Beth El Synagogue, הדגמה: אוהל אברהם / Ohel Avraham)
// - Owner 2 (Ben): 2 venues (Beit Midrash, Chagat House)
// - Various event lists and events (including Hebrew prayer times for הדגמה: אוהל אברהם)
//
// Password for all test owners is "demo" (hashed with bcrypt).
func SeedTestData(ctx context.Context, db Execer) (*TestData, error) {
	now := time.Now()
	data := &TestData{}

	// Hash password for test users (password: "demo")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("demo"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Owner 1: Abe - insert or get existing (is_demo = true for locking and clear-demo-only)
	data.Owner1UUID = uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO venue_owners (owner_uuid, name, mobile, email, password_hash, is_demo, created_at, modified_at)
		VALUES ($1, $2, $3, $4, $5, true, $6, $7)
		ON CONFLICT (email) DO UPDATE SET is_demo = true, modified_at = EXCLUDED.modified_at
	`, data.Owner1UUID, "Abe", "+1-555-0199", "abe@demo.org", passwordHash, now, now)
	if err != nil {
		return nil, err
	}
	// Get the actual UUID (either the one we just inserted or the existing one)
	err = db.QueryRowContext(ctx, `
		SELECT owner_uuid FROM venue_owners WHERE email = $1
	`, "abe@demo.org").Scan(&data.Owner1UUID)
	if err != nil {
		return nil, err
	}

	// Owner 2: Ben - insert or get existing (is_demo = true for locking and clear-demo-only)
	data.Owner2UUID = uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO venue_owners (owner_uuid, name, mobile, email, password_hash, is_demo, created_at, modified_at)
		VALUES ($1, $2, $3, $4, $5, true, $6, $7)
		ON CONFLICT (email) DO UPDATE SET is_demo = true, modified_at = EXCLUDED.modified_at
	`, data.Owner2UUID, "Ben", "+1-555-0200", "ben@demo.org", passwordHash, now, now)
	if err != nil {
		return nil, err
	}
	// Get the actual UUID (either the one we just inserted or the existing one)
	err = db.QueryRowContext(ctx, `
		SELECT owner_uuid FROM venue_owners WHERE email = $1
	`, "ben@demo.org").Scan(&data.Owner2UUID)
	if err != nil {
		return nil, err
	}

	// Owner 1 - Venue 1: Beth El Synagogue - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
	`, data.Owner1UUID, "DEMO: Beth El Synagogue").Scan(&data.Venue1UUID)
	if err == sql.ErrNoRows {
		// Venue doesn't exist, insert it
		data.Venue1UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO venues (venue_uuid, owner_uuid, name, banner_image, address, geolocation, comment, timezone, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Venue1UUID, data.Owner1UUID, "DEMO: Beth El Synagogue", "https://images.unsplash.com/photo-1486718448742-163732cd1544?w=600&h=200&fit=crop",
			"15 King George Street, Jerusalem", "31.7787,35.2175", "A welcoming community in the heart of the city.",
			"Asia/Jerusalem", now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Owner 1 - Venue 2: הדגמה: אוהל אברהם (Ohel Avraham) - Hebrew community venue - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
	`, data.Owner1UUID, "הדגמה: אוהל אברהם").Scan(&data.Venue2UUID)
	if err == sql.ErrNoRows {
		// Try legacy name "אוהל אברהם" (without prefix) and update in place if present
		err = db.QueryRowContext(ctx, `
			SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
		`, data.Owner1UUID, "אוהל אברהם").Scan(&data.Venue2UUID)
		if err == nil {
			_, err = db.ExecContext(ctx, `
				UPDATE venues SET name = $1, modified_at = $2 WHERE venue_uuid = $3
			`, "הדגמה: אוהל אברהם", now, data.Venue2UUID)
			if err != nil {
				return nil, err
			}
		} else if err == sql.ErrNoRows {
			// Try legacy name "DEMO: Community Center" and update in place if present
			err = db.QueryRowContext(ctx, `
				SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
			`, data.Owner1UUID, "DEMO: Community Center").Scan(&data.Venue2UUID)
			if err == nil {
				_, err = db.ExecContext(ctx, `
					UPDATE venues SET name = $1, banner_image = $2, address = $3, comment = $4, modified_at = $5
					WHERE venue_uuid = $6
				`, "הדגמה: אוהל אברהם", "https://images.unsplash.com/photo-1512917774080-9991f1c4c750?w=600&h=200&fit=crop",
					"רחוב בן יהודה 42, ירושלים", "מרחב קהילתי רב-שימושי.", now, data.Venue2UUID)
				if err != nil {
					return nil, err
				}
			} else if err == sql.ErrNoRows {
				// Venue doesn't exist, insert it
				data.Venue2UUID = uuid.New()
				_, err = db.ExecContext(ctx, `
					INSERT INTO venues (venue_uuid, owner_uuid, name, banner_image, address, geolocation, comment, timezone, created_at, modified_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`, data.Venue2UUID, data.Owner1UUID, "הדגמה: אוהל אברהם", "https://images.unsplash.com/photo-1512917774080-9991f1c4c750?w=600&h=200&fit=crop",
				"רחוב בן יהודה 42, ירושלים", "31.7800,35.2167", "מרחב קהילתי רב-שימושי.",
					"Asia/Jerusalem", now, now)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Owner 2 - Venue 1: Beit Midrash - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
	`, data.Owner2UUID, "DEMO: Beit Midrash").Scan(&data.Venue3UUID)
	if err == sql.ErrNoRows {
		// Venue doesn't exist, insert it
		data.Venue3UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO venues (venue_uuid, owner_uuid, name, banner_image, address, geolocation, comment, timezone, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Venue3UUID, data.Owner2UUID, "DEMO: Beit Midrash", "https://placehold.co/600x200?text=Beit+Midrash",
			"28 Jaffa Road, Jerusalem", "31.7820,35.2180", "Study hall and prayer space.",
			"Asia/Jerusalem", now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Owner 2 - Venue 2: Chagat House - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
	`, data.Owner2UUID, "DEMO: Chagat House").Scan(&data.Venue4UUID)
	if err == sql.ErrNoRows {
		// Venue doesn't exist, insert it
		data.Venue4UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO venues (venue_uuid, owner_uuid, name, banner_image, address, geolocation, comment, timezone, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Venue4UUID, data.Owner2UUID, "DEMO: Chagat House", "https://placehold.co/600x200?text=Chagat+House",
			"12 Rechov Agron, Jerusalem", "31.7750,35.2200", "Warm and welcoming Chagat center.",
			"Asia/Jerusalem", now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event List 1 for Venue 1: Daily Minyan - PUBLIC - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_list_uuid FROM event_lists WHERE venue_uuid = $1 AND name = $2
	`, data.Venue1UUID, "Daily Minyan").Scan(&data.EventList1UUID)
	if err == sql.ErrNoRows {
		// Event list doesn't exist, insert it
		data.EventList1UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO event_lists (event_list_uuid, venue_uuid, name, date, comment, visibility, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.EventList1UUID, data.Venue1UUID, "Daily Minyan", "2025-12-25",
			"Morning and afternoon prayers", "public", 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event List 2 for Venue 1: Shabbat Services - PUBLIC - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_list_uuid FROM event_lists WHERE venue_uuid = $1 AND name = $2
	`, data.Venue1UUID, "Shabbat Services").Scan(&data.EventList2UUID)
	if err == sql.ErrNoRows {
		// Event list doesn't exist, insert it
		data.EventList2UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO event_lists (event_list_uuid, venue_uuid, name, date, comment, visibility, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.EventList2UUID, data.Venue1UUID, "Shabbat Services", "2025-12-26",
			"Friday evening and Saturday morning services", "public", 1, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event List for Venue 2 (הדגמה: אוהל אברהם): זמני תפילות חול - public so the venue appears in the public venues list
	err = db.QueryRowContext(ctx, `
		SELECT event_list_uuid FROM event_lists WHERE venue_uuid = $1 AND name = $2
	`, data.Venue2UUID, "זמני תפילות חול").Scan(&data.EventList4UUID)
	if err == sql.ErrNoRows {
		// Try legacy name "Community Events" and update in place if present
		err = db.QueryRowContext(ctx, `
			SELECT event_list_uuid FROM event_lists WHERE venue_uuid = $1 AND name = $2
		`, data.Venue2UUID, "Community Events").Scan(&data.EventList4UUID)
		if err == nil {
			_, err = db.ExecContext(ctx, `
				UPDATE event_lists SET name = $1, comment = $2, modified_at = $3 WHERE event_list_uuid = $4
			`, "זמני תפילות חול", "תפילות ימי חול", now, data.EventList4UUID)
			if err != nil {
				return nil, err
			}
		} else if err == sql.ErrNoRows {
			data.EventList4UUID = uuid.New()
			_, err = db.ExecContext(ctx, `
				INSERT INTO event_lists (event_list_uuid, venue_uuid, name, date, comment, visibility, sort_order, created_at, modified_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, data.EventList4UUID, data.Venue2UUID, "זמני תפילות חול", "2025-12-25",
				"תפילות ימי חול", "public", 0, now, now)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event List 3 for Venue 3: Weekly Schedule - PRIVATE - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_list_uuid FROM event_lists WHERE venue_uuid = $1 AND name = $2
	`, data.Venue3UUID, "Weekly Schedule").Scan(&data.EventList3UUID)
	if err == sql.ErrNoRows {
		// Event list doesn't exist, insert it
		data.EventList3UUID = uuid.New()
		privateToken := uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO event_lists (event_list_uuid, venue_uuid, name, date, comment, visibility, private_link_token, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.EventList3UUID, data.Venue3UUID, "Weekly Schedule", "2025-12-25",
			"Regular weekly learning sessions", "private", privateToken, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 1: Shacharis - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND datetime = $3
	`, data.EventList1UUID, "Shacharis", "2025-12-25T06:00:00+02:00").Scan(&data.Event1UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event1UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, datetime, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.Event1UUID, data.EventList1UUID, "Shacharis", "2025-12-25T06:00:00+02:00",
			"we start on time", 0, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 2: Mincha - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND datetime = $3
	`, data.EventList1UUID, "Mincha", "2025-12-25T16:30:00+02:00").Scan(&data.Event2UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event2UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, datetime, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.Event2UUID, data.EventList1UUID, "Mincha", "2025-12-25T16:30:00+02:00",
			"", 0, 1, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 3: Kabbalat Shabbat - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND datetime = $3
	`, data.EventList2UUID, "Kabbalat Shabbat", "2025-12-26T17:30:00+02:00").Scan(&data.Event3UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event3UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, datetime, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.Event3UUID, data.EventList2UUID, "Kabbalat Shabbat", "2025-12-26T17:30:00+02:00",
			"", 60, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 4: Shabbat Morning - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND datetime = $3
	`, data.EventList2UUID, "Shabbat Morning", "2025-12-27T09:00:00+02:00").Scan(&data.Event4UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event4UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, datetime, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.Event4UUID, data.EventList2UUID, "Shabbat Morning", "2025-12-27T09:00:00+02:00",
			"", 120, 1, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 5: Morning Learning - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND datetime = $3
	`, data.EventList3UUID, "Morning Learning", "2025-12-25T08:00:00+02:00").Scan(&data.Event5UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event5UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, datetime, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.Event5UUID, data.EventList3UUID, "Morning Learning", "2025-12-25T08:00:00+02:00",
			"", 90, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 6: Evening Shiur - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND datetime = $3
	`, data.EventList3UUID, "Evening Shiur", "2025-12-25T19:30:00+02:00").Scan(&data.Event6UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event6UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, datetime, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.Event6UUID, data.EventList3UUID, "Evening Shiur", "2025-12-25T19:30:00+02:00",
			"", 60, 1, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 7: שחרית (Shacharit) - הדגמה: אוהל אברהם - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND datetime = $3
	`, data.EventList4UUID, "שחרית", "2025-12-25T06:30:00+02:00").Scan(&data.Event7UUID)
	if err == sql.ErrNoRows {
		data.Event7UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, datetime, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.Event7UUID, data.EventList4UUID, "שחרית", "2025-12-25T06:30:00+02:00",
			"", 0, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 8: מנחה (Mincha) - הדגמה: אוהל אברהם - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND datetime = $3
	`, data.EventList4UUID, "מנחה", "2025-12-25T16:30:00+02:00").Scan(&data.Event8UUID)
	if err == sql.ErrNoRows {
		data.Event8UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, datetime, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.Event8UUID, data.EventList4UUID, "מנחה", "2025-12-25T16:30:00+02:00",
			"", 0, 1, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 9: מעריב (Ma'ariv) - הדגמה: אוהל אברהם - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND datetime = $3
	`, data.EventList4UUID, "מעריב", "2025-12-25T18:00:00+02:00").Scan(&data.Event9UUID)
	if err == sql.ErrNoRows {
		data.Event9UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, datetime, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, data.Event9UUID, data.EventList4UUID, "מעריב", "2025-12-25T18:00:00+02:00",
			"", 0, 2, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return data, nil
}

// ClearTestData removes all test data from the database.
// This deletes all records from all tables (in dependency order).
// Use with caution - this will delete ALL data in the database.
func ClearTestData(ctx context.Context, db Execer) error {
	// Delete in reverse dependency order
	queries := []string{
		"DELETE FROM refresh_tokens",
		"DELETE FROM events",
		"DELETE FROM event_lists",
		"DELETE FROM venues",
		"DELETE FROM venue_owners",
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

// ClearDemoDataOnly removes only demo (seeded) owners and their data.
// Deletes venue_owners where is_demo = true; CASCADE removes their
// venues, event_lists, events, and refresh_tokens. Real users' data is untouched.
func ClearDemoDataOnly(ctx context.Context, db Execer) error {
	_, err := db.ExecContext(ctx, `DELETE FROM venue_owners WHERE is_demo = true`)
	return err
}
