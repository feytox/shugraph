package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type GroupScheme string

const (
	GroupSchemeFarm    GroupScheme = "farm"
	GroupSchemeAddress GroupScheme = "address"
	GroupSchemeChain   GroupScheme = "chain"
	GroupSchemeRing    GroupScheme = "ring"
)

type GroupRules struct {
	Window time.Duration
	Farm   FarmRules
}

type TimeWindow struct {
	Start time.Time
	End   time.Time
}

func (w TimeWindow) Contains(t time.Time) bool {
	return !t.Before(w.Start) && !t.After(w.End)
}

type GroupStats struct {
	FirstEventAt  time.Time
	LastEventAt   time.Time
	AccountsCount int
	TotalAmount   int64
}

type Risk struct {
	Score int
	Parts json.RawMessage
}

type GroupStatus string

const (
	GroupStatusNew      GroupStatus = "new"
	GroupStatusApproved GroupStatus = "approved"
	GroupStatusDeclined GroupStatus = "declined"
)

type Group struct {
	ID            uuid.UUID
	Scheme        GroupScheme
	Status        GroupStatus
	RiskScore     int
	RiskParts     json.RawMessage
	DetectedAt    time.Time
	FirstEventAt  time.Time
	LastEventAt   time.Time
	AccountsCount int
	TotalAmount   int64
	IsFrozen      bool
	Decision      *Decision
}

type Decision struct {
	Comment   string
	DecidedAt time.Time
}

type GroupCursor struct {
	DetectedAt time.Time
	ID         uuid.UUID
}

type GroupFilter struct {
	Status  *GroupStatus
	Schemes []GroupScheme
	After   *GroupCursor
	Limit   int32
}
