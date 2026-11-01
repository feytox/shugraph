package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

type groupRow struct {
	ID              uuid.UUID       `db:"id"`
	Scheme          string          `db:"scheme"`
	Status          string          `db:"status"`
	RiskScore       int32           `db:"risk_score"`
	RiskParts       json.RawMessage `db:"risk_parts"`
	DetectedAt      time.Time       `db:"detected_at"`
	FirstEventAt    time.Time       `db:"first_event_at"`
	LastEventAt     time.Time       `db:"last_event_at"`
	AccountsCount   int32           `db:"accounts_count"`
	TotalAmount     int64           `db:"total_amount"`
	IsFrozen        bool            `db:"is_frozen"`
	DecisionComment *string         `db:"decision_comment"`
	DecidedAt       *time.Time      `db:"decided_at"`
}

var groupStruct = sqlbuilder.NewStruct(new(groupRow)).For(sqlbuilder.PostgreSQL)

type PgGroupRepository struct {
	pool db.DBTX
}

func NewPgGroupRepository(pool db.DBTX) *PgGroupRepository {
	return &PgGroupRepository{
		pool: pool,
	}
}

func (p *PgGroupRepository) List(ctx context.Context, filter domain.GroupFilter) ([]domain.Group, error) {
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

	rows, err := p.pool.Query(ctx, sql, args...)
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

func groupFromRow(row groupRow) domain.Group {
	group := domain.Group{
		ID:            row.ID,
		Scheme:        domain.GroupScheme(row.Scheme),
		Status:        domain.GroupStatus(row.Status),
		RiskScore:     int(row.RiskScore),
		RiskParts:     row.RiskParts,
		DetectedAt:    row.DetectedAt,
		FirstEventAt:  row.FirstEventAt,
		LastEventAt:   row.LastEventAt,
		AccountsCount: int(row.AccountsCount),
		TotalAmount:   row.TotalAmount,
		IsFrozen:      row.IsFrozen,
	}

	if row.DecidedAt != nil {
		group.Decision = &domain.Decision{
			Comment:   *row.DecisionComment,
			DecidedAt: *row.DecidedAt,
		}
	}

	return group
}
