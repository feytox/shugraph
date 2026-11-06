-- name: InsertEvent :execrows
INSERT INTO events (event_id, type, occurred_at)
VALUES ($1, $2, $3)
    ON CONFLICT (event_id) DO NOTHING;

-- name: InsertClient :exec
INSERT INTO clients (id)
VALUES ($1)
    ON CONFLICT (id) DO NOTHING;

-- name: InsertAccount :exec
INSERT INTO accounts (id, client_id)
VALUES ($1, $2)
    ON CONFLICT (id) DO NOTHING;

-- name: GetAccountOwner :one
SELECT client_id
FROM accounts
WHERE id = $1;

-- name: InsertDevice :exec
INSERT INTO devices (device_id)
VALUES ($1)
    ON CONFLICT (device_id) DO NOTHING;

-- name: InsertNetworkAddress :exec
INSERT INTO network_addresses (address_id)
VALUES ($1)
    ON CONFLICT (address_id) DO NOTHING;

-- name: InsertTransfer :exec
INSERT INTO transfers (event_id, from_account_id, to_account_id, amount)
VALUES ($1, $2, $3, $4);

-- name: InsertLoginEvent :exec
INSERT INTO login_events (event_id, account_id, device_id, address_id)
VALUES ($1, $2, $3, $4);

-- name: IsSameLoginStored :one
SELECT EXISTS (
    SELECT 1
    FROM events e
    JOIN login_events l ON l.event_id = e.event_id
    JOIN accounts a ON a.id = l.account_id
    WHERE e.event_id = $1
      AND e.occurred_at = $2
      AND l.account_id = $3
      AND a.client_id = $4
      AND l.device_id = $5
      AND l.address_id = $6
);

-- name: IsSameTransferStored :one
SELECT EXISTS (
    SELECT 1
    FROM events e
    JOIN transfers t ON t.event_id = e.event_id
    JOIN accounts from_account ON from_account.id = t.from_account_id
    JOIN accounts to_account ON to_account.id = t.to_account_id
    WHERE e.event_id = sqlc.arg(event_id)
      AND e.occurred_at = sqlc.arg(occurred_at)
      AND t.from_account_id = sqlc.arg(from_account_id)
      AND from_account.client_id = sqlc.arg(from_client_id)
      AND t.to_account_id = sqlc.arg(to_account_id)
      AND to_account.client_id = sqlc.arg(to_client_id)
      AND t.amount = sqlc.arg(amount)
);
