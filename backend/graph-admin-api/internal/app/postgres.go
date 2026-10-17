package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"github.com/feytox/shugraph/backend/graph-admin-api/db/migrations"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/config"
)

func connectPostgres(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, log *slog.Logger) error {
	cfg, err := config.Load[config.Migrate]()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pool, err := connectPostgres(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()

	return migrate(ctx, pool, log)
}

func migrate(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("create migration lock: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)

	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS, goose.WithSessionLocker(locker))
	if err != nil {
		_ = sqlDB.Close()
		return fmt.Errorf("create migration provider: %w", err)
	}
	defer func() { _ = provider.Close() }()

	results, err := provider.Up(ctx)
	for _, r := range results {
		log.Info("migration applied", "migration.version", r.Source.Version, "migration.file", r.Source.Path, "duration_ms", r.Duration.Milliseconds())
	}

	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	if len(results) == 0 {
		log.Info("database schema is up to date")
	}

	return nil
}
