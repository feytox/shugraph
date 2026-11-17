package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

var groupStruct = sqlbuilder.NewStruct(new(db.Group)).For(sqlbuilder.PostgreSQL)

type PgGroupRepository struct {
	conn    db.DBTX
	queries *db.Queries
}

func NewPgGroupRepository(conn db.DBTX) *PgGroupRepository {
	return &PgGroupRepository{
		conn:    conn,
		queries: db.New(conn),
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

	rows, err := p.conn.Query(ctx, query, args...)
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

func (p *PgGroupRepository) Get(ctx context.Context, id uuid.UUID) (domain.Group, error) {
	record, err := p.queries.GetGroup(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Group{}, domain.ErrGroupNotFound
	}

	if err != nil {
		return domain.Group{}, fmt.Errorf("select group: %w", err)
	}

	return groupFromRecord(record), nil
}

func (p *PgGroupRepository) ListMembers(ctx context.Context, groupID uuid.UUID) ([]domain.Account, error) {
	rows, err := p.queries.ListGroupMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("select group members: %w", err)
	}

	members := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		members = append(members, domain.Account{ID: row.ID, ClientID: row.ClientID})
	}

	return members, nil
}

func (p *PgGroupRepository) ListLogins(ctx context.Context, groupID uuid.UUID) ([]domain.LoginEvent, error) {
	rows, err := p.queries.ListGroupLogins(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("select group logins: %w", err)
	}

	logins := make([]domain.LoginEvent, 0, len(rows))
	for _, row := range rows {
		logins = append(logins, domain.LoginEvent{
			EventID:    row.EventID,
			OccurredAt: row.OccurredAt,
			AccountID:  row.AccountID,
			ClientID:   row.ClientID,
			DeviceID:   row.DeviceID,
			AddressID:  row.AddressID,
		})
	}

	return logins, nil
}

func (p *PgGroupRepository) ListTransfers(ctx context.Context, groupID uuid.UUID) ([]domain.TransferEvent, error) {
	rows, err := p.queries.ListGroupTransfers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("select group transfers: %w", err)
	}

	transfers := make([]domain.TransferEvent, 0, len(rows))
	for _, row := range rows {
		transfers = append(transfers, domain.TransferEvent{
			EventID:       row.EventID,
			OccurredAt:    row.OccurredAt,
			FromAccountID: row.FromAccountID,
			FromClientID:  row.FromClientID,
			ToAccountID:   row.ToAccountID,
			ToClientID:    row.ToClientID,
			Amount:        row.Amount,
		})
	}

	return transfers, nil
}

func groupFromRecord(record db.Group) domain.Group {
	group := domain.Group{
		ID:            record.ID,
		Scheme:        domain.GroupScheme(record.Scheme),
		Status:        domain.GroupStatus(record.Status),
		RiskScore:     int(record.RiskScore),
		RiskParts:     record.RiskParts,
		RiskLevel:     domain.RiskLevel(record.RiskLevel),
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
