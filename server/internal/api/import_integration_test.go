package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WorldOccupier/finny/server/internal/database"
	"github.com/WorldOccupier/finny/server/internal/domain"
	"github.com/WorldOccupier/finny/server/internal/importservice"
)

func TestImportAPIIntegrationPreviewConfirmSearchAndSummary(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "finny.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	store := database.NewSQLiteStore(db)
	if err := store.SaveAccount(ctx, domain.Account{ID: "checking", BankSource: "Test bank", AccountLabel: "Checking", Currency: domain.CURRENCY_GBP, Owner: domain.OWNERSHIP_JOINT}); err != nil {
		t.Fatal(err)
	}
	handler := NewImportHandler(importservice.New(store))

	preview := httptest.NewRecorder()
	handler.ServeHTTP(preview, importRequest(t, "statement.csv", "date,description,amount,currency\n2026-07-20,Coffee,-3.50,GBP\ninvalid,Missing date,2,GBP\n"))
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status = %d, body = %s", preview.Code, preview.Body)
	}
	var previewBody struct {
		Token        string `json:"token"`
		ValidRows    int    `json:"validRows"`
		InvalidCount int    `json:"invalidCount"`
	}
	if err := json.NewDecoder(preview.Body).Decode(&previewBody); err != nil {
		t.Fatal(err)
	}
	if previewBody.Token == "" || previewBody.ValidRows != 1 || previewBody.InvalidCount != 1 {
		t.Fatalf("preview = %+v", previewBody)
	}

	confirmBody := `{"token":"` + previewBody.Token + `"}`
	confirm := httptest.NewRecorder()
	handler.ServeHTTP(confirm, httptest.NewRequest(http.MethodPost, STATEMENTS_CONFIRM_ROUTE, strings.NewReader(confirmBody)))
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, body = %s", confirm.Code, confirm.Body)
	}

	transactions := httptest.NewRecorder()
	handler.ServeHTTP(transactions, httptest.NewRequest(http.MethodGet, TRANSACTIONS_ROUTE+"?accountId=checking&user=user_two&q=coffee&page=1&pageSize=1", nil))
	if transactions.Code != http.StatusOK || !strings.Contains(transactions.Body.String(), `"amount":"-3.5"`) {
		t.Fatalf("transactions status/body = %d/%s", transactions.Code, transactions.Body)
	}

	for _, period := range []string{"day", "week", "month", "year"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, SPENDING_SUMMARY_ROUTE+"?period="+period+"&date=2026-07-20&user=user_two", nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"period":"`+period+`"`) {
			t.Fatalf("summary %s = %d/%s", period, response.Code, response.Body)
		}
	}
}

func TestImportAPIIntegrationRejectsUnsupportedFileAndUnknownRoute(t *testing.T) {
	store := testStore(t)
	handler := NewImportHandler(importservice.New(store))
	unsupported := httptest.NewRecorder()
	handler.ServeHTTP(unsupported, importRequest(t, "statement.txt", "not a statement"))
	if unsupported.Code != http.StatusBadRequest {
		t.Fatalf("unsupported file status = %d", unsupported.Code)
	}
	unknown := httptest.NewRecorder()
	handler.ServeHTTP(unknown, httptest.NewRequest(http.MethodGet, "/api/unknown", nil))
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown route status = %d", unknown.Code)
	}
}

func importRequest(t *testing.T, filename, content string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(file, content)
	for key, value := range map[string]string{"accountId": "checking", "date": "0", "description": "1", "amount": "2", "debit": "-1", "credit": "-1", "currency": "3"} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, STATEMENTS_PREVIEW_ROUTE, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
