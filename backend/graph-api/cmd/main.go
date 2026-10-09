package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"graph-api/internal/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	command := "serve"
	args := os.Args[1:]
	if len(args) > 0 {
		command = args[0]
	}
	cnf, err := app.LoadCnf()
	if err != nil {
		slog.Error("failed load", "error", err)
		stop()
		os.Exit(1)
	}
	switch command {
	case "serve":
		if err := app.Serve(ctx, cnf); err != nil {
			slog.Error("command failed", "error", err)
			stop()
			os.Exit(1)
		}
	case "migrate":
		slog.Info("migrate started")
		err := app.Migrate(ctx, cnf)
		if err != nil {
			slog.Error("migration failed", "error", err)
			stop()
			os.Exit(1)
		}
		os.Exit(0)
	default:
		slog.Error("unknown command expected serve or migrate", "command", command)
		stop()
		os.Exit(1)
	}
}
