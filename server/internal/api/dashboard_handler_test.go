package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
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

func TestDashboardHandlerPostSavesAndReplaysIdempotently(t *testing.T) {
	store := testStore(t)
	handler := NewDashboardHandler(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := []byte(`{"revision":0,"assets":[{"id":0,"name":"Savings","values":[{"type":"UKGBP","value":"100"},{"type":"INDIAINR","value":"5000"}]}],"fxRate":"100","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`)
	request := httptest.NewRequest(http.MethodPost, "/api/dashboard", bytes.NewReader(body))
	request.Header.Set(IDEMPOTENCY_KEY_HEADER, "save-1")
	first := httptest.NewRecorder()
	handler.ServeHTTP(first, request)
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d, body = %s", first.Code, first.Body)
	}
	var response DashboardResponse
	if err := json.Unmarshal(first.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Revision != 1 || len(response.History) != 1 {
		t.Fatalf("response = %+v", response)
	}
	retry := httptest.NewRecorder()
	retryRequest := httptest.NewRequest(http.MethodPost, "/api/dashboard", bytes.NewReader(body))
	retryRequest.Header.Set(IDEMPOTENCY_KEY_HEADER, "save-1")
	handler.ServeHTTP(retry, retryRequest)
	if retry.Code != http.StatusOK || retry.Body.String() != first.Body.String() {
		t.Fatalf("retry status/body = %d/%s, want %d/%s", retry.Code, retry.Body, first.Code, first.Body)
	}
}

func TestDashboardHandlerPostRejectsStaleRevisionAndKeyReuse(t *testing.T) {
	store := testStore(t)
	handler := NewDashboardHandler(store, slog.Default())
	body := []byte(`{"revision":0,"assets":[],"fxRate":"1","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`)
	first := httptest.NewRequest(http.MethodPost, "/api/dashboard", bytes.NewReader(body))
	first.Header.Set(IDEMPOTENCY_KEY_HEADER, "save-2")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, first)
	if response.Code != http.StatusOK {
		t.Fatalf("first status = %d, body = %s", response.Code, response.Body)
	}
	changed := []byte(`{"revision":0,"assets":[],"fxRate":"2","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`)
	reuse := httptest.NewRequest(http.MethodPost, "/api/dashboard", bytes.NewReader(changed))
	reuse.Header.Set(IDEMPOTENCY_KEY_HEADER, "save-2")
	reuseResponse := httptest.NewRecorder()
	handler.ServeHTTP(reuseResponse, reuse)
	if reuseResponse.Code != http.StatusConflict {
		t.Fatalf("reuse status = %d, want %d", reuseResponse.Code, http.StatusConflict)
	}
	stale := httptest.NewRequest(http.MethodPost, "/api/dashboard", bytes.NewReader(body))
	stale.Header.Set(IDEMPOTENCY_KEY_HEADER, "save-3")
	staleResponse := httptest.NewRecorder()
	handler.ServeHTTP(staleResponse, stale)
	if staleResponse.Code != http.StatusConflict {
		t.Fatalf("stale status = %d, want %d", staleResponse.Code, http.StatusConflict)
	}
}

