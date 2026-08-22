package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetConfig returns the value for a given config key.
func GetConfig(ctx context.Context, pool *pgxpool.Pool, key string) (string, error) {
	var value string
	err := pool.QueryRow(ctx, `SELECT value FROM config WHERE key = $1`, key).Scan(&value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("GetConfig: %w", err)
	}
	return value, nil
}

// SetConfig sets the value for a given config key (upsert).
func SetConfig(ctx context.Context, pool *pgxpool.Pool, key, value string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO config (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value = $2`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("SetConfig: %w", err)
	}
	return nil
}

// GetFetcherMode returns the current fetcher mode. Defaults to "simple".
func GetFetcherMode(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	mode, err := GetConfig(ctx, pool, "fetcher_mode")
	if err != nil {
		return "", err
	}
	if mode == "" {
		return "simple", nil
	}
	return mode, nil
}
