package app

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/feytox/shugraph/backend/graph-api/internal/config"
)

const timestampLayout = "2006-01-02T15:04:05.000Z"

func NewLogger() (*slog.Logger, error) {
	logCfg, err := config.Load[config.Log]()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       logCfg.Level,
		ReplaceAttr: renameStandardKeys,
	})

	return slog.New(handler).With(
		slog.String("service.name", logCfg.ServiceName),
		slog.String("deployment.environment.name", logCfg.Environment),
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
