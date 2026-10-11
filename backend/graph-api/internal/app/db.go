package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToPg(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DB.URL)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, readinessTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return pool, nil
}
