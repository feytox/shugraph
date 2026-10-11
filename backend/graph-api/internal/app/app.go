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
	"github.com/feytox/shugraph/backend/graph-api/internal/handler"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository"
	"github.com/feytox/shugraph/backend/graph-api/internal/service"
	"github.com/go-chi/chi/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Serve(ctx context.Context, cfg *Config) error {
	pool, err := ConnectToPg(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := repository.NewPgNoteRepository(pool)
	noteService := service.NewNoteService(repo)
	router := chi.NewMux()
	_ = handler.NewAPI(router, noteService, pool)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router); err != nil {
		return err
	}

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

func Migrate(ctx context.Context, cfg *Config) error {
	pool, err := ConnectToPg(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	goose.SetBaseFS(migrations.FS)
	gooseDB := stdlib.OpenDBFromPool(pool)
	defer func(gooseDB *sql.DB) {
		err := gooseDB.Close()
		if err != nil {
			slog.Error("error closing goose db", "err", err.Error())
			return
		}
	}(gooseDB)

	if err := goose.UpContext(ctx, gooseDB, "."); err != nil {

		return err
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

func ConnectToPg(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DB.URL)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
