-- +goose Up
CREATE TYPE group_scheme AS ENUM ('farm', 'address', 'chain', 'ring');
CREATE TYPE group_status AS ENUM ('new', 'approved', 'declined');

CREATE TABLE groups (
                        id               UUID         PRIMARY KEY,
                        scheme           group_scheme NOT NULL,
                        status           group_status NOT NULL DEFAULT 'new',
                        risk_score       INTEGER      NOT NULL CHECK (risk_score BETWEEN 0 AND 100),
                        risk_parts       JSONB        NOT NULL,
                        detected_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                        first_event_at   TIMESTAMPTZ  NOT NULL,
                        last_event_at    TIMESTAMPTZ  NOT NULL,
                        accounts_count   INTEGER      NOT NULL CHECK (accounts_count >= 0),
                        total_amount     BIGINT       NOT NULL CHECK (total_amount >= 0),
                        is_frozen        BOOLEAN      NOT NULL DEFAULT FALSE,
                        decision_comment TEXT,
                        decided_at       TIMESTAMPTZ,

                        CHECK (first_event_at <= last_event_at),
                        CHECK ((status = 'new') = (decided_at IS NULL)),
                        CHECK ((decided_at IS NULL) = (decision_comment IS NULL))
);

CREATE INDEX groups_status_detected_at_idx ON groups (status, detected_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS groups;
DROP TYPE IF EXISTS group_status;
DROP TYPE IF EXISTS group_scheme;