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

## Phase 13 — Transaction domain and ownership model

- [x] Define persistent user records for the two household users.
- [x] Define account records with bank/source name, account label, currency, and ownership scope.
- [x] Support `user_one`, `user_two`, and `joint` account ownership.
- [x] Define importer identity separately from account ownership.
- [x] Define transaction fields: date, signed decimal amount, currency, description, reference, account, statement, and source row.
- [x] Define statement fields: filename, format, checksum, account, importer, period, status, and row counts.
- [x] Define transaction visibility rules for individual and joint accounts.
- [x] Define validation for supported currencies, dates, amounts, and required fields.
- [x] Define fingerprint normalization rules.
- [x] Add domain tests for ownership, visibility, validation, and fingerprint generation.
- [x] Review checkpoint: approve transaction and ownership domain boundaries before persistence work.

Acceptance checks:

- Domain tests pass.
- GBP and INR are accepted.
- Unsupported currencies and invalid amounts are rejected.
- Joint-account transactions are visible to both users.
- Identical rows in the same statement retain separate source-row identity.

## Phase 14 — CSV/XLSX parsing and import preview

- [x] Select existing or standard-library-compatible CSV parsing support.
- [x] Add XLSX parsing using the smallest suitable existing dependency or approved dependency.
- [x] Define column mapping for date, description, debit, credit, signed amount, currency, and reference.
- [x] Support signed-amount files.
- [x] Support separate debit and credit columns.
- [x] Normalize debits as negative and credits as positive.
- [x] Reject rows where both debit and credit are populated.
- [x] Validate required mappings before parsing.
- [x] Return row-level validation errors without aborting valid rows.
- [x] Return parsed rows with source row numbers.
- [x] Calculate file checksum and statement period.
- [x] Build a preview result containing valid rows, invalid rows, and summary counts.
- [x] Add parser tests for representative CSV and XLSX files.
- [x] Add tests for malformed files, missing columns, invalid dates, invalid amounts, and unsupported currencies.
- [x] Review checkpoint: approve parser behavior and preview contract before database integration.

Acceptance checks:

- CSV and XLSX fixtures parse successfully.
- Debit/credit normalization is correct.
- Invalid rows are reported with row numbers.
- Previewing an import does not persist data.
- Parsed financial values use `decimal`, never `float64`.

## Phase 15 — SQLite schema and persistence

- [x] Add the next ordered database migration.
- [x] Add `users` table and seed the two initial users.
- [x] Add `accounts` table with ownership scope and supported currency constraints.
- [x] Add `statements` table with checksum and import metadata.
- [x] Add `transactions` table with decimal amount and source-row identity.
- [x] Add uniqueness constraints and indexes for transaction lookup.
- [x] Add indexes for account, date, currency, direction, fingerprint, and statement.
- [x] Add foreign keys between users, accounts, statements, and transactions.
- [x] Extend the domain-oriented persistence interface.
- [x] Add account and statement persistence operations.
- [x] Add transaction persistence and query operations.
- [x] Add transaction summary query operations.
- [x] Keep SQLite-specific types and SQL inside the database package.
- [x] Test empty-database migration.
- [x] Test ordered and repeatable migrations.
- [x] Test decimal round-tripping.
- [x] Test foreign-key and uniqueness constraints.
- [x] Test persistence of individual and joint accounts.
- [x] Review checkpoint: approve schema, indexes, constraints, and repository boundaries.

Acceptance checks:

- Existing dashboard migrations and tests remain green.
- A clean database initializes with the new tables and users.
- Financial amounts round-trip exactly.
- Invalid ownership, currency, and foreign-key records are rejected.

## Phase 16 — Import commit, deduplication, and transaction service

- [x] Implement preview token creation and validation.
- [x] Prevent preview-token reuse after confirmation.
- [x] Prevent duplicate imports by file checksum.
- [x] Implement transaction fingerprint comparison.
- [x] Preserve identical occurrences within one statement using source-row identity.
- [x] Deduplicate matching no-reference rows from overlapping statement periods.
- [x] Preserve matching no-reference rows from non-overlapping periods.
- [x] Report duplicate rows during confirmation.
- [x] Persist statement metadata and accepted transactions atomically.
- [x] Persist invalid-row and duplicate counts on the statement record.
- [x] Record importing user independently from account ownership.
- [x] Ensure failed commits roll back all statement and transaction writes.
- [x] Add service tests for retries, duplicates, partial imports, and rollback.
- [x] Review checkpoint: approve import semantics and failure behavior before HTTP endpoints.

Acceptance checks:

