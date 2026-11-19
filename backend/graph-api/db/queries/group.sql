-- name: GetGroupStatus :one
SELECT status
FROM groups
WHERE id = sqlc.arg(group_id);

-- name: UpdateGroupSolution :one
UPDATE groups
SET is_frozen = TRUE,
    status = sqlc.arg(status),
    decision_comment = sqlc.arg(decision_comment),
    decided_at = NOW()
WHERE id = sqlc.arg(group_id) AND status = 'new'
RETURNING id, status, decision_comment, decided_at;
