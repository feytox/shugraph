-- +goose Up
CREATE TABLE groups (
    id               UUID        PRIMARY KEY,
    scheme           TEXT        NOT NULL,
    status           TEXT        NOT NULL DEFAULT 'new',
    risk_score       INTEGER     NOT NULL,
    risk_parts       JSONB       NOT NULL,
    detected_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    first_event_at   TIMESTAMPTZ NOT NULL,
    last_event_at    TIMESTAMPTZ NOT NULL,
    accounts_count   INTEGER     NOT NULL,
    total_amount     BIGINT      NOT NULL,
    is_frozen        BOOLEAN     NOT NULL DEFAULT FALSE,
    decision_comment TEXT,
    decided_at       TIMESTAMPTZ,

    CONSTRAINT groups_scheme_chk CHECK (scheme IN ('farm', 'address', 'chain', 'ring')),
    CONSTRAINT groups_status_chk CHECK (status IN ('new', 'approved', 'declined')),
    CONSTRAINT groups_risk_score_chk CHECK (risk_score BETWEEN 0 AND 100),
    CONSTRAINT groups_risk_parts_chk CHECK (jsonb_typeof(risk_parts) = 'object'),
    CONSTRAINT groups_accounts_count_chk CHECK (accounts_count >= 0),
    CONSTRAINT groups_total_amount_chk CHECK (total_amount >= 0),
    CONSTRAINT groups_event_range_chk CHECK (first_event_at <= last_event_at),
    CONSTRAINT groups_decision_status_chk CHECK ((status = 'new') = (decided_at IS NULL)),
    CONSTRAINT groups_decision_fields_chk CHECK ((decided_at IS NULL) = (decision_comment IS NULL))
);

CREATE INDEX groups_status_detected_at_id_idx ON groups (status, detected_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS groups;
