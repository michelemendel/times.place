package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/michelemendel/times.place/internal/testdata"
)

// TestDB wraps a database connection and transaction for test isolation
type TestDB struct {
	DB        *sql.DB
	Tx        *sql.Tx
	Ctx       context.Context
	TestData  *testdata.TestData
	Cleanup   func() // Call this to rollback and cleanup
}

// SetupTestDB creates a test database connection and starts a transaction.
// The transaction is rolled back after the test completes, ensuring test isolation.
//
// Usage:
//
//	func TestMyFeature(t *testing.T) {
//		testDB := SetupTestDB(t)
//		defer testDB.Cleanup()
//
//		// Use testDB.DB or testDB.Tx for database operations
//		// All changes will be rolled back after the test
//	}
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default to devcontainer connection
		dbURL = "postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("Failed to ping database: %v", err)
	}

	ctx := context.Background()

	// Start a transaction for test isolation
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to start transaction: %v", err)
	}

	// Seed test data
	testData, err := testdata.SeedTestData(ctx, tx)
	if err != nil {
		tx.Rollback()
		db.Close()
		t.Fatalf("Failed to seed test data: %v", err)
	}

	cleanup := func() {
		if err := tx.Rollback(); err != nil {
			t.Logf("Warning: failed to rollback transaction: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}

	return &TestDB{
		DB:       db,
		Tx:       tx,
		Ctx:      ctx,
		TestData: testData,
		Cleanup:  cleanup,
	}
}

// SetupTestDBWithoutTransaction creates a test database connection WITHOUT a transaction.
// Use this when you need to test transaction behavior or when using transactions
// would interfere with your test (e.g., testing transaction rollbacks).
//
// WARNING: This does NOT provide automatic cleanup. You must manually clean up test data
// or use a separate test database that gets reset between test runs.
//
// Usage:
//
//	func TestTransactionBehavior(t *testing.T) {
//		testDB := SetupTestDBWithoutTransaction(t)
//		defer testDB.Cleanup()
//
//		// Manually seed data if needed
//		testData, err := testdata.SeedTestData(testDB.Ctx, testDB.DB)
//		if err != nil {
//			t.Fatalf("Failed to seed test data: %v", err)
//		}
//		defer testdata.ClearTestData(testDB.Ctx, testDB.DB)
//
//		// Test transaction behavior
//	}
func SetupTestDBWithoutTransaction(t *testing.T) *TestDB {
	t.Helper()

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default to devcontainer connection
		dbURL = "postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("Failed to ping database: %v", err)
	}

	ctx := context.Background()

	cleanup := func() {
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}

	return &TestDB{
		DB:      db,
		Tx:      nil, // No transaction
		Ctx:     ctx,
		Cleanup: cleanup,
	}
}

// GetDatabaseURL returns the database connection URL for testing.
// It checks environment variables and provides sensible defaults.
func GetDatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	// Default to devcontainer connection
	return "postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable"
}

// RequireDatabase skips the test if DATABASE_URL is not set and we can't connect.
// Use this for integration tests that require a live database.
func RequireDatabase(t *testing.T) {
	t.Helper()

	dbURL := GetDatabaseURL()
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot open database connection: %v", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
		return
	}
}

// Example test function showing how to use the test helpers
func ExampleTestUsage(t *testing.T) {
	// This is just an example - remove or rename if you want to use it
	testDB := SetupTestDB(t)
	defer testDB.Cleanup()

	// Use testDB.Tx for queries (all changes will be rolled back)
	var count int
	err := testDB.Tx.QueryRow("SELECT COUNT(*) FROM venue_owners").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Access seeded test data
	ownerUUID := testDB.TestData.Owner1UUID
	fmt.Printf("Test owner UUID: %s\n", ownerUUID)
}
