CREATE TABLE assets (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    active INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1))
);

CREATE TABLE snapshots (
    id INTEGER PRIMARY KEY,
    committed_at TEXT NOT NULL UNIQUE,
    fx_rate NUMERIC NOT NULL CHECK (fx_rate >= 0)
);

CREATE TABLE snapshot_asset_values (
    snapshot_id INTEGER NOT NULL,
    asset_id INTEGER NOT NULL,
    value_type TEXT NOT NULL CHECK (value_type IN ('UKGBP', 'INDIAINR')),
    value NUMERIC NOT NULL CHECK (value >= 0),
    PRIMARY KEY (snapshot_id, asset_id, value_type),
    FOREIGN KEY (snapshot_id) REFERENCES snapshots (id),
    FOREIGN KEY (asset_id) REFERENCES assets (id)
);

CREATE TABLE snapshot_totals (
    id INTEGER PRIMARY KEY,
    snapshot_id INTEGER NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN ('country', 'combined')),
    value_type TEXT,
    currency TEXT,
    value NUMERIC NOT NULL CHECK (value >= 0),
    CHECK (
        (scope = 'country' AND value_type IN ('UKGBP', 'INDIAINR') AND currency IS NULL)
        OR
        (scope = 'combined' AND value_type IS NULL AND currency IN ('GBP', 'INR'))
    ),
    UNIQUE (snapshot_id, scope, value_type, currency),
    FOREIGN KEY (snapshot_id) REFERENCES snapshots (id)
);

CREATE TABLE spending_limits (
    limit_key TEXT PRIMARY KEY CHECK (length(trim(limit_key)) > 0),
    amount NUMERIC NOT NULL CHECK (amount >= 0),
    currency TEXT NOT NULL CHECK (currency IN ('GBP', 'INR'))
);

CREATE TABLE income_totals (
    user_key TEXT PRIMARY KEY CHECK (length(trim(user_key)) > 0),
    amount NUMERIC NOT NULL CHECK (amount >= 0),
    currency TEXT NOT NULL DEFAULT 'GBP' CHECK (currency = 'GBP')
);

CREATE TABLE current_fx (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    fx_rate NUMERIC NOT NULL CHECK (fx_rate >= 0)
);

CREATE TABLE dashboard_revision (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    revision INTEGER NOT NULL CHECK (revision >= 0)
);

CREATE TABLE idempotency_keys (
    idempotency_key TEXT PRIMARY KEY CHECK (length(trim(idempotency_key)) > 0),
    request_hash TEXT NOT NULL,
    response_json TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_assets_active ON assets (active);
CREATE INDEX idx_snapshots_committed_at ON snapshots (committed_at);
CREATE INDEX idx_snapshot_asset_values_asset ON snapshot_asset_values (asset_id);
CREATE INDEX idx_snapshot_totals_snapshot ON snapshot_totals (snapshot_id);
CREATE INDEX idx_idempotency_keys_created_at ON idempotency_keys (created_at);
