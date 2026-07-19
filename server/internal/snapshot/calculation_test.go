package snapshot

import (
	"testing"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func TestCalculateTotalsConvertsBothCombinedCurrencies(t *testing.T) {
	assets := []domain.Asset{{
		ID:   1,
		Name: "Mixed savings",
		Values: []domain.AssetValue{
			{Type: domain.UK_GBP, Value: mustDecimal(t, "100")},
			{Type: domain.INDIA_INR, Value: mustDecimal(t, "500")},
		},
	}}

	totals, err := CalculateTotals(assets, mustDecimal(t, "4"))
	if err != nil {
		t.Fatal(err)
	}

	if got := totals.Combined[0].Value.String(); got != "225" {
		t.Fatalf("combined GBP = %s, want 225", got)
	}
	if got := totals.Combined[1].Value.String(); got != "900" {
		t.Fatalf("combined INR = %s, want 900", got)
	}
	if totals.Combined[0].Value.String() == totals.Combined[1].Value.String() {
		t.Fatal("combined totals should retain their converted currency values")
	}
}
