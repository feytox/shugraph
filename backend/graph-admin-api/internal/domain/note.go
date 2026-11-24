package domain

import (
	"time"
)

type ProxiedNote struct {
	ID          int64
	GraphNoteID string
	Text        string
	CreatedAt   time.Time
}
