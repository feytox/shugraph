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
			INSERT INTO groups (id, scheme, status, risk_score, risk_parts, risk_level, detected_at,
			                    first_event_at, last_event_at, accounts_count, total_amount,
			                    is_frozen, decision_comment, decided_at)
			VALUES ($1, $2, $3, $4, '{"scheme": 55, "speed": 15}', 'low', $5, $6, $5, $7, $8, $9, $10, $11)`,
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

		next := page[len(page)-1].Cursor()
		filter.After = &next
	}
}

func ids(groups []domain.Group) []int {
	result := make([]int, 0, len(groups))
	for _, g := range groups {
		result = append(result, int(g.ID[15]))
	}

	return result
}

func requireGroup4(t *testing.T, got domain.Group) {
	t.Helper()

	require.JSONEq(t, `{"scheme": 55, "speed": 15}`, string(got.RiskParts))

	got.RiskParts = nil
	detectedAt := seedBase.Add(4 * time.Minute)
	require.Equal(t, domain.Group{
		ID:            groupID(4),
		Scheme:        domain.GroupSchemeRing,
		Status:        domain.GroupStatusDeclined,
		RiskScore:     40,
		RiskLevel:     domain.RiskLevelLow,
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
		requireGroup4(t, groups[4])
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

func insertGroupEvents(t *testing.T) {
	t.Helper()
	insertGroups(t)
	truncate(t, "events", "accounts", "clients", "devices", "network_addresses")

	_, err := testPool.Exec(t.Context(), `
		INSERT INTO clients (id) VALUES ('cli-c'), ('cli-b'), ('cli-a');
		INSERT INTO accounts (id, client_id) VALUES ('acc-c', 'cli-c'), ('acc-b', 'cli-b'), ('acc-a', 'cli-a');
		INSERT INTO devices (device_id) VALUES ('dev-1'), ('dev-2');
		INSERT INTO network_addresses (address_id) VALUES ('10.0.0.1'), ('10.0.0.2');
		INSERT INTO events (event_id, type, occurred_at) VALUES
			('transfer-1805', 'transfer', '2026-10-12 18:05:00Z'),
			('transfer-1804', 'transfer', '2026-10-12 18:04:00Z'),
			('login-1803', 'login', '2026-10-12 18:03:00Z'),
			('login-1802', 'login', '2026-10-12 18:02:00Z'),
			('login-1801', 'login', '2026-10-12 18:01:00Z');
		INSERT INTO login_events (event_id, account_id, device_id, address_id) VALUES
			('login-1803', 'acc-a', 'dev-2', '10.0.0.1'),
			('login-1802', 'acc-b', 'dev-1', '10.0.0.2'),
			('login-1801', 'acc-a', 'dev-1', '10.0.0.1');
		INSERT INTO transfers (event_id, from_account_id, to_account_id, amount) VALUES
			('transfer-1805', 'acc-b', 'acc-c', 7000),
			('transfer-1804', 'acc-a', 'acc-c', 5000);
		INSERT INTO group_members (group_id, account_id) VALUES
			('00000000-0000-0000-0000-000000000001', 'acc-c'),
			('00000000-0000-0000-0000-000000000001', 'acc-b'),
			('00000000-0000-0000-0000-000000000001', 'acc-a'),
			('00000000-0000-0000-0000-000000000002', 'acc-b');
		INSERT INTO group_events (group_id, event_id) VALUES
			('00000000-0000-0000-0000-000000000001', 'transfer-1804'),
			('00000000-0000-0000-0000-000000000001', 'login-1802'),
			('00000000-0000-0000-0000-000000000001', 'login-1801'),
			('00000000-0000-0000-0000-000000000002', 'transfer-1805'),
			('00000000-0000-0000-0000-000000000002', 'login-1803');`)
	require.NoError(t, err)
}

func TestGroupRepository_Get(t *testing.T) {
	insertGroups(t)

	repo := repository.NewPgGroupRepository(testPool)

	t.Run("Found", func(t *testing.T) {
		got, err := repo.Get(t.Context(), groupID(4))
		require.NoError(t, err)
		requireGroup4(t, got)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := repo.Get(t.Context(), groupID(99))
		require.ErrorIs(t, err, domain.ErrGroupNotFound)
	})
}

func TestGroupRepository_Contents(t *testing.T) {
	insertGroupEvents(t)

	repo := repository.NewPgGroupRepository(testPool)

	t.Run("Members", func(t *testing.T) {
		members, err := repo.ListMembers(t.Context(), groupID(1))
		require.NoError(t, err)
		require.Equal(t, []domain.Account{
			{ID: "acc-a", ClientID: "cli-a"},
			{ID: "acc-b", ClientID: "cli-b"},
			{ID: "acc-c", ClientID: "cli-c"},
		}, members)
	})

	t.Run("Logins", func(t *testing.T) {
		logins, err := repo.ListLogins(t.Context(), groupID(1))
		require.NoError(t, err)
		require.Equal(t, []domain.LoginEvent{
			{
				EventID:    "login-1801",
				OccurredAt: seedBase.Add(time.Minute),
				AccountID:  "acc-a",
				ClientID:   "cli-a",
				DeviceID:   "dev-1",
				AddressID:  "10.0.0.1",
			},
			{
				EventID:    "login-1802",
				OccurredAt: seedBase.Add(2 * time.Minute),
				AccountID:  "acc-b",
				ClientID:   "cli-b",
				DeviceID:   "dev-1",
				AddressID:  "10.0.0.2",
			},
		}, logins)
	})

	t.Run("Transfers", func(t *testing.T) {
		transfers, err := repo.ListTransfers(t.Context(), groupID(1))
		require.NoError(t, err)
		require.Equal(t, []domain.TransferEvent{{
			EventID:       "transfer-1804",
			OccurredAt:    seedBase.Add(4 * time.Minute),
			FromAccountID: "acc-a",
			FromClientID:  "cli-a",
			ToAccountID:   "acc-c",
			ToClientID:    "cli-c",
			Amount:        5000,
		}}, transfers)
	})

	t.Run("EmptyGroup", func(t *testing.T) {
		members, err := repo.ListMembers(t.Context(), groupID(3))
		require.NoError(t, err)
		require.Empty(t, members)

		logins, err := repo.ListLogins(t.Context(), groupID(3))
		require.NoError(t, err)
		require.Empty(t, logins)

		transfers, err := repo.ListTransfers(t.Context(), groupID(3))
		require.NoError(t, err)
		require.Empty(t, transfers)
	})
}
