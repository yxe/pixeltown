package db

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	Pool *pgxpool.Pool
}

func Open(ctx context.Context) (*DB, error) {
	url := os.Getenv("DATABASE_URL")

	if url == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	cfg, err := pgxpool.ParseConfig(url)

	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	// connection caps
	cfg.MaxConns = 25
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 15 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	d := &DB{Pool: pool}

	if err := d.migrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return d, nil
}

func (d *DB) migrate(ctx context.Context) error {
	sqlDB := stdlib.OpenDBFromPool(d.Pool)
	defer sqlDB.Close()

	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	goose.SetLogger(goose.NopLogger())
	return goose.UpContext(ctx, sqlDB, "migrations")
}

func (d *DB) Close() {
	d.Pool.Close()
}
