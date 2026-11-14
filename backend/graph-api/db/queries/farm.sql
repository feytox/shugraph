-- name: FindOpenFarmGroups :many
SELECT DISTINCT g.id
FROM groups g
JOIN group_events ge ON ge.group_id = g.id
JOIN login_events l ON l.event_id = ge.event_id
WHERE g.scheme = 'farm'
  AND NOT g.is_frozen
  AND l.device_id = sqlc.arg(device_id)
LIMIT 2;

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
SELECT e.event_id, e.occurred_at, l.account_id, a.client_id
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
