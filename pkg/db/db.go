// Package db centralizes PostgreSQL connection setup for the services.
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pgx pool and waits (up to ~30s) for the database to accept
// connections, so a service can start slightly before Postgres is ready.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for {
		if err = pool.Ping(waitCtx); err == nil {
			return pool, nil
		}
		select {
		case <-waitCtx.Done():
			pool.Close()
			return nil, err
		case <-time.After(time.Second):
		}
	}
}

// RowScanner is satisfied by both pgx.Row and pgx.Rows, so a single scan helper
// can serve QueryRow and Query results.
type RowScanner interface {
	Scan(dest ...any) error
}

// Nullable maps an empty string to nil for nullable text columns.
func Nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
