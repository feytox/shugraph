package service

import (
	"context"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

type eventRepository interface {
	SaveLogin(ctx context.Context, event domain.LoginEvent) (created bool, err error)
	SaveTransfer(ctx context.Context, event domain.TransferEvent) (created bool, err error)
}

type Repositories struct {
	Events eventRepository
	Groups groupAssemblyRepository
	Farms  farmRepository
}

type Transactor func(ctx context.Context, fn func(Repositories) error) error

type EventService struct {
	transactor Transactor
	rules      domain.GroupRules
}

func NewEventService(transactor Transactor, rules domain.GroupRules) *EventService {
	return &EventService{transactor: transactor, rules: rules}
}

func (s *EventService) AcceptLogin(ctx context.Context, event domain.LoginEvent) (created bool, err error) {
	err = s.transactor(ctx, func(r Repositories) error {
		created, err = r.Events.SaveLogin(ctx, event)
		if err != nil || !created {
			return err
		}

		return s.assembly(r).afterLogin(ctx, event)
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

	err = s.transactor(ctx, func(r Repositories) error {
		created, err = r.Events.SaveTransfer(ctx, event)
		if err != nil || !created {
			return err
		}

		return s.assembly(r).afterTransfer(ctx, event)
	})
	if err != nil {
		return false, err
	}

	return created, nil
}

func (s *EventService) assembly(r Repositories) groupAssembly {
	return groupAssembly{rules: s.rules, groups: r.Groups, farms: r.Farms}
}
