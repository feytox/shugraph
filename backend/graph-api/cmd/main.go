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
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	command := "serve"
	if len(args) > 0 {
		command = args[0]
	}
	switch command {
	case "serve":
		cnf, err := app.LoadConfig()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if err := app.Serve(ctx, cnf); err != nil {
			return fmt.Errorf("serve: %w", err)
		}
	case "migrate":
		cnf, err := app.LoadConfig()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		slog.Info("migrate started")
		if err := app.Migrate(ctx, cnf); err != nil {
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
