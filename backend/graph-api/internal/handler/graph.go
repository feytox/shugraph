package handler

import (
	"time"
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
