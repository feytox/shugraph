-- name: ListGroups :many
SELECT *
FROM groups
WHERE (sqlc.narg('status')::group_status IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('schemes')::text[] IS NULL OR scheme::text = ANY (sqlc.narg('schemes')::text[]))
  AND (sqlc.narg('cursor_detected_at')::timestamptz IS NULL
       OR (detected_at, id) < (sqlc.narg('cursor_detected_at')::timestamptz, sqlc.narg('cursor_id')::uuid))
ORDER BY detected_at DESC, id DESC
LIMIT sqlc.arg('limit');
