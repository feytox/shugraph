package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

const groupAssemblyLockKey int64 = 1

type PgGroupAssemblyRepository struct {
	queries *db.Queries
}

func NewPgGroupAssemblyRepository(tx pgx.Tx) *PgGroupAssemblyRepository {
	return &PgGroupAssemblyRepository{queries: db.New(tx)}
}

func (p *PgGroupAssemblyRepository) LockAssembly(ctx context.Context) error {
	if err := p.queries.LockGroupAssembly(ctx, groupAssemblyLockKey); err != nil {
		return fmt.Errorf("lock group assembly: %w", err)
	}

	return nil
}

func (p *PgGroupAssemblyRepository) FreezeStale(ctx context.Context, staleBefore time.Time) error {
	if err := p.queries.FreezeStaleGroups(ctx, staleBefore); err != nil {
		return fmt.Errorf("freeze stale groups: %w", err)
	}

	return nil
}

func (p *PgGroupAssemblyRepository) Create(ctx context.Context, group domain.Group) error {
	riskScore, err := toInt4(group.RiskScore)
	if err != nil {
		return fmt.Errorf("insert group risk score: %w", err)
	}

	accountsCount, err := toInt4(group.AccountsCount)
	if err != nil {
		return fmt.Errorf("insert group accounts count: %w", err)
	}

	err = p.queries.InsertGroup(ctx, db.InsertGroupParams{
		ID:            group.ID,
		Scheme:        string(group.Scheme),
		RiskScore:     riskScore,
		RiskParts:     group.RiskParts,
		FirstEventAt:  group.FirstEventAt,
		LastEventAt:   group.LastEventAt,
		AccountsCount: accountsCount,
		TotalAmount:   group.TotalAmount,
	})
	if err != nil {
		return fmt.Errorf("insert group: %w", err)
	}

	return nil
}

func (p *PgGroupAssemblyRepository) AddEvidence(ctx context.Context, groupID uuid.UUID, evidence domain.GroupEvidence) error {
	err := p.queries.AddGroupEvents(ctx, db.AddGroupEventsParams{GroupID: groupID, EventIds: evidence.EventIDs})
	if err != nil {
		return fmt.Errorf("add group events: %w", err)
	}

	err = p.queries.AddGroupMembers(ctx, db.AddGroupMembersParams{GroupID: groupID, AccountIds: evidence.AccountIDs})
	if err != nil {
		return fmt.Errorf("add group members: %w", err)
	}

	return nil
}

func (p *PgGroupAssemblyRepository) RefreshStats(ctx context.Context, groupID uuid.UUID) (domain.GroupStats, error) {
	row, err := p.queries.RefreshGroupStats(ctx, groupID)
	if err != nil {
		return domain.GroupStats{}, fmt.Errorf("refresh group stats: %w", err)
	}

	return domain.GroupStats{
		FirstEventAt:  row.FirstEventAt,
		LastEventAt:   row.LastEventAt,
		AccountsCount: int(row.AccountsCount),
		TotalAmount:   row.TotalAmount,
	}, nil
}

func (p *PgGroupAssemblyRepository) UpdateRisk(ctx context.Context, groupID uuid.UUID, risk domain.Risk) error {
	score, err := toInt4(risk.Score)
	if err != nil {
		return fmt.Errorf("update group risk score: %w", err)
	}

	err = p.queries.UpdateGroupRisk(ctx, db.UpdateGroupRiskParams{
		GroupID:   groupID,
		RiskScore: score,
		RiskParts: risk.Parts,
	})
	if err != nil {
		return fmt.Errorf("update group risk: %w", err)
	}

	return nil
}

func toInt4(n int) (int32, error) {
	if n < math.MinInt32 || n > math.MaxInt32 {
		return 0, fmt.Errorf("%d is out of int4 range", n)
	}

	return int32(n), nil
}
