package database

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func TestOpenAndMigrateEmptyDatabase(t *testing.T) {
	db := openMigratedDatabase(t)

	var foreignKeys int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys = %d, want 1", foreignKeys)
	}

	for _, table := range []string{
		"schema_migrations", "assets", "snapshots", "snapshot_asset_values",
		"snapshot_totals", "spending_limits", "income_totals", "current_fx",
		"dashboard_revision", "idempotency_keys",
	} {
		assertExists(t, db, "table", table)
	}
	for _, index := range []string{
		"idx_assets_active", "idx_snapshots_committed_at",
		"idx_snapshot_asset_values_asset", "idx_snapshot_totals_snapshot",
		"idx_idempotency_keys_created_at",
	} {
		assertExists(t, db, "index", index)
	}
}

func TestMigrationsAreOrderedAndRepeatable(t *testing.T) {
	db := openDatabase(t)
	files := fstest.MapFS{
		"migrations/0002_second.sql": {Data: []byte("CREATE TABLE second (id INTEGER PRIMARY KEY); INSERT INTO second (id) VALUES (2);")},
		"migrations/0001_first.sql":  {Data: []byte("CREATE TABLE first (id INTEGER PRIMARY KEY); INSERT INTO first (id) VALUES (1);")},
	}

	if err := migrateFS(context.Background(), db, files); err != nil {
		t.Fatal(err)
	}
	if err := migrateFS(context.Background(), db, files); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("migration count = %d, want 2", count)
	}
	var first, second int
	if err := db.QueryRow("SELECT id FROM first").Scan(&first); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT id FROM second").Scan(&second); err != nil {
		t.Fatal(err)
	}
	if first != 1 || second != 2 {
		t.Fatalf("migration data = %d, %d", first, second)
	}
}

func TestFailedMigrationIsNotRecorded(t *testing.T) {
	db := openDatabase(t)
	files := fstest.MapFS{
		"migrations/0001_broken.sql": {Data: []byte("CREATE TABLE broken (;")},
	}

	if err := migrateFS(context.Background(), db, files); err == nil {
		t.Fatal("broken migration succeeded")
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("failed migration count = %d, want 0", count)
	}
}

func TestForeignKeysAndHistoricalInactiveAssets(t *testing.T) {
	db := openMigratedDatabase(t)

	if _, err := db.Exec(`INSERT INTO snapshot_asset_values (snapshot_id, asset_id, value_type, value) VALUES (1, 99, 'UKGBP', '1')`); err == nil {
		t.Fatal("orphan snapshot asset value was accepted")
	}
	if _, err := db.Exec(`INSERT INTO snapshot_totals (snapshot_id, scope, value_type, value) VALUES (99, 'country', 'UKGBP', '1')`); err == nil {
		t.Fatal("orphan snapshot total was accepted")
	}
	if _, err := db.Exec(`INSERT INTO assets (id, name, active) VALUES (1, 'Old savings', 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO snapshots (id, committed_at, fx_rate) VALUES (1, '2026-01-01T12:00:00Z', '1.25')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO snapshot_asset_values (snapshot_id, asset_id, value_type, value) VALUES (1, 1, 'UKGBP', '100.50')`); err != nil {
		t.Fatal(err)
	}
}

func TestSaveDashboardSnapshotRollsBackOnFailure(t *testing.T) {
	db := openMigratedDatabase(t)
	store := NewSQLiteStore(db)
	value, err := domain.NewDecimal("100")
	if err != nil {
		t.Fatal(err)
	}
	fxRate, err := domain.NewDecimal("2")
	if err != nil {
		t.Fatal(err)
	}
	snapshot := domain.Snapshot{
		ID:          1,
		CommittedAt: time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC),
		FXRate:      fxRate,
		Assets: []domain.Asset{{ID: 1, Name: "Savings", Values: []domain.AssetValue{
			{Type: domain.UK_GBP, Value: value},
			{Type: domain.INDIA_INR, Value: value},
		}}},
		Totals: domain.DashboardTotals{
			Country:  []domain.TotalValue{{Type: domain.UK_GBP, Value: value}, {Type: domain.INDIA_INR, Value: value}},
			Combined: []domain.CombinedTotal{{Currency: domain.CURRENCY_GBP, Value: value}, {Currency: domain.CURRENCY_INR, Value: value}},
		},
	}
	err = store.SaveDashboardSnapshot(context.Background(), DashboardSnapshotSave{
		Assets:         snapshot.Assets,
		Snapshot:       snapshot,
		SpendingLimits: []domain.SpendingLimit{{Key: "broken", Amount: value, Currency: domain.Currency("USD")}},
		CurrentFXRate:  fxRate,
	})
	if err == nil {
		t.Fatal("invalid aggregate save succeeded")
	}
	var assets, snapshots int
	if err := db.QueryRow("SELECT COUNT(*) FROM assets").Scan(&assets); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM snapshots").Scan(&snapshots); err != nil {
		t.Fatal(err)
	}
	if assets != 0 || snapshots != 0 {
		t.Fatalf("rollback left assets=%d snapshots=%d", assets, snapshots)
	}
}

