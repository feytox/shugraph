package domain

import (
	"encoding/json"
	"errors"
)

const (
	EventTypeLogin    = "login"
	EventTypeTransfer = "transfer"
)

type Scenario struct {
	Name   string            `json:"name"`
	Events []json.RawMessage `json:"events"`
}

var ErrScenarioNotFound = errors.New("scenario not found")
