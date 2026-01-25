package test

import (
	"testing"

	"github.com/google/uuid"
)

// Example test demonstrating the test helpers
// This test will be skipped if the database is not available
func TestExampleUsage(t *testing.T) {
	// Skip if database is not available (useful for CI/CD)
	RequireDatabase(t)

	// Setup test database with transaction isolation
	testDB := SetupTestDB(t)
	defer testDB.Cleanup() // Always cleanup (rolls back transaction)

	// Access seeded test data
	ownerUUID := testDB.TestData.Owner1UUID
	if ownerUUID == (uuid.UUID{}) {
		t.Error("Expected Owner1UUID to be set")
	}

	// Use testDB.Tx for database queries
	// All changes will be automatically rolled back after the test
	var count int
	err := testDB.Tx.QueryRow("SELECT COUNT(*) FROM venue_owners").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 venue owners, got %d", count)
	}

	// You can make changes in the transaction - they'll be rolled back
	_, err = testDB.Tx.Exec("INSERT INTO venue_owners (owner_uuid, name, mobile, email, password_hash) VALUES (gen_random_uuid(), 'Test', '123', 'test@test.com', 'hash')")
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify the change is visible within the transaction
	err = testDB.Tx.QueryRow("SELECT COUNT(*) FROM venue_owners").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 venue owners after insert, got %d", count)
	}

	// After testDB.Cleanup() is called, the transaction is rolled back
	// and the database returns to its original state
}
