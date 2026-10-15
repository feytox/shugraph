package domain

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID        uuid.UUID
	Text      string
	CreatedAt time.Time
}
