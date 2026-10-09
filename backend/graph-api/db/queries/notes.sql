-- name: CreateNote :exec
INSERT INTO notes (id, text, created_at)
VALUES ($1, $2, $3);

-- name: GetNotes :many
SELECT * FROM notes;