package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type EventHandler struct{}

func (h *EventHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "login-event",
		Summary:       "Create an login event",
		Method:        http.MethodPut,
		Path:          "/api/v1/events/logins/{event_id}",
		DefaultStatus: http.StatusCreated,
	}, h.Login)

	huma.Register(api, huma.Operation{
		OperationID:   "transfer-event",
		Summary:       "Create an transfer event",
		Method:        http.MethodPut,
		Path:          "/api/v1/events/transfers/{event_id}",
		DefaultStatus: http.StatusCreated,
	}, h.Transfer)
}

type LoginInput struct {
	EventID string `path:"event_id"`

	Body struct {
		OccurredAt time.Time

		AccountID string
		ClientID  string
		DeviceID  string
		AddressID string
	}
}

type TransferInput struct {
	EventID string `path:"event_id"`

	Body struct {
		OccurredAt time.Time

		FromAccountID string
		FromClientID  string
		ToAccountID   string
		ToClientID    string
		Amount        int64
	}
}

func (h *EventHandler) Login(_ context.Context, _ *LoginInput) (*struct{}, error) {
	return nil, nil
}

func (h *EventHandler) Transfer(_ context.Context, _ *TransferInput) (*struct{}, error) {
	return nil, nil
}
