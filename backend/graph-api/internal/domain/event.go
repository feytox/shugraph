package domain

import (
	"errors"
	"time"
)

type EventType string

const (
	EventTypeLogin    EventType = "login"
	EventTypeTransfer EventType = "transfer"
)

type LoginEvent struct {
	EventID    string
	OccurredAt time.Time
	AccountID  string
	ClientID   string
	DeviceID   string
	AddressID  string
}

type TransferEvent struct {
	EventID       string
	OccurredAt    time.Time
	FromAccountID string
	FromClientID  string
	ToAccountID   string
	ToClientID    string
	Amount        int64
}

var (
	ErrEventConflict        = errors.New("event id already used with different fields")
	ErrAccountOwnerMismatch = errors.New("account belongs to another client")
	ErrSameAccounts         = errors.New("transfer accounts must differ")
)
