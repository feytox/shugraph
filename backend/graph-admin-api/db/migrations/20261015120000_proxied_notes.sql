-- +goose Up
CREATE TABLE proxied_notes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    graph_note_id TEXT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS proxied_notes;
