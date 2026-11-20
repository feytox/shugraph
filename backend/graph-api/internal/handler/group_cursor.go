package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

var errInvalidCursor = errors.New("invalid cursor")

type groupCursor struct {
	DetectedAt time.Time `json:"detected_at"`
	ID         uuid.UUID `json:"id"`
}

type cursorParam struct {
	after *domain.GroupCursor
}

func (p *cursorParam) UnmarshalText(text []byte) error {
	raw, err := base64.RawURLEncoding.DecodeString(string(text))
	if err != nil {
		return errInvalidCursor
	}

	var cursor groupCursor
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return errInvalidCursor
	}

	if cursor.DetectedAt.IsZero() || cursor.ID == uuid.Nil {
		return errInvalidCursor
	}

	p.after = &domain.GroupCursor{DetectedAt: cursor.DetectedAt, ID: cursor.ID}

	return nil
}

func encodeCursor(cursor domain.GroupCursor) string {
	raw, _ := json.Marshal(groupCursor{DetectedAt: cursor.DetectedAt, ID: cursor.ID})

	return base64.RawURLEncoding.EncodeToString(raw)
}
