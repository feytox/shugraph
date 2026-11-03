package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/config"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/postgres"
)

func Migrate(ctx context.Context, log *slog.Logger) error {
	cfg, err := config.Load[config.Migrate]()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pool, err := postgres.Connect(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()

	return postgres.Migrate(ctx, pool, log)
}
