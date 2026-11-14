//go:build integration

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/feytox/shugraph/backend/graph-api/internal/handler"
)

const (
	loginsPath    = "/api/v1/events/logins/"
	transfersPath = "/api/v1/events/transfers/"
)

var eventTables = []string{
	"clients",
	"accounts",
	"devices",
	"network_addresses",
	"events",
	"transfers",
	"login_events",
}

var eventTime = time.Date(2042, time.October, 06, 07, 0, 0, 0, time.UTC)

func newLogin() handler.LoginPayload {
	return handler.LoginPayload{
		OccurredAt: eventTime,
		AccountID:  "account-1",
		ClientID:   "client-1",
		DeviceID:   "device-1",
		AddressID:  "10.0.0.1",
	}
}

func newTransfer() handler.TransferPayload {
	return handler.TransferPayload{
		OccurredAt:    eventTime,
		FromAccountID: "account-1",
		FromClientID:  "client-1",
		ToAccountID:   "account-2",
		ToClientID:    "client-2",
		Amount:        100,
	}
}

func TestEvents_TransferBetweenNewAccountsCreatesClientsAccountsAndEvent(t *testing.T) {
	clearEventTables(t)

	require.Equal(t, http.StatusCreated, putEvent(t, transfersPath+"transfer-1", newTransfer()))

	require.Equal(t, map[string]int{
		"clients":           2,
		"accounts":          2,
		"devices":           0,
		"network_addresses": 0,
		"events":            1,
		"transfers":         1,
		"login_events":      0,
	}, countEventRows(t))
}

func TestEvents_LoginFromKnownDeviceDoesNotCreateDevice(t *testing.T) {
	clearEventTables(t)

	require.Equal(t, http.StatusCreated, putEvent(t, loginsPath+"login-1", newLogin()))
	require.Equal(t, http.StatusCreated, putEvent(t, loginsPath+"login-2", newLogin()))

	rows := countEventRows(t)
	require.Equal(t, 1, rows["devices"])
	require.Equal(t, 1, rows["network_addresses"])
	require.Equal(t, 2, rows["login_events"])
}

func TestEvents_BrokenEventWritesNothing(t *testing.T) {
	tests := []struct {
		name string
		path string
		body any
	}{
		{
			name: "LoginWithAnotherAccountOwner",
			path: loginsPath + "login-2",
			body: func() handler.LoginPayload {
				login := newLogin()
				login.ClientID = "client-new"
				login.DeviceID = "device-new"
				login.AddressID = "10.0.0.2"

				return login
			}(),
		},
		{
			name: "TransferWithAnotherSenderOwner",
			path: transfersPath + "transfer-1",
			body: func() handler.TransferPayload {
				transfer := newTransfer()
				transfer.FromClientID = "client-new"
				transfer.ToAccountID = "account-new"
				transfer.ToClientID = "client-other"

				return transfer
			}(),
		},
		{
			name: "TransferWithAnotherReceiverOwner",
			path: transfersPath + "transfer-1",
			body: func() handler.TransferPayload {
				transfer := newTransfer()
				transfer.FromAccountID = "account-new"
				transfer.FromClientID = "client-other"
				transfer.ToAccountID = "account-1"
				transfer.ToClientID = "client-new"

				return transfer
			}(),
		},
		{
			name: "TransferToSameAccount",
			path: transfersPath + "transfer-1",
			body: func() handler.TransferPayload {
				transfer := newTransfer()
				transfer.FromAccountID = "account-new"
				transfer.ToAccountID = "account-new"
				transfer.ToClientID = transfer.FromClientID

				return transfer
			}(),
		},
		{
			name: "TransferWithZeroAmount",
			path: transfersPath + "transfer-1",
			body: func() handler.TransferPayload {
				transfer := newTransfer()
				transfer.FromAccountID = "account-new"
				transfer.Amount = 0

				return transfer
			}(),
		},
		{
			name: "LoginWithoutRequiredFields",
			path: loginsPath + "login-2",
			body: map[string]string{"account_id": "account-new"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearEventTables(t)
			require.Equal(t, http.StatusCreated, putEvent(t, loginsPath+"login-1", newLogin()))

			rowsBefore := countEventRows(t)

			require.Equal(t, http.StatusUnprocessableEntity, putEvent(t, tc.path, tc.body))
			require.Equal(t, rowsBefore, countEventRows(t))
		})
	}
}

