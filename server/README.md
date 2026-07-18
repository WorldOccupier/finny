# Finny Server

The `server/` directory contains Finny’s Go HTTP API. It is responsible for persistence, validation, calculations, and business rules.

## Responsibilities

The server owns:

- JSON HTTP endpoints.
- SQLite persistence and database migrations.
- Household net-worth snapshot rules.
- GBP and INR calculations.
- Snapshot FX-rate preservation.
- Monthly spending-limit persistence.
- Income-total persistence.
- The future authentication middleware boundary.

The server should remain independent of browser rendering and React component behavior.

## Planned structure

```text
server/
├── cmd/
│   └── finny/
│       └── main.go              # Process entrypoint
├── internal/
│   ├── database/                # Persistence interface and SQLite implementation
│   ├── networth/                # Assets, snapshots, and calculations
│   ├── spendinglimits/          # Monthly limit logic
│   ├── income/                  # Income totals
│   └── api/                     # HTTP handlers and middleware
├── migrations/
├── go.mod
└── README.md
```

Keep domain logic separate from HTTP handlers. Handlers should decode requests, call domain operations, and encode responses. They should not contain database queries or calculation rules directly.

The `database/` package contains both the application-facing persistence interface and its initial SQLite implementation. Keep the interface expressed in domain operations rather than generic CRUD or SQLite-specific types so a different data source can be introduced later with limited impact.

## Dashboard API

For the POC, the server exposes one complete dashboard resource:

```text
GET  /api/dashboard
POST /api/dashboard
```

The frontend has two routes:

```text
/       # read-only dashboard
/edit   # editable dashboard form
```

The API intentionally transfers the complete dashboard because most information is updated together during an edit. Separate resource requests are not needed for the POC.

### GET `/api/dashboard`

Return the complete dashboard read model, including:

- Current UK and India net-worth values.
- Current combined totals in GBP and INR.
- Current FX rate.
- Shared assets and their current values and currencies.
- Net-worth history points.
- Spending limits.
- GBP income totals for both users.
- A dashboard revision for optimistic concurrency.

Decimal values are serialized as strings. Timestamps use ISO 8601 with the `Europe/London` offset.

Historical snapshots are read-only to API consumers and are included for graph rendering.

### POST `/api/dashboard`

The client submits the complete editable dashboard graph and includes:

- The revision returned by `GET`.
- Numeric client-generated IDs for all submitted records, beginning at `0`.
- Current values for every asset in both countries.
- Spending limits.
- GBP income totals for both users.
- The snapshot FX rate.

The request uses the standard HTTP header:

```text
Idempotency-Key: <client-generated-key>
```

The server:

1. Checks the idempotency key.
2. Rejects stale revisions before changing data.
3. Validates the complete submitted graph.
4. Creates one new server-timestamped household snapshot.
5. Creates newly submitted assets.
6. Removes omitted assets from the current template while retaining their historical values.
7. Replaces mutable spending limits and income totals.
8. Calculates and stores historical totals.
9. Updates the current FX rate to the submitted snapshot FX rate.
10. Increments the dashboard revision.
11. Returns the complete newly committed dashboard read model.

All steps occur in one database transaction. If any step fails, no part of the request is persisted.

### Concurrency and retries

Every successful `GET` and `POST` includes the current dashboard revision. A `POST` with a stale revision returns `409 Conflict` and does not modify data.

If a committed `POST` is retried with the same idempotency key, the server returns the original committed dashboard response without creating another snapshot. Reusing an idempotency key with a different request body returns `409 Conflict`.

### Error responses

Errors use a consistent JSON shape:

```json
{
  "error": {
    "code": "revision_conflict",
    "message": "The dashboard changed after it was loaded. Reload before saving."
  }
}
```

Initial error categories include invalid JSON, validation errors, invalid decimals or currencies, missing asset values, duplicate IDs, revision conflicts, idempotency conflicts, not-found errors, and internal errors.

The schema package defines the exact JSON contract used by the handlers in later phases. `GET /api/dashboard` returns `revision`, active `assets`, `currentFxRate`, `currentTotals`, `history`, `spendingLimits`, and `income`. History entries contain numeric IDs, `committedAt` timestamps formatted with the `Europe/London` offset, the saved `fxRate`, native asset values, and frozen totals. `POST /api/dashboard` accepts the revision, numeric client-generated asset IDs beginning at `0`, current asset values, the submitted snapshot `fxRate`, spending limits, and GBP income totals; computed history and totals are not client-editable.

All financial values are JSON strings. Asset values identify their country and currency through `UKGBP`/`GBP` and `INDIAINR`/`INR`; spending limits carry an explicit `GBP` or `INR` currency. The `Idempotency-Key` header is required for POST and accepts a trimmed value from 1 through 255 characters. Error bodies always use `{ "error": { "code": "...", "message": "..." } }`. Invalid request, validation, currency, decimal, asset-value, and idempotency-key errors use `400`; missing resources use `404`; revision and idempotency conflicts use `409`; unexpected failures use `500`.

Bank transaction import and transaction categorization are intentionally outside the POC API.

## Planned domain rules

- One household snapshot contains both UK and India values.
- Shared asset rows appear in both country sections. Each asset has an ID, name, value, and currency.
- UK assets use GBP; India assets use INR.
- The first snapshot requires a value for every asset in both countries.
- Later snapshots carry forward the previous values.
- New assets and their values are created as part of saving a household snapshot.
- Snapshot records are append-only.
- Each snapshot stores the manually entered INR-per-GBP rate used for its totals.
- Each snapshot stores native asset values, calculated historical totals, and the FX rate used for those calculations.
- Historical totals are not recalculated when later exchange rates are entered.
- Current totals use the latest snapshot values and the selected display currency.
- Income consists of one GBP total per user; INR income totals are not part of the POC.

## Persistence and calculations

Use SQLite as the initial data source, with decimal values stored using SQLite `NUMERIC` columns. Use `shopspring/decimal` in Go for asset values, totals, income, spending limits, and exchange rates. Do not use `float64` for persisted financial values or financial calculations.

Saving a snapshot is one atomic operation. It should validate the timestamp, shared asset values, country currencies, and FX rate; calculate totals; persist new assets and snapshot data; and update the current FX rate. If any part fails, the complete operation rolls back.

Use `Europe/London` for user-facing timestamp interpretation. Multiple snapshots may share a calendar date, but each snapshot must have a distinct timestamp.

## Development

The Go API runs in its own Docker container, separately from the Vite container:

```text
Go API:  http://localhost:8080
Web app: http://localhost:5173
```

Docker Compose runs both services during development. The Vite development server proxies `/api/*` requests to the Go API service. The Go process serves JSON API responses only and does not serve frontend assets.

## Planned commands

Once the server is implemented, the expected development command will be:

```bash
docker compose up web server
```

Database setup and migration commands will be added when the persistence implementation is introduced. SQLite data must be stored in a host-mounted volume rather than only inside the server container.

## Future boundaries

Authentication is intentionally deferred, but the server should keep authentication and authorization as middleware/domain boundaries so they can be added without redesigning every handler. Remote access, deployment, automatic exchange rates, and bank imports will be decided separately.
