package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/feytox/shugraph/backend/graph-api/db/migrations"
	"github.com/feytox/shugraph/backend/graph-api/internal/config"
	"github.com/feytox/shugraph/backend/graph-api/internal/handler"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository"
	"github.com/feytox/shugraph/backend/graph-api/internal/service"
	"github.com/go-chi/chi/v5"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const readinessTimeout = 2 * time.Second

func Serve(ctx context.Context) error {
	cfg, err := config.Load[config.Serve]()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pool, err := ConnectToPg(ctx, cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()

	repo := repository.NewPgNoteRepository(pool)
	noteService := service.NewNoteService(repo)
	router := chi.NewMux()
	_ = handler.NewAPI(router, noteService, pool)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func Migrate(ctx context.Context) error {
	cfg, err := config.Load[config.DBConfig]()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pool, err := ConnectToPg(ctx, cfg.URL)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()
	gooseDB := stdlib.OpenDBFromPool(pool)
	provider, err := goose.NewProvider(goose.DialectPostgres, gooseDB, migrations.FS)
	if err != nil {
		return fmt.Errorf("goose.NewProvider: %w", err)
	}

	defer func(gooseDB *sql.DB) {
		err := gooseDB.Close()
		if err != nil {
			slog.Error("error closing goose db", "err", err.Error())
			return
		}
	}(gooseDB)

	upResult, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("provider.Up: %w", err)
	}

	for _, migrateRes := range upResult {
		slog.Info("migration up", "up", migrateRes)
	}
	if len(upResult) == 0 {
		slog.Info("database schema is up to date")
	}

	return nil
}

func GenerateOpenAPI() error {
	r := chi.NewRouter()
	api := handler.NewAPI(r, nil, nil)
	b, err := api.OpenAPI().YAML()
	if err != nil {
		return fmt.Errorf("error generating open api: %w", err)
	}
	_, err = os.Stdout.Write(b)
	if err != nil {
		return fmt.Errorf("error stdout write api: %w", err)
	}
	return nil
}
