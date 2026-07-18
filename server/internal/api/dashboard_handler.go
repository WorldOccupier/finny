package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/WorldOccupier/finny/server/internal/database"
)

const (
	DASHBOARD_ROUTE   = "GET /api/dashboard"
	JSON_CONTENT_TYPE = "application/json"
)

type DashboardHandler struct {
	store  database.Store
	logger *slog.Logger
}

func NewDashboardHandler(store database.Store, logger *slog.Logger) *DashboardHandler {
	return &DashboardHandler{store: store, logger: logger}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, ERROR_CODE_INTERNAL, "the dashboard could not be loaded")
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

func writeError(w http.ResponseWriter, code ErrorCode, message string) {
	w.Header().Set("Content-Type", JSON_CONTENT_TYPE)
	w.WriteHeader(StatusCodeForError(code))
	_ = json.NewEncoder(w).Encode(NewError(code, message))
}

var _ http.Handler = (*DashboardHandler)(nil)
