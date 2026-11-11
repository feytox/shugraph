package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

type groupAssemblyRepository interface {
	LockAssembly(ctx context.Context) error
	FreezeStale(ctx context.Context, staleBefore time.Time) error
	Create(ctx context.Context, group domain.Group) error
	AddEvidence(ctx context.Context, groupID uuid.UUID, evidence domain.GroupEvidence) error
	RefreshStats(ctx context.Context, groupID uuid.UUID) (domain.GroupStats, error)
	UpdateRisk(ctx context.Context, groupID uuid.UUID, risk domain.Risk) error
}

type farmRepository interface {
	FindOpenGroup(ctx context.Context, deviceID string) (id uuid.UUID, found bool, err error)
	ListOpenGroupDevices(ctx context.Context, accountID string) ([]string, error)
	ListLogins(ctx context.Context, deviceID string, window domain.TimeWindow) ([]domain.FarmLogin, error)
	ListTransfers(ctx context.Context, fromAccountIDs []string, window domain.TimeWindow) ([]domain.FarmTransfer, error)
}

type groupAssembly struct {
	rules  domain.GroupRules
	groups groupAssemblyRepository
	farms  farmRepository
}

func (a groupAssembly) afterLogin(ctx context.Context, event domain.LoginEvent) error {
	if err := a.prepare(ctx, event.OccurredAt); err != nil {
		return err
	}

	return a.assembleFarm(ctx, event.DeviceID, event.OccurredAt)
}

func (a groupAssembly) afterTransfer(ctx context.Context, event domain.TransferEvent) error {
	if err := a.prepare(ctx, event.OccurredAt); err != nil {
		return err
	}

	devices, err := a.farms.ListOpenGroupDevices(ctx, event.FromAccountID)
	if err != nil {
		return err
	}

	for _, device := range devices {
		if err := a.assembleFarm(ctx, device, event.OccurredAt); err != nil {
			return err
		}
	}

	return nil
}

func (a groupAssembly) prepare(ctx context.Context, eventAt time.Time) error {
	if err := a.groups.LockAssembly(ctx); err != nil {
		return err
	}

	return a.groups.FreezeStale(ctx, eventAt.Add(-a.rules.Window))
}

func (a groupAssembly) assembleFarm(ctx context.Context, deviceID string, eventAt time.Time) error {
	around := domain.TimeWindow{Start: eventAt.Add(-a.rules.Window), End: eventAt.Add(a.rules.Window)}

	logins, err := a.farms.ListLogins(ctx, deviceID, around)
	if err != nil {
		return err
	}

	farm, ok := a.rules.DetectFarm(logins, eventAt)
	if !ok {
		return nil
	}

	transfers, err := a.farms.ListTransfers(ctx, farm.AccountIDs(), farm.Window)
	if err != nil {
		return err
	}

	groupID, err := a.ensureFarmGroup(ctx, deviceID, eventAt)
	if err != nil {
		return err
	}

	if err := a.groups.AddEvidence(ctx, groupID, a.rules.FarmEvidence(farm, transfers)); err != nil {
		return err
	}

	stats, err := a.groups.RefreshStats(ctx, groupID)
	if err != nil {
		return err
	}

	return a.groups.UpdateRisk(ctx, groupID, domain.FarmRisk(stats))
}

func (a groupAssembly) ensureFarmGroup(ctx context.Context, deviceID string, eventAt time.Time) (uuid.UUID, error) {
	id, found, err := a.farms.FindOpenGroup(ctx, deviceID)
	if err != nil || found {
		return id, err
	}

	stats := domain.GroupStats{FirstEventAt: eventAt, LastEventAt: eventAt}
	risk := domain.FarmRisk(stats)

	group := domain.Group{
		ID:           uuid.New(),
		Scheme:       domain.GroupSchemeFarm,
		RiskScore:    risk.Score,
		RiskParts:    risk.Parts,
		FirstEventAt: stats.FirstEventAt,
		LastEventAt:  stats.LastEventAt,
	}

	if err := a.groups.Create(ctx, group); err != nil {
		return uuid.Nil, err
	}

	return group.ID, nil
}
