package app

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feytox/shugraph/backend/graph-api/internal/repository"
	"github.com/feytox/shugraph/backend/graph-api/internal/service"
)

var snapshotTx = pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}

func inTx[R any](pool *pgxpool.Pool, opts pgx.TxOptions, repositories func(pgx.Tx) R) service.Transactor[R] {
	return func(ctx context.Context, fn func(R) error) error {
		return pgx.BeginTxFunc(ctx, pool, opts, func(tx pgx.Tx) error {
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

func groupRepositories(tx pgx.Tx) service.GroupRepositories {
	return service.GroupRepositories{Groups: repository.NewPgGroupRepository(tx)}
}
