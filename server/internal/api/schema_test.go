package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func TestValidDashboardFixtures(t *testing.T) {
	for _, name := range []string{"dashboard-response.json", "dashboard-request.json"} {
		data, err := os.ReadFile(filepath.Join("testdata", "valid", name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "dashboard-response.json" {
			var response DashboardResponse
			if err := json.Unmarshal(data, &response); err != nil {
				t.Fatalf("decode %s: %v", name, err)
			}
			if response.History[0].CommittedAt != "2026-07-15T13:00:00+01:00" {
				t.Fatalf("timestamp = %q", response.History[0].CommittedAt)
			}
			continue
		}
		var request DashboardRequest
		if err := json.Unmarshal(data, &request); err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
		if request.Revision != 3 || request.Assets[0].ID != 0 {
			t.Fatalf("request = %+v", request)
		}
	}
}

func TestInvalidDashboardFixtures(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("testdata", "invalid"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join("testdata", "invalid", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var request DashboardRequest
		if err := json.Unmarshal(data, &request); err == nil {
			t.Errorf("invalid fixture %s decoded successfully", entry.Name())
		}
	}
}

func TestDashboardResponseFormatsLondonTimestampsAndPreservesDecimals(t *testing.T) {
	value := mustDecimal(t, "100.50")
	dashboard := domain.Dashboard{
		CurrentFXRate: value,
		History: []domain.Snapshot{{
			CommittedAt: time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC),
			FXRate:      value,
		}},
	}
	data, err := json.Marshal(NewDashboardResponse(dashboard))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || !strings.Contains(string(data), `"committedAt":"2026-07-15T13:00:00+01:00"`) || !strings.Contains(string(data), `"currentFxRate":"100.5"`) {
		t.Fatalf("response JSON = %s", data)
	}
}

func TestNewDashboardResponseRecalculatesCombinedTotalsForCurrentAndHistory(t *testing.T) {
	uk := mustDecimal(t, "225")
	india := mustDecimal(t, "900")
	fxRate := mustDecimal(t, "100")
	assets := []domain.Asset{{
		ID:   1,
		Name: "Mixed savings",
		Values: []domain.AssetValue{
			{Type: domain.UK_GBP, Value: uk},
			{Type: domain.INDIA_INR, Value: india},
		},
	}}
	staleTotals := domain.DashboardTotals{Combined: []domain.CombinedTotal{
		{Currency: domain.CURRENCY_GBP, Value: uk},
		{Currency: domain.CURRENCY_INR, Value: uk},
	}}
	dashboard := domain.Dashboard{
		Assets:        assets,
		CurrentFXRate: fxRate,
		CurrentTotals: staleTotals,
		History: []domain.Snapshot{{
			ID:          1,
			CommittedAt: time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC),
			FXRate:      fxRate,
			Assets:      assets,
			Totals:      staleTotals,
		}},
	}

	response := NewDashboardResponse(dashboard)
	assertCombinedTotals(t, response.CurrentTotals.Combined, "234", "23400")
	assertCombinedTotals(t, response.History[0].Totals.Combined, "234", "23400")
}

func assertCombinedTotals(t *testing.T, totals []domain.CombinedTotal, wantGBP, wantINR string) {
	t.Helper()
	if len(totals) != 2 || totals[0].Value.String() != wantGBP || totals[1].Value.String() != wantINR {
		t.Fatalf("combined totals = %+v, want GBP %s and INR %s", totals, wantGBP, wantINR)
	}
}

func TestErrorStatusMapping(t *testing.T) {
	for _, test := range []struct {
		code   ErrorCode
		status int
	}{
		{ERROR_CODE_REVISION_CONFLICT, 409},
		{ERROR_CODE_IDEMPOTENCY_CONFLICT, 409},
		{ERROR_CODE_NOT_FOUND, 404},
		{ERROR_CODE_INTERNAL, 500},
		{ERROR_CODE_INVALID_JSON, 400},
	} {
		if got := StatusCodeForError(test.code); got != test.status {
			t.Errorf("StatusCodeForError(%q) = %d, want %d", test.code, got, test.status)
		}
	}
}

func TestIdempotencyKeyValidation(t *testing.T) {
	if err := ValidateIdempotencyKey("save-123"); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"", "   ", string(make([]byte, MAX_IDEMPOTENCY_KEY_LENGTH+1))} {
		if err := ValidateIdempotencyKey(key); err == nil {
			t.Errorf("ValidateIdempotencyKey(%q) succeeded", key)
		}
	}
}

func mustDecimal(t *testing.T, value string) domain.Decimal {
	t.Helper()
	parsed, err := domain.NewDecimal(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
