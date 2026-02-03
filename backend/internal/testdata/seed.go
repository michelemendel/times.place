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
	Owner1UUID     uuid.UUID
	Owner2UUID     uuid.UUID
	Venue1UUID     uuid.UUID
	Venue2UUID     uuid.UUID
	Venue3UUID     uuid.UUID
	Venue4UUID     uuid.UUID
	EventList1UUID uuid.UUID
	EventList2UUID uuid.UUID
	EventList3UUID uuid.UUID // Ben's house - Birthday
	EventList4UUID uuid.UUID // הדגמה: אוהל אברהם - זמני תפילות חול, public so venue appears in public list
	EventList5UUID uuid.UUID // After School Math - Daily Schedule
	Event1UUID     uuid.UUID
	Event2UUID     uuid.UUID
	Event3UUID     uuid.UUID
	Event4UUID     uuid.UUID
	Event5UUID     uuid.UUID // Welcome - Ben's house Birthday
	Event6UUID     uuid.UUID // unused (was Evening Shiur)
	Event7UUID     uuid.UUID // שחרית - Ohel Avraham
	Event8UUID     uuid.UUID // מנחה - Ohel Avraham
	Event9UUID     uuid.UUID // מעריב - Ohel Avraham
	Event10UUID    uuid.UUID // Algebra - After School Math
	Event11UUID    uuid.UUID // Calculus - After School Math
	Event12UUID    uuid.UUID // Lunch - After School Math
	Event13UUID    uuid.UUID // Fourier Transformations - After School Math
}

