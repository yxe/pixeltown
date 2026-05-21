// Package store holds typed pgx-backed helpers for each table.
// Handlers call Store methods instead of writing SQL inline.
package store

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store carries the postgres pool every helper method uses.
type Store struct {
	pool *pgxpool.Pool
}

// New returns a Store backed by pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}
