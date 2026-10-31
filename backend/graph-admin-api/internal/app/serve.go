package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/config"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/graphapi"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/handler"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/logattr"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/postgres"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/repository"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/service"
)

func Serve(ctx context.Context, log *slog.Logger) error {
	cfg, err := config.Load[config.Serve]()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	srv, err := newServer(ctx, cfg, log)
	if err != nil {
		return err
	}
	defer srv.close()

	return srv.run(ctx)
}

type server struct {
	log             *slog.Logger
	pool            *pgxpool.Pool
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

func newServer(ctx context.Context, cfg config.Serve, log *slog.Logger) (*server, error) {
	pool, err := postgres.Connect(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	graph, err := graphapi.New(cfg.GraphAPI.URL, &http.Client{Timeout: cfg.GraphAPI.Timeout})
	if err != nil {
		pool.Close()
		return nil, err
	}

	notes := service.NewNoteService(graph, repository.NewPgNoteRepository(pool))

	httpServer := &http.Server{
		Addr: net.JoinHostPort("", strconv.Itoa(cfg.HTTP.Port)),
		Handler: handler.New(handler.Deps{
			Notes:    notes,
			Database: pool,
			Log:      log,
		}),
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	return &server{
		log:             log,
		pool:            pool,
		httpServer:      httpServer,
		shutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, nil
}

func (s *server) run(ctx context.Context) error {
	serveErr := make(chan error, 1)

	go func() {
		s.log.Info("http server listening", "addr", s.httpServer.Addr)

		serveErr <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
	case <-ctx.Done():
		s.log.Info("shutdown signal received")
	}

	s.shutdown()

	return nil
}

func (s *server) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("http server shutdown failed", logattr.Error(err))
	}

	s.log.Info("shutdown complete")
}

func (s *server) close() {
	s.pool.Close()
}