// SeedTestData inserts test data into the database and returns references to the created records.
// This is designed for use in tests and development, NOT for production.
//
// The seeded data matches the structure from frontend/src/lib/demo_data.ts:
// - Owner 1 (Abe): 2 venues (Beth El Synagogue, הדגמה: אוהל אברהם / Ohel Avraham)
// - Owner 2 (Ben): 2 venues (Ben's house, After School Math)
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

	// Owner 2 - Venue 1: Ben's house (private event demo) - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
	`, data.Owner2UUID, "DEMO: Ben's house").Scan(&data.Venue3UUID)
	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, `
			SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
		`, data.Owner2UUID, "DEMO: Beit Midrash").Scan(&data.Venue3UUID)
		if err == nil {
			_, err = db.ExecContext(ctx, `
				UPDATE venues SET name = $1, banner_image = $2, modified_at = $3 WHERE venue_uuid = $4
			`, "DEMO: Ben's house", "https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?w=600&h=200&fit=crop", now, data.Venue3UUID)
			if err != nil {
				return nil, err
			}
		} else if err == sql.ErrNoRows {
			data.Venue3UUID = uuid.New()
			_, err = db.ExecContext(ctx, `
				INSERT INTO venues (venue_uuid, owner_uuid, name, banner_image, address, geolocation, comment, timezone, created_at, modified_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`, data.Venue3UUID, data.Owner2UUID, "DEMO: Ben's house", "https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?w=600&h=200&fit=crop",
				"28 Jaffa Road, Jerusalem", "31.7820,35.2180", "Private event demo.",
				"Asia/Jerusalem", now, now)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Owner 2 - Venue 2: After School Math - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
	`, data.Owner2UUID, "DEMO: After School Math").Scan(&data.Venue4UUID)
	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, `
			SELECT venue_uuid FROM venues WHERE owner_uuid = $1 AND name = $2
		`, data.Owner2UUID, "DEMO: Chagat House").Scan(&data.Venue4UUID)
		if err == nil {
			_, err = db.ExecContext(ctx, `
				UPDATE venues SET name = $1, banner_image = $2, modified_at = $3 WHERE venue_uuid = $4
			`, "DEMO: After School Math", "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?w=600&h=200&fit=crop", now, data.Venue4UUID)
			if err != nil {
				return nil, err
			}
		} else if err == sql.ErrNoRows {
			data.Venue4UUID = uuid.New()
			_, err = db.ExecContext(ctx, `
				INSERT INTO venues (venue_uuid, owner_uuid, name, banner_image, address, geolocation, comment, timezone, created_at, modified_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`, data.Venue4UUID, data.Owner2UUID, "DEMO: After School Math", "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?w=600&h=200&fit=crop",
				"12 Rechov Agron, Jerusalem", "31.7750,35.2200", "Daily math schedule.",
				"Asia/Jerusalem", now, now)
			if err != nil {
				return nil, err
			}
		} else {
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

	// Event List 3 for Venue 3 (Ben's house): Birthday - PUBLIC - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_list_uuid FROM event_lists WHERE venue_uuid = $1 AND name = $2
	`, data.Venue3UUID, "Birthday").Scan(&data.EventList3UUID)
	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, `
			SELECT event_list_uuid FROM event_lists WHERE venue_uuid = $1 AND name = $2
		`, data.Venue3UUID, "Weekly Schedule").Scan(&data.EventList3UUID)
		if err == nil {
			_, err = db.ExecContext(ctx, `
				UPDATE event_lists SET name = $1, comment = $2, visibility = $3, private_link_token = NULL, modified_at = $4 WHERE event_list_uuid = $5
			`, "Birthday", "Come as you are", "public", now, data.EventList3UUID)
			if err != nil {
				return nil, err
			}
			_, err = db.ExecContext(ctx, `
				DELETE FROM events WHERE event_list_uuid = $1 AND event_name IN ('Morning Learning', 'Evening Shiur')
			`, data.EventList3UUID)
			if err != nil {
				return nil, err
			}
		} else if err == sql.ErrNoRows {
			data.EventList3UUID = uuid.New()
			_, err = db.ExecContext(ctx, `
				INSERT INTO event_lists (event_list_uuid, venue_uuid, name, date, comment, visibility, private_link_token, sort_order, created_at, modified_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`, data.EventList3UUID, data.Venue3UUID, "Birthday", "2025-12-25",
				"Come as you are", "public", nil, 0, now, now)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `UPDATE event_lists SET visibility = 'public', private_link_token = NULL, modified_at = $1 WHERE event_list_uuid = $2`, now, data.EventList3UUID)
	if err != nil {
		return nil, err
	}

	// Event 1: Shacharis - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList1UUID, "Shacharis", "06:00:00").Scan(&data.Event1UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event1UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event1UUID, data.EventList1UUID, "Shacharis", nil, "06:00:00",
			"we start on time", 0, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 2: Mincha - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList1UUID, "Mincha", "16:30:00").Scan(&data.Event2UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event2UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event2UUID, data.EventList1UUID, "Mincha", nil, "16:30:00",
			"", 0, 1, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 3: Kabbalat Shabbat - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList2UUID, "Kabbalat Shabbat", "17:30:00").Scan(&data.Event3UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event3UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event3UUID, data.EventList2UUID, "Kabbalat Shabbat", nil, "17:30:00",
			"", 60, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 4: Shabbat Morning - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList2UUID, "Shabbat Morning", "09:00:00").Scan(&data.Event4UUID)
	if err == sql.ErrNoRows {
		// Event doesn't exist, insert it
		data.Event4UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event4UUID, data.EventList2UUID, "Shabbat Morning", nil, "09:00:00",
			"", 120, 1, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 5: Welcome (Ben's house Birthday) - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList3UUID, "Welcome", "19:00:00").Scan(&data.Event5UUID)
	if err == sql.ErrNoRows {
		data.Event5UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event5UUID, data.EventList3UUID, "Welcome", nil, "19:00:00",
			"There will be barbecue and cake", 0, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `UPDATE events SET comment = $1, modified_at = $2 WHERE event_uuid = $3`, "There will be barbecue and cake", now, data.Event5UUID)
	if err != nil {
		return nil, err
	}

	// Event List 5 for Venue 4 (After School Math): Daily Schedule - PUBLIC - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_list_uuid FROM event_lists WHERE venue_uuid = $1 AND name = $2
	`, data.Venue4UUID, "Daily Schedule").Scan(&data.EventList5UUID)
	if err == sql.ErrNoRows {
		data.EventList5UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO event_lists (event_list_uuid, venue_uuid, name, date, comment, visibility, private_link_token, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.EventList5UUID, data.Venue4UUID, "Daily Schedule", "2025-12-25",
			"Daily math classes", "public", nil, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `UPDATE event_lists SET visibility = 'public', private_link_token = NULL, modified_at = $1 WHERE event_list_uuid = $2`, now, data.EventList5UUID)
	if err != nil {
		return nil, err
	}

	// Event 10: Algebra - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList5UUID, "Algebra", "09:00:00").Scan(&data.Event10UUID)
	if err == sql.ErrNoRows {
		data.Event10UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event10UUID, data.EventList5UUID, "Algebra", nil, "09:00:00",
			"", 45, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 11: Calculus - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList5UUID, "Calculus", "10:00:00").Scan(&data.Event11UUID)
	if err == sql.ErrNoRows {
		data.Event11UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event11UUID, data.EventList5UUID, "Calculus", nil, "10:00:00",
			"", 45, 1, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 12: Lunch - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList5UUID, "Lunch", "11:00:00").Scan(&data.Event12UUID)
	if err == sql.ErrNoRows {
		data.Event12UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event12UUID, data.EventList5UUID, "Lunch", nil, "11:00:00",
			"", 50, 2, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 13: Fourier Transformations - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList5UUID, "Fourier Transformations", "12:00:00").Scan(&data.Event13UUID)
	if err == sql.ErrNoRows {
		data.Event13UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event13UUID, data.EventList5UUID, "Fourier Transformations", nil, "12:00:00",
			"", 45, 3, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 7: שחרית (Shacharit) - הדגמה: אוהל אברהם - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList4UUID, "שחרית", "06:30:00").Scan(&data.Event7UUID)
	if err == sql.ErrNoRows {
		data.Event7UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event7UUID, data.EventList4UUID, "שחרית", nil, "06:30:00",
			"", 0, 0, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 8: מנחה (Mincha) - הדגמה: אוהל אברהם - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList4UUID, "מנחה", "16:30:00").Scan(&data.Event8UUID)
	if err == sql.ErrNoRows {
		data.Event8UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event8UUID, data.EventList4UUID, "מנחה", nil, "16:30:00",
			"", 0, 1, now, now)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Event 9: מעריב (Ma'ariv) - הדגמה: אוהל אברהם - insert or get existing
	err = db.QueryRowContext(ctx, `
		SELECT event_uuid FROM events WHERE event_list_uuid = $1 AND event_name = $2 AND event_time = $3
	`, data.EventList4UUID, "מעריב", "18:00:00").Scan(&data.Event9UUID)
	if err == sql.ErrNoRows {
		data.Event9UUID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO events (event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, data.Event9UUID, data.EventList4UUID, "מעריב", nil, "18:00:00",
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
