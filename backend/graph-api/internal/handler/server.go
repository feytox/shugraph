package handler

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func NewAPI(router chi.Router, service noteService, pinger pinger) huma.API {
	api := humachi.New(router, huma.DefaultConfig("Graph API", "1.0.0"))
	healthHandler := NewHealthHandler(pinger)
	healthHandler.Registry(api)

	noteHandler := NewNoteHandler(service)
	noteHandler.Register(api)

	return api
}