func TestDashboardHandlerRejectsConcurrentSavesWithSameRevision(t *testing.T) {
	store := testStore(t)
	handler := NewDashboardHandler(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := []byte(`{"revision":0,"assets":[],"fxRate":"1","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`)
	start := make(chan struct{})
	responses := make(chan int, 2)
	var wait sync.WaitGroup
	for _, key := range []string{"concurrent-1", "concurrent-2"} {
		wait.Add(1)
		go func(key string) {
			defer wait.Done()
			<-start
			request := httptest.NewRequest(http.MethodPost, DASHBOARD_ROUTE, bytes.NewReader(body))
			request.Header.Set(IDEMPOTENCY_KEY_HEADER, key)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			responses <- recorder.Code
		}(key)
	}
	close(start)
	wait.Wait()
	close(responses)
	var success, conflict int
	for status := range responses {
		switch status {
		case http.StatusOK:
			success++
		case http.StatusConflict:
			conflict++
		default:
			t.Fatalf("concurrent status = %d", status)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("concurrent results = success %d, conflict %d", success, conflict)
	}
	dashboard, err := store.LoadDashboard(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Revision != 1 || len(dashboard.History) != 1 {
		t.Fatalf("concurrent dashboard = revision %d, history %d", dashboard.Revision, len(dashboard.History))
	}
}

func TestDashboardHandlerConcurrentSameKeyCommitsOnceAndReplays(t *testing.T) {
	store := testStore(t)
	handler := NewDashboardHandler(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := []byte(`{"revision":0,"assets":[],"fxRate":"1","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`)
	start := make(chan struct{})
	responses := make(chan string, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			request := httptest.NewRequest(http.MethodPost, DASHBOARD_ROUTE, bytes.NewReader(body))
			request.Header.Set(IDEMPOTENCY_KEY_HEADER, "concurrent-same-key")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				responses <- fmt.Sprintf("status %d: %s", recorder.Code, recorder.Body.String())
				return
			}
			responses <- recorder.Body.String()
		}()
	}
	close(start)
	wait.Wait()
	close(responses)
	var bodies []string
	for response := range responses {
		bodies = append(bodies, response)
	}
	if len(bodies) != 2 || bodies[0] != bodies[1] {
		t.Fatalf("same-key responses = %q", bodies)
	}
	dashboard, err := store.LoadDashboard(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Revision != 1 || len(dashboard.History) != 1 {
		t.Fatalf("same-key dashboard = revision %d, history %d", dashboard.Revision, len(dashboard.History))
	}
}

func TestDashboardHandlerIdempotencyFailureRollsBackDashboard(t *testing.T) {
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "finny.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TRIGGER fail_idempotency_insert BEFORE INSERT ON idempotency_keys BEGIN SELECT RAISE(FAIL, 'idempotency unavailable'); END`)
	if err != nil {
		t.Fatal(err)
	}
	store := database.NewSQLiteStore(db)
	handler := NewDashboardHandler(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, DASHBOARD_ROUTE, strings.NewReader(`{"revision":0,"assets":[],"fxRate":"1","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`))
	request.Header.Set(IDEMPOTENCY_KEY_HEADER, "failure-1")
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	dashboard, err := store.LoadDashboard(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Revision != 0 || len(dashboard.History) != 0 {
		t.Fatalf("failed save changed dashboard = revision %d, history %d", dashboard.Revision, len(dashboard.History))
	}
}

func TestDashboardHandlerMapsPostStoreFailureToInternalError(t *testing.T) {
	handler := NewDashboardHandler(failingSaveStore{Store: testStore(t)}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, DASHBOARD_ROUTE, strings.NewReader(`{"revision":0,"assets":[],"fxRate":"1","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"}}`))
	request.Header.Set(IDEMPOTENCY_KEY_HEADER, "failure-2")
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestDashboardHandlerRejectsOversizedBody(t *testing.T) {
	handler := NewDashboardHandler(testStore(t), slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := append([]byte(`{"revision":0,"assets":[],"fxRate":"1","spendingLimits":[],"income":{"userOneGBP":"0","userTwoGBP":"0"},"padding":"`), bytes.Repeat([]byte("x"), MAX_REQUEST_BODY)...)
	body = append(body, []byte(`"}`)...)
	request := httptest.NewRequest(http.MethodPost, DASHBOARD_ROUTE, bytes.NewReader(body))
	request.Header.Set(IDEMPOTENCY_KEY_HEADER, "large-1")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

type failingStore struct{ database.Store }

type failingSaveStore struct{ database.Store }

func (f failingSaveStore) SaveDashboardSnapshot(context.Context, database.DashboardSnapshotSave) (database.DashboardSnapshotCommit, error) {
	return database.DashboardSnapshotCommit{}, fmt.Errorf("database exploded")
}

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
