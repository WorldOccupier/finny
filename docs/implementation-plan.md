# Finny Implementation Plan

This is the staged implementation plan for Finny. Complete phases in order and check off each item only after it has been implemented and verified.

## Phase 1 — Go server foundation

- [x] Create `server/go.mod`.
- [x] Create the Go development Dockerfile.
- [x] Create `server/cmd/finny/main.go`.
- [x] Add configurable server port.
- [x] Add `GET /health`.
- [x] Add graceful shutdown.
- [x] Test that the server starts.
- [x] Test that `/health` returns `200`.
- [x] Review Go version, module name, port, and configuration style.

## Phase 2 — React/Vite foundation

- [x] Create `web/package.json`.
- [x] Create the web development Dockerfile.
- [x] Create the Vite React TypeScript entrypoint.
- [x] Add the `/` route.
- [x] Add the `/edit` route.
- [x] Add the Vite `/api` proxy to the Go server.
- [x] Add Docker Compose development services for `web` and `server`.
- [x] Mount source directories for development hot reload.
- [x] Add a host-mounted SQLite data volume.
- [x] Render a dashboard placeholder.
- [x] Render an edit placeholder.
- [x] Verify a frontend request can reach `/health` through the proxy between containers.
- [x] Review frontend tooling and routing.

## Phase 3 — Domain types and decimal conventions

- [x] Add domain types for assets.
- [x] Add domain types for UK and India values.
- [x] Add domain types for snapshots and totals.
- [x] Add domain types for spending limits.
- [x] Add domain types for income totals.
- [x] Add the dashboard read-model type.
- [x] Add `shopspring/decimal`.
- [x] Define decimal JSON serialization as strings.
- [x] Define `Europe/London` timestamp handling.
- [x] Test valid decimal parsing.
- [x] Test invalid decimal rejection.
- [x] Test GBP and INR validation.
- [x] Verify no financial calculation uses `float64`.
- [x] Review the domain types before creating the schema.

## Phase 4 — SQLite schema and migrations

- [x] Create the `server/internal/database` package.
- [x] Add SQLite connection setup.
- [x] Add migration execution.
- [x] Add asset storage.
- [x] Add snapshot storage.
- [x] Add snapshot asset-value storage.
- [x] Add snapshot-total storage.
- [x] Add spending-limit storage.
- [x] Add income storage.
- [x] Add current-FX storage.
- [x] Add dashboard-revision storage.
- [x] Add idempotency-key storage.
- [x] Add foreign keys and required indexes.
- [x] Test initialization from an empty database.
- [x] Test ordered and repeatable migrations.
- [x] Test decimal round-tripping through SQLite `NUMERIC` columns.
- [x] Review the schema and migration SQL.

## Phase 5 — Persistence interface and SQLite adapter

- [x] Define one domain-oriented persistence interface.
- [x] Add dashboard read operations.
- [x] Add atomic dashboard-save operation.
- [x] Add revision operations.
- [x] Add current-FX operations.
- [x] Add asset operations.
- [x] Add snapshot operations.
- [x] Add spending-limit operations.
- [x] Add income operations.
- [x] Add idempotency-result operations.
- [x] Keep SQL and SQLite types inside the database package.
- [x] Add repository tests against SQLite.
- [x] Test empty-dashboard loading.
- [x] Test persistence of new records.
- [x] Test historical asset values after current-template removal.
- [x] Review the interface for domain-level boundaries.

## Phase 6 — Snapshot calculation and atomic save

- [x] Validate the first snapshot.
- [x] Implement carry-forward values.
- [x] Implement shared asset creation.
- [x] Implement removal from the current asset template.
- [x] Preserve removed assets in historical snapshots.
- [x] Validate UK and India currencies.
- [x] Calculate UK totals.
- [x] Calculate India totals.
- [x] Calculate combined GBP and INR totals.
- [x] Store snapshot FX values and historical totals.
- [x] Update current FX when a snapshot is saved.
- [x] Assign the server commit timestamp.
- [x] Execute the complete save in one transaction.
- [x] Test first-snapshot completeness.
- [x] Test carry-forward behavior.
- [x] Test new-asset validation.
- [x] Test historical preservation after removal.
- [x] Test frozen historical totals.
- [x] Test current-rate calculations.
- [x] Test rollback after validation or persistence failure.
- [x] Review calculation examples.

