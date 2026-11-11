package service

import (
	"context"
	"fmt"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

type eventRepository interface {
	SaveLogin(ctx context.Context, event domain.LoginEvent) (created bool, err error)
	SaveTransfer(ctx context.Context, event domain.TransferEvent) (created bool, err error)
}

type EventRepositories struct {
	Events eventRepository
	Groups groupAssemblyRepository
	Farms  farmRepository
}

type EventService struct {
	transactor Transactor[EventRepositories]
	rules      domain.GroupRules
}

func NewEventService(transactor Transactor[EventRepositories], rules domain.GroupRules) *EventService {
	return &EventService{transactor: transactor, rules: rules}
}

func (s *EventService) AcceptLogin(ctx context.Context, event domain.LoginEvent) (created bool, err error) {
	err = s.transactor(ctx, func(r EventRepositories) error {
		created, err = r.Events.SaveLogin(ctx, event)
		if err != nil {
			return fmt.Errorf("save login %s: %w", event.EventID, err)
		}

		if !created {
			return nil
		}

		if err := s.assembly(r).afterLogin(ctx, event); err != nil {
			return fmt.Errorf("assemble groups after login %s: %w", event.EventID, err)
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return created, nil
}

func (s *EventService) AcceptTransfer(ctx context.Context, event domain.TransferEvent) (created bool, err error) {
	if event.FromAccountID == event.ToAccountID {
		return false, domain.ErrSameAccounts
	}

	err = s.transactor(ctx, func(r EventRepositories) error {
		created, err = r.Events.SaveTransfer(ctx, event)
		if err != nil {
			return fmt.Errorf("save transfer %s: %w", event.EventID, err)
		}

		if !created {
			return nil
		}

		if err := s.assembly(r).afterTransfer(ctx, event); err != nil {
			return fmt.Errorf("assemble groups after transfer %s: %w", event.EventID, err)
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return created, nil
}

func (s *EventService) assembly(r EventRepositories) groupAssembly {
	return groupAssembly{rules: s.rules, groups: r.Groups, farms: r.Farms}
}
