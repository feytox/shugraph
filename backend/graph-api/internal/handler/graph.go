package handler

import (
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type GroupGraph struct {
	GroupID string      `json:"group_id" format:"uuid"`
	Nodes   []GraphNode `json:"nodes" nullable:"false"`
	Edges   []GraphEdge `json:"edges" nullable:"false"`
}

type GraphNode struct {
	Data GraphNodeData `json:"data"`
}

type GraphNodeData struct {
	ID       string `json:"id" minLength:"1" doc:"Type-prefixed ID, e.g. account:acc-1."`
	Type     string `json:"type" enum:"client,account,device,network_address"`
	EntityID string `json:"entity_id" minLength:"1" doc:"Original source identifier."`
}

type GraphEdge struct {
	Data GraphEdgeData `json:"data"`
}

type EdgeEndpoints struct {
	ID     string `json:"id" minLength:"1" doc:"Unique type-prefixed edge ID. Login IDs additionally include device/address suffix."`
	Source string `json:"source" minLength:"1"`
	Target string `json:"target" minLength:"1"`
}

type OwnershipEdge struct {
	EdgeEndpoints
	Type string `json:"type" enum:"ownership"`
}

type LoginEdge struct {
	EdgeEndpoints
	Type       string    `json:"type" enum:"login"`
	EventID    string    `json:"event_id" minLength:"1"`
	OccurredAt time.Time `json:"occurred_at"`
}

type TransferEdge struct {
	EdgeEndpoints
	Type       string    `json:"type" enum:"transfer"`
	EventID    string    `json:"event_id" minLength:"1"`
	OccurredAt time.Time `json:"occurred_at"`
	Amount     int64     `json:"amount" minimum:"1" doc:"Amount in kopecks."`
}

// Optional Go fields serve the three wire variants; oneOf enforces their shape.
type GraphEdgeData struct {
	EdgeEndpoints
	Type       string     `json:"type"`
	EventID    string     `json:"event_id,omitempty"`
	OccurredAt *time.Time `json:"occurred_at,omitempty"`
	Amount     *int64     `json:"amount,omitempty"`
}

func (GraphEdgeData) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{OneOf: []*huma.Schema{
		r.Schema(reflect.TypeFor[OwnershipEdge](), true, ""),
		r.Schema(reflect.TypeFor[LoginEdge](), true, ""),
		r.Schema(reflect.TypeFor[TransferEdge](), true, ""),
	}}
}
