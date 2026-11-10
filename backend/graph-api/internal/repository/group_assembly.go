package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

const groupAssemblyLockKey int64 = 1

type groupAssembler struct {
	rules domain.GroupRules
}

type timeWindow struct {
	start time.Time
	end   time.Time
}

func (a groupAssembler) afterLogin(ctx context.Context, q *db.Queries, event domain.LoginEvent) error {
	if err := a.prepare(ctx, q, event.OccurredAt); err != nil {
		return err
	}

	return a.assembleFarm(ctx, q, event.DeviceID, event.OccurredAt)
}

func (a groupAssembler) afterTransfer(ctx context.Context, q *db.Queries, event domain.TransferEvent) error {
	if err := a.prepare(ctx, q, event.OccurredAt); err != nil {
		return err
	}

	devices, err := q.ListOpenFarmGroupDevices(ctx, event.FromAccountID)
	if err != nil {
		return fmt.Errorf("list open farm group devices: %w", err)
	}

	for _, device := range devices {
		if err := a.assembleFarm(ctx, q, device, event.OccurredAt); err != nil {
			return err
		}
	}

	return nil
}

func (a groupAssembler) prepare(ctx context.Context, q *db.Queries, eventAt time.Time) error {
	if err := q.LockGroupAssembly(ctx, groupAssemblyLockKey); err != nil {
		return fmt.Errorf("lock group assembly: %w", err)
	}

	if err := q.FreezeStaleGroups(ctx, eventAt.Add(-a.rules.Window)); err != nil {
		return fmt.Errorf("freeze stale groups: %w", err)
	}

	return nil
}

func (a groupAssembler) assembleFarm(ctx context.Context, q *db.Queries, deviceID string, eventAt time.Time) error {
	open, err := findOpenFarmGroup(ctx, q, deviceID)
	if err != nil {
		return err
	}

	windowEnd := eventAt
	if open != nil && open.LastEventAt.After(eventAt) {
		windowEnd = open.LastEventAt
	}

	window := timeWindow{start: windowEnd.Add(-a.rules.Window), end: windowEnd}

	logins, err := listFarmLogins(ctx, q, deviceID, window)
	if err != nil {
		return err
	}

	if open == nil && !a.rules.Farm.Detected(logins) {
		return nil
	}

	transfers, err := listFarmTransfers(ctx, q, logins, window)
	if err != nil {
		return err
	}

	groupID, err := ensureGroup(ctx, q, open, domain.GroupSchemeFarm, eventAt)
	if err != nil {
		return err
	}

	return saveGroupEvidence(ctx, q, groupID, a.rules.Farm.Evidence(logins, transfers))
}

func findOpenFarmGroup(ctx context.Context, q *db.Queries, deviceID string) (*db.FindOpenFarmGroupRow, error) {
	group, err := q.FindOpenFarmGroup(ctx, deviceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("find open farm group: %w", err)
	}

	return &group, nil
}

func listFarmLogins(ctx context.Context, q *db.Queries, deviceID string, window timeWindow) ([]domain.FarmLogin, error) {
	rows, err := q.ListFarmLogins(ctx, db.ListFarmLoginsParams{
		DeviceID:    deviceID,
		WindowStart: window.start,
		WindowEnd:   window.end,
	})
	if err != nil {
		return nil, fmt.Errorf("list farm logins: %w", err)
	}

	logins := make([]domain.FarmLogin, 0, len(rows))
	for _, row := range rows {
		logins = append(logins, domain.FarmLogin{
			EventID:   row.EventID,
			AccountID: row.AccountID,
			ClientID:  row.ClientID,
		})
	}

	return logins, nil
}

func listFarmTransfers(
	ctx context.Context,
	q *db.Queries,
	logins []domain.FarmLogin,
	window timeWindow,
) ([]domain.FarmTransfer, error) {
	accounts := make([]string, 0, len(logins))
	for _, login := range logins {
		accounts = append(accounts, login.AccountID)
	}

	rows, err := q.ListFarmTransfers(ctx, db.ListFarmTransfersParams{
		FromAccountIds: accounts,
		WindowStart:    window.start,
		WindowEnd:      window.end,
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

func ensureGroup(
	ctx context.Context,
	q *db.Queries,
	open *db.FindOpenFarmGroupRow,
	scheme domain.GroupScheme,
	eventAt time.Time,
) (uuid.UUID, error) {
	if open != nil {
		return open.ID, nil
	}

	id := uuid.New()

	err := q.InsertGroup(ctx, db.InsertGroupParams{ID: id, Scheme: string(scheme), EventAt: eventAt})
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert group: %w", err)
	}

	return id, nil
}

func saveGroupEvidence(ctx context.Context, q *db.Queries, groupID uuid.UUID, evidence domain.GroupEvidence) error {
	err := q.AddGroupEvents(ctx, db.AddGroupEventsParams{GroupID: groupID, EventIds: evidence.EventIDs})
	if err != nil {
		return fmt.Errorf("add group events: %w", err)
	}

	err = q.AddGroupMembers(ctx, db.AddGroupMembersParams{GroupID: groupID, AccountIds: evidence.AccountIDs})
	if err != nil {
		return fmt.Errorf("add group members: %w", err)
	}

	if err := q.RefreshGroupStats(ctx, groupID); err != nil {
		return fmt.Errorf("refresh group stats: %w", err)
	}

	return nil
}
