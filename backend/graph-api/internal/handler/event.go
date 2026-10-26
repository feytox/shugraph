package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type EventHandler struct{}

func (h *EventHandler) Register(api huma.API) {
	const ingest = "Event IDs are globally unique across both event types. New events return 201 with no body. An identical retry returns 204 without recomputing groups; an existing ID with different type or payload returns 409. Compare parsed values, not JSON formatting. Persistence/deduplication are implemented separately. Amounts are integer kopecks."
	for _, op := range []struct {
		id, path, summary string
		login             bool
	}{
		{"login-event", "/api/v1/events/logins/{event_id}", "Accept a login event", true},
		{"transfer-event", "/api/v1/events/transfers/{event_id}", "Accept a transfer event", false},
	} {
		operation := huma.Operation{OperationID: op.id, Summary: op.summary, Method: http.MethodPut,
			Path: op.path, DefaultStatus: http.StatusCreated, Description: ingest,
			Errors: []int{400, 409, 422, 501}, Responses: map[string]*huma.Response{"204": {Description: "Identical event already accepted; no changes"}},
		}
		if op.login {
			huma.Register(api, operation, h.Login)
		} else {
			huma.Register(api, operation, h.Transfer)
		}
	}
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
	Type    string `json:"type" enum:"login"`
	LoginPayload
}
type TransferEvent struct {
	EventID string `json:"event_id" minLength:"1"`
	Type    string `json:"type" enum:"transfer"`
	TransferPayload
}

// Event contains exactly one typed evidence event, flattened on the wire.
type Event struct {
	Login    *LoginEvent
	Transfer *TransferEvent
}

func (Event) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{OneOf: []*huma.Schema{
		r.Schema(reflect.TypeFor[LoginEvent](), true, ""),
		r.Schema(reflect.TypeFor[TransferEvent](), true, ""),
	}}
}
func (e Event) MarshalJSON() ([]byte, error) {
	if (e.Login == nil) == (e.Transfer == nil) {
		return nil, fmt.Errorf("event requires exactly one payload")
	}
	if e.Login != nil {
		return json.Marshal(e.Login)
	}
	return json.Marshal(e.Transfer)
}
func (h *EventHandler) Login(context.Context, *LoginInput) (*struct{}, error) {
	return nil, huma.Error501NotImplemented("Event ingestion is not implemented yet.")
}
func (h *EventHandler) Transfer(context.Context, *TransferInput) (*struct{}, error) {
	return nil, huma.Error501NotImplemented("Event ingestion is not implemented yet.")
}
