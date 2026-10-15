package handler

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
)

func NewAPI(router chi.Router, service noteService, pinger pinger) huma.API {
	router.Use(middleware.Recoverer, middleware.RequestID)

	router.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{}))

	api := humachi.New(router, huma.DefaultConfig("Graph API", "1.0.0"))
	healthHandler := NewHealthHandler(pinger)
	healthHandler.Register(api)

	noteHandler := NewNoteHandler(service)
	noteHandler.Register(api)

	return api
}
