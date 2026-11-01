package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

type PgGroupRepository struct {
	queries *db.Queries
}

func NewPgGroupRepository(pool db.DBTX) *PgGroupRepository {
	return &PgGroupRepository{
		queries: db.New(pool),
	}
}

func (p *PgGroupRepository) List(ctx context.Context, filter domain.GroupFilter) ([]domain.Group, error) {
	params := db.ListGroupsParams{
		Limit: filter.Limit,
	}

	if filter.Status != nil {
		params.Status = db.NullGroupStatus{GroupStatus: db.GroupStatus(*filter.Status), Valid: true}
	}

	for _, scheme := range filter.Schemes {
		params.Schemes = append(params.Schemes, string(scheme))
	}

	if filter.After != nil {
		params.CursorDetectedAt = pgtype.Timestamptz{Time: filter.After.DetectedAt, Valid: true}
		params.CursorID = pgtype.UUID{Bytes: filter.After.ID, Valid: true}
	}

	records, err := p.queries.ListGroups(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("select groups: %w", err)
	}

	groups := make([]domain.Group, 0, len(records))
	for _, record := range records {
		groups = append(groups, groupFromRecord(record))
	}

	return groups, nil
}

func groupFromRecord(record db.Group) domain.Group {
	group := domain.Group{
		ID:            record.ID,
		Scheme:        domain.GroupScheme(record.Scheme),
		Status:        domain.GroupStatus(record.Status),
		RiskScore:     int(record.RiskScore),
		RiskParts:     record.RiskParts,
		DetectedAt:    record.DetectedAt,
		FirstEventAt:  record.FirstEventAt,
		LastEventAt:   record.LastEventAt,
		AccountsCount: int(record.AccountsCount),
		TotalAmount:   record.TotalAmount,
		IsFrozen:      record.IsFrozen,
	}

	if record.DecidedAt.Valid {
		group.Decision = &domain.Decision{
			Comment:   record.DecisionComment.String,
			DecidedAt: record.DecidedAt.Time,
		}
	}

	return group
}
