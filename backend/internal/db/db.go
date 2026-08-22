package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver for database/sql
	"github.com/pressly/goose/v3"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Connect creates a pgx connection pool and runs migrations.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pool.Ping: %w", err)
	}

	if err := runMigrations(databaseURL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	if err := runRiverMigrations(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("river migrations: %w", err)
	}

	return pool, nil
}

func runMigrations(databaseURL string) error {
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(migrationsFS)
	goose.SetTableName("goose_migrations")

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("goose.Up: %w", err)
	}

	return nil
}

func runRiverMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrator, err := rivermigrate.New[pgx.Tx](riverpgxv5.New(pool), nil)
	if err != nil {
		return fmt.Errorf("rivermigrate.New: %w", err)
	}
	res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return fmt.Errorf("rivermigrate.Migrate: %w", err)
	}
	for _, m := range res.Versions {
		log.Printf("River migration %d applied", m.Version)
	}
	return nil
}
