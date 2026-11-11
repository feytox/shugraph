package app

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feytox/shugraph/backend/graph-api/internal/repository"
	"github.com/feytox/shugraph/backend/graph-api/internal/service"
)

func inTx[R any](pool *pgxpool.Pool, repositories func(pgx.Tx) R) service.Transactor[R] {
	return func(ctx context.Context, fn func(R) error) error {
		return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
			return fn(repositories(tx))
		})
	}
}

func eventRepositories(tx pgx.Tx) service.EventRepositories {
	return service.EventRepositories{
		Events: repository.NewPgEventRepository(tx),
		Groups: repository.NewPgGroupAssemblyRepository(tx),
		Farms:  repository.NewPgFarmRepository(tx),
	}
}