func TestEvents_SameIDWithDifferentLoginFieldsConflicts(t *testing.T) {
	tests := []struct {
		name   string
		change func(*handler.LoginPayload)
	}{
		{"OccurredAt", func(l *handler.LoginPayload) { l.OccurredAt = eventTime.Add(time.Second) }},
		{"AccountID", func(l *handler.LoginPayload) { l.AccountID = "account-new" }},
		{"ClientID", func(l *handler.LoginPayload) { l.ClientID = "client-new" }},
		{"DeviceID", func(l *handler.LoginPayload) { l.DeviceID = "device-new" }},
		{"AddressID", func(l *handler.LoginPayload) { l.AddressID = "10.0.0.2" }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearEventTables(t)
			require.Equal(t, http.StatusCreated, putEvent(t, loginsPath+"login-1", newLogin()))

			rowsBefore := countEventRows(t)

			login := newLogin()
			tc.change(&login)

			require.Equal(t, http.StatusConflict, putEvent(t, loginsPath+"login-1", login))
			require.Equal(t, rowsBefore, countEventRows(t))
		})
	}
}

func TestEvents_SameIDWithDifferentTransferFieldsConflicts(t *testing.T) {
	tests := []struct {
		name   string
		change func(*handler.TransferPayload)
	}{
		{"OccurredAt", func(tr *handler.TransferPayload) { tr.OccurredAt = eventTime.Add(time.Second) }},
		{"FromAccountID", func(tr *handler.TransferPayload) { tr.FromAccountID = "account-new" }},
		{"FromClientID", func(tr *handler.TransferPayload) { tr.FromClientID = "client-new" }},
		{"ToAccountID", func(tr *handler.TransferPayload) { tr.ToAccountID = "account-new" }},
		{"ToClientID", func(tr *handler.TransferPayload) { tr.ToClientID = "client-new" }},
		{"Amount", func(tr *handler.TransferPayload) { tr.Amount = 101 }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearEventTables(t)
			require.Equal(t, http.StatusCreated, putEvent(t, transfersPath+"transfer-1", newTransfer()))

			rowsBefore := countEventRows(t)

			transfer := newTransfer()
			tc.change(&transfer)

			require.Equal(t, http.StatusConflict, putEvent(t, transfersPath+"transfer-1", transfer))
			require.Equal(t, rowsBefore, countEventRows(t))
		})
	}
}

func TestEvents_SameIDWithAnotherEventTypeConflicts(t *testing.T) {
	clearEventTables(t)

	require.Equal(t, http.StatusCreated, putEvent(t, loginsPath+"event-1", newLogin()))
	require.Equal(t, http.StatusCreated, putEvent(t, transfersPath+"event-2", newTransfer()))

	rowsBefore := countEventRows(t)

	require.Equal(t, http.StatusConflict, putEvent(t, transfersPath+"event-1", newTransfer()))
	require.Equal(t, http.StatusConflict, putEvent(t, loginsPath+"event-2", newLogin()))
	require.Equal(t, rowsBefore, countEventRows(t))
}

func TestEvents_ConflictTakesPriorityOverOwnerMismatch(t *testing.T) {
	clearEventTables(t)

	require.Equal(t, http.StatusCreated, putEvent(t, loginsPath+"login-1", newLogin()))

	login := newLogin()
	login.ClientID = "client-new"

	require.Equal(t, http.StatusConflict, putEvent(t, loginsPath+"login-1", login))
}

