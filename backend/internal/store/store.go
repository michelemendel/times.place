package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/michelemendel/times.place/db/sqlc"
)

// Store wraps the database connection and sqlc queries
type Store struct {
	db      *pgxpool.Pool
	Queries *sqlc.Queries
}

// NewStore creates a new store with a database connection
func NewStore(dbURL string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	queries := sqlc.New(pool)

	return &Store{
		db:      pool,
		Queries: queries,
	}, nil
}

// Close closes the database connection pool
func (s *Store) Close() {
	s.db.Close()
}

// DB returns the underlying database pool (for transactions if needed)
func (s *Store) DB() *pgxpool.Pool {
	return s.db
}
