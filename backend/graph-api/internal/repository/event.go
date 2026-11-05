package repository

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

type PgEventRepository struct {
	pool *pgxpool.Pool
}

func NewPgEventRepository(pool *pgxpool.Pool) *PgEventRepository {
	return &PgEventRepository{pool: pool}
}

func (p *PgEventRepository) SaveLogin(ctx context.Context, event domain.LoginEvent) (created bool, err error) {
	params := eventParams(event.EventID, domain.EventTypeLogin, event.OccurredAt)

	return saveEvent(ctx, p.pool, params, event, insertLoginDetails)
}

func (p *PgEventRepository) SaveTransfer(ctx context.Context, event domain.TransferEvent) (created bool, err error) {
	params := eventParams(event.EventID, domain.EventTypeTransfer, event.OccurredAt)

	return saveEvent(ctx, p.pool, params, event, insertTransferDetails)
}

func eventParams(eventID string, eventType domain.EventType, occurredAt time.Time) db.InsertEventParams {
	return db.InsertEventParams{
		EventID:    eventID,
		Type:       string(eventType),
		OccurredAt: occurredAt.Truncate(time.Microsecond),
	}
}

func saveEvent[Event any](
	ctx context.Context,
	pool *pgxpool.Pool,
	params db.InsertEventParams,
	event Event,
	insertDetails func(context.Context, *db.Queries, Event) error,
) (created bool, err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin transaction: %w", err)
	}
	defer rollback(ctx, tx)

	q := db.New(tx)

	rows, err := q.InsertEvent(ctx, params)
	if err != nil {
		return false, fmt.Errorf("insert event: %w", err)
	}

	if rows == 0 {
		return false, nil
	}

	if err := insertDetails(ctx, q, event); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit transaction: %w", err)
	}

	return true, nil
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func insertLoginDetails(ctx context.Context, q *db.Queries, event domain.LoginEvent) error {
	err := ensureAccounts(ctx, q, accountOwner{accountID: event.AccountID, clientID: event.ClientID})
	if err != nil {
		return err
	}

	if err := q.InsertDevice(ctx, event.DeviceID); err != nil {
		return fmt.Errorf("insert device: %w", err)
	}

	if err := q.InsertNetworkAddress(ctx, event.AddressID); err != nil {
		return fmt.Errorf("insert network address: %w", err)
	}

	err = q.InsertLoginEvent(ctx, db.InsertLoginEventParams{
		EventID:   event.EventID,
		AccountID: event.AccountID,
		DeviceID:  event.DeviceID,
		AddressID: event.AddressID,
	})
	if err != nil {
		return fmt.Errorf("insert login event: %w", err)
	}

	return nil
}

func insertTransferDetails(ctx context.Context, q *db.Queries, event domain.TransferEvent) error {
	err := ensureAccounts(ctx, q,
		accountOwner{accountID: event.FromAccountID, clientID: event.FromClientID},
		accountOwner{accountID: event.ToAccountID, clientID: event.ToClientID},
	)
	if err != nil {
		return err
	}

	err = q.InsertTransfer(ctx, db.InsertTransferParams{
		EventID:       event.EventID,
		FromAccountID: event.FromAccountID,
		ToAccountID:   event.ToAccountID,
		Amount:        event.Amount,
	})
	if err != nil {
		return fmt.Errorf("insert transfer: %w", err)
	}

	return nil
}

type accountOwner struct {
	accountID string
	clientID  string
}

func ensureAccounts(ctx context.Context, q *db.Queries, accounts ...accountOwner) error {
	if err := insertClientsInLockOrder(ctx, q, accounts); err != nil {
		return err
	}

	if err := insertAccountsInLockOrder(ctx, q, accounts); err != nil {
		return err
	}

	return checkAccountOwners(ctx, q, accounts)
}

func insertClientsInLockOrder(ctx context.Context, q *db.Queries, accounts []accountOwner) error {
	slices.SortFunc(accounts, func(a, b accountOwner) int {
		return cmp.Compare(a.clientID, b.clientID)
	})

	for _, account := range accounts {
		if err := q.InsertClient(ctx, account.clientID); err != nil {
			return fmt.Errorf("insert client: %w", err)
		}
	}

	return nil
}

func insertAccountsInLockOrder(ctx context.Context, q *db.Queries, accounts []accountOwner) error {
	slices.SortFunc(accounts, func(a, b accountOwner) int {
		return cmp.Compare(a.accountID, b.accountID)
	})

	for _, account := range accounts {
		err := q.InsertAccount(ctx, db.InsertAccountParams{
			ID:       account.accountID,
			ClientID: account.clientID,
		})
		if err != nil {
			return fmt.Errorf("insert account: %w", err)
		}
	}

	return nil
}

func checkAccountOwners(ctx context.Context, q *db.Queries, accounts []accountOwner) error {
	for _, account := range accounts {
		owner, err := q.GetAccountOwner(ctx, account.accountID)
		if err != nil {
			return fmt.Errorf("get account owner: %w", err)
		}

		if owner != account.clientID {
			return domain.ErrAccountOwnerMismatch
		}
	}

	return nil
}
