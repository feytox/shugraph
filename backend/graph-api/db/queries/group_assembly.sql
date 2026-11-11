-- name: LockGroupAssembly :exec
SELECT pg_advisory_xact_lock(sqlc.arg(lock_key)::bigint);

-- name: FreezeStaleGroups :exec
UPDATE groups
SET is_frozen = TRUE
WHERE NOT is_frozen
  AND last_event_at < sqlc.arg(stale_before);

-- name: InsertGroup :exec
INSERT INTO groups (id, scheme, risk_score, risk_parts, first_event_at, last_event_at, accounts_count, total_amount)
VALUES (
    sqlc.arg(id), sqlc.arg(scheme), sqlc.arg(risk_score), sqlc.arg(risk_parts),
    sqlc.arg(first_event_at), sqlc.arg(last_event_at), sqlc.arg(accounts_count), sqlc.arg(total_amount)
);

-- name: AddGroupEvents :exec
INSERT INTO group_events (group_id, event_id)
SELECT sqlc.arg(group_id), unnest(sqlc.arg(event_ids)::text[])
ON CONFLICT DO NOTHING;

-- name: AddGroupMembers :exec
INSERT INTO group_members (group_id, account_id)
SELECT sqlc.arg(group_id), unnest(sqlc.arg(account_ids)::text[])
ON CONFLICT DO NOTHING;

-- name: RefreshGroupStats :one
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
WHERE g.id = sqlc.arg(group_id)
RETURNING g.first_event_at, g.last_event_at, g.accounts_count, g.total_amount;

-- name: UpdateGroupRisk :exec
UPDATE groups
SET risk_score = sqlc.arg(risk_score),
    risk_parts = sqlc.arg(risk_parts)
WHERE id = sqlc.arg(group_id);
