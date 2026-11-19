package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/google/uuid"
)

type groupAssemblyLocker interface {
	LockAssembly(ctx context.Context) error
}

type groupSolutionRepository interface {
	GetGroupStatus(ctx context.Context, groupID uuid.UUID) (domain.GroupStatus, error)
	UpdateGroupSolution(ctx context.Context, solution domain.GroupSolution) (domain.GroupSolution, error)
}

type GroupSolutionRepositories struct {
	GroupAssemblyLocker     groupAssemblyLocker
	GroupSolutionRepository groupSolutionRepository
}

type GroupSolutionService struct {
	transactor Transactor[GroupSolutionRepositories]
}

func NewGroupSolutionService(transactor Transactor[GroupSolutionRepositories]) *GroupSolutionService {
	return &GroupSolutionService{
		transactor: transactor,
	}
}

func (srv GroupSolutionService) ApplySolution(ctx context.Context, solution domain.GroupSolution) (domain.GroupSolution, error) {
	if solution.Decision.Comment == "" {
		return domain.GroupSolution{}, errors.New("comment cannot be empty")
	}

	if solution.Status == domain.GroupStatusNew {
		return domain.GroupSolution{}, errors.New("solution status must be approved or declined")
	}

	var appliedSolution domain.GroupSolution

	if err := srv.transactor(ctx, func(r GroupSolutionRepositories) error {
		err := r.GroupAssemblyLocker.LockAssembly(ctx)
		if err != nil {
			return fmt.Errorf("taking assembly lock: %w", err)
		}

		appliedSolution, err = r.GroupSolutionRepository.UpdateGroupSolution(ctx, solution)
		if err != nil {
			status, err := r.GroupSolutionRepository.GetGroupStatus(ctx, solution.GroupID)
			if err != nil {
				if notFound, ok := errors.AsType[domain.ErrGroupNotFound](err); ok {
					return notFound
				}
				return fmt.Errorf("get group status: %w", err)
			}

			return domain.ErrGroupSolutionAlreadyExists{GroupID: solution.GroupID, Status: status}
		}

		return nil
	}); err != nil {
		return domain.GroupSolution{}, err
	}

	return appliedSolution, nil
}
