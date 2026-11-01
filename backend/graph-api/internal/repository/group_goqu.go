package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

var pg = goqu.Dialect("postgres")

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

type GoquGroupRepository struct {
	pool db.DBTX
}

func NewGoquGroupRepository(pool db.DBTX) *GoquGroupRepository {
	return &GoquGroupRepository{
		pool: pool,
	}
}

func (g *GoquGroupRepository) List(ctx context.Context, filter domain.GroupFilter) ([]domain.Group, error) {
	query := pg.From("groups").
		Select(&groupRow{}).
		Order(goqu.C("detected_at").Desc(), goqu.C("id").Desc())

	if filter.Limit > 0 {
		query = query.Limit(uint(filter.Limit))
	}

	if filter.Status != nil {
		query = query.Where(goqu.C("status").Eq(string(*filter.Status)))
	}

	if len(filter.Schemes) > 0 {
		query = query.Where(goqu.C("scheme").In(filter.Schemes))
	}

	if filter.After != nil {
		query = query.Where(goqu.L("(detected_at, id) < (?, ?)", filter.After.DetectedAt, filter.After.ID))
	}

	sql, args, err := query.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build select groups: %w", err)
	}

	rows, err := g.pool.Query(ctx, sql, args...)
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
