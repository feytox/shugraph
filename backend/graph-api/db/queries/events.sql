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

-- name: GetStoredLogin :one
SELECT e.occurred_at, l.account_id, a.client_id, l.device_id, l.address_id
FROM events e
JOIN login_events l ON l.event_id = e.event_id
JOIN accounts a ON a.id = l.account_id
WHERE e.event_id = $1;

-- name: GetStoredTransfer :one
SELECT
    e.occurred_at,
    t.from_account_id,
    from_account.client_id AS from_client_id,
    t.to_account_id,
    to_account.client_id AS to_client_id,
    t.amount
FROM events e
JOIN transfers t ON t.event_id = e.event_id
JOIN accounts from_account ON from_account.id = t.from_account_id
JOIN accounts to_account ON to_account.id = t.to_account_id
WHERE e.event_id = $1;
