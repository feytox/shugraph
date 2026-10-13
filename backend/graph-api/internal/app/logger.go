package app

import (
	"log/slog"
	"os"

	"github.com/feytox/shugraph/backend/graph-api/internal/config"
)

const timestampLayout = "2006-01-02T15:04:05.000Z"

func NewLogger(cfg config.Log) (*slog.Logger, error) {

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       cfg.Level,
		ReplaceAttr: renameStandardKeys,
	})

	return slog.New(handler).With(
		slog.String("service.name", cfg.ServiceName),
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