## Phase 7 — Dashboard API schema

- [x] Define the `GET /api/dashboard` response.
- [x] Define the `POST /api/dashboard` request.
- [x] Define revision handling.
- [x] Define numeric client-generated IDs.
- [x] Define asset values and currencies.
- [x] Define spending-limit values and currencies.
- [x] Define GBP income totals.
- [x] Define snapshot FX data.
- [x] Define `Idempotency-Key` behavior.
- [x] Define JSON error shape.
- [x] Define status-code mapping.
- [x] Create valid JSON fixtures.
- [x] Create invalid JSON fixtures.
- [x] Review and approve the API contract.

## Phase 8 — GET dashboard endpoint

- [x] Implement `GET /api/dashboard`.
- [x] Load the complete dashboard read model.
- [x] Include current UK and India values.
- [x] Include combined GBP and INR totals.
- [x] Include current FX and revision.
- [x] Include net-worth history.
- [x] Include spending limits and income.
- [x] Serialize decimals as strings.
- [x] Serialize timestamps with the documented timezone behavior.
- [x] Map database failures to safe errors.
- [x] Test empty-dashboard responses.
- [x] Test populated-dashboard responses.
- [x] Test history ordering.
- [x] Test current total calculations.
- [x] Review the JSON response manually.

## Phase 9 — POST dashboard endpoint

- [x] Implement request decoding.
- [x] Validate `Idempotency-Key`.
- [x] Check the submitted revision.
- [x] Reject stale revisions with `409 Conflict`.
- [x] Validate the complete editable graph.
- [x] Invoke the atomic save operation.
- [x] Store the committed idempotency result.
- [x] Return the complete committed dashboard.
- [x] Test valid saves.
- [x] Test stale revisions.
- [x] Test idempotent retries.
- [x] Test idempotency-key reuse with a different body.
- [x] Test rollback on invalid data.
- [x] Test new and removed assets.
- [x] Review concurrent-save behavior.

## Phase 10 — Read-only dashboard UI

- [x] Add the dashboard API client.
- [x] Add loading, empty, and error states.
- [x] Add the combined net-worth card.
- [x] Add the GBP/INR display toggle.
- [x] Add the historical net-worth graph.
- [x] Add UK and India sections or tabs.
- [x] Add spending-limit display.
- [x] Add income display.
- [x] Test rendering of a complete API response.
- [x] Test empty and error states.
- [x] Test currency switching.
- [x] Review visual hierarchy.

## Phase 11 — Edit dashboard UI

- [x] Load the complete dashboard into an editable form.
- [x] Add asset creation.
- [x] Add asset rename and removal.
- [x] Add UK and India asset-value editing.
- [x] Add FX-rate editing.
- [x] Add spending-limit editing.
- [x] Add income editing for both users.
- [x] Submit the complete graph.
- [x] Submit the revision.
- [x] Generate and submit an idempotency key.
- [x] Handle validation errors without losing form data.
- [x] Handle revision conflicts with a reload flow.
- [x] Refresh from the successful POST response.
- [x] Test first-snapshot editing.
- [x] Test carry-forward editing.
- [x] Test new and removed assets.
- [x] Test retry after a lost response.
- [x] Review the full edit workflow.

## Phase 12 — Integration and end-to-end verification

- [x] Add Go API integration tests.
- [x] Add frontend API-client tests.
- [x] Add browser-level tests for `/`.
- [x] Add browser-level tests for `/edit`.
- [x] Test creation of the first snapshot.
- [x] Test creation of a later snapshot.
- [x] Test FX changes with frozen history.
- [x] Test asset removal with retained history.
- [x] Test stale concurrent edits.
- [x] Test idempotent retries.
- [x] Verify a clean checkout can run the documented commands. (Verified from an isolated dependency-free copy with `docker compose up web server --build -d`; Go `/health` returned `{"status":"ok"}` and Vite returned the HTML shell. Temporary host ports were used because the default ports were occupied.)
- [x] Update documentation to match the implementation.
- [x] Review the POC before adding authentication or bank imports.

## Deferred work

- [ ] Add authentication.
- [ ] Add automatic FX rates.
- [ ] Add bank `.xlsx` imports.
- [ ] Add transaction categorization.
- [ ] Decide remote access and deployment.
- [ ] Add liabilities or debt tracking if required.