func TestEvents_RepeatedDeliveryDoesNotCreateSecondRecord(t *testing.T) {
	const moscowOffsetSeconds = 3 * 60 * 60

	tests := []struct {
		name       string
		occurredAt time.Time
		repeatedAt time.Time
	}{
		{"SameTime", eventTime, eventTime},
		{"NanosecondPrecision", eventTime.Add(123456789), eventTime.Add(123456789)},
		{"SameMomentInAnotherTimeZone", eventTime, eventTime.In(time.FixedZone("MSK", moscowOffsetSeconds))},
	}

	for _, tc := range tests {
		t.Run("Login"+tc.name, func(t *testing.T) {
			clearEventTables(t)

			login := newLogin()
			login.OccurredAt = tc.occurredAt
			require.Equal(t, http.StatusCreated, putEvent(t, loginsPath+"login-1", login))

			rowsBefore := countEventRows(t)

			login.OccurredAt = tc.repeatedAt
			require.Equal(t, http.StatusNoContent, putEvent(t, loginsPath+"login-1", login))
			require.Equal(t, rowsBefore, countEventRows(t))
		})

		t.Run("Transfer"+tc.name, func(t *testing.T) {
			clearEventTables(t)

			transfer := newTransfer()
			transfer.OccurredAt = tc.occurredAt
			require.Equal(t, http.StatusCreated, putEvent(t, transfersPath+"transfer-1", transfer))

			rowsBefore := countEventRows(t)

			transfer.OccurredAt = tc.repeatedAt
			require.Equal(t, http.StatusNoContent, putEvent(t, transfersPath+"transfer-1", transfer))
			require.Equal(t, rowsBefore, countEventRows(t))
		})
	}
}

func TestEvents_ConcurrentDuplicatesCreateOneRecord(t *testing.T) {
	const requests = 50

	clearEventTables(t)

	statuses := make(chan int, requests)

	var wg sync.WaitGroup
	for range requests {
		wg.Go(func() {
			statuses <- sendEvent(t.Context(), transfersPath+"transfer-1", newTransfer())
		})
	}

	wg.Wait()
	close(statuses)

	statusCounts := map[int]int{}
	for status := range statuses {
		statusCounts[status]++
	}

	require.Equal(t, map[int]int{http.StatusCreated: 1, http.StatusNoContent: requests - 1}, statusCounts)
	require.Equal(t, 1, countEventRows(t)["transfers"])
}

func TestEvents_ConcurrentOppositeTransfersBetweenNewAccountsDoNotDeadlock(t *testing.T) {
	const pairs = 100

	clearEventTables(t)

	statuses := make(chan int, 2*pairs)

	var wg sync.WaitGroup

	for pair := range pairs {
		accountA := fmt.Sprintf("account-a-%d", pair)
		accountB := fmt.Sprintf("account-b-%d", pair)

		wg.Go(func() {
			path := fmt.Sprintf("%sforward-%d", transfersPath, pair)
			statuses <- sendEvent(t.Context(), path, transferBetween(accountA, accountB))
		})
		wg.Go(func() {
			path := fmt.Sprintf("%sbackward-%d", transfersPath, pair)
			statuses <- sendEvent(t.Context(), path, transferBetween(accountB, accountA))
		})
	}

	wg.Wait()
	close(statuses)

	for status := range statuses {
		require.Equal(t, http.StatusCreated, status)
	}

	require.Equal(t, 2*pairs, countEventRows(t)["transfers"])
}

func transferBetween(fromAccountID, toAccountID string) handler.TransferPayload {
	transfer := newTransfer()
	transfer.FromAccountID = fromAccountID
	transfer.FromClientID = "client-of-" + fromAccountID
	transfer.ToAccountID = toAccountID
	transfer.ToClientID = "client-of-" + toAccountID

	return transfer
}

func putEvent(t *testing.T, path string, body any) int {
	t.Helper()

	status := sendEvent(t.Context(), path, body)
	require.NotZero(t, status, "request failed")

	return status
}

func sendEvent(ctx context.Context, path string, body any) int {
	payload, err := json.Marshal(body)
	if err != nil {
		return 0
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, serverURL+path, bytes.NewReader(payload))
	if err != nil {
		return 0
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode
}

func clearEventTables(t *testing.T) {
	t.Helper()

	_, err := testPool.Exec(t.Context(), "TRUNCATE groups, clients, devices, network_addresses, events CASCADE")
	require.NoError(t, err)
}

func countEventRows(t *testing.T) map[string]int {
	t.Helper()

	rows := make(map[string]int, len(eventTables))

	for _, table := range eventTables {
		var count int

		err := testPool.QueryRow(t.Context(), "SELECT count(*) FROM "+table).Scan(&count)
		require.NoError(t, err)

		rows[table] = count
	}

	return rows
}
