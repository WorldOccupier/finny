package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/WorldOccupier/finny/server/internal/database"
	"github.com/WorldOccupier/finny/server/internal/domain"
)

func TestDashboardHandlerReturnsEmptyDashboard(t *testing.T) {
	handler := NewDashboardHandler(testStore(t), slog.New(slog.NewTextHandler(io.Discard, nil)))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/dashboard", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != JSON_CONTENT_TYPE {
		t.Fatalf("content type = %q, want %q", contentType, JSON_CONTENT_TYPE)
	}
	var response DashboardResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Revision != 0 || len(response.Assets) != 0 || len(response.History) != 0 || len(response.SpendingLimits) != 0 {
		t.Fatalf("empty response = %+v", response)
	}
	if response.Assets == nil || response.History == nil || response.SpendingLimits == nil {
		t.Fatal("empty collections must be JSON arrays")
	}
	if response.CurrentTotals.Country == nil || response.CurrentTotals.Combined == nil {
		t.Fatal("empty total collections must be JSON arrays")
	}
	if response.CurrentFXRate.String() != "0" {
		t.Fatalf("empty FX rate = %s", response.CurrentFXRate)
	}
}

func TestDashboardHandlerReturnsCompleteDashboardInHistoryOrder(t *testing.T) {
	store := testStore(t)
	value := mustDecimal(t, "100.50")
	indiaValue := mustDecimal(t, "5000")
	assets := []domain.Asset{{ID: 0, Name: "Savings", Values: []domain.AssetValue{
		{Type: domain.UK_GBP, Value: value},
		{Type: domain.INDIA_INR, Value: indiaValue},
	}}}
	firstTotals := domain.DashboardTotals{
		Country:  []domain.TotalValue{{Type: domain.UK_GBP, Value: value}, {Type: domain.INDIA_INR, Value: indiaValue}},
		Combined: []domain.CombinedTotal{{Currency: domain.CURRENCY_GBP, Value: mustDecimal(t, "150.50")}, {Currency: domain.CURRENCY_INR, Value: mustDecimal(t, "15050")}},
	}
	secondTotals := domain.DashboardTotals{
		Country:  []domain.TotalValue{{Type: domain.UK_GBP, Value: mustDecimal(t, "200")}, {Type: domain.INDIA_INR, Value: mustDecimal(t, "6000")}},
		Combined: []domain.CombinedTotal{{Currency: domain.CURRENCY_GBP, Value: mustDecimal(t, "260")}, {Currency: domain.CURRENCY_INR, Value: mustDecimal(t, "26000")}},
	}
	dashboard := domain.Dashboard{
		Revision:      7,
		Assets:        assets,
		CurrentFXRate: mustDecimal(t, "100"),
		CurrentTotals: secondTotals,
		History: []domain.Snapshot{
			{ID: 2, CommittedAt: time.Date(2026, time.July, 16, 13, 0, 0, 0, domain.LondonLocation()), FXRate: mustDecimal(t, "100"), Assets: assets, Totals: secondTotals},
			{ID: 1, CommittedAt: time.Date(2026, time.July, 15, 13, 0, 0, 0, domain.LondonLocation()), FXRate: value, Assets: assets, Totals: firstTotals},
		},
		SpendingLimits: []domain.SpendingLimit{{Key: "rent", Amount: value, Currency: domain.CURRENCY_GBP}},
		Income:         domain.IncomeTotals{UserOneGBP: mustDecimal(t, "3000"), UserTwoGBP: mustDecimal(t, "2500")},
	}
	if err := store.SaveDashboard(context.Background(), dashboard); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	NewDashboardHandler(store, slog.New(slog.NewTextHandler(io.Discard, nil))).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/dashboard", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body, err := io.ReadAll(recorder.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"currentFxRate":"100"`) || !strings.Contains(string(body), `"value":"100.5"`) {
		t.Fatalf("response did not preserve decimal strings: %s", body)
	}
	var response DashboardResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Revision != 7 || len(response.Assets) != 1 || len(response.SpendingLimits) != 1 || response.Income.UserOneGBP.String() != "3000" {
		t.Fatalf("incomplete response = %+v", response)
	}
	if len(response.History) != 2 || response.History[0].ID != 1 || response.History[1].ID != 2 {
		t.Fatalf("history order = %+v", response.History)
	}
	if response.History[0].CommittedAt != "2026-07-15T13:00:00+01:00" {
		t.Fatalf("timestamp = %q", response.History[0].CommittedAt)
	}
	if response.CurrentTotals.Combined[0].Value.String() != "260" || response.CurrentTotals.Combined[1].Value.String() != "26000" {
		t.Fatalf("current totals = %+v", response.CurrentTotals)
	}
}

func TestDashboardHandlerMapsStoreFailuresToSafeInternalError(t *testing.T) {
	handler := NewDashboardHandler(failingStore{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/dashboard", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(recorder.Body.String(), "database exploded") {
		t.Fatal("internal database error leaked into response")
	}
	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Error.Code != ERROR_CODE_INTERNAL {
		t.Fatalf("error code = %q", response.Error.Code)
	}
}

type failingStore struct{ database.Store }

func (failingStore) LoadDashboard(context.Context) (domain.Dashboard, error) {
	return domain.Dashboard{}, &testError{"database exploded"}
}

type testError struct{ message string }

func (e *testError) Error() string { return e.message }

func testStore(t *testing.T) *database.SQLiteStore {
	t.Helper()
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "finny.db"))
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
