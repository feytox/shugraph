package service

import (
	"context"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

type eventRepository interface {
	SaveLogin(ctx context.Context, event domain.LoginEvent) (created bool, err error)
	SaveTransfer(ctx context.Context, event domain.TransferEvent) (created bool, err error)
}

type EventService struct {
	eventRepository eventRepository
}

func NewEventService(eventRepository eventRepository) *EventService {
	return &EventService{eventRepository: eventRepository}
}

func (s *EventService) AcceptLogin(ctx context.Context, event domain.LoginEvent) (created bool, err error) {
	return s.eventRepository.SaveLogin(ctx, event)
}

func (s *EventService) AcceptTransfer(ctx context.Context, event domain.TransferEvent) (created bool, err error) {
	if event.FromAccountID == event.ToAccountID {
		return false, domain.ErrSameAccounts
	}

	return s.eventRepository.SaveTransfer(ctx, event)
}
