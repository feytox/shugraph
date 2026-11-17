-- +goose Up
ALTER TABLE groups ADD COLUMN risk_level TEXT;

UPDATE groups
SET risk_level = CASE
    WHEN risk_score >= 75 THEN 'high'
    WHEN risk_score >= 60 THEN 'medium'
    ELSE 'low'
END;

ALTER TABLE groups
    ALTER COLUMN risk_level SET NOT NULL,
    ADD CONSTRAINT groups_risk_level_chk CHECK (risk_level IN ('low', 'medium', 'high'));

-- +goose Down
ALTER TABLE groups DROP COLUMN IF EXISTS risk_level;
