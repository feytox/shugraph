//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/feytox/shugraph/backend/graph-api/internal/config"
	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	pgdb "github.com/feytox/shugraph/backend/graph-api/internal/postgres"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository"
)

type groupLister interface {
	List(ctx context.Context, filter domain.GroupFilter) ([]domain.Group, error)
}

type seedGroup struct {
	n        int
	scheme   domain.GroupScheme
	status   domain.GroupStatus
	detected int
}

var seedBase = time.Date(2026, 10, 12, 18, 0, 0, 0, time.UTC)

// Группы 7 и 8 обнаружены в одну минуту: порядок между ними решает id.
var seedGroups = []seedGroup{
	{n: 1, scheme: domain.GroupSchemeFarm, status: domain.GroupStatusNew, detected: 1},
	{n: 2, scheme: domain.GroupSchemeRing, status: domain.GroupStatusNew, detected: 2},
	{n: 3, scheme: domain.GroupSchemeChain, status: domain.GroupStatusApproved, detected: 3},
	{n: 4, scheme: domain.GroupSchemeRing, status: domain.GroupStatusDeclined, detected: 4},
	{n: 5, scheme: domain.GroupSchemeAddress, status: domain.GroupStatusNew, detected: 5},
	{n: 6, scheme: domain.GroupSchemeRing, status: domain.GroupStatusNew, detected: 6},
	{n: 7, scheme: domain.GroupSchemeFarm, status: domain.GroupStatusNew, detected: 7},
	{n: 8, scheme: domain.GroupSchemeChain, status: domain.GroupStatusNew, detected: 7},
}

func groupID(n int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
}

func startPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := t.Context()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("graph_api"),
		postgres.WithUsername("graph_api"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, pgContainer.Terminate(context.Background()))
	})

	dbURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	dbCfg, err := env.ParseAsWithOptions[config.Database](env.Options{Environment: map[string]string{
		"DATABASE_URL": dbURL,
	}})
	require.NoError(t, err)

	pool, err := pgdb.Connect(ctx, dbCfg)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, pgdb.Migrate(ctx, pool, slog.Default()))

	return pool
}

func insertGroups(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	for _, g := range seedGroups {
		detectedAt := seedBase.Add(time.Duration(g.detected) * time.Minute)

		var (
			comment   *string
			decidedAt *time.Time
		)

		if g.status != domain.GroupStatusNew {
			text := "решение по группе"
			at := detectedAt.Add(time.Hour)
			comment, decidedAt = &text, &at
		}

		_, err := pool.Exec(t.Context(), `
			INSERT INTO groups (id, scheme, status, risk_score, risk_parts, detected_at,
			                    first_event_at, last_event_at, accounts_count, total_amount,
			                    decision_comment, decided_at)
			VALUES ($1, $2, $3, 80, '{"scheme": 55, "speed": 15}', $4, $5, $4, 3, 100000, $6, $7)`,
			groupID(g.n), string(g.scheme), string(g.status), detectedAt,
			detectedAt.Add(-10*time.Minute), comment, decidedAt,
		)
		require.NoError(t, err)
	}
}

// listAll листает страницы по курсору, пока они не закончатся.
func listAll(t *testing.T, repo groupLister, filter domain.GroupFilter) []domain.Group {
	t.Helper()

	var all []domain.Group

	for {
		page, err := repo.List(t.Context(), filter)
		require.NoError(t, err)

		all = append(all, page...)

		if len(page) < int(filter.Limit) {
			return all
		}

		last := page[len(page)-1]
		filter.After = &domain.GroupCursor{DetectedAt: last.DetectedAt, ID: last.ID}
	}
}

func ids(groups []domain.Group) []int {
	result := make([]int, 0, len(groups))
	for _, g := range groups {
		result = append(result, int(g.ID[15]))
	}

	return result
}

func TestGroupRepository_List(t *testing.T) {
	pool := startPostgres(t)
	insertGroups(t, pool)

	repos := map[string]groupLister{
		"sqlc":       repository.NewPgGroupRepository(pool),
		"goqu":       repository.NewGoquGroupRepository(pool),
		"sqlbuilder": repository.NewSQLBuilderGroupRepository(pool),
	}

	statusNew := domain.GroupStatusNew

	tests := []struct {
		name     string
		filter   domain.GroupFilter
		expected []int
	}{
		{
			name:     "NoFilters",
			filter:   domain.GroupFilter{Limit: 100},
			expected: []int{8, 7, 6, 5, 4, 3, 2, 1},
		},
		{
			name:     "ByStatus",
			filter:   domain.GroupFilter{Status: &statusNew, Limit: 100},
			expected: []int{8, 7, 6, 5, 2, 1},
		},
		{
			name:     "BySchemes",
			filter:   domain.GroupFilter{Schemes: []domain.GroupScheme{domain.GroupSchemeRing}, Limit: 100},
			expected: []int{6, 4, 2},
		},
		{
			name: "ByStatusAndSchemes",
			filter: domain.GroupFilter{
				Status:  &statusNew,
				Schemes: []domain.GroupScheme{domain.GroupSchemeRing, domain.GroupSchemeFarm},
				Limit:   100,
			},
			expected: []int{7, 6, 2, 1},
		},
		{
			name:     "PagesOfOne",
			filter:   domain.GroupFilter{Limit: 1},
			expected: []int{8, 7, 6, 5, 4, 3, 2, 1},
		},
		{
			name:     "PagesOfThreeByStatus",
			filter:   domain.GroupFilter{Status: &statusNew, Limit: 3},
			expected: []int{8, 7, 6, 5, 2, 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results := make(map[string][]domain.Group, len(repos))

			for name, repo := range repos {
				results[name] = listAll(t, repo, tc.filter)
				require.Equal(t, tc.expected, ids(results[name]), name)
			}

			for name, result := range results {
				require.Equal(t, results["sqlc"], result, name)
			}
		})
	}
}

func TestGroupRepository_ListDecision(t *testing.T) {
	pool := startPostgres(t)
	insertGroups(t, pool)

	repos := map[string]groupLister{
		"sqlc":       repository.NewPgGroupRepository(pool),
		"goqu":       repository.NewGoquGroupRepository(pool),
		"sqlbuilder": repository.NewSQLBuilderGroupRepository(pool),
	}

	for name, repo := range repos {
		t.Run(name, func(t *testing.T) {
			groups, err := repo.List(t.Context(), domain.GroupFilter{Limit: 100})
			require.NoError(t, err)

			for _, g := range groups {
				if g.Status == domain.GroupStatusNew {
					require.Nil(t, g.Decision, "group %s", g.ID)
					continue
				}

				require.NotNil(t, g.Decision, "group %s", g.ID)
				require.Equal(t, "решение по группе", g.Decision.Comment)
				require.Equal(t, g.DetectedAt.Add(time.Hour), g.Decision.DecidedAt)
			}
		})
	}
}
