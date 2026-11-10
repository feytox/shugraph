package domain

import (
	"time"
)

type Scenario struct {
	Name   string          `json:"name"`
	Events []ScenarioEvent `json:"events"`
}

type ScenarioEvent struct {
	EventID    string    `json:"event_id"`
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
	Amount     int64     `json:"amount"`
}
