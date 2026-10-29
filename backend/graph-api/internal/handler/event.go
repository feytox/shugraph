package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

const eventsTag = "Events"

type eventHandlers struct{}

func registerEvents(api huma.API, h *eventHandlers) {
	const ingestDescription = `Event IDs are globally unique across both event types.
New events return 201 with no body.
An identical retry returns 204 without recomputing groups;
an existing ID with different type or payload returns 409.`

	huma.Register(api, huma.Operation{
		OperationID:   "login-event",
		Summary:       "Accept a login event",
		Method:        http.MethodPut,
		Path:          "/api/v1/events/logins/{event_id}",
		Tags:          []string{eventsTag},
		DefaultStatus: http.StatusCreated,
		Description:   ingestDescription,
		Errors: []int{
			http.StatusBadRequest,
			http.StatusConflict,
			http.StatusUnprocessableEntity,
			http.StatusNotImplemented,
		},
		Responses: map[string]*huma.Response{
			"204": {Description: "Identical event already accepted; no changes"},
		},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID:   "transfer-event",
		Summary:       "Accept a transfer event",
		Method:        http.MethodPut,
		Path:          "/api/v1/events/transfers/{event_id}",
		Tags:          []string{eventsTag},
		DefaultStatus: http.StatusCreated,
		Description:   ingestDescription,
		Errors: []int{
			http.StatusBadRequest,
			http.StatusConflict,
			http.StatusUnprocessableEntity,
			http.StatusNotImplemented,
		},
		Responses: map[string]*huma.Response{
			"204": {Description: "Identical event already accepted; no changes"},
		},
	}, h.transfer)
}

type Account struct {
	ID       string `json:"id" minLength:"1"`
	ClientID string `json:"client_id" minLength:"1"`
}
type LoginPayload struct {
	OccurredAt time.Time `json:"occurred_at"`
	AccountID  string    `json:"account_id" minLength:"1"`
	ClientID   string    `json:"client_id" minLength:"1"`
	DeviceID   string    `json:"device_id" minLength:"1"`
	AddressID  string    `json:"address_id" minLength:"1" doc:"Source IP/network address; retained as a source identifier."`
}
type TransferPayload struct {
	OccurredAt    time.Time `json:"occurred_at"`
	FromAccountID string    `json:"from_account_id" minLength:"1"`
	FromClientID  string    `json:"from_client_id" minLength:"1"`
	ToAccountID   string    `json:"to_account_id" minLength:"1"`
	ToClientID    string    `json:"to_client_id" minLength:"1"`
	Amount        int64     `json:"amount" minimum:"1" doc:"Amount in kopecks."`
}
type LoginInput struct {
	EventID string `path:"event_id" minLength:"1"`
	Body    LoginPayload
}

type TransferInput struct {
	EventID string `path:"event_id" minLength:"1"`
	Body    TransferPayload
}
type LoginEvent struct {
	EventID string `json:"event_id" minLength:"1"`
	LoginPayload
}
type TransferEvent struct {
	EventID string `json:"event_id" minLength:"1"`
	TransferPayload
}

func (h *eventHandlers) login(context.Context, *LoginInput) (*struct{}, error) {
	return nil, huma.Error501NotImplemented("Event ingestion is not implemented yet.")
}
func (h *eventHandlers) transfer(context.Context, *TransferInput) (*struct{}, error) {
	return nil, huma.Error501NotImplemented("Event ingestion is not implemented yet.")
}
