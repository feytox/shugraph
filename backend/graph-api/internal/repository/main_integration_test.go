//go:build integration

package repository_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/feytox/shugraph/backend/graph-api/internal/config"
	pgdb "github.com/feytox/shugraph/backend/graph-api/internal/postgres"
)

// testPool - один Postgres на все тесты пакета, данные между тестами чистит truncate.
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("graph_api"),
		postgres.WithUsername("graph_api"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		slog.Error("could not start postgres", "err", err)
		return 1
	}

	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			slog.Error("could not terminate postgres container", "err", err)
		}
	}()

	dbURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		slog.Error("could not get connection string to postgres", "err", err)
		return 1
	}

	dbCfg, err := env.ParseAsWithOptions[config.Database](env.Options{Environment: map[string]string{
		"DATABASE_URL": dbURL,
	}})
	if err != nil {
		slog.Error("could not parse database config", "err", err)
		return 1
	}

	testPool, err = pgdb.Connect(ctx, dbCfg)
	if err != nil {
		slog.Error("could not connect to postgres", "err", err)
		return 1
	}
	defer testPool.Close()

	if err := pgdb.Migrate(ctx, testPool, slog.Default()); err != nil {
		slog.Error("could not migrate", "err", err)
		return 1
	}

	return m.Run()
}

func truncate(t *testing.T, tables ...string) {
	t.Helper()

	for _, table := range tables {
		_, err := testPool.Exec(t.Context(), "TRUNCATE "+table+" CASCADE")
		require.NoError(t, err)
	}
}
