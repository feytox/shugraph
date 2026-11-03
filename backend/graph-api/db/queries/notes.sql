-- name: CreateNote :one
INSERT INTO notes (id, text)
VALUES ($1, $2)
RETURNING *;
-- name: GetNotes :many
SELECT * FROM notes;