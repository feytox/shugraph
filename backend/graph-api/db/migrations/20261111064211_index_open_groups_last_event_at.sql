-- +goose Up
CREATE INDEX groups_open_last_event_at_idx ON groups (last_event_at) WHERE NOT is_frozen;

-- +goose Down
DROP INDEX IF EXISTS groups_open_last_event_at_idx;
