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

- [ ] Create `web/package.json`.
- [ ] Create the web development Dockerfile.
- [ ] Create the Vite React TypeScript entrypoint.
- [ ] Add the `/` route.
- [ ] Add the `/edit` route.
- [ ] Add the Vite `/api` proxy to the Go server.
- [ ] Add Docker Compose development services for `web` and `server`.
- [ ] Mount source directories for development hot reload.
- [ ] Add a host-mounted SQLite data volume.
- [ ] Render a dashboard placeholder.
- [ ] Render an edit placeholder.
- [ ] Verify a frontend request can reach `/health` through the proxy between containers.
- [ ] Review frontend tooling and routing.

## Phase 3 — Domain types and decimal conventions

- [ ] Add domain types for assets.
- [ ] Add domain types for UK and India values.
- [ ] Add domain types for snapshots and totals.
- [ ] Add domain types for spending limits.
- [ ] Add domain types for income totals.
- [ ] Add the dashboard read-model type.
- [ ] Add `shopspring/decimal`.
- [ ] Define decimal JSON serialization as strings.
- [ ] Define `Europe/London` timestamp handling.
- [ ] Test valid decimal parsing.
- [ ] Test invalid decimal rejection.
- [ ] Test GBP and INR validation.
- [ ] Verify no financial calculation uses `float64`.
- [ ] Review the domain types before creating the schema.

## Phase 4 — SQLite schema and migrations

- [ ] Create the `server/internal/database` package.
- [ ] Add SQLite connection setup.
- [ ] Add migration execution.
- [ ] Add asset storage.
- [ ] Add snapshot storage.
- [ ] Add snapshot asset-value storage.
- [ ] Add snapshot-total storage.
- [ ] Add spending-limit storage.
- [ ] Add income storage.
- [ ] Add current-FX storage.
- [ ] Add dashboard-revision storage.
- [ ] Add idempotency-key storage.
- [ ] Add foreign keys and required indexes.
- [ ] Test initialization from an empty database.
- [ ] Test ordered and repeatable migrations.
- [ ] Test decimal round-tripping through SQLite `NUMERIC` columns.
- [ ] Review the schema and migration SQL.

## Phase 5 — Persistence interface and SQLite adapter

- [ ] Define one domain-oriented persistence interface.
- [ ] Add dashboard read operations.
- [ ] Add atomic dashboard-save operation.
- [ ] Add revision operations.
- [ ] Add current-FX operations.
- [ ] Add asset operations.
- [ ] Add snapshot operations.
- [ ] Add spending-limit operations.
- [ ] Add income operations.
- [ ] Add idempotency-result operations.
- [ ] Keep SQL and SQLite types inside the database package.
- [ ] Add repository tests against SQLite.
- [ ] Test empty-dashboard loading.
- [ ] Test persistence of new records.
- [ ] Test historical asset values after current-template removal.
- [ ] Review the interface for domain-level boundaries.

## Phase 6 — Snapshot calculation and atomic save

- [ ] Validate the first snapshot.
- [ ] Implement carry-forward values.
- [ ] Implement shared asset creation.
- [ ] Implement removal from the current asset template.
- [ ] Preserve removed assets in historical snapshots.
- [ ] Validate UK and India currencies.
- [ ] Calculate UK totals.
- [ ] Calculate India totals.
- [ ] Calculate combined GBP and INR totals.
- [ ] Store snapshot FX values and historical totals.
- [ ] Update current FX when a snapshot is saved.
- [ ] Assign the server commit timestamp.
- [ ] Execute the complete save in one transaction.
- [ ] Test first-snapshot completeness.
- [ ] Test carry-forward behavior.
- [ ] Test new-asset validation.
- [ ] Test historical preservation after removal.
- [ ] Test frozen historical totals.
- [ ] Test current-rate calculations.
- [ ] Test rollback after validation or persistence failure.
- [ ] Review calculation examples.

## Phase 7 — Dashboard API schema

- [ ] Define the `GET /api/dashboard` response.
- [ ] Define the `POST /api/dashboard` request.
- [ ] Define revision handling.
- [ ] Define numeric client-generated IDs.
- [ ] Define asset values and currencies.
- [ ] Define spending-limit values and currencies.
- [ ] Define GBP income totals.
- [ ] Define snapshot FX data.
- [ ] Define `Idempotency-Key` behavior.
- [ ] Define JSON error shape.
- [ ] Define status-code mapping.
- [ ] Create valid JSON fixtures.
- [ ] Create invalid JSON fixtures.
- [ ] Review and approve the API contract.

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
