//go:build integration

package app

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/feytox/shugraph/backend/graph-api/internal/handler"
)

const scenariosDir = "../../../graph-admin-api/scenarios"

type scenario struct {
	Events   []json.RawMessage `json:"events"`
	Expected struct {
		Groups []struct {
			Scheme           string   `json:"scheme"`
			MemberAccountIDs []string `json:"member_account_ids"`
			EventIDs         []string `json:"event_ids"`
		} `json:"groups"`
	} `json:"expected"`
}

type scenarioEvent struct {
	EventID    string    `json:"event_id"`
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
	Amount     int64     `json:"amount"`
}

type storedGroup struct {
	Scheme        string
	Members       []string
	Events        []string
	FirstEventAt  time.Time
	LastEventAt   time.Time
	AccountsCount int
	TotalAmount   int64
	IsFrozen      bool
}

func TestGroups_Scenarios(t *testing.T) {
	for _, name := range []string{"device_farm_with_group", "device_farm_without_group"} {
		t.Run(name, func(t *testing.T) {
			sc := loadScenario(t, name)

			clearEventTables(t)
			playScenario(t, sc)

			require.Equal(t, expectedGroups(t, sc), storedGroups(t))
		})
	}
}

func TestGroups_ScenarioReplayOnEmptyDatabaseGivesSameGroup(t *testing.T) {
	sc := loadScenario(t, "device_farm_with_group")

	clearEventTables(t)
	playScenario(t, sc)

	first := storedGroups(t)
	require.Len(t, first, 1)

	clearEventTables(t)
	playScenario(t, sc)

	require.Equal(t, first, storedGroups(t))
}

func TestGroups_ScenarioInReverseOrderGivesSameGroup(t *testing.T) {
	sc := loadScenario(t, "device_farm_with_group")
	slices.Reverse(sc.Events)

	clearEventTables(t)
	playScenario(t, sc)

	require.Equal(t, expectedGroups(t, sc), storedGroups(t))
}

func TestGroups_FarmGrowsUntilFrozen(t *testing.T) {
	clearEventTables(t)

	putLogins(t, "dev-farm", 0, "a", "b")
	require.Empty(t, storedGroups(t))

	putLogins(t, "dev-farm", 2, "c")
	require.Equal(t, []string{"acc-a", "acc-b", "acc-c"}, onlyGroup(t).Members)

	putLogins(t, "dev-farm", 3, "d")
	require.Equal(t, []string{"acc-a", "acc-b", "acc-c", "acc-d"}, onlyGroup(t).Members)

	putLogins(t, "dev-other", 3+24*60+1, "x")
	require.True(t, onlyGroup(t).IsFrozen)

	putLogins(t, "dev-farm", 3+24*60+2, "e")
	require.Len(t, storedGroups(t), 1)
}

func TestGroups_LateLoginAfterFreezeStaysOutOfGroup(t *testing.T) {
	clearEventTables(t)

	putLogins(t, "dev-farm", 3*60+27, "a", "b", "c")
	putLogins(t, "dev-other", 24*60+3*60+30, "x")
	putLogins(t, "dev-farm", 5*60, "d")

	group := onlyGroup(t)
	require.True(t, group.IsFrozen)
	require.Equal(t, []string{"acc-a", "acc-b", "acc-c"}, group.Members)
}

func TestGroups_ParallelLoginsOnOneDeviceGiveOneGroup(t *testing.T) {
	clearEventTables(t)

	clients := []string{"a", "b", "c", "d", "e", "f"}
	statuses := make([]int, len(clients))

	var wg sync.WaitGroup

	for i, client := range clients {
		wg.Go(func() {
			statuses[i] = sendEvent(t.Context(), loginsPath+"login-"+client, deviceLogin("dev-farm", i, client))
		})
	}

	wg.Wait()

	for i, status := range statuses {
		require.Equal(t, http.StatusCreated, status, clients[i])
	}

	require.Equal(t, []string{"acc-a", "acc-b", "acc-c", "acc-d", "acc-e", "acc-f"}, onlyGroup(t).Members)
}

func TestGroups_FrozenGroupEventsDoNotCountForNewGroup(t *testing.T) {
	clearEventTables(t)

	putLogins(t, "dev-farm", 0, "a", "b", "c")

	_, err := testPool.Exec(t.Context(), "UPDATE groups SET is_frozen = TRUE")
	require.NoError(t, err)

	putLogins(t, "dev-farm", 3, "d", "e")
	require.Len(t, storedGroups(t), 1)

	putLogins(t, "dev-farm", 5, "f")

	groups := storedGroups(t)
	require.Len(t, groups, 2)
	require.Equal(t, []string{"acc-d", "acc-e", "acc-f"}, groups[1].Members)
	require.Equal(t, []string{"login-d", "login-e", "login-f"}, groups[1].Events)
}

