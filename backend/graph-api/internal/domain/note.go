package domain

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	Id        uuid.UUID
	Text      string
	CreatedAt time.Time
}
