package snapshot

import (
	"context"
	"testing"
	"time"

	"github.com/WorldOccupier/finny/server/internal/database"
	"github.com/WorldOccupier/finny/server/internal/domain"
)

func TestFirstSnapshotRequiresBothCountryValuesAndCalculatesTotals(t *testing.T) {
	store := newTestStore(t)
	now := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	service := NewService(store, func() time.Time { return now })
	value100 := mustDecimal(t, "100")
	value500 := mustDecimal(t, "500")
	fxRate := mustDecimal(t, "2")

	dashboard, err := service.Save(context.Background(), SnapshotInput{
		Assets: []domain.Asset{{ID: 1, Name: "Savings", Values: []domain.AssetValue{
			{Type: domain.UK_GBP, Value: value100},
			{Type: domain.INDIA_INR, Value: value500},
		}}},
		FXRate: fxRate,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(dashboard.History) != 1 || len(dashboard.History[0].Assets) != 1 {
		t.Fatalf("dashboard history = %+v", dashboard.History)
	}
	if dashboard.History[0].CommittedAt.Location() != domain.LondonLocation() {
		t.Fatalf("snapshot location = %v", dashboard.History[0].CommittedAt.Location())
	}
	totals := dashboard.History[0].Totals
	if totals.Country[0].Value.String() != "100" || totals.Country[1].Value.String() != "500" {
		t.Fatalf("country totals = %+v", totals.Country)
	}
	if totals.Combined[0].Value.String() != "350" || totals.Combined[1].Value.String() != "700" {
		t.Fatalf("combined totals = %+v", totals.Combined)
	}
}

func TestFirstSnapshotRejectsMissingValue(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store, time.Now)
	value := mustDecimal(t, "100")
	_, err := service.Save(context.Background(), SnapshotInput{
		Assets: []domain.Asset{{ID: 1, Name: "Incomplete", Values: []domain.AssetValue{{Type: domain.UK_GBP, Value: value}}}},
		FXRate: mustDecimal(t, "2"),
	})
	if err == nil {
		t.Fatal("incomplete first snapshot succeeded")
	}
}

func TestLaterSnapshotCarriesForwardAndNewAssetsRequireBothValues(t *testing.T) {
	store := newTestStore(t)
	now := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	service := NewService(store, func() time.Time { return now })
	uk := mustDecimal(t, "100")
	india := mustDecimal(t, "500")
	if _, err := service.Save(context.Background(), SnapshotInput{
		Assets: []domain.Asset{{ID: 1, Name: "Savings", Values: []domain.AssetValue{{Type: domain.UK_GBP, Value: uk}, {Type: domain.INDIA_INR, Value: india}}}},
		FXRate: mustDecimal(t, "2"),
	}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Hour)
	updated, err := service.Save(context.Background(), SnapshotInput{
		Assets: []domain.Asset{{ID: 1, Name: "Savings", Values: []domain.AssetValue{{Type: domain.UK_GBP, Value: mustDecimal(t, "150")}}}},
		FXRate: mustDecimal(t, "2"),
	})
	if err != nil {
		t.Fatal(err)
	}
	carriedForward := domain.Decimal{}
	for _, assetValue := range updated.Assets[0].Values {
		if assetValue.Type == domain.INDIA_INR {
			carriedForward = assetValue.Value
		}
	}
	if carriedForward.String() != "500" {
		t.Fatalf("carried-forward asset = %+v", updated.Assets[0])
	}
	now = now.Add(time.Hour)
	_, err = service.Save(context.Background(), SnapshotInput{
		Assets: []domain.Asset{{ID: 1, Name: "Savings", Values: []domain.AssetValue{{Type: domain.UK_GBP, Value: uk}, {Type: domain.INDIA_INR, Value: india}}}, {ID: 2, Name: "New asset", Values: []domain.AssetValue{{Type: domain.UK_GBP, Value: uk}}}},
		FXRate: mustDecimal(t, "2"),
	})
	if err == nil {
		t.Fatal("incomplete new asset succeeded")
	}
}

func TestRemovedAssetsRemainInHistoryAndHistoricalTotalsStayFrozen(t *testing.T) {
	store := newTestStore(t)
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	service := NewService(store, func() time.Time { return now })
	assets := []domain.Asset{{ID: 1, Name: "Savings", Values: []domain.AssetValue{{Type: domain.UK_GBP, Value: mustDecimal(t, "100")}, {Type: domain.INDIA_INR, Value: mustDecimal(t, "500")}}}}
	if _, err := service.Save(context.Background(), SnapshotInput{Assets: assets, FXRate: mustDecimal(t, "2")}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Hour)
	updated, err := service.Save(context.Background(), SnapshotInput{Assets: nil, FXRate: mustDecimal(t, "4")})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Assets) != 0 || len(updated.History[0].Assets) != 1 {
		t.Fatalf("removed asset state = %+v", updated)
	}
	if updated.History[0].Totals.Combined[0].Value.String() != "350" {
		t.Fatalf("historical total changed = %+v", updated.History[0].Totals)
	}
	if updated.History[1].Totals.Combined[0].Value.String() != "0" {
		t.Fatalf("empty current total = %+v", updated.History[1].Totals)
	}
}

func TestSnapshotInputRejectsDuplicatesAndInvalidFX(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store, time.Now)
	value := mustDecimal(t, "1")
	input := SnapshotInput{Assets: []domain.Asset{
		{ID: 1, Name: "Same", Values: []domain.AssetValue{{Type: domain.UK_GBP, Value: value}, {Type: domain.INDIA_INR, Value: value}}},
		{ID: 1, Name: "Other", Values: []domain.AssetValue{{Type: domain.UK_GBP, Value: value}, {Type: domain.INDIA_INR, Value: value}}},
	}, FXRate: value}
	if _, err := service.Save(context.Background(), input); err == nil {
		t.Fatal("duplicate asset ID succeeded")
	}
	input.Assets = []domain.Asset{{ID: 1, Name: "Same", Values: []domain.AssetValue{{Type: domain.UK_GBP, Value: value}, {Type: domain.INDIA_INR, Value: value}}}}
	input.FXRate = domain.Decimal{}
	if _, err := service.Save(context.Background(), input); err == nil {
		t.Fatal("zero FX rate succeeded")
	}
}

func newTestStore(t *testing.T) *database.SQLiteStore {
	t.Helper()
	db, err := database.Open(context.Background(), t.TempDir()+"/finny.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(context.Background(), db); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return database.NewSQLiteStore(db)
}

func mustDecimal(t *testing.T, value string) domain.Decimal {
	t.Helper()
	parsed, err := domain.NewDecimal(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
