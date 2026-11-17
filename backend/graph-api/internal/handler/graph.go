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

	return GroupGraph{
		GroupID:    details.Group.ID.String(),
		Nodes:      graphNodes(accounts, details.Logins),
		Ownerships: ownershipEdges(accounts),
		Logins:     loginEdges(details.Logins),
		Transfers:  transferEdges(details.Transfers),
	}
}

func graphNodes(accounts []domain.Account, logins []domain.LoginEvent) []GraphNode {
	clients := uniqueBy(accounts, func(a domain.Account) string { return a.ClientID })
	devices := uniqueBy(logins, func(l domain.LoginEvent) string { return l.DeviceID })
	addresses := uniqueBy(logins, func(l domain.LoginEvent) string { return l.AddressID })

	nodes := make([]GraphNode, 0, len(clients)+len(accounts)+len(devices)+len(addresses))

	for _, client := range clients {
		nodes = append(nodes, graphNode(nodeClient, client.ClientID))
	}

	for _, account := range accounts {
		nodes = append(nodes, graphNode(nodeAccount, account.ID))
	}

	for _, login := range devices {
		nodes = append(nodes, graphNode(nodeDevice, login.DeviceID))
	}

	for _, login := range addresses {
		nodes = append(nodes, graphNode(nodeNetworkAddress, login.AddressID))
	}

	return nodes
}

func ownershipEdges(accounts []domain.Account) []OwnershipEdge {
	edges := make([]OwnershipEdge, 0, len(accounts))

	for _, account := range accounts {
		edges = append(edges, OwnershipEdge{EdgeEndpoints{
			ID:     "ownership:" + account.ClientID + ":" + account.ID,
			Source: nodeID(nodeClient, account.ClientID),
			Target: nodeID(nodeAccount, account.ID),
		}})
	}

	return edges
}

func loginEdges(logins []domain.LoginEvent) []LoginEdge {
	edges := make([]LoginEdge, 0, 2*len(logins))

	for _, login := range logins {
		edges = append(edges,
			loginEdge(login, "device", nodeID(nodeDevice, login.DeviceID)),
			loginEdge(login, "address", nodeID(nodeNetworkAddress, login.AddressID)),
		)
	}

	return edges
}

func transferEdges(transfers []domain.TransferEvent) []TransferEdge {
	edges := make([]TransferEdge, 0, len(transfers))

	for _, transfer := range transfers {
		edges = append(edges, TransferEdge{
			ID:         "transfer:" + transfer.EventID,
			Source:     nodeID(nodeAccount, transfer.FromAccountID),
			Target:     nodeID(nodeAccount, transfer.ToAccountID),
			EventID:    transfer.EventID,
			OccurredAt: transfer.OccurredAt,
			Amount:     transfer.Amount,
		})
	}

	return edges
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
		ID:         "login:" + login.EventID + ":" + target,
		Source:     nodeID(nodeAccount, login.AccountID),
		Target:     targetID,
		EventID:    login.EventID,
		OccurredAt: login.OccurredAt,
	}
}
