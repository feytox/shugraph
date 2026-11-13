package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

const eventsTag = "Events"

type EventService interface {
	AcceptLogin(ctx context.Context, event domain.LoginEvent) (created bool, err error)
	AcceptTransfer(ctx context.Context, event domain.TransferEvent) (created bool, err error)
}

type eventHandlers struct {
	events EventService
	log    *slog.Logger
}

func registerEvents(api huma.API, h *eventHandlers) {
	huma.Register(api, huma.Operation{
		OperationID:   "login-event",
		Summary:       "Accept a login event",
		Method:        http.MethodPut,
		Path:          "/api/v1/events/logins/{event_id}",
		Tags:          []string{eventsTag},
		DefaultStatus: http.StatusCreated,
		Errors: []int{
			http.StatusBadRequest,
			http.StatusConflict,
			http.StatusUnprocessableEntity,
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
		Errors: []int{
			http.StatusBadRequest,
			http.StatusConflict,
			http.StatusUnprocessableEntity,
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

type EventAcceptedOutput struct {
	Status int
}

func (h *eventHandlers) login(ctx context.Context, in *LoginInput) (*EventAcceptedOutput, error) {
	created, err := h.events.AcceptLogin(ctx, domain.LoginEvent{
		EventID:    in.EventID,
		OccurredAt: in.Body.OccurredAt,
		AccountID:  in.Body.AccountID,
		ClientID:   in.Body.ClientID,
		DeviceID:   in.Body.DeviceID,
		AddressID:  in.Body.AddressID,
	})
	if err != nil {
		return nil, h.eventProblem(ctx, err)
	}

	return eventAccepted(created), nil
}

func (h *eventHandlers) transfer(ctx context.Context, in *TransferInput) (*EventAcceptedOutput, error) {
	created, err := h.events.AcceptTransfer(ctx, domain.TransferEvent{
		EventID:       in.EventID,
		OccurredAt:    in.Body.OccurredAt,
		FromAccountID: in.Body.FromAccountID,
		FromClientID:  in.Body.FromClientID,
		ToAccountID:   in.Body.ToAccountID,
		ToClientID:    in.Body.ToClientID,
		Amount:        in.Body.Amount,
	})
	if err != nil {
		return nil, h.eventProblem(ctx, err)
	}

	return eventAccepted(created), nil
}

func eventAccepted(created bool) *EventAcceptedOutput {
	if created {
		return &EventAcceptedOutput{Status: http.StatusCreated}
	}

	return &EventAcceptedOutput{Status: http.StatusNoContent}
}

func (h *eventHandlers) eventProblem(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrEventConflict):
		return huma.Error409Conflict(domain.ErrEventConflict.Error())
	case errors.Is(err, domain.ErrAccountOwnerMismatch):
		return huma.Error422UnprocessableEntity(domain.ErrAccountOwnerMismatch.Error())
	case errors.Is(err, domain.ErrSameAccounts):
		return huma.Error422UnprocessableEntity(domain.ErrSameAccounts.Error())
	default:
		return problem(ctx, h.log, err)
	}
}
