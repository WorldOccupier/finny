# Finny Architectural Decisions

This file records decisions that should remain stable across implementation sessions. Update it when a decision changes, and include the reason and date of the change.

## Current decisions

### Repository and runtime

- One repository contains `server/` and `web/`.
- The server is Go and exposes JSON HTTP APIs.
- The web app is React, TypeScript, and Vite.
- Vite and Go run as separate processes.
- Development runs Vite and Go in separate Docker containers orchestrated by Docker Compose.
- Go does not serve the React build output.
- SQLite data is stored in a host-mounted volume, not only inside the server container.

### Persistence

- SQLite is the initial data source.
- The Go SQLite driver is `modernc.org/sqlite` so the server remains CGO-free in Docker and CI.
- Persistence is isolated in the `database` package.
- One domain-oriented persistence interface is sufficient for the POC.
- Financial values use `shopspring/decimal` in Go.
- SQLite values use `NUMERIC` columns.
- Financial values must not use `float64`.

### Net worth

- UK values are in GBP.
- India values are in INR.
- A household snapshot contains both countries.
- Snapshots are append-only.
- Later snapshots carry forward previous asset values.
- New assets are created as part of saving a snapshot.
- Removed assets leave the current template but retain historical values.
- The server assigns the snapshot commit timestamp.
- User-facing time handling uses `Europe/London`.
- Historical totals use the FX rate saved with their snapshot.
- FX-derived combined totals use fixed 18-decimal-place division for deterministic results.
- The current dashboard uses the latest native values and the current FX rate.

### Dashboard API

- The POC uses `GET /api/dashboard` and `POST /api/dashboard`.
- Both operations work with the complete dashboard resource.
- POST creates one new snapshot and replaces mutable configuration atomically.
- Optimistic concurrency uses a dashboard revision.
- Client-generated numeric IDs begin at `0`.
- The `Idempotency-Key` HTTP header prevents duplicate saves after retries.
- Successful POST returns the complete newly committed dashboard read model.

### Income and spending limits

- Income is one GBP total per user.
- INR income totals are not included in the POC.
- Spending limits are persisted configurable key/value records with a currency.

### Transactions and statement imports

- Transaction imports are introduced after the net-worth POC as a separate domain.
- The initial household has two persistent users, represented by `user_one` and `user_two` until authentication is added.
- Accounts have `user_one`, `user_two`, or `joint` ownership; joint-account transactions are visible to both users.
- Statement importer identity is stored separately from account ownership.
- Imported transactions support GBP and INR and use signed `shopspring/decimal` amounts.
- Debits are negative and credits are positive.
- Transaction dates and statement periods use `Europe/London`.
- Transaction fingerprints use unambiguous normalized field encoding and retain source-row identity for repeated rows within one statement.
- CSV parsing uses the Go standard library; XLSX parsing uses `github.com/xuri/excelize/v2`.
- CSV/XLSX parsing, persistence, API endpoints, search, aggregation, and UI work proceed in the ordered implementation phases.
- Transaction categorization remains deferred.
- Import previews are held in the API process with single-use random tokens; confirmation persists statement metadata and accepted transactions atomically.
- File checksums reject repeated statement imports. No-reference transactions with matching fingerprints are suppressed when reimported; source rows within one statement remain distinct.
- Transaction listing supports account, text, currency, and page/page-size filters. Summary responses accept day, week, month, and year period labels.

## Change log

- 2026-07-18: Selected `modernc.org/sqlite` for the SQLite driver to keep Docker and CI builds CGO-free.
- 2026-07-18: Snapshot calculations use `shopspring/decimal` with fixed 18-place division for deterministic FX conversions.
- 2026-07-19: London snapshot timestamps serialize with RFC3339 nanosecond precision so rapid consecutive snapshots remain distinct while preserving their London offset.
