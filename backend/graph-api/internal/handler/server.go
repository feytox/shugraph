package handler

import (
	"graph-api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	Router chi.Router
}

func NewServer(service *service.NoteService, pool *pgxpool.Pool) *Server {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Graph API", "1.0.0"))

	healthHandler := NewHealthHandler(pool)
	healthHandler.Registry(api)

	noteHandler := NewNoteHandler(service)
	noteHandler.Register(api)

	server := &Server{
		Router: router,
	}
	return server
}
