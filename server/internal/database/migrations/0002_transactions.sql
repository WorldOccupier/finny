CREATE TABLE users (
    id TEXT PRIMARY KEY CHECK (id IN ('user_one', 'user_two')),
    display_name TEXT NOT NULL
);

INSERT INTO users (id, display_name) VALUES ('user_one', 'User One'), ('user_two', 'User Two');

CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    bank_source TEXT NOT NULL CHECK (length(trim(bank_source)) > 0),
    account_label TEXT NOT NULL CHECK (length(trim(account_label)) > 0),
    currency TEXT NOT NULL CHECK (currency IN ('GBP', 'INR')),
    owner TEXT NOT NULL CHECK (owner IN ('user_one', 'user_two', 'joint'))
);

CREATE TABLE statements (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    imported_by TEXT NOT NULL,
    filename TEXT NOT NULL,
    format TEXT NOT NULL CHECK (format IN ('csv', 'xlsx')),
    checksum TEXT NOT NULL UNIQUE CHECK (length(checksum) = 64),
    period_start TEXT NOT NULL,
    period_end TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status = 'imported'),
    imported_rows INTEGER NOT NULL CHECK (imported_rows >= 0),
    invalid_rows INTEGER NOT NULL CHECK (invalid_rows >= 0),
    duplicate_rows INTEGER NOT NULL CHECK (duplicate_rows >= 0),
    FOREIGN KEY (account_id) REFERENCES accounts (id),
    FOREIGN KEY (imported_by) REFERENCES users (id)
);

CREATE TABLE transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    statement_id TEXT NOT NULL,
    transaction_date TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    currency TEXT NOT NULL CHECK (currency IN ('GBP', 'INR')),
    description TEXT NOT NULL CHECK (length(trim(description)) > 0),
    reference TEXT NOT NULL DEFAULT '',
    source_row INTEGER NOT NULL CHECK (source_row > 0),
    fingerprint TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts (id),
    FOREIGN KEY (statement_id) REFERENCES statements (id),
    UNIQUE (statement_id, source_row)
);

CREATE INDEX idx_accounts_owner ON accounts (owner);
CREATE INDEX idx_statements_account ON statements (account_id);
CREATE INDEX idx_transactions_account_date ON transactions (account_id, transaction_date);
CREATE INDEX idx_transactions_currency ON transactions (currency);
CREATE INDEX idx_transactions_fingerprint ON transactions (fingerprint);
CREATE INDEX idx_transactions_statement ON transactions (statement_id);
