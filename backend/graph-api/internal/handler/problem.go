package handler

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"

	"github.com/feytox/shugraph/backend/graph-api/internal/logattr"
)

func problem(ctx context.Context, log *slog.Logger, err error) error {
	log.ErrorContext(ctx, "request failed", logattr.Error(err))

	return huma.Error500InternalServerError("internal error")
}
