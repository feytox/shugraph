package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/graphapi/oapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validScenarioJSON = `{
  "name": "test_scenario",
  "events": [
    {
      "event_id": "evt-login-1",
      "type": "login",
      "occurred_at": "2026-11-12T13:20:00Z",
      "account_id": "acc-1",
      "client_id": "cli-1",
      "device_id": "dev-1",
      "address_id": "192.168.1.1"
    },
    {
      "event_id": "evt-transfer-1",
      "type": "transfer",
      "occurred_at": "2026-11-12T13:30:00Z",
      "from_account_id": "acc-1",
      "from_client_id": "cli-1",
      "to_account_id": "acc-2",
      "to_client_id": "cli-2",
      "amount": 50000
    }
  ]
}`

const invalidScenarioJSON = `{
  "name": "test_invalid_scenario",
  "events": [
    {
      "event_id": "evt-exit-1",
      "type": "exit",
      "occurred_at": "2026-11-12T13:20:00Z",
      "account_id": "acc-1",
      "client_id": "cli-1",
      "device_id": "dev-1",
      "address_id": "192.168.1.1"
    }
  ]
}`

type eventAPIMock struct {
	login    func(ctx context.Context, eventID string, payload oapi.LoginPayload) error
	transfer func(ctx context.Context, eventID string, payload oapi.TransferPayload) error
}

func (m eventAPIMock) LoginEvent(ctx context.Context, eventID string, payload oapi.LoginPayload) error {
	if m.login != nil {
		return m.login(ctx, eventID, payload)
	}

	return nil
}

func (m eventAPIMock) TransferEvent(ctx context.Context, eventID string, payload oapi.TransferPayload) error {
	if m.transfer != nil {
		return m.transfer(ctx, eventID, payload)
	}

	return nil
}

type scenarioRepositoryMock struct {
	scenarioRepository
	getScenarios func(name string) (*domain.Scenario, error)
}

func (r scenarioRepositoryMock) GetByName(name string) (*domain.Scenario, error) {
	if r.getScenarios != nil {
		return r.getScenarios(name)
	}

	return nil, nil
}

func TestScenarioServiceRunSuccess(t *testing.T) {
	domainScenario := testScenario(t, validScenarioJSON)

	var wg sync.WaitGroup
	wg.Add(2)

	api := &eventAPIMock{
		login: func(ctx context.Context, eventID string, payload oapi.LoginPayload) error {
			assert.Equal(t, "evt-login-1", eventID)
			assert.Equal(t, "acc-1", payload.AccountId)
			assert.Equal(t, "cli-1", payload.ClientId)
			assert.Equal(t, "dev-1", payload.DeviceId)
			assert.Equal(t, "192.168.1.1", payload.AddressId)
			wg.Done()

			return nil
		},
		transfer: func(ctx context.Context, eventID string, payload oapi.TransferPayload) error {
			assert.Equal(t, "evt-transfer-1", eventID)
			assert.Equal(t, "acc-1", payload.FromAccountId)
			assert.Equal(t, "cli-1", payload.FromClientId)
			assert.Equal(t, "acc-2", payload.ToAccountId)
			assert.Equal(t, "cli-2", payload.ToClientId)
			assert.Equal(t, int64(50000), payload.Amount)
			wg.Done()

			return nil
		},
	}

	repo := scenarioRepositoryMock{
		getScenarios: func(name string) (*domain.Scenario, error) {
			assert.Equal(t, domainScenario.Name, name)

			return domainScenario, nil
		},
	}

	logger := slog.New(slog.DiscardHandler)
	svc := NewScenarioService(api, repo, logger)
	t.Cleanup(svc.Stop)

	err := svc.Run(domainScenario.Name, 0)
	require.NoError(t, err)

	wg.Wait()
}

func TestScenarioServiceRunInvalidType(t *testing.T) {
	domainScenario := testScenario(t, invalidScenarioJSON)
	api := &eventAPIMock{}
	repo := scenarioRepositoryMock{
		getScenarios: func(name string) (*domain.Scenario, error) {
			return domainScenario, nil
		},
	}
	logger := slog.New(slog.DiscardHandler)
	svc := NewScenarioService(api, repo, logger)
	t.Cleanup(svc.Stop)

	err := svc.Run(domainScenario.Name, 0)
	require.Error(t, err)
}

func TestScenarioServiceRunStop(t *testing.T) {
	domainScenario := testScenario(t, validScenarioJSON)

	firstEventStarted := make(chan struct{})
	transferCalled := false

	api := &eventAPIMock{
		login: func(ctx context.Context, eventID string, payload oapi.LoginPayload) error {
			close(firstEventStarted)

			return nil
		},
		transfer: func(ctx context.Context, eventID string, payload oapi.TransferPayload) error {
			transferCalled = true

			return nil
		},
	}
	repo := scenarioRepositoryMock{
		getScenarios: func(name string) (*domain.Scenario, error) {
			return domainScenario, nil
		},
	}

	logger := slog.New(slog.DiscardHandler)
	svc := NewScenarioService(api, repo, logger)
	t.Cleanup(svc.Stop)

	err := svc.Run(domainScenario.Name, 1)
	require.NoError(t, err)

	<-firstEventStarted
	svc.Stop()

	assert.False(t, transferCalled)
}

func testScenario(t *testing.T, scenarioJSON string) *domain.Scenario {
	t.Helper()

	var s domain.Scenario

	err := json.Unmarshal([]byte(scenarioJSON), &s)
	require.NoError(t, err)

	return &s
}
