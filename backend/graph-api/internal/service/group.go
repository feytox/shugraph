package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

type groupRepository interface {
	List(ctx context.Context, filter domain.GroupFilter) ([]domain.Group, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Group, error)
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]domain.Account, error)
	ListLogins(ctx context.Context, groupID uuid.UUID) ([]domain.LoginEvent, error)
	ListTransfers(ctx context.Context, groupID uuid.UUID) ([]domain.TransferEvent, error)
}

type GroupRepositories struct {
	Groups groupRepository
}

type GroupPage struct {
	Groups []domain.Group
	Next   *domain.GroupCursor
}

type GroupDetails struct {
	Group     domain.Group
	Members   []domain.Account
	Logins    []domain.LoginEvent
	Transfers []domain.TransferEvent
}

type GroupService struct {
	groups   groupRepository
	snapshot Transactor[GroupRepositories]
}

func NewGroupService(groups groupRepository, snapshot Transactor[GroupRepositories]) *GroupService {
	return &GroupService{groups: groups, snapshot: snapshot}
}

func (s *GroupService) List(ctx context.Context, filter domain.GroupFilter) (GroupPage, error) {
	limit := filter.Limit
	if limit <= 0 {
		return GroupPage{}, fmt.Errorf("list groups: limit must be positive, got %d", limit)
	}

	filter.Limit++

	groups, err := s.groups.List(ctx, filter)
	if err != nil {
		return GroupPage{}, err
	}

	if len(groups) <= limit {
		return GroupPage{Groups: groups}, nil
	}

	groups = groups[:limit]
	next := groups[limit-1].Cursor()

	return GroupPage{Groups: groups, Next: &next}, nil
}

func (s *GroupService) Get(ctx context.Context, id uuid.UUID) (GroupDetails, error) {
	var details GroupDetails

	err := s.snapshot(ctx, func(r GroupRepositories) error {
		var err error

		if details.Group, err = r.Groups.Get(ctx, id); err != nil {
			return err
		}

		if details.Members, err = r.Groups.ListMembers(ctx, id); err != nil {
			return err
		}

		if details.Logins, err = r.Groups.ListLogins(ctx, id); err != nil {
			return err
		}

		details.Transfers, err = r.Groups.ListTransfers(ctx, id)

		return err
	})
	if err != nil {
		return GroupDetails{}, fmt.Errorf("get group %s: %w", id, err)
	}

	return details, nil
}
