package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/logattr"
)

const (
	healthTag        = "Health"
	readinessTimeout = 2 * time.Second
)

type StatusOutput struct {
	Body struct {
		Status string `json:"status" example:"ok"`
	}
}

type healthHandlers struct {
	database Pinger
	log      *slog.Logger
}

func registerHealth(api huma.API, h *healthHandlers) {
	huma.Register(api, huma.Operation{
		OperationID: "health",
		Summary:     "Check that the process is alive",
		Method:      http.MethodGet,
		Path:        "/healthz",
		Tags:        []string{healthTag},
	}, h.healthz)

	huma.Register(api, huma.Operation{
		OperationID: "ready",
		Summary:     "Check that the service can handle requests",
		Method:      http.MethodGet,
		Path:        "/readyz",
		Tags:        []string{healthTag},
		Errors:      []int{http.StatusServiceUnavailable},
	}, h.readyz)
}

func (h *healthHandlers) healthz(_ context.Context, _ *struct{}) (*StatusOutput, error) {
	return statusOK(), nil
}

func (h *healthHandlers) readyz(ctx context.Context, _ *struct{}) (*StatusOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, readinessTimeout)
	defer cancel()

	if err := h.database.Ping(ctx); err != nil {
		h.log.WarnContext(ctx, "database is unavailable", logattr.Error(err))
		return nil, huma.Error503ServiceUnavailable("database is unavailable")
	}

	return statusOK(), nil
}

func statusOK() *StatusOutput {
	out := &StatusOutput{}
	out.Body.Status = "ok"

	return out
}
