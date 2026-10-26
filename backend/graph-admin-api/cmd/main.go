package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/app"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/logattr"
)

var version = "dev"

func main() {
	log, err := app.NewLogger(version)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}

	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, log, os.Args[1:]); err != nil {
		log.Error("command failed", logattr.Error(err))
		stop()
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger, args []string) error {
	command := "serve"
	if len(args) > 0 {
		command = args[0]
	}

	switch command {
	case "serve":
		return app.Serve(ctx, log)
	case "migrate":
		return app.Migrate(ctx, log)
	case "openapi":
		return app.WriteOpenAPI(os.Stdout)
	case "healthcheck":
		return app.Healthcheck(ctx)
	default:
		return fmt.Errorf("unknown command %q, expected serve, migrate, openapi or healthcheck", command)
	}
}
