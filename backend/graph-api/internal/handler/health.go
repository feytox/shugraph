package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(repo *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: repo}
}

func (h *HealthHandler) Registry(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "health",
		Summary:     "Checks the application's availability",
		Method:      http.MethodGet,
		Path:        "/healthz",
	}, h.Healthz)

	huma.Register(api, huma.Operation{
		OperationID: "ready",
		Summary:     "Checks the application's readiness for operation",
		Method:      http.MethodGet,
		Path:        "/readyz",
	}, h.Readyz)
}

type HealthCheckOutput struct {
	Body struct {
		Status string `json:"status" example:"ok" doc:"Operational status"`
	}
}

func (h *HealthHandler) Healthz(_ context.Context, _ *struct{}) (*HealthCheckOutput, error) {
	response := &HealthCheckOutput{}
	response.Body.Status = "ok"
	return response, nil
}

type ReadyCheckOutput struct {
	Body struct {
		Status string `json:"status" example:"ok" doc:"Availability status"`
	}
}

func (h *HealthHandler) Readyz(ctx context.Context, _ *struct{}) (*ReadyCheckOutput, error) {
	if err := h.pool.Ping(ctx); err != nil {
		return nil, huma.Error503ServiceUnavailable("database is unavailable", err)
	}
	response := &ReadyCheckOutput{}
	response.Body.Status = "ok"
	return response, nil
}
