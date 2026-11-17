-- +goose Up
ALTER TABLE groups
    ADD COLUMN risk_level TEXT NOT NULL DEFAULT 'low'
        CONSTRAINT groups_risk_level_chk CHECK (risk_level IN ('low', 'medium', 'high'));

ALTER TABLE groups ALTER COLUMN risk_level DROP DEFAULT;

-- +goose Down
ALTER TABLE groups DROP COLUMN IF EXISTS risk_level;