func putLogins(t *testing.T, deviceID string, minute int, clients ...string) {
	t.Helper()

	for i, client := range clients {
		require.Equal(t, http.StatusCreated, putEvent(t, loginsPath+"login-"+client, deviceLogin(deviceID, minute+i, client)))
	}
}

func deviceLogin(deviceID string, minute int, client string) handler.LoginPayload {
	return handler.LoginPayload{
		OccurredAt: eventTime.Add(time.Duration(minute) * time.Minute),
		AccountID:  "acc-" + client,
		ClientID:   "cli-" + client,
		DeviceID:   deviceID,
		AddressID:  "10.0.0.1",
	}
}

func onlyGroup(t *testing.T) storedGroup {
	t.Helper()

	groups := storedGroups(t)
	require.Len(t, groups, 1)

	return groups[0]
}

func loadScenario(t *testing.T, name string) scenario {
	t.Helper()

	data, err := fs.ReadFile(os.DirFS(scenariosDir), name+".json")
	require.NoError(t, err)

	var sc scenario
	require.NoError(t, json.Unmarshal(data, &sc))

	return sc
}

func playScenario(t *testing.T, sc scenario) {
	t.Helper()

	for _, raw := range sc.Events {
		event := decodeScenarioEvent(t, raw)

		var (
			path string
			body any
		)

		switch event.Type {
		case "login":
			var login handler.LoginPayload
			require.NoError(t, json.Unmarshal(raw, &login))

			path, body = loginsPath+event.EventID, login
		case "transfer":
			var transfer handler.TransferPayload
			require.NoError(t, json.Unmarshal(raw, &transfer))

			path, body = transfersPath+event.EventID, transfer
		default:
			require.FailNow(t, fmt.Sprintf("unknown event type %q", event.Type))
		}

		require.Equal(t, http.StatusCreated, putEvent(t, path, body), event.EventID)
	}
}

func decodeScenarioEvent(t *testing.T, raw json.RawMessage) scenarioEvent {
	t.Helper()

	var event scenarioEvent
	require.NoError(t, json.Unmarshal(raw, &event))

	return event
}

func expectedGroups(t *testing.T, sc scenario) []storedGroup {
	t.Helper()

	events := make(map[string]scenarioEvent, len(sc.Events))

	for _, raw := range sc.Events {
		event := decodeScenarioEvent(t, raw)
		events[event.EventID] = event
	}

	groups := make([]storedGroup, 0, len(sc.Expected.Groups))

	for _, expected := range sc.Expected.Groups {
		group := storedGroup{
			Scheme:        expected.Scheme,
			Members:       slices.Sorted(slices.Values(expected.MemberAccountIDs)),
			Events:        slices.Sorted(slices.Values(expected.EventIDs)),
			AccountsCount: len(expected.MemberAccountIDs),
		}

		for i, id := range group.Events {
			event, ok := events[id]
			require.True(t, ok, "expected event %s is not in scenario", id)

			if i == 0 || event.OccurredAt.Before(group.FirstEventAt) {
				group.FirstEventAt = event.OccurredAt
			}

			if event.OccurredAt.After(group.LastEventAt) {
				group.LastEventAt = event.OccurredAt
			}

			group.TotalAmount += event.Amount
		}

		groups = append(groups, group)
	}

	return groups
}

func storedGroups(t *testing.T) []storedGroup {
	t.Helper()

	rows, err := testPool.Query(t.Context(), `
		SELECT g.scheme,
		       ARRAY(SELECT account_id FROM group_members WHERE group_id = g.id ORDER BY account_id),
		       ARRAY(SELECT event_id FROM group_events WHERE group_id = g.id ORDER BY event_id),
		       g.first_event_at, g.last_event_at, g.accounts_count, g.total_amount, g.is_frozen
		FROM groups g
		ORDER BY g.first_event_at, g.id`)
	require.NoError(t, err)

	defer rows.Close()

	groups := []storedGroup{}

	for rows.Next() {
		var g storedGroup

		err := rows.Scan(&g.Scheme, &g.Members, &g.Events, &g.FirstEventAt, &g.LastEventAt,
			&g.AccountsCount, &g.TotalAmount, &g.IsFrozen)
		require.NoError(t, err)

		groups = append(groups, g)
	}

	require.NoError(t, rows.Err())

	return groups
}
