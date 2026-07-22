package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/WorldOccupier/finny/server/internal/domain"
	"github.com/WorldOccupier/finny/server/internal/importpreview"
	"github.com/WorldOccupier/finny/server/internal/importservice"
)

const (
	STATEMENTS_PREVIEW_ROUTE = "/api/statements/preview"
	STATEMENTS_CONFIRM_ROUTE = "/api/statements/confirm"
	STATEMENTS_ROUTE         = "/api/statements"
	TRANSACTIONS_ROUTE       = "/api/transactions"
	SPENDING_SUMMARY_ROUTE   = "/api/spending/summary"
	MAX_IMPORT_SIZE          = 10 << 20
)

type ImportHandler struct{ service *importservice.Service }

func NewImportHandler(service *importservice.Service) *ImportHandler {
	return &ImportHandler{service: service}
}

func (h *ImportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case STATEMENTS_PREVIEW_ROUTE:
		h.preview(w, r)
	case STATEMENTS_CONFIRM_ROUTE:
		h.confirm(w, r)
	case STATEMENTS_ROUTE:
		h.statements(w, r)
	case SPENDING_SUMMARY_ROUTE:
		h.summary(w, r)
	default:
		h.transactions(w, r)
	}
}

func (h *ImportHandler) summary(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Summary(r.Context(), r.URL.Query().Get("accountId"))
	if err != nil {
		writeError(w, ERROR_CODE_INTERNAL, "summary could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"summary": items})
}

func (h *ImportHandler) preview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ERROR_CODE_INVALID_JSON, "method not allowed")
		return
	}
	if err := r.ParseMultipartForm(MAX_IMPORT_SIZE); err != nil {
		writeError(w, ERROR_CODE_INVALID_JSON, "invalid multipart upload")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, ERROR_CODE_INVALID_JSON, "file is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, MAX_IMPORT_SIZE+1))
	if err != nil || len(data) > MAX_IMPORT_SIZE {
		writeError(w, ERROR_CODE_INVALID_JSON, "file is too large")
		return
	}
	request, err := mappingFromForm(r, header, file)
	if err != nil {
		writeError(w, ERROR_CODE_VALIDATION, err.Error())
		return
	}
	p, err := h.service.Preview(data, request, domain.UserID(valueOr(r.FormValue("importedBy"), string(domain.USER_ONE))))
	if err != nil {
		writeError(w, ERROR_CODE_VALIDATION, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": p.Token, "checksum": p.Result.Checksum, "periodStart": p.Result.PeriodStart, "periodEnd": p.Result.PeriodEnd, "transactions": p.Result.Transactions, "invalidRows": p.Result.InvalidRows, "validRows": p.Result.ValidRows, "invalidCount": p.Result.InvalidCount})
}

func (h *ImportHandler) confirm(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"token"`
	}
	if json.NewDecoder(r.Body).Decode(&request) != nil || strings.TrimSpace(request.Token) == "" {
		writeError(w, ERROR_CODE_INVALID_JSON, "token is required")
		return
	}
	statement, count, err := h.service.Confirm(r.Context(), request.Token)
	if errors.Is(err, importservice.ErrDuplicateStatement) {
		writeError(w, ERROR_CODE_VALIDATION, "statement already imported")
		return
	}
	if err != nil {
		writeError(w, ERROR_CODE_VALIDATION, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"statement": statement, "importedRows": count})
}
func (h *ImportHandler) statements(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Statements(r.Context())
	if err != nil {
		writeError(w, ERROR_CODE_INTERNAL, "statements could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"statements": items})
}
func (h *ImportHandler) transactions(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Transactions(r.Context(), r.URL.Query().Get("accountId"))
	if err != nil {
		writeError(w, ERROR_CODE_INTERNAL, "transactions could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": items})
}

func mappingFromForm(r *http.Request, header *multipart.FileHeader, _ multipart.File) (importpreview.ImportRequest, error) {
	get := func(name string) int { value, _ := strconv.Atoi(r.FormValue(name)); return value }
	return importpreview.ImportRequest{Filename: header.Filename, AccountID: r.FormValue("accountId"), StatementID: r.FormValue("statementId"), Mapping: importpreview.ColumnMapping{Date: get("date"), Description: get("description"), Amount: get("amount"), Debit: get("debit"), Credit: get("credit"), Currency: get("currency"), Reference: get("reference")}}, nil
}
func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", JSON_CONTENT_TYPE)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
