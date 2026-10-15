package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

const readinessTimeout = 2 * time.Second

type pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	pinger pinger
}

func NewHealthHandler(pinger pinger) *HealthHandler {
	return &HealthHandler{pinger: pinger}
}

func (h *HealthHandler) Register(api huma.API) {
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
	ctxWithTimeout, cancel := context.WithTimeout(ctx, readinessTimeout)
	defer cancel()
	if err := h.pinger.Ping(ctxWithTimeout); err != nil {
		slog.Error("database is unavailable", "err", err)
		return nil, huma.Error503ServiceUnavailable("database is unavailable")
	}
	response := &ReadyCheckOutput{}
	response.Body.Status = "ok"
	return response, nil
}
