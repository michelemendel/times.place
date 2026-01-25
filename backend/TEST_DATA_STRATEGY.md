# Test Data Strategy - Summary

This document directly answers your questions about test data management.

## 1. Where should we handle insertion of test data?

### Answer: `backend/internal/testdata/seed.go`

**Location**: `backend/internal/testdata/seed.go`

This package provides:
- `SeedTestData(ctx, db)` - Inserts test data matching the frontend demo data
- `ClearTestData(ctx, db)` - Removes all test data
- `TestData` struct - Holds UUIDs of seeded records for use in tests

### Usage in Tests

```go
import (
    "github.com/michelemendel/times.place/internal/test"
)

func TestMyFeature(t *testing.T) {
    testDB := test.SetupTestDB(t)  // Automatically seeds test data
    defer testDB.Cleanup()
    
    // Access seeded data
    ownerUUID := testDB.TestData.Owner1UUID
    venueUUID := testDB.TestData.Venue1UUID
}
```

### Manual Seeding for Development

To manually seed test data for local development:

```bash
# Seed test data (adds to existing data)
make dbseed

# Clear existing data and seed fresh test data
make dbseedclear
```

Or run directly:
```bash
cd backend
go run ./cmd/cli/seed/main.go        # Seed data
go run ./cmd/cli/seed/main.go -clear # Clear and reseed
```

**Test credentials:**
- Owner 1: `abe@demo.org` / `demo`
- Owner 2: `ben@demo.org` / `demo`

### Production Safety

**✅ Safe**: Test data utilities are in `internal/` package (not imported by production code)

**✅ Safe**: Test helpers only seed data when explicitly called in tests

**✅ Safe**: No automatic seeding - must be explicitly invoked

**❌ Never**: Don't call `SeedTestData()` in production code paths

### Temporary for Production

The test data is **never** inserted in production:
- Test data seeding is only called from test code
- Production code never imports `internal/testdata` or `internal/test`
- No environment variables or flags enable test data in production

## 2. How to handle running tests against a live database

### Strategy: Use Transactions for Isolation

We run tests against a **live database** but use **database transactions** for isolation:

1. Each test starts a transaction
2. Test data is seeded within the transaction
3. Test runs and makes changes
4. Transaction is **rolled back** after test completes
5. Database returns to clean state

### Local Development

**Setup:**
```bash
# 1. Start devcontainer (database included)
make devcontainerup

# 2. Run migrations
make dbgooseup

# 3. Run tests
make btest
```

**Connection:**
- Uses `DATABASE_URL` environment variable (if set)
- Defaults to devcontainer connection: `postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable`
- Tests run inside the devcontainer or with `DATABASE_URL` set

**Isolation:**
- Each test uses its own transaction
- Automatic rollback after test completes
- No manual cleanup required
- Tests can run in parallel (separate transactions)

### CI/CD

**Strategy:** Use PostgreSQL service container (same as local)

**GitHub Actions Example:**
```yaml
services:
  postgres:
    image: postgres:16-alpine
    env:
      POSTGRES_USER: timesplace
      POSTGRES_PASSWORD: timesplace
      POSTGRES_DB: timesplace
    ports:
      - 5432:5432

steps:
  - name: Run migrations
    env:
      DATABASE_URL: postgres://timesplace:timesplace@localhost:5432/timesplace?sslmode=disable
    run: |
      cd backend
      goose -dir db/migrations postgres "$DATABASE_URL" up
  
  - name: Run tests
    env:
      DATABASE_URL: postgres://timesplace:timesplace@localhost:5432/timesplace?sslmode=disable
    run: |
      cd backend
      go test ./... -v
```

**Key Points:**
- ✅ Same database structure as production
- ✅ Transaction isolation (no test interference)
- ✅ No mocks or test doubles needed
- ✅ Fast execution (rollback is quick)
- ✅ Real database behavior and constraints

### Alternative: Dedicated Test Database

If you prefer a separate test database that gets reset:

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

**Note:** The test helpers support `TEST_DATABASE_URL` environment variable for this use case.

## Summary

### Test Data Location
- **Backend**: `backend/internal/testdata/seed.go`
- **Frontend**: `frontend/src/lib/demo_data.ts` (already exists)
- **Usage**: Only in tests, never in production

### Database Testing
- **Local**: Run against devcontainer database with transaction isolation
- **CI/CD**: Run against PostgreSQL service container with transaction isolation
- **Isolation**: Database transactions (automatic rollback)
- **No mocks**: Real database for realistic testing

### Files Created
1. `backend/internal/testdata/seed.go` - Test data seeding
2. `backend/internal/test/helpers.go` - Test database helpers
3. `backend/TESTING.md` - Complete testing guide
4. `backend/TEST_DATA_STRATEGY.md` - This summary (answers your questions)
5. `.github/workflows/backend-tests.yml.example` - CI/CD example

### Next Steps
1. Install Go dependencies: `cd backend && go mod tidy` (inside devcontainer)
2. Run tests: `make btest`
3. See `TESTING.md` for detailed documentation
