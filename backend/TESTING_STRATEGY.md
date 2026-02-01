# Testing Strategy

This document describes the backend testing strategy: test data management, database testing approach, and how to run tests locally and in CI.

---

## 1. Test Data

### Where test data is defined and inserted

**Location:** `backend/internal/testdata/seed.go`

This package provides:

- `SeedTestData(ctx, db)` – inserts test data matching the frontend demo data
- `ClearTestData(ctx, db)` – removes all test data
- `TestData` struct – holds UUIDs of seeded records for use in tests

### When to use test data

| Context              | Use test data? |
|----------------------|----------------|
| **Integration tests**| Yes – use helpers that seed automatically |
| **Local development**| Optional – seed manually for dev |
| **Production**       | Never – do not seed test data in production |

### Test data shape

Seeded data aligns with the frontend demo:

- **Owner 1 (Abe):** 2 venues (Beth El Synagogue, Community Center)
- **Owner 2 (Ben):** 2 venues (Ben's house, After School Math)
- Event lists and events under those venues
- All test owners use password `"demo"` (bcrypt hashed)

**Test credentials:**

- Owner 1: `abe@demo.org` / `demo`
- Owner 2: `ben@demo.org` / `demo`

### Production safety

- Test data lives in `internal/testdata` and is not imported by production code.
- Seeding only runs when explicitly called (tests or CLI).
- Do not call `SeedTestData()` (or similar) from production code paths.

---

## 2. Database Testing Strategy

### Live database, no mocks

Tests run against a **real database** (not mocks). That gives:

- Real SQL, constraints, and foreign keys
- Migration compatibility checks
- Realistic behavior

### Isolation via transactions

We use **database transactions** so tests don’t leave data behind:

1. Each test starts a transaction.
2. Test data is seeded inside the transaction.
3. The test runs and may change data.
4. The transaction is **rolled back** when the test finishes.
5. The database is left clean.

Effects:

- Tests don’t affect each other.
- No manual cleanup.
- Rollback is fast.
- Safe to run tests in parallel (each has its own transaction).

### Test helpers

**Package:** `backend/internal/test/helpers.go`

- `SetupTestDB(t *testing.T)` – connection with a transaction and seeded data; use `defer testDB.Cleanup()`.
- `SetupTestDBWithoutTransaction(t *testing.T)` – connection without a transaction (only when you need to test transaction behavior).
- `RequireDatabase(t *testing.T)` – skips the test if the database is unavailable (e.g. in environments without DB).

### Example test

```go
package mypackage_test

import (
    "testing"
    "github.com/michelemendel/times.place/internal/test"
)

func TestMyFeature(t *testing.T) {
    testDB := test.SetupTestDB(t)
    defer testDB.Cleanup()

    // testDB.Tx for DB operations; all changes rolled back after test
    ownerUUID := testDB.TestData.Owner1UUID
    // ... test code
}
```

---

## 3. Local Testing

### Prerequisites

1. Devcontainer (or Postgres) running.
2. Migrations applied: `make dbup`.
3. `DATABASE_URL` set if not using devcontainer defaults.

### Manual seeding (development only)

```bash
# Add test data (on top of existing data)
make dbseed

# Clear demo data only (does not seed; run make dbseed to re-seed)
make dbseedclear
```

Or via CLI:

```bash
cd backend
go run ./cmd/cli/seed/main.go              # seed only
go run ./cmd/cli/seed/main.go -clear-demo-only  # clear demo only (no seed)
```

### Running tests

```bash
# All tests
make btest
# or
cd backend && go test ./...

# Single package, verbose
go test ./internal/test -v

# With coverage
make btestcover
# or
go test ./... -cover
```

### Database connection for tests

Tests use, in order:

1. `TEST_DATABASE_URL` (if set)
2. `DATABASE_URL` (if set)
3. Devcontainer default: `postgres://timesplace:timesplace@postgres:5432/timesplace?sslmode=disable`

---

## 4. CI/CD Testing

### Approach

Same as local: real Postgres, migrations, transaction-isolated tests. No mocks.

### Workflow

The repo uses `.github/workflows/ci.yml`, which:

- Starts a PostgreSQL service container.
- Runs migrations.
- Runs backend tests with `DATABASE_URL` and `TEST_DATABASE_URL` set to the same DB (transaction rollback keeps it clean).
- Optionally uploads coverage (e.g. Codecov).

See `.github/workflows/ci.yml` for the full workflow.

### Optional: dedicated test database

If you want a separate DB that is recreated each run:

```yaml
- name: Setup test database
  run: |
    psql "$DATABASE_URL" -c "DROP DATABASE IF EXISTS timesplace_test;"
    psql "$DATABASE_URL" -c "CREATE DATABASE timesplace_test;"
    goose -dir backend/db/migrations postgres "postgres://timesplace:timesplace@localhost:5432/timesplace_test?sslmode=disable" up

- name: Run tests
  env:
    TEST_DATABASE_URL: postgres://timesplace:timesplace@localhost:5432/timesplace_test?sslmode=disable
  run: cd backend && go test ./... -v
```

Test helpers support `TEST_DATABASE_URL` for this.

---

## 5. Best Practices

1. **Use transaction-backed helpers in tests**  
   Prefer `test.SetupTestDB(t)` and `defer testDB.Cleanup()` instead of opening a raw connection and seeding manually.

2. **Use seeded test data**  
   Use `testDB.TestData` (e.g. `Owner1UUID`) instead of creating new UUIDs and inserting one-off data in every test.

3. **Never seed in production**  
   Do not call test data seeding from production code or based on production env vars.

4. **Skip when DB is unavailable**  
   In integration tests, call `test.RequireDatabase(t)` so tests are skipped (e.g. in environments without a DB) instead of failing.

5. **Use TEST_DATABASE_URL in CI if needed**  
   You can point tests at a dedicated test DB via `TEST_DATABASE_URL` while keeping `DATABASE_URL` for migrations.

---

## 6. Troubleshooting

| Problem | What to check |
|--------|----------------|
| **"connection refused"** | Devcontainer/Postgres running (`make devcontainerup`), DB healthy (`make dbverify`), `DATABASE_URL` correct. |
| **Tests affect each other** | Use `SetupTestDB()` and `defer testDB.Cleanup()`. Avoid `SetupTestDBWithoutTransaction()` unless you need it. |
| **Test data not found** | Migrations applied (`make dbup`), seeding ran (e.g. via `SetupTestDB`), correct DB in use. |
| **Slow tests** | Check indexes (`make dbverify`), size of seeded data, and that you’re using a local DB. |

---

## Summary

| Topic | Summary |
|-------|--------|
| **Test data** | `backend/internal/testdata/seed.go`; use in tests only, never in production. |
| **Isolation** | Transactions (rollback after each test) via `test.SetupTestDB(t)`. |
| **Local** | Devcontainer DB, `make dbup` then `make btest` / `make btestcover`. |
| **CI** | Postgres service in workflow; same strategy as local (see `.github/workflows/ci.yml`). |
| **Production** | Do not seed or depend on test data in production. |
