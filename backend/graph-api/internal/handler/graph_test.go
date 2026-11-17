package handler

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/service"
)

func TestGroupGraph_Farm(t *testing.T) {
	groupID := uuid.MustParse("6f1c2a3e-1b7d-4c1a-9e51-0a1b2c3d4e03")
	at := func(minute int) time.Time { return time.Date(2026, 11, 12, 10, minute, 0, 0, time.UTC) }

	graph := groupGraph(service.GroupDetails{
		Group: domain.Group{ID: groupID, Scheme: domain.GroupSchemeFarm},
		Members: []domain.Account{
			{ID: "acc-collector", ClientID: "cli-organizer"},
			{ID: "acc-ravil", ClientID: "cli-ravil"},
			{ID: "acc-sasha", ClientID: "cli-sasha"},
		},
		Logins: []domain.LoginEvent{
			{EventID: "login-1", OccurredAt: at(1), AccountID: "acc-ravil", ClientID: "cli-ravil", DeviceID: "dev-1", AddressID: "10.0.0.1"},
			{EventID: "login-2", OccurredAt: at(2), AccountID: "acc-sasha", ClientID: "cli-sasha", DeviceID: "dev-1", AddressID: "10.0.0.2"},
			{EventID: "login-3", OccurredAt: at(3), AccountID: "acc-sasha", ClientID: "cli-sasha", DeviceID: "dev-1", AddressID: "10.0.0.1"},
		},
		Transfers: []domain.TransferEvent{
			{
				EventID: "transfer-1", OccurredAt: at(4), Amount: 5000,
				FromAccountID: "acc-ravil", FromClientID: "cli-ravil",
				ToAccountID: "acc-collector", ToClientID: "cli-organizer",
			},
		},
	})

	edge := func(id, source, target string) EdgeEndpoints {
		return EdgeEndpoints{ID: id, Source: source, Target: target}
	}

	require.Equal(t, GroupGraph{
		GroupID: groupID.String(),
		Nodes: []GraphNode{
			{ID: "client:cli-organizer", Type: "client", EntityID: "cli-organizer"},
			{ID: "client:cli-ravil", Type: "client", EntityID: "cli-ravil"},
			{ID: "client:cli-sasha", Type: "client", EntityID: "cli-sasha"},
			{ID: "account:acc-collector", Type: "account", EntityID: "acc-collector"},
			{ID: "account:acc-ravil", Type: "account", EntityID: "acc-ravil"},
			{ID: "account:acc-sasha", Type: "account", EntityID: "acc-sasha"},
			{ID: "device:dev-1", Type: "device", EntityID: "dev-1"},
			{ID: "network_address:10.0.0.1", Type: "network_address", EntityID: "10.0.0.1"},
			{ID: "network_address:10.0.0.2", Type: "network_address", EntityID: "10.0.0.2"},
		},
		Ownerships: []OwnershipEdge{
			{edge("ownership:cli-organizer:acc-collector", "client:cli-organizer", "account:acc-collector")},
			{edge("ownership:cli-ravil:acc-ravil", "client:cli-ravil", "account:acc-ravil")},
			{edge("ownership:cli-sasha:acc-sasha", "client:cli-sasha", "account:acc-sasha")},
		},
		Logins: []LoginEdge{
			{edge("login:login-1:device", "account:acc-ravil", "device:dev-1"), "login-1", at(1)},
			{edge("login:login-1:address", "account:acc-ravil", "network_address:10.0.0.1"), "login-1", at(1)},
			{edge("login:login-2:device", "account:acc-sasha", "device:dev-1"), "login-2", at(2)},
			{edge("login:login-2:address", "account:acc-sasha", "network_address:10.0.0.2"), "login-2", at(2)},
			{edge("login:login-3:device", "account:acc-sasha", "device:dev-1"), "login-3", at(3)},
			{edge("login:login-3:address", "account:acc-sasha", "network_address:10.0.0.1"), "login-3", at(3)},
		},
		Transfers: []TransferEdge{
			{edge("transfer:transfer-1", "account:acc-ravil", "account:acc-collector"), "transfer-1", at(4), 5000},
		},
	}, graph)
}
