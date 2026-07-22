package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WorldOccupier/finny/server/internal/importservice"
)

func TestImportHandlerRejectsInvalidConfirmJSON(t *testing.T) {
	handler := NewImportHandler(importservice.New(testStore(t)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, STATEMENTS_CONFIRM_ROUTE, strings.NewReader(`{"token":7}`)))
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_json"`) {
		t.Fatalf("response = %d/%s", response.Code, response.Body)
	}
}

func TestImportHandlerValidatesSummaryAndMethods(t *testing.T) {
	handler := NewImportHandler(importservice.New(testStore(t)))
	invalidPeriod := httptest.NewRecorder()
	handler.ServeHTTP(invalidPeriod, httptest.NewRequest(http.MethodGet, SPENDING_SUMMARY_ROUTE+"?period=quarter", nil))
	if invalidPeriod.Code != http.StatusBadRequest {
		t.Fatalf("invalid period status = %d", invalidPeriod.Code)
	}

	wrongMethod := httptest.NewRecorder()
	handler.ServeHTTP(wrongMethod, httptest.NewRequest(http.MethodPost, TRANSACTIONS_ROUTE, nil))
	if wrongMethod.Code != http.StatusBadRequest {
		t.Fatalf("wrong method status = %d", wrongMethod.Code)
	}
}
