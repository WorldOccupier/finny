package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/WorldOccupier/finny/server/internal/database"
)

func TestDashboardAPIIntegrationCreatesLaterSnapshotAndPreservesHistory(t *testing.T) {
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "finny.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	handler := NewDashboardHandler(database.NewSQLiteStore(db), slog.New(slog.NewTextHandler(io.Discard, nil)))

	firstBody := []byte(`{"revision":0,"assets":[{"id":0,"name":"Savings","values":[{"type":"UKGBP","value":"100"},{"type":"INDIAINR","value":"5000"}]}],"fxRate":"100","spendingLimits":[],"income":{"userOneGBP":"3000","userTwoGBP":"2500"}}`)
	first := postDashboard(t, handler, "integration-first", firstBody)
	if first.Revision != 1 || len(first.History) != 1 {
		t.Fatalf("first dashboard = revision %d, history %d", first.Revision, len(first.History))
	}

	secondBody := []byte(`{"revision":1,"assets":[],"fxRate":"110","spendingLimits":[],"income":{"userOneGBP":"3100","userTwoGBP":"2500"}}`)
	second := postDashboard(t, handler, "integration-second", secondBody)
	if second.Revision != 2 || len(second.History) != 2 || len(second.Assets) != 0 {
		t.Fatalf("second dashboard = revision %d, history %d, assets %d", second.Revision, len(second.History), len(second.Assets))
	}
	if second.History[0].Assets[0].Name != "Savings" {
		t.Fatal("removed asset was not retained in historical snapshot")
	}

	request := httptest.NewRequest(http.MethodGet, DASHBOARD_ROUTE, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("GET status = %d", response.Code)
	}
	var loaded DashboardResponse
	if err := json.NewDecoder(response.Body).Decode(&loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.CurrentFXRate.String() != "110" || loaded.History[0].FXRate.String() != "100" {
		t.Fatalf("FX history was not frozen: current %s, first %s", loaded.CurrentFXRate, loaded.History[0].FXRate)
	}
}

func TestDashboardAPIIntegrationConvertsAssetCurrencyMemberships(t *testing.T) {
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "finny.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	handler := NewDashboardHandler(database.NewSQLiteStore(db), slog.New(slog.NewTextHandler(io.Discard, nil)))

	first := postDashboard(t, handler, "membership-first", []byte(`{"revision":0,"assets":[{"id":0,"name":"Savings","valueTypes":["UKGBP","INDIAINR"],"values":[{"type":"UKGBP","value":"100"},{"type":"INDIAINR","value":"5000"}]}],"fxRate":"100","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`))
	if len(first.Assets[0].Values) != 2 {
		t.Fatalf("first asset values = %+v", first.Assets[0].Values)
	}
	assertSavedAssetTypes(t, handler, []string{"UKGBP", "INDIAINR"})

	gbpOnly := postDashboard(t, handler, "membership-gbp", []byte(`{"revision":1,"assets":[{"id":0,"name":"Savings","valueTypes":["UKGBP"],"values":[{"type":"UKGBP","value":"125"}]}],"fxRate":"100","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`))
	if len(gbpOnly.Assets[0].Values) != 1 || gbpOnly.Assets[0].Values[0].Type != "UKGBP" {
		t.Fatalf("GBP-only asset = %+v", gbpOnly.Assets[0].Values)
	}
	if len(gbpOnly.History[0].Assets[0].Values) != 2 {
		t.Fatalf("GBP conversion changed history = %+v", gbpOnly.History[0].Assets[0].Values)
	}
	assertSavedAssetTypes(t, handler, []string{"UKGBP"})

	inrOnly := postDashboard(t, handler, "membership-inr", []byte(`{"revision":2,"assets":[{"id":0,"name":"Savings","valueTypes":["INDIAINR"],"values":[{"type":"INDIAINR","value":"6500"}]}],"fxRate":"100","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`))
	if len(inrOnly.Assets[0].Values) != 1 || inrOnly.Assets[0].Values[0].Type != "INDIAINR" {
		t.Fatalf("INR-only asset = %+v", inrOnly.Assets[0].Values)
	}
	assertSavedAssetTypes(t, handler, []string{"INDIAINR"})
}

func assertSavedAssetTypes(t *testing.T, handler http.Handler, expectedTypes []string) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, DASHBOARD_ROUTE, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("GET status = %d", response.Code)
	}
	var dashboard DashboardResponse
	if err := json.NewDecoder(response.Body).Decode(&dashboard); err != nil {
		t.Fatal(err)
	}
	if len(dashboard.Assets) != 1 || len(dashboard.Assets[0].Values) != len(expectedTypes) {
		t.Fatalf("saved asset values = %+v", dashboard.Assets)
	}
	actualTypes := make(map[string]struct{}, len(dashboard.Assets[0].Values))
	for _, value := range dashboard.Assets[0].Values {
		actualTypes[string(value.Type)] = struct{}{}
	}
	for _, expectedType := range expectedTypes {
		if _, found := actualTypes[expectedType]; !found {
			t.Fatalf("saved asset value types = %v, missing %s", actualTypes, expectedType)
		}
	}
}

func postDashboard(t *testing.T, handler http.Handler, key string, body []byte) DashboardResponse {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, DASHBOARD_ROUTE, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(IDEMPOTENCY_KEY_HEADER, key)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("POST status = %d, body = %s", response.Code, response.Body)
	}
	var dashboard DashboardResponse
	if err := json.NewDecoder(response.Body).Decode(&dashboard); err != nil {
		t.Fatal(err)
	}
	return dashboard
}
