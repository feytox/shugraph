package domain

import (
	"encoding/json"
	"errors"
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

type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

type Risk struct {
	Score int
	Parts json.RawMessage
	Level RiskLevel
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
	RiskLevel     RiskLevel
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

type Account struct {
	ID       string
	ClientID string
}

var ErrGroupNotFound = errors.New("group not found")

type GroupCursor struct {
	DetectedAt time.Time
	ID         uuid.UUID
}

func (g Group) SpanMinutes() int {
	return int(g.LastEventAt.Sub(g.FirstEventAt) / time.Minute)
}

func (g Group) Cursor() GroupCursor {
	return GroupCursor{DetectedAt: g.DetectedAt, ID: g.ID}
}

type GroupFilter struct {
	Status  *GroupStatus
	Schemes []GroupScheme
	After   *GroupCursor
	Limit   int
}
