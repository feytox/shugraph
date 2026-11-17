package handler

import (
	"time"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/service"
)

const (
	nodeClient         = "client"
	nodeAccount        = "account"
	nodeDevice         = "device"
	nodeNetworkAddress = "network_address"
)

type GroupGraph struct {
	GroupID    string          `json:"group_id" format:"uuid"`
	Nodes      []GraphNode     `json:"nodes" nullable:"false"`
	Ownerships []OwnershipEdge `json:"ownerships" nullable:"false"`
	Logins     []LoginEdge     `json:"logins" nullable:"false"`
	Transfers  []TransferEdge  `json:"transfers" nullable:"false"`
}

type GraphNode struct {
	ID       string `json:"id" minLength:"1" doc:"Type-prefixed ID, e.g. account:acc-1."`
	Type     string `json:"type" enum:"client,account,device,network_address"`
	EntityID string `json:"entity_id" minLength:"1" doc:"Original source identifier."`
}

type EdgeEndpoints struct {
	ID     string `json:"id" minLength:"1" doc:"Unique type-prefixed edge ID."`
	Source string `json:"source" minLength:"1"`
	Target string `json:"target" minLength:"1"`
}

type OwnershipEdge struct {
	EdgeEndpoints
}

type LoginEdge struct {
	EdgeEndpoints
	EventID    string    `json:"event_id" minLength:"1"`
	OccurredAt time.Time `json:"occurred_at"`
}

type TransferEdge struct {
	EdgeEndpoints
	EventID    string    `json:"event_id" minLength:"1"`
	OccurredAt time.Time `json:"occurred_at"`
	Amount     int64     `json:"amount" minimum:"1" doc:"Amount in kopecks."`
}

func groupGraph(details service.GroupDetails) GroupGraph {
	accounts := graphAccounts(details)
	clients := uniqueBy(accounts, func(a domain.Account) string { return a.ClientID })
	devices := uniqueBy(details.Logins, func(l domain.LoginEvent) string { return l.DeviceID })
	addresses := uniqueBy(details.Logins, func(l domain.LoginEvent) string { return l.AddressID })

	graph := GroupGraph{
		GroupID:    details.Group.ID.String(),
		Nodes:      make([]GraphNode, 0, len(clients)+len(accounts)+len(devices)+len(addresses)),
		Ownerships: make([]OwnershipEdge, 0, len(accounts)),
		Logins:     make([]LoginEdge, 0, 2*len(details.Logins)),
		Transfers:  make([]TransferEdge, 0, len(details.Transfers)),
	}

	for _, client := range clients {
		graph.Nodes = append(graph.Nodes, graphNode(nodeClient, client.ClientID))
	}

	for _, account := range accounts {
		graph.Nodes = append(graph.Nodes, graphNode(nodeAccount, account.ID))
	}

	for _, login := range devices {
		graph.Nodes = append(graph.Nodes, graphNode(nodeDevice, login.DeviceID))
	}

	for _, login := range addresses {
		graph.Nodes = append(graph.Nodes, graphNode(nodeNetworkAddress, login.AddressID))
	}

	for _, account := range accounts {
		graph.Ownerships = append(graph.Ownerships, OwnershipEdge{EdgeEndpoints{
			ID:     "ownership:" + account.ClientID + ":" + account.ID,
			Source: nodeID(nodeClient, account.ClientID),
			Target: nodeID(nodeAccount, account.ID),
		}})
	}

	for _, login := range details.Logins {
		graph.Logins = append(graph.Logins,
			loginEdge(login, "device", nodeID(nodeDevice, login.DeviceID)),
			loginEdge(login, "address", nodeID(nodeNetworkAddress, login.AddressID)),
		)
	}

	for _, transfer := range details.Transfers {
		graph.Transfers = append(graph.Transfers, TransferEdge{
			EdgeEndpoints: EdgeEndpoints{
				ID:     "transfer:" + transfer.EventID,
				Source: nodeID(nodeAccount, transfer.FromAccountID),
				Target: nodeID(nodeAccount, transfer.ToAccountID),
			},
			EventID:    transfer.EventID,
			OccurredAt: transfer.OccurredAt,
			Amount:     transfer.Amount,
		})
	}

	return graph
}

func graphAccounts(details service.GroupDetails) []domain.Account {
	accounts := make([]domain.Account, 0, len(details.Members)+len(details.Logins)+2*len(details.Transfers))
	accounts = append(accounts, details.Members...)

	for _, login := range details.Logins {
		accounts = append(accounts, domain.Account{ID: login.AccountID, ClientID: login.ClientID})
	}

	for _, transfer := range details.Transfers {
		accounts = append(accounts,
			domain.Account{ID: transfer.FromAccountID, ClientID: transfer.FromClientID},
			domain.Account{ID: transfer.ToAccountID, ClientID: transfer.ToClientID},
		)
	}

	return uniqueBy(accounts, func(a domain.Account) string { return a.ID })
}

func uniqueBy[T any](items []T, key func(T) string) []T {
	seen := make(map[string]bool, len(items))
	result := make([]T, 0, len(items))

	for _, item := range items {
		k := key(item)
		if seen[k] {
			continue
		}

		seen[k] = true

		result = append(result, item)
	}

	return result
}

func nodeID(nodeType, entityID string) string {
	return nodeType + ":" + entityID
}

func graphNode(nodeType, entityID string) GraphNode {
	return GraphNode{ID: nodeID(nodeType, entityID), Type: nodeType, EntityID: entityID}
}

func loginEdge(login domain.LoginEvent, target, targetID string) LoginEdge {
	return LoginEdge{
		EdgeEndpoints: EdgeEndpoints{
			ID:     "login:" + login.EventID + ":" + target,
			Source: nodeID(nodeAccount, login.AccountID),
			Target: targetID,
		},
		EventID:    login.EventID,
		OccurredAt: login.OccurredAt,
	}
}
