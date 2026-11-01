//go:build integration

package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository"
)

type seedGroup struct {
	n        int
	scheme   domain.GroupScheme
	status   domain.GroupStatus
	detected int
}

var seedBase = time.Date(2026, 10, 12, 18, 0, 0, 0, time.UTC)

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

func insertGroups(t *testing.T) {
	t.Helper()

	truncate(t, "groups")

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

		_, err := testPool.Exec(t.Context(), `
			INSERT INTO groups (id, scheme, status, risk_score, risk_parts, detected_at,
			                    first_event_at, last_event_at, accounts_count, total_amount,
			                    is_frozen, decision_comment, decided_at)
			VALUES ($1, $2, $3, $4, '{"scheme": 55, "speed": 15}', $5, $6, $5, $7, $8, $9, $10, $11)`,
			groupID(g.n), string(g.scheme), string(g.status), 10*g.n, detectedAt,
			detectedAt.Add(-10*time.Minute), g.n, int64(g.n)*1000, g.n%2 == 0, comment, decidedAt,
		)
		require.NoError(t, err)
	}
}

func listAll(t *testing.T, repo *repository.PgGroupRepository, filter domain.GroupFilter) []domain.Group {
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
	insertGroups(t)

	repo := repository.NewPgGroupRepository(testPool)
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
			require.Equal(t, tc.expected, ids(listAll(t, repo, tc.filter)))
		})
	}

	t.Run("AllFields", func(t *testing.T) {
		groups, err := repo.List(t.Context(), domain.GroupFilter{Limit: 100})
		require.NoError(t, err)

		got := groups[4]
		require.JSONEq(t, `{"scheme": 55, "speed": 15}`, string(got.RiskParts))

		got.RiskParts = nil
		detectedAt := seedBase.Add(4 * time.Minute)
		require.Equal(t, domain.Group{
			ID:            groupID(4),
			Scheme:        domain.GroupSchemeRing,
			Status:        domain.GroupStatusDeclined,
			RiskScore:     40,
			DetectedAt:    detectedAt,
			FirstEventAt:  detectedAt.Add(-10 * time.Minute),
			LastEventAt:   detectedAt,
			AccountsCount: 4,
			TotalAmount:   4000,
			IsFrozen:      true,
			Decision: &domain.Decision{
				Comment:   "решение по группе",
				DecidedAt: detectedAt.Add(time.Hour),
			},
		}, got)
	})

	t.Run("NonPositiveLimit", func(t *testing.T) {
		for _, limit := range []int32{0, -1} {
			_, err := repo.List(t.Context(), domain.GroupFilter{Limit: limit})
			require.Error(t, err, "limit %d", limit)
		}
	})

	t.Run("Decision", func(t *testing.T) {
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
