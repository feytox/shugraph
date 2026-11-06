package repository

import (
	"context"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

var groupStruct = sqlbuilder.NewStruct(new(db.Group)).For(sqlbuilder.PostgreSQL)

type PgGroupRepository struct {
	pool db.DBTX
}

func NewPgGroupRepository(pool db.DBTX) *PgGroupRepository {
	return &PgGroupRepository{
		pool: pool,
	}
}

func (p *PgGroupRepository) List(ctx context.Context, filter domain.GroupFilter) ([]domain.Group, error) {
	if filter.Limit <= 0 {
		return nil, fmt.Errorf("list groups: limit must be positive, got %d", filter.Limit)
	}

	sb := groupStruct.SelectFrom("groups")
	sb.OrderBy("detected_at DESC", "id DESC")
	sb.Limit(int(filter.Limit))

	if filter.Status != nil {
		sb.Where(sb.Equal("status", string(*filter.Status)))
	}

	if len(filter.Schemes) > 0 {
		sb.Where(sb.In("scheme", sqlbuilder.Flatten(filter.Schemes)...))
	}

	if filter.After != nil {
		sb.Where(sb.LessThan(
			sqlbuilder.TupleNames("detected_at", "id"),
			sqlbuilder.Tuple(filter.After.DetectedAt, filter.After.ID),
		))
	}

	query, args := sb.Build()

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select groups: %w", err)
	}

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[db.Group])
	if err != nil {
		return nil, fmt.Errorf("scan groups: %w", err)
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
