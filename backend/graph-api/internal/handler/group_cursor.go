package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

type groupCursor struct {
	DetectedAt time.Time `json:"detected_at"`
	ID         uuid.UUID `json:"id"`
}

func encodeCursor(cursor domain.GroupCursor) string {
	raw, _ := json.Marshal(groupCursor{DetectedAt: cursor.DetectedAt, ID: cursor.ID})

	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeCursor(token string) (domain.GroupCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return domain.GroupCursor{}, err
	}

	var cursor groupCursor
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return domain.GroupCursor{}, err
	}

	if cursor.DetectedAt.IsZero() || cursor.ID == uuid.Nil {
		return domain.GroupCursor{}, errors.New("incomplete cursor")
	}

	return domain.GroupCursor{DetectedAt: cursor.DetectedAt, ID: cursor.ID}, nil
}
