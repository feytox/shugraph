package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PgGroupSolutionRepository struct {
	queries *db.Queries
}

func NewPgGroupSolutionRepository(tx pgx.Tx) *PgGroupSolutionRepository {
	return &PgGroupSolutionRepository{
		queries: db.New(tx),
	}
}

func (p *PgGroupSolutionRepository) GetGroupStatus(ctx context.Context, groupID uuid.UUID) (domain.GroupStatus, error) {
	statusStr, err := p.queries.GetGroupStatus(ctx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrGroupNotFound{GroupID: groupID}
		}
		return "", fmt.Errorf("get group status: %w", err)
	}

	return domain.GroupStatus(statusStr), nil
}

func (p *PgGroupSolutionRepository) UpdateGroupSolution(ctx context.Context, solution domain.GroupSolution) (domain.GroupSolution, error) {
	updateGroup := db.UpdateGroupSolutionParams{
		DecisionComment: pgtype.Text{String: solution.Decision.Comment, Valid: true},
		Status:          string(solution.Status),
		GroupID:         solution.GroupID,
	}

	row, err := p.queries.UpdateGroupSolution(ctx, updateGroup)
	if err != nil {
		return domain.GroupSolution{}, fmt.Errorf("update group solution: %w", err)
	}

	status := domain.GroupStatus(row.Status)

	return domain.GroupSolution{
		GroupID: row.ID,
		Status:  status,
		Decision: &domain.Decision{
			Comment:   row.DecisionComment.String,
			DecidedAt: row.DecidedAt.Time,
		},
	}, nil
}