- Confirming a preview persists valid rows exactly once.
- Reconfirming a preview does not create another statement.
- Reimporting the same file is rejected or reported as already imported.
- Duplicate transactions are not counted in spending totals.
- Failed persistence leaves no partial statement or transaction data.

## Phase 17 — Transaction and import APIs

- [x] Add `POST /api/statements/preview`.
- [x] Add `POST /api/statements/confirm`.
- [x] Add `GET /api/statements`.
- [x] Add `GET /api/transactions`.
- [x] Add `GET /api/spending/summary`.
- [x] Define request and response schemas.
- [x] Define multipart upload limits and supported file types.
- [x] Define statement preview response with valid, invalid, and duplicate rows.
- [x] Define transaction search filters.
- [x] Define pagination and deterministic ordering.
- [x] Define summary period parameters for day, week, month, and year.
- [x] Define API error codes for invalid files, invalid mappings, expired previews, duplicate imports, and validation failures.
- [x] Map database and service errors to safe HTTP responses.
- [x] Add valid and invalid JSON fixtures.
- [x] Add handler unit tests.
- [x] Add API integration tests against SQLite.
- [x] Review checkpoint: approve API contract and error behavior before frontend implementation.

Acceptance checks:

- Preview, confirmation, listing, search, and summary endpoints work against a real SQLite database.
- API responses serialize decimal values as strings.
- Pagination, filters, and ownership visibility behave correctly.
- Invalid requests return the documented error shape.
- Existing dashboard endpoints remain unchanged and passing.

## Phase 18 — Spending and transaction frontend

- [x] Add frontend API client types for statements, previews, transactions, and summaries.
- [x] Add a dedicated spending/transactions route.
- [x] Add account and ownership selection.
- [x] Add CSV/XLSX upload form.
- [x] Add column-mapping controls.
- [x] Add preview table with row-level validation errors.
- [x] Add duplicate and invalid-row counts.
- [x] Add confirmation flow.
- [x] Preserve preview state when confirmation fails.
- [x] Add searchable transaction table.
- [x] Add date, user, account, currency, direction, and text filters.
- [x] Add pagination controls.
- [x] Add daily, weekly, monthly, and yearly summary views.
- [x] Display signed amounts with correct currency.
- [x] Add loading, empty, error, and success states.
- [x] Add frontend unit/component tests.
- [x] Review checkpoint: approve complete import and spending user flow.

Acceptance checks:

- A user can upload, map, preview, and confirm a CSV/XLSX statement.
- Invalid rows are visible before confirmation.
- Joint-account transactions appear to both users.
- Search and filters update results correctly.
- All four summary periods render correct totals.
- Existing dashboard routes and UI tests remain passing.

## Phase 19 — End-to-end verification and documentation

- [x] Add browser-level coverage for CSV import.
- [x] Add browser-level coverage for XLSX import.
- [x] Test invalid-row preview.
- [x] Test duplicate-file reimport.
- [x] Test duplicate transactions across overlapping statements.
- [x] Test non-overlapping identical transactions.
- [x] Test individual and joint account visibility.
- [x] Test transaction search and pagination.
- [x] Test daily, weekly, monthly, and yearly totals.
- [x] Test Europe/London date boundaries.
- [x] Run Go formatting, tests, vet, and build.
- [x] Run frontend formatting, tests, build, and browser tests.
- [x] Run database migration and clean-checkout verification.
- [x] Update root, server, and web README files.
- [x] Update `docs/decisions.md` with transaction, ownership, import, deduplication, and aggregation decisions.
- [x] Replace the deferred bank-import item in `docs/implementation-plan.md` with completed phase checkboxes only after verification.
- [x] Review checkpoint: confirm all implementation phases are complete and no Critical or Warning findings remain.

Acceptance checks:

- All required checks pass from a clean checkout.
- The full import-to-summary workflow works through the browser.
- Existing net-worth functionality remains unaffected.
- Documentation matches the implemented API and user workflow.

## Phase gates for transaction work

- [x] Work proceeds strictly in numerical order.
- [x] A phase starts only after every checkbox and its review checkpoint are complete.
- [x] A phase is not complete based on code alone; its tests and acceptance checks must pass.
- [x] Every Critical or Warning review finding is fixed before advancing.
- [x] Update this file immediately after each completed phase.
- [x] Update `docs/decisions.md` in the same change whenever an architectural decision changes.

## Deferred work

- [x] Add authentication.
- [x] Add automatic FX rates.
- [x] Add bank statement imports beyond the CSV/XLSX transaction-import phases above.
- [x] Add transaction categorization.
- [x] Decide remote access and deployment.
- [x] Add liabilities or debt tracking if required.
