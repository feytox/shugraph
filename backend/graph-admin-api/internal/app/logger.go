package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/config"
)

const timestampLayout = "2006-01-02T15:04:05.000Z"

func NewLogger(version string) (*slog.Logger, error) {
	cfg, err := config.Load[config.Log]()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       cfg.Level,
		ReplaceAttr: renameStandardKeys,
	})

	return slog.New(requestIDHandler{handler}).With(
		slog.String("service.name", cfg.ServiceName),
		slog.String("service.version", version),
		slog.String("deployment.environment.name", cfg.Environment),
	), nil
}

func renameStandardKeys(groups []string, a slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return a
	}

	switch a.Key {
	case slog.TimeKey:
		return slog.String("timestamp", a.Value.Time().UTC().Format(timestampLayout))
	case slog.MessageKey:
		a.Key = "message"
	}

	return a
}

type requestIDHandler struct {
	slog.Handler
}

func (h requestIDHandler) Handle(ctx context.Context, record slog.Record) error {
	if id := middleware.GetReqID(ctx); id != "" {
		record.AddAttrs(slog.String("request_id", id))
	}

	return h.Handler.Handle(ctx, record)
}

func (h requestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return requestIDHandler{h.Handler.WithAttrs(attrs)}
}

func (h requestIDHandler) WithGroup(name string) slog.Handler {
	return requestIDHandler{h.Handler.WithGroup(name)}
}
