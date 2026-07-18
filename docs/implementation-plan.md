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

- [ ] Implement `GET /api/dashboard`.
- [ ] Load the complete dashboard read model.
- [ ] Include current UK and India values.
- [ ] Include combined GBP and INR totals.
- [ ] Include current FX and revision.
- [ ] Include net-worth history.
- [ ] Include spending limits and income.
- [ ] Serialize decimals as strings.
- [ ] Serialize timestamps with the documented timezone behavior.
- [ ] Map database failures to safe errors.
- [ ] Test empty-dashboard responses.
- [ ] Test populated-dashboard responses.
- [ ] Test history ordering.
- [ ] Test current total calculations.
- [ ] Review the JSON response manually.

## Phase 9 — POST dashboard endpoint

- [ ] Implement request decoding.
- [ ] Validate `Idempotency-Key`.
- [ ] Check the submitted revision.
- [ ] Reject stale revisions with `409 Conflict`.
- [ ] Validate the complete editable graph.
- [ ] Invoke the atomic save operation.
- [ ] Store the committed idempotency result.
- [ ] Return the complete committed dashboard.
- [ ] Test valid saves.
- [ ] Test stale revisions.
- [ ] Test idempotent retries.
- [ ] Test idempotency-key reuse with a different body.
- [ ] Test rollback on invalid data.
- [ ] Test new and removed assets.
- [ ] Review concurrent-save behavior.

## Phase 10 — Read-only dashboard UI

- [ ] Add the dashboard API client.
- [ ] Add loading, empty, and error states.
- [ ] Add the combined net-worth card.
- [ ] Add the GBP/INR display toggle.
- [ ] Add the historical net-worth graph.
- [ ] Add UK and India sections or tabs.
- [ ] Add spending-limit display.
- [ ] Add income display.
- [ ] Test rendering of a complete API response.
- [ ] Test empty and error states.
- [ ] Test currency switching.
- [ ] Review visual hierarchy.

## Phase 11 — Edit dashboard UI

- [ ] Load the complete dashboard into an editable form.
- [ ] Add asset creation.
- [ ] Add asset rename and removal.
- [ ] Add UK and India asset-value editing.
- [ ] Add FX-rate editing.
- [ ] Add spending-limit editing.
- [ ] Add income editing for both users.
- [ ] Submit the complete graph.
- [ ] Submit the revision.
- [ ] Generate and submit an idempotency key.
- [ ] Handle validation errors without losing form data.
- [ ] Handle revision conflicts with a reload flow.
- [ ] Refresh from the successful POST response.
- [ ] Test first-snapshot editing.
- [ ] Test carry-forward editing.
- [ ] Test new and removed assets.
- [ ] Test retry after a lost response.
- [ ] Review the full edit workflow.

## Phase 12 — Integration and end-to-end verification

- [ ] Add Go API integration tests.
- [ ] Add frontend API-client tests.
- [ ] Add browser-level tests for `/`.
- [ ] Add browser-level tests for `/edit`.
- [ ] Test creation of the first snapshot.
- [ ] Test creation of a later snapshot.
- [ ] Test FX changes with frozen history.
- [ ] Test asset removal with retained history.
- [ ] Test stale concurrent edits.
- [ ] Test idempotent retries.
- [ ] Verify a clean checkout can run the documented commands.
- [ ] Update documentation to match the implementation.
- [ ] Review the POC before adding authentication or bank imports.

## Deferred work

- [ ] Add authentication.
- [ ] Add automatic FX rates.
- [ ] Add bank `.xlsx` imports.
- [ ] Add transaction categorization.
- [ ] Decide remote access and deployment.
- [ ] Add liabilities or debt tracking if required.
