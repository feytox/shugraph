package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

type PgFarmRepository struct {
	queries *db.Queries
}

func NewPgFarmRepository(tx pgx.Tx) *PgFarmRepository {
	return &PgFarmRepository{queries: db.New(tx)}
}

func (p *PgFarmRepository) FindOpenGroup(ctx context.Context, deviceID string) (uuid.UUID, bool, error) {
	ids, err := p.queries.FindOpenFarmGroups(ctx, deviceID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("find open farm group: %w", err)
	}

	switch len(ids) {
	case 0:
		return uuid.Nil, false, nil
	case 1:
		return ids[0], true, nil
	default:
		return uuid.Nil, false, fmt.Errorf("find open farm group: device %s has several open farm groups", deviceID)
	}
}

func (p *PgFarmRepository) ListOpenGroupDevices(ctx context.Context, accountID string) ([]string, error) {
	devices, err := p.queries.ListOpenFarmGroupDevices(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list open farm group devices: %w", err)
	}

	return devices, nil
}

func (p *PgFarmRepository) ListLogins(
	ctx context.Context,
	deviceID string,
	window domain.TimeWindow,
) ([]domain.FarmLogin, error) {
	rows, err := p.queries.ListFarmLogins(ctx, db.ListFarmLoginsParams{
		DeviceID:    deviceID,
		WindowStart: window.Start,
		WindowEnd:   window.End,
	})
	if err != nil {
		return nil, fmt.Errorf("list farm logins: %w", err)
	}

	logins := make([]domain.FarmLogin, 0, len(rows))
	for _, row := range rows {
		logins = append(logins, domain.FarmLogin{
			EventID:    row.EventID,
			OccurredAt: row.OccurredAt,
			AccountID:  row.AccountID,
			ClientID:   row.ClientID,
		})
	}

	return logins, nil
}

func (p *PgFarmRepository) ListTransfers(
	ctx context.Context,
	fromAccountIDs []string,
	window domain.TimeWindow,
) ([]domain.FarmTransfer, error) {
	rows, err := p.queries.ListFarmTransfers(ctx, db.ListFarmTransfersParams{
		FromAccountIds: fromAccountIDs,
		WindowStart:    window.Start,
		WindowEnd:      window.End,
	})
	if err != nil {
		return nil, fmt.Errorf("list farm transfers: %w", err)
	}

	transfers := make([]domain.FarmTransfer, 0, len(rows))
	for _, row := range rows {
		transfers = append(transfers, domain.FarmTransfer{
			EventID:       row.EventID,
			FromAccountID: row.FromAccountID,
			ToAccountID:   row.ToAccountID,
		})
	}

	return transfers, nil
}
