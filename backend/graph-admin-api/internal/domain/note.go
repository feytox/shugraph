package domain

import (
	"errors"
	"time"
)

var ErrGraphAPIUnavailable = errors.New("graph-api is unavailable")

type ProxiedNote struct {
	ID          int64
	GraphNoteID string
	Text        string
	CreatedAt   time.Time
}
