# Testing Guide

This document describes the testing strategy for the backend, including test data management and database testing practices.

## Test Data Management

### Overview

Test data is managed through the `backend/internal/testdata` package, which provides utilities to seed and clear test data. This mirrors the frontend demo data structure (`frontend/src/lib/demo_data.ts`) but works with the actual database.

### Test Data Location

**Backend test data**: `backend/internal/testdata/seed.go`

This package provides:
- `SeedTestData()` - Inserts test data into the database
- `ClearTestData()` - Removes all test data from the database
- `TestData` struct - Holds UUIDs of seeded records for use in tests

### When to Use Test Data

1. **In Tests**: Always use test data seeding in integration tests
2. **Local Development**: Optional - you can seed test data manually for development
3. **Production**: NEVER - Test data should never be seeded in production

### Test Data Structure

The seeded data matches the frontend demo data:
- **Owner 1 (Abe)**: 2 venues (Beth El Synagogue, Community Center)
- **Owner 2 (Ben)**: 2 venues (Beit Midrash, Chagat House)
- Various event lists and events
- All test owners have password: `"demo"` (bcrypt hashed)

## Database Testing Strategy

### Running Tests Against Live Database

We run tests against a **live database** (not mocks). This provides:
- Real database behavior and constraints
- Actual SQL query validation
- Foreign key and constraint testing
- Migration compatibility verification

### Test Isolation

Tests use **database transactions** for isolation:

1. Each test starts a transaction
2. Test data is seeded within the transaction
3. Test runs and makes changes
4. Transaction is rolled back after test completes
5. Database returns to clean state

This ensures:
- Tests don't interfere with each other
- No manual cleanup required
- Fast test execution (rollback is quick)
- Tests can run in parallel (if using separate transactions)

### Test Helpers

The `backend/internal/test/helpers.go` package provides:

- `SetupTestDB(t *testing.T)` - Creates a test DB connection with transaction isolation
- `SetupTestDBWithoutTransaction(t *testing.T)` - Creates a connection without transaction (for testing transaction behavior)
- `RequireDatabase(t *testing.T)` - Skips test if database is unavailable

## Local Testing

### Prerequisites

1. **Devcontainer running**: Database must be available
2. **Migrations applied**: Run `make dbgooseup` before tests
3. **Environment variables**: `DATABASE_URL` should be set (or use defaults)

### Seeding Test Data Manually

For local development, you can manually seed test data:

```bash
# Seed test data (adds to existing data)
make dbseed

# Clear existing data and seed fresh test data
make dbseedclear
```

Or run directly:

```bash
# From inside devcontainer
cd backend
go run ./cmd/cli/seed/main.go

# Clear and reseed
go run ./cmd/cli/seed/main.go -clear
```

**Test credentials:**
- Owner 1: `abe@demo.org` / `demo`
- Owner 2: `ben@demo.org` / `demo`

### Running Tests

```bash
# From inside devcontainer or with DATABASE_URL set
cd backend
go test ./...

# Run specific test
go test ./internal/test -v

# Run with coverage
go test ./... -cover
```

### Test Database Connection

Tests automatically use:
1. `DATABASE_URL` environment variable (if set)
2. `TEST_DATABASE_URL` environment variable (if set)
3. Default devcontainer connection: `postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable`

### Example Test

```go
package mypackage_test

import (
    "testing"
    "github.com/michelemendel/times.place/internal/test"
)

func TestMyFeature(t *testing.T) {
    testDB := test.SetupTestDB(t)
    defer testDB.Cleanup()

    // Use testDB.Tx for database operations
    // All changes will be rolled back after the test
    
    // Access seeded test data
    ownerUUID := testDB.TestData.Owner1UUID
    
    // Your test code here
}
```

## CI/CD Testing

### Strategy

In CI/CD, we use the **same approach** as local testing:
- Run tests against a live database
- Use transactions for test isolation
- No mocks or test databases required

### GitHub Actions Example

See `.github/workflows/backend-tests.yml.example` for a complete example workflow.

**Quick setup:**

1. Copy the example workflow:
   ```bash
   cp .github/workflows/backend-tests.yml.example .github/workflows/backend-tests.yml
   ```

2. The workflow will:
   - Start a PostgreSQL service container
   - Run database migrations
   - Execute all tests
   - Generate coverage reports

**Key points:**
- Uses PostgreSQL service container (no external database needed)
- Tests run against the same database structure as production
- Transaction isolation ensures tests don't interfere with each other
- Coverage reports can be uploaded to codecov or similar services

### Alternative: Dedicated Test Database

If you prefer a separate test database that gets reset between runs:

```yaml
- name: Setup test database
  run: |
    psql "$DATABASE_URL" -c "DROP DATABASE IF EXISTS timesplace_test;"
    psql "$DATABASE_URL" -c "CREATE DATABASE timesplace_test;"
    TEST_DATABASE_URL="postgres://timesplace:timesplace@localhost:5432/timesplace_test?sslmode=disable"
    goose -dir backend/db/migrations postgres "$TEST_DATABASE_URL" up

- name: Run tests
  env:
    TEST_DATABASE_URL: postgres://timesplace:timesplace@localhost:5432/timesplace_test?sslmode=disable
  run: |
    cd backend
    go test ./... -v
```

## Best Practices

### 1. Always Use Transactions in Tests

```go
// ✅ Good: Uses transaction isolation
testDB := test.SetupTestDB(t)
defer testDB.Cleanup()

// ❌ Bad: No isolation, requires manual cleanup
db := sql.Open(...)
testdata.SeedTestData(ctx, db)
// Must manually clear data
```

### 2. Use Seeded Test Data

```go
// ✅ Good: Uses pre-seeded test data
ownerUUID := testDB.TestData.Owner1UUID

// ❌ Bad: Creates new test data in every test
ownerUUID := uuid.New()
// Insert owner...
```

### 3. Don't Seed in Production

```go
// ❌ Never do this in production code
if os.Getenv("ENV") == "production" {
    testdata.SeedTestData(ctx, db) // NO!
}
```

### 4. Skip Tests if Database Unavailable

```go
func TestIntegration(t *testing.T) {
    test.RequireDatabase(t) // Skips if DB unavailable
    
    testDB := test.SetupTestDB(t)
    defer testDB.Cleanup()
    // ...
}
```

### 5. Use Test-Specific Environment Variables

For CI/CD, you can use `TEST_DATABASE_URL` to point to a separate test database if desired, while keeping `DATABASE_URL` for the main database.

## Troubleshooting

### Tests Fail with "connection refused"

- Ensure devcontainer is running: `make devcontainerup`
- Check database is healthy: `make dbverify`
- Verify `DATABASE_URL` is set correctly

### Tests Interfere with Each Other

- Ensure you're using `SetupTestDB()` which provides transaction isolation
- Don't use `SetupTestDBWithoutTransaction()` unless testing transaction behavior
- Check that `defer testDB.Cleanup()` is called

### Test Data Not Found

- Ensure migrations are applied: `make dbgooseup`
- Check that `SeedTestData()` completed successfully
- Verify you're using the correct database connection

### Slow Tests

- Transaction rollbacks should be fast
- If tests are slow, check for:
  - Missing indexes (run `make dbverify`)
  - Large test datasets (reduce if possible)
  - Network latency (use local database)

## Summary

- **Test data**: Managed in `backend/internal/testdata/seed.go`
- **Test isolation**: Use transactions (automatic rollback)
- **Local testing**: Run against devcontainer database
- **CI/CD testing**: Run against service database (PostgreSQL service)
- **Production**: Never seed test data in production code
