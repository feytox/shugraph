package repository

import (
	"context"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

var groupStruct = sqlbuilder.NewStruct(new(groupRow)).For(sqlbuilder.PostgreSQL)

type SQLBuilderGroupRepository struct {
	pool db.DBTX
}

func NewSQLBuilderGroupRepository(pool db.DBTX) *SQLBuilderGroupRepository {
	return &SQLBuilderGroupRepository{
		pool: pool,
	}
}

func (s *SQLBuilderGroupRepository) List(ctx context.Context, filter domain.GroupFilter) ([]domain.Group, error) {
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
		sb.Where(fmt.Sprintf("(detected_at, id) < (%s, %s)", sb.Var(filter.After.DetectedAt), sb.Var(filter.After.ID)))
	}

	sql, args := sb.Build()

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("select groups: %w", err)
	}

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[groupRow])
	if err != nil {
		return nil, fmt.Errorf("scan groups: %w", err)
	}

	groups := make([]domain.Group, 0, len(records))
	for _, record := range records {
		groups = append(groups, groupFromRow(record))
	}

	return groups, nil
}
