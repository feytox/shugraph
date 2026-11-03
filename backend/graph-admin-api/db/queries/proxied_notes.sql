-- name: CreateProxiedNote :one
INSERT INTO proxied_notes (graph_note_id, text)
VALUES ($1, $2)
RETURNING id, graph_note_id, text, created_at;
