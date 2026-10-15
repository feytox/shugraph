package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	apiTitle   = "Graph Admin API"
	apiVersion = "1.0.0"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Deps struct {
	Database Pinger
	Log      *slog.Logger
}

func New(deps Deps) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID, accessLog(deps.Log), recoverPanics(deps.Log))
	registerAPI(router, deps)

	return router
}

func OpenAPISpec() ([]byte, error) {
	api := registerAPI(chi.NewRouter(), Deps{Log: slog.New(slog.DiscardHandler)})
	return api.OpenAPI().YAML()
}

func registerAPI(router chi.Router, deps Deps) huma.API {
	config := huma.DefaultConfig(apiTitle, apiVersion)
	config.CreateHooks = nil

	api := humachi.New(router, config)
	registerHealth(api, &healthHandlers{database: deps.Database, log: deps.Log})

	return api
}
