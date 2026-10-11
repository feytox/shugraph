package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/feytox/shugraph/backend/graph-api/db/migrations"
	"github.com/feytox/shugraph/backend/graph-api/internal/handler"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository"
	"github.com/feytox/shugraph/backend/graph-api/internal/service"

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
	server := handler.NewServer(noteService, pool)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), server.Router); err != nil {
		return err
	}

	return nil
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

func ConnectToPg(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DB.URL)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
