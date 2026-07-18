package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/WorldOccupier/finny/server/internal/database"
	"github.com/WorldOccupier/finny/server/internal/snapshot"
)

const (
	DASHBOARD_ROUTE   = "/api/dashboard"
	JSON_CONTENT_TYPE = "application/json"
	MAX_REQUEST_BODY  = 1 << 20
)

type DashboardHandler struct {
	store    database.Store
	logger   *slog.Logger
	snapshot *snapshot.Service
}

func NewDashboardHandler(store database.Store, logger *slog.Logger) *DashboardHandler {
	return &DashboardHandler{store: store, logger: logger, snapshot: snapshot.NewService(store, time.Now)}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, ERROR_CODE_INTERNAL, "the dashboard could not be loaded")
		return
	}
	if r.Method == http.MethodPost {
		h.handlePost(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, ERROR_CODE_INTERNAL, "method not allowed")
		return
	}

	dashboard, err := h.store.LoadDashboard(r.Context())
	if err != nil {
		if h.logger != nil {
			h.logger.Error("dashboard read failed", "error", err)
		}
		writeError(w, ERROR_CODE_INTERNAL, "the dashboard could not be loaded")
		return
	}

	response := NewDashboardResponse(dashboard)
	w.Header().Set("Content-Type", JSON_CONTENT_TYPE)
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		if h.logger == nil {
			return
		}
		h.logger.Error("dashboard response failed", "error", err)
	}
}

func (h *DashboardHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get(IDEMPOTENCY_KEY_HEADER)
	if err := ValidateIdempotencyKey(key); err != nil {
		writeError(w, ERROR_CODE_INVALID_IDEMPOTENCY_KEY, "a valid Idempotency-Key header is required")
		return
	}
	key = NormalizeIdempotencyKey(key)
	body, err := io.ReadAll(io.LimitReader(r.Body, MAX_REQUEST_BODY+1))
	if err != nil {
		writeError(w, ERROR_CODE_INVALID_JSON, "the request body could not be read")
		return
	}
	if len(body) > MAX_REQUEST_BODY {
		writeError(w, ERROR_CODE_INVALID_JSON, "the request body is too large")
		return
	}
	var request DashboardRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&request)
	if err != nil {
		writeError(w, ERROR_CODE_INVALID_JSON, "the request body is invalid")
		return
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		writeError(w, ERROR_CODE_INVALID_JSON, "the request body must contain one JSON object")
		return
	}
	canonical, err := json.Marshal(request)
	if err != nil {
		writeError(w, ERROR_CODE_INTERNAL, "the request could not be normalized")
		return
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(canonical))
	existing, err := h.store.GetIdempotencyResult(r.Context(), key)
	if err == nil {
		if existing.RequestHash != hash {
			writeError(w, ERROR_CODE_IDEMPOTENCY_CONFLICT, "the idempotency key was used with a different request")
			return
		}
		w.Header().Set("Content-Type", JSON_CONTENT_TYPE)
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, bytes.NewReader([]byte(existing.ResponseJSON)))
		return
	}
	if !errors.Is(err, database.ErrNotFound) {
		writeError(w, ERROR_CODE_INTERNAL, "the dashboard could not be loaded")
		return
	}
	if h.snapshot == nil {
		writeError(w, ERROR_CODE_INTERNAL, "the dashboard could not be saved")
		return
	}
	prepared, err := h.snapshot.Prepare(r.Context(), snapshot.SnapshotInput{Assets: request.Assets, FXRate: request.FXRate, SpendingLimits: request.SpendingLimits, Income: request.Income})
	if err != nil {
		if snapshot.IsValidationError(err) {
			writeError(w, ERROR_CODE_VALIDATION, "the dashboard could not be saved")
			return
		}
		writeError(w, ERROR_CODE_INTERNAL, "the dashboard could not be saved")
		return
	}
	response := NewDashboardResponse(prepared.Dashboard)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		writeError(w, ERROR_CODE_INTERNAL, "the dashboard response could not be created")
		return
	}
	result := database.IdempotencyResult{Key: key, RequestHash: hash, ResponseJSON: string(responseJSON), CreatedAt: time.Now()}
	prepared.Save.ExpectedRevision = request.Revision
	prepared.Save.Revision = request.Revision + 1
	prepared.Save.Idempotency = result
	commit, err := h.store.SaveDashboardSnapshot(r.Context(), prepared.Save)
	if errors.Is(err, database.ErrRevisionConflict) {
		writeError(w, ERROR_CODE_REVISION_CONFLICT, "the dashboard revision is stale")
		return
	}
	if errors.Is(err, database.ErrIdempotencyConflict) {
		writeError(w, ERROR_CODE_IDEMPOTENCY_CONFLICT, "the idempotency key was used with a different request")
		return
	}
	if err != nil {
		if h.logger != nil {
			h.logger.Error("dashboard atomic save failed", "error", err)
		}
		writeError(w, ERROR_CODE_INTERNAL, "the dashboard could not be saved")
		return
	}
	if commit.Replayed {
		responseJSON = []byte(commit.Idempotency.ResponseJSON)
	}
	w.Header().Set("Content-Type", JSON_CONTENT_TYPE)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseJSON)
}

func writeError(w http.ResponseWriter, code ErrorCode, message string) {
	w.Header().Set("Content-Type", JSON_CONTENT_TYPE)
	w.WriteHeader(StatusCodeForError(code))
	_ = json.NewEncoder(w).Encode(NewError(code, message))
}

var _ http.Handler = (*DashboardHandler)(nil)
