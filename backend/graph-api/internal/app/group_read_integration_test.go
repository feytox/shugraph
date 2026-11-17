//go:build integration

package app

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/feytox/shugraph/backend/graph-api/internal/handler"
)

const groupsPath = "/api/v1/groups"

type groupList struct {
	Items      []handler.GroupItem `json:"items"`
	NextCursor *string             `json:"next_cursor"`
}

func TestGroupRead_ScenarioGroupInListAndCard(t *testing.T) {
	sc := loadScenario(t, "device_farm")

	clearEventTables(t)
	playScenario(t, sc)

	var id uuid.UUID

	var detectedAt time.Time

	err := testPool.QueryRow(t.Context(), "SELECT id, detected_at FROM groups").Scan(&id, &detectedAt)
	require.NoError(t, err)

	stored := expectedGroups(t, sc)[0]
	expected := handler.GroupItem{
		ID:           id.String(),
		Scheme:       handler.GroupSchemeFarm,
		Status:       handler.GroupStatusNew,
		RiskLevel:    handler.GroupRiskLow,
		RiskParts:    handler.RiskParts{Farm: &handler.FarmScheme{}},
		DetectedAt:   detectedAt.UTC(),
		FirstEventAt: stored.FirstEventAt,
		LastEventAt:  stored.LastEventAt,
		Features: handler.Feature{
			AccountsCount: stored.AccountsCount,
			SpanMinutes:   28,
			TotalAmount:   stored.TotalAmount,
		},
	}

	var list groupList
	require.Equal(t, http.StatusOK, getJSON(t, groupsPath, &list))
	require.Nil(t, list.NextCursor)
	require.Len(t, list.Items, 1)
	require.Equal(t, expected, utcItem(list.Items[0]))

	var card handler.Group
	require.Equal(t, http.StatusOK, getJSON(t, groupsPath+"/"+id.String(), &card))
	require.Equal(t, expected, utcItem(card.GroupItem))
	require.False(t, card.IsFrozen)
	require.Nil(t, card.Decision)

	events := slices.Concat(
		idsOf(card.Logins, func(e handler.LoginEvent) string { return e.EventID }),
		idsOf(card.Transfers, func(e handler.TransferEvent) string { return e.EventID }),
	)

	require.Equal(t, stored.Members, idsOf(card.Members, func(a handler.Account) string { return a.ID }))
	require.Equal(t, stored.Events, slices.Sorted(slices.Values(events)))
}

func TestGroupRead_PagesOfOneGiveEveryGroupOnce(t *testing.T) {
	clearEventTables(t)

	putLogins(t, "dev-1", 0, "a", "b", "c")
	putLogins(t, "dev-2", 10, "d", "e", "f")
	putLogins(t, "dev-3", 20, "g", "h", "i")

	rows, err := testPool.Query(t.Context(), "SELECT id::text FROM groups ORDER BY detected_at DESC, id DESC")
	require.NoError(t, err)

	expected := []string{}

	for rows.Next() {
		var id string
		require.NoError(t, rows.Scan(&id))

		expected = append(expected, id)
	}

	require.NoError(t, rows.Err())
	require.Len(t, expected, 3)

	var (
		seen   []string
		cursor *string
	)

	for range expected {
		query := url.Values{"limit": {"1"}}
		if cursor != nil {
			query.Set("cursor", *cursor)
		}

		var page groupList
		require.Equal(t, http.StatusOK, getJSON(t, groupsPath+"?"+query.Encode(), &page))
		require.Len(t, page.Items, 1)

		seen = append(seen, page.Items[0].ID)
		cursor = page.NextCursor
	}

	require.Nil(t, cursor)
	require.Equal(t, expected, seen)
}

func TestGroupRead_BrokenCursorIsUnprocessable(t *testing.T) {
	require.Equal(t, http.StatusUnprocessableEntity, getJSON(t, groupsPath+"?cursor=broken", nil))
}

func TestGroupRead_GraphHasOnlyGroupEvidence(t *testing.T) {
	clearEventTables(t)
	playScenario(t, loadScenario(t, "device_farm"))

	var id string
	require.NoError(t, testPool.QueryRow(t.Context(), "SELECT id::text FROM groups").Scan(&id))

	var graph handler.GroupGraph
	require.Equal(t, http.StatusOK, getJSON(t, groupsPath+"/"+id+"/graph", &graph))

	require.Equal(t, id, graph.GroupID)
	require.ElementsMatch(t, []string{
		"client:cli-ravil", "client:cli-sasha", "client:cli-dima", "client:cli-organizer",
		"account:acc-ravil", "account:acc-sasha", "account:acc-dima", "account:acc-collector",
		"device:dev-common",
		"network_address:198.51.100.21", "network_address:198.51.100.22", "network_address:198.51.100.23",
	}, idsOf(graph.Nodes, func(n handler.GraphNode) string { return n.ID }))
	require.ElementsMatch(t, []string{
		"ownership:cli-ravil:acc-ravil", "ownership:cli-sasha:acc-sasha",
		"ownership:cli-dima:acc-dima", "ownership:cli-organizer:acc-collector",
	}, idsOf(graph.Ownerships, func(e handler.OwnershipEdge) string { return e.ID }))
	require.ElementsMatch(t, []string{
		"login:evt-df-with-0001:device", "login:evt-df-with-0001:address",
		"login:evt-df-with-0002:device", "login:evt-df-with-0002:address",
		"login:evt-df-with-0003:device", "login:evt-df-with-0003:address",
	}, idsOf(graph.Logins, func(e handler.LoginEdge) string { return e.ID }))
	require.ElementsMatch(t, []string{
		"transfer:evt-df-with-0004", "transfer:evt-df-with-0005", "transfer:evt-df-with-0006",
	}, idsOf(graph.Transfers, func(e handler.TransferEdge) string { return e.ID }))
}

func TestGroupRead_UnknownGroupIsNotFound(t *testing.T) {
	path := groupsPath + "/" + uuid.NewString()

	require.Equal(t, http.StatusNotFound, getJSON(t, path, nil))
	require.Equal(t, http.StatusNotFound, getJSON(t, path+"/graph", nil))
}

func getJSON(t *testing.T, path string, out any) int {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	if out != nil && resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(out))
	}

	return resp.StatusCode
}

func utcItem(item handler.GroupItem) handler.GroupItem {
	item.DetectedAt = item.DetectedAt.UTC()
	item.FirstEventAt = item.FirstEventAt.UTC()
	item.LastEventAt = item.LastEventAt.UTC()

	return item
}

func idsOf[T any](items []T, id func(T) string) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, id(item))
	}

	return ids
}
