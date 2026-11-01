-- +goose Up
CREATE TABLE clients (
    id TEXT PRIMARY KEY
);

CREATE TABLE accounts (
    id        TEXT PRIMARY KEY,
    client_id TEXT NOT NULL REFERENCES clients (id)
);

CREATE INDEX accounts_client_id_idx ON accounts (client_id);

CREATE TABLE devices (
    device_id TEXT PRIMARY KEY
);

CREATE TABLE network_addresses (
    address_id TEXT PRIMARY KEY
);

CREATE TABLE events (
    event_id    TEXT        PRIMARY KEY,
    type        TEXT        NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT events_type_chk CHECK (type IN ('transfer', 'login'))
);

CREATE INDEX events_occurred_at_idx ON events (occurred_at);

CREATE TABLE transfers (
    event_id        TEXT   PRIMARY KEY REFERENCES events (event_id) ON DELETE CASCADE,
    from_account_id TEXT   NOT NULL REFERENCES accounts (id),
    to_account_id   TEXT   NOT NULL REFERENCES accounts (id),
    amount          BIGINT NOT NULL,

    CONSTRAINT transfers_amount_chk CHECK (amount > 0),
    CONSTRAINT transfers_accounts_chk CHECK (from_account_id <> to_account_id)
);

CREATE INDEX transfers_from_account_id_idx ON transfers (from_account_id);
CREATE INDEX transfers_to_account_id_idx ON transfers (to_account_id);

CREATE TABLE login_events (
    event_id   TEXT PRIMARY KEY REFERENCES events (event_id) ON DELETE CASCADE,
    account_id TEXT NOT NULL REFERENCES accounts (id),
    device_id  TEXT NOT NULL REFERENCES devices (device_id),
    address_id TEXT NOT NULL REFERENCES network_addresses (address_id)
);

CREATE INDEX login_events_account_id_idx ON login_events (account_id);
CREATE INDEX login_events_device_id_idx ON login_events (device_id);
CREATE INDEX login_events_address_id_idx ON login_events (address_id);

CREATE TABLE group_members (
    group_id   UUID NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    account_id TEXT NOT NULL REFERENCES accounts (id),

    PRIMARY KEY (group_id, account_id)
);

CREATE INDEX group_members_account_id_idx ON group_members (account_id);

CREATE TABLE group_events (
    group_id UUID NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    event_id TEXT NOT NULL REFERENCES events (event_id),

    PRIMARY KEY (group_id, event_id)
);

CREATE INDEX group_events_event_id_idx ON group_events (event_id);

-- +goose Down
DROP TABLE IF EXISTS group_events;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS login_events;
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS network_addresses;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS clients;