func TestDecimalRoundTripsThroughNumericColumns(t *testing.T) {
	db := openMigratedDatabase(t)
	if _, err := db.Exec(`INSERT INTO assets (id, name) VALUES (1, 'Precise savings')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO snapshots (id, committed_at, fx_rate) VALUES (1, '2026-01-01T12:00:00Z', '1.23456789')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO snapshot_asset_values (snapshot_id, asset_id, value_type, value) VALUES (1, 1, 'UKGBP', '123456.789012')`); err != nil {
		t.Fatal(err)
	}

	var value string
	if err := db.QueryRow(`SELECT CAST(value AS TEXT) FROM snapshot_asset_values WHERE snapshot_id = 1`).Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "123456.789012" {
		t.Fatalf("round-tripped value = %q", value)
	}
}

func TestSchemaConstraints(t *testing.T) {
	db := openMigratedDatabase(t)

	assertRejected(t, db, `INSERT INTO assets (id, name) VALUES (1, '')`)
	assertRejected(t, db, `INSERT INTO assets (id, name) VALUES (1, NULL)`)
	assertRejected(t, db, `INSERT INTO assets (id, name, active) VALUES (2, 'Bad active', 2)`)
	assertRejected(t, db, `INSERT INTO snapshots (id, committed_at, fx_rate) VALUES (1, '2026-01-01T12:00:00Z', '-1')`)
	assertRejected(t, db, `INSERT INTO snapshots (id, committed_at, fx_rate) VALUES (1, NULL, '1')`)
	assertRejected(t, db, `INSERT INTO spending_limits (limit_key, amount, currency) VALUES ('food', '1', 'USD')`)
	assertRejected(t, db, `INSERT INTO income_totals (user_key, amount, currency) VALUES ('user1', '1', 'INR')`)
	assertRejected(t, db, `INSERT INTO snapshot_totals (snapshot_id, scope, value_type, currency, value) VALUES (1, 'country', 'UKGBP', 'GBP', '1')`)
	assertRejected(t, db, `INSERT INTO current_fx (id, fx_rate) VALUES (2, '1')`)
	assertRejected(t, db, `INSERT INTO dashboard_revision (id, revision) VALUES (2, 0)`)
	assertRejected(t, db, `INSERT INTO idempotency_keys (idempotency_key, request_hash, response_json, created_at) VALUES ('', 'hash', '{}', 'now')`)

	if _, err := db.Exec(`INSERT INTO snapshots (id, committed_at, fx_rate) VALUES (1, '2026-01-01T12:00:00Z', '1')`); err != nil {
		t.Fatal(err)
	}
	assertRejected(t, db, `INSERT INTO snapshots (id, committed_at, fx_rate) VALUES (2, '2026-01-01T12:00:00Z', '1')`)
	if _, err := db.Exec(`INSERT INTO current_fx (id, fx_rate) VALUES (1, '1.2')`); err != nil {
		t.Fatal(err)
	}
	assertRejected(t, db, `INSERT INTO current_fx (id, fx_rate) VALUES (1, '1.3')`)
	if _, err := db.Exec(`INSERT INTO dashboard_revision (id, revision) VALUES (1, 0)`); err != nil {
		t.Fatal(err)
	}
	assertRejected(t, db, `INSERT INTO dashboard_revision (id, revision) VALUES (1, 1)`)
	if _, err := db.Exec(`INSERT INTO idempotency_keys (idempotency_key, request_hash, response_json, created_at) VALUES ('key', 'hash', '{}', 'now')`); err != nil {
		t.Fatal(err)
	}
	assertRejected(t, db, `INSERT INTO idempotency_keys (idempotency_key, request_hash, response_json, created_at) VALUES ('key', 'other', '{}', 'later')`)
}

func openMigratedDatabase(t *testing.T) *sql.DB {
	db := openDatabase(t)
	if err := Migrate(context.Background(), db); err != nil {
		db.Close()
		t.Fatal(err)
	}
	return db
}

func openDatabase(t *testing.T) *sql.DB {
	db, err := Open(context.Background(), t.TempDir()+"/finny.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func assertExists(t *testing.T, db *sql.DB, objectType, name string) {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = ? AND name = ?`, objectType, name).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("%s %q does not exist", objectType, name)
	}
}

func assertRejected(t *testing.T, db *sql.DB, statement string) {
	t.Helper()
	if _, err := db.Exec(statement); err == nil {
		t.Fatalf("statement was accepted: %s", statement)
	}
}
