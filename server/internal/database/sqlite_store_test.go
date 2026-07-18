package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func TestLoadDashboardFromEmptyDatabase(t *testing.T) {
	store := newTestStore(t)
	dashboard, err := store.LoadDashboard(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Revision != 0 || len(dashboard.Assets) != 0 || len(dashboard.History) != 0 || len(dashboard.SpendingLimits) != 0 {
		t.Fatalf("empty dashboard = %+v", dashboard)
	}
	if !dashboard.CurrentFXRate.IsZero() {
		t.Fatalf("empty FX rate = %s", dashboard.CurrentFXRate.String())
	}
}

func TestStorePersistsDashboardRecords(t *testing.T) {
	store := newTestStore(t)
	value := mustDecimal(t, "100.25")
	indiaValue := mustDecimal(t, "5000")
	snapshotTime := time.Date(2026, time.January, 15, 12, 0, 0, 0, domain.LondonLocation())

	dashboard := domain.Dashboard{
		Revision: 3,
		Assets: []domain.Asset{{
			ID:   1,
			Name: "Savings",
			Values: []domain.AssetValue{
				{AssetID: 1, Type: domain.UK_GBP, Value: value},
				{AssetID: 1, Type: domain.INDIA_INR, Value: indiaValue},
			},
		}},
		CurrentFXRate: value,
		History: []domain.Snapshot{{
			ID:          1,
			CommittedAt: snapshotTime,
			FXRate:      value,
			AssetValues: []domain.AssetValue{{AssetID: 1, Type: domain.UK_GBP, Value: value}, {AssetID: 1, Type: domain.INDIA_INR, Value: indiaValue}},
			Totals: domain.DashboardTotals{
				Country:  []domain.TotalValue{{Type: domain.UK_GBP, Value: value}, {Type: domain.INDIA_INR, Value: indiaValue}},
				Combined: []domain.CombinedTotal{{Currency: domain.CURRENCY_GBP, Value: value}, {Currency: domain.CURRENCY_INR, Value: indiaValue}},
			},
		}},
		SpendingLimits: []domain.SpendingLimit{{Key: "food", Amount: value, Currency: domain.CURRENCY_GBP}},
		Income:         domain.IncomeTotals{UserOneGBP: value, UserTwoGBP: indiaValue},
	}
	if err := store.SaveDashboard(context.Background(), dashboard); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.LoadDashboard(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Revision != 3 || len(loaded.Assets) != 1 || len(loaded.History) != 1 || len(loaded.SpendingLimits) != 1 {
		t.Fatalf("loaded dashboard = %+v", loaded)
	}
	if len(loaded.Assets[0].Values) != 2 || loaded.Assets[0].Values[0].AssetID != 1 {
		t.Fatalf("loaded asset values = %+v", loaded.Assets[0].Values)
	}
	if loaded.Income.UserOneGBP.String() != "100.25" || loaded.Income.UserTwoGBP.String() != "5000" {
		t.Fatalf("loaded income = %+v", loaded.Income)
	}
	if len(loaded.CurrentTotals.Country) != 2 || len(loaded.CurrentTotals.Combined) != 2 {
		t.Fatalf("loaded current totals = %+v", loaded.CurrentTotals)
	}
}

func TestHistoricalValuesRemainAfterAssetRemovalFromCurrentTemplate(t *testing.T) {
	store := newTestStore(t)
	value := mustDecimal(t, "100")
	asset := domain.Asset{ID: 1, Name: "Old savings"}
	if err := store.SaveAsset(context.Background(), asset); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveSnapshot(context.Background(), domain.Snapshot{
		ID:          1,
		CommittedAt: time.Date(2026, time.January, 1, 12, 0, 0, 0, domain.LondonLocation()),
		FXRate:      value,
		AssetValues: []domain.AssetValue{{AssetID: 1, Type: domain.UK_GBP, Value: value}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveDashboard(context.Background(), domain.Dashboard{}); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.LoadDashboard(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Assets) != 0 {
		t.Fatalf("removed asset still current: %+v", loaded.Assets)
	}
	if len(loaded.History) != 1 || len(loaded.History[0].AssetValues) != 1 {
		t.Fatalf("historical asset value was lost: %+v", loaded.History)
	}
}

func TestRevisionFXAndIdempotencyOperations(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	if _, err := store.GetRevision(ctx); !errors.Is(err, ErrNotFound) {
		t.Fatalf("initial revision error = %v", err)
	}
	if err := store.SetRevision(ctx, 4); err != nil {
		t.Fatal(err)
	}
	revision, err := store.GetRevision(ctx)
	if err != nil || revision != 4 {
		t.Fatalf("revision = %d, error = %v", revision, err)
	}
	rate := mustDecimal(t, "1.25")
	if err := store.SetCurrentFX(ctx, rate); err != nil {
		t.Fatal(err)
	}
	loadedRate, err := store.GetCurrentFX(ctx)
	if err != nil || !loadedRate.Equal(rate.Decimal) {
		t.Fatalf("FX rate = %s, error = %v", loadedRate.String(), err)
	}

	result := IdempotencyResult{Key: "save-1", RequestHash: "hash", ResponseJSON: `{"ok":true}`, CreatedAt: time.Now().UTC()}
	if err := store.SaveIdempotencyResult(ctx, result); err != nil {
		t.Fatal(err)
	}
	loadedResult, err := store.GetIdempotencyResult(ctx, result.Key)
	if err != nil || loadedResult.RequestHash != result.RequestHash || loadedResult.ResponseJSON != result.ResponseJSON {
		t.Fatalf("idempotency result = %+v, error = %v", loadedResult, err)
	}
}

func TestSaveDashboardRollsBackOnPersistenceFailure(t *testing.T) {
	store := newTestStore(t)
	value := mustDecimal(t, "100")
	dashboard := domain.Dashboard{
		Assets: []domain.Asset{{ID: 1, Name: "Should roll back"}},
		History: []domain.Snapshot{{
			ID:          1,
			CommittedAt: time.Date(2026, time.January, 1, 12, 0, 0, 0, domain.LondonLocation()),
			FXRate:      value,
			AssetValues: []domain.AssetValue{{AssetID: 999, Type: domain.UK_GBP, Value: value}},
		}},
	}
	if err := store.SaveDashboard(context.Background(), dashboard); err == nil {
		t.Fatal("invalid dashboard save succeeded")
	}
	assets, err := store.ListAssets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 0 {
		t.Fatalf("rolled-back assets = %+v", assets)
	}
}

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	db, err := Open(context.Background(), t.TempDir()+"/finny.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(context.Background(), db); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return NewSQLiteStore(db)
}

func mustDecimal(t *testing.T, value string) domain.Decimal {
	t.Helper()
	parsed, err := domain.NewDecimal(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
