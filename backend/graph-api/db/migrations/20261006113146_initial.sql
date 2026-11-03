-- +goose Up
CREATE TABLE notes (
       id UUID PRIMARY KEY,
       text TEXT NOT NULL,
       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS notes;
