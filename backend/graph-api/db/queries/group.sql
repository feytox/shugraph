-- name: GetGroup :one
SELECT *
FROM groups
WHERE id = $1;

-- name: ListGroupMembers :many
SELECT a.id, a.client_id
FROM group_members m
JOIN accounts a ON a.id = m.account_id
WHERE m.group_id = $1
ORDER BY a.id;

-- name: ListGroupLogins :many
SELECT e.event_id, e.occurred_at, l.account_id, a.client_id, l.device_id, l.address_id
FROM group_events ge
JOIN events e ON e.event_id = ge.event_id
JOIN login_events l ON l.event_id = ge.event_id
JOIN accounts a ON a.id = l.account_id
WHERE ge.group_id = $1
ORDER BY e.occurred_at, e.event_id;

-- name: ListGroupTransfers :many
SELECT e.event_id,
       e.occurred_at,
       t.from_account_id,
       from_account.client_id AS from_client_id,
       t.to_account_id,
       to_account.client_id AS to_client_id,
       t.amount
FROM group_events ge
JOIN events e ON e.event_id = ge.event_id
JOIN transfers t ON t.event_id = ge.event_id
JOIN accounts from_account ON from_account.id = t.from_account_id
JOIN accounts to_account ON to_account.id = t.to_account_id
WHERE ge.group_id = $1
ORDER BY e.occurred_at, e.event_id;
