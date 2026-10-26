package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/logattr"
)

func problem(ctx context.Context, log *slog.Logger, err error) error {
	if errors.Is(err, domain.ErrGraphAPIUnavailable) {
		log.WarnContext(ctx, "graph-api request failed", logattr.Error(err))
		return huma.Error502BadGateway(domain.ErrGraphAPIUnavailable.Error())
	}

	log.ErrorContext(ctx, "request failed", logattr.Error(err))

	return huma.Error500InternalServerError("internal error")
}
