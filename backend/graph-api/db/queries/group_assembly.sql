-- name: LockGroupAssembly :exec
SELECT pg_advisory_xact_lock(sqlc.arg(lock_key)::bigint);

-- name: FreezeStaleGroups :exec
UPDATE groups
SET is_frozen = TRUE
WHERE NOT is_frozen
  AND last_event_at < sqlc.arg(stale_before);

-- name: FindOpenFarmGroup :one
SELECT DISTINCT g.id, g.last_event_at
FROM groups g
JOIN group_events ge ON ge.group_id = g.id
JOIN login_events l ON l.event_id = ge.event_id
WHERE g.scheme = 'farm'
  AND NOT g.is_frozen
  AND l.device_id = sqlc.arg(device_id);

-- name: ListOpenFarmGroupDevices :many
SELECT DISTINCT l.device_id
FROM groups g
JOIN group_events ge ON ge.group_id = g.id
JOIN login_events l ON l.event_id = ge.event_id
WHERE g.scheme = 'farm'
  AND NOT g.is_frozen
  AND l.account_id = sqlc.arg(account_id)
ORDER BY l.device_id;

-- name: ListFarmLogins :many
SELECT e.event_id, l.account_id, a.client_id
FROM login_events l
JOIN events e ON e.event_id = l.event_id
JOIN accounts a ON a.id = l.account_id
WHERE l.device_id = sqlc.arg(device_id)
  AND e.occurred_at BETWEEN sqlc.arg(window_start) AND sqlc.arg(window_end)
  AND NOT EXISTS (
      SELECT 1
      FROM group_events ge
      JOIN groups g ON g.id = ge.group_id
      WHERE ge.event_id = e.event_id
        AND g.scheme = 'farm'
        AND g.is_frozen
  );

-- name: ListFarmTransfers :many
SELECT e.event_id, t.from_account_id, t.to_account_id
FROM transfers t
JOIN events e ON e.event_id = t.event_id
WHERE t.from_account_id = ANY (sqlc.arg(from_account_ids)::text[])
  AND e.occurred_at BETWEEN sqlc.arg(window_start) AND sqlc.arg(window_end)
  AND NOT EXISTS (
      SELECT 1
      FROM group_events ge
      JOIN groups g ON g.id = ge.group_id
      WHERE ge.event_id = e.event_id
        AND g.scheme = 'farm'
        AND g.is_frozen
  );

-- name: InsertGroup :exec
INSERT INTO groups (id, scheme, risk_score, risk_parts, first_event_at, last_event_at, accounts_count, total_amount)
VALUES (sqlc.arg(id), sqlc.arg(scheme), 0, '{}', sqlc.arg(event_at), sqlc.arg(event_at), 0, 0);

-- name: AddGroupEvents :exec
INSERT INTO group_events (group_id, event_id)
SELECT sqlc.arg(group_id), unnest(sqlc.arg(event_ids)::text[])
ON CONFLICT DO NOTHING;

-- name: AddGroupMembers :exec
INSERT INTO group_members (group_id, account_id)
SELECT sqlc.arg(group_id), unnest(sqlc.arg(account_ids)::text[])
ON CONFLICT DO NOTHING;

-- name: RefreshGroupStats :exec
UPDATE groups g
SET first_event_at = s.first_event_at,
    last_event_at  = s.last_event_at,
    total_amount   = s.total_amount,
    accounts_count = (SELECT count(*) FROM group_members m WHERE m.group_id = g.id)
FROM (
    SELECT min(e.occurred_at)::timestamptz AS first_event_at,
           max(e.occurred_at)::timestamptz AS last_event_at,
           coalesce(sum(t.amount), 0)::bigint AS total_amount
    FROM group_events ge
    JOIN events e ON e.event_id = ge.event_id
    LEFT JOIN transfers t ON t.event_id = ge.event_id
    WHERE ge.group_id = sqlc.arg(group_id)
) s
WHERE g.id = sqlc.arg(group_id);
