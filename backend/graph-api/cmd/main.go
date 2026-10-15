package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/feytox/shugraph/backend/graph-api/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Println("command failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	command := "serve"
	if len(args) > 0 {
		command = args[0]
	}

	logger, err := app.NewLogger()
	if err != nil {
		return fmt.Errorf("init default logger: %w", err)
	}
	slog.SetDefault(logger)
	switch command {
	case "serve":
		if err := app.Serve(ctx); err != nil {
			return fmt.Errorf("serve: %w", err)
		}
	case "migrate":
		slog.Info("migrate started")
		if err := app.Migrate(ctx); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	case "openapi":
		if err := app.GenerateOpenAPI(); err != nil {
			return fmt.Errorf("openapi: %w", err)
		}
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
	return nil
}
