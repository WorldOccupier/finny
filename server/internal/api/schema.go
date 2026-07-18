package api

import (
	"fmt"
	"strings"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

const (
	IDEMPOTENCY_KEY_HEADER     = "Idempotency-Key"
	MIN_IDEMPOTENCY_KEY_LENGTH = 1
	MAX_IDEMPOTENCY_KEY_LENGTH = 255
)

type DashboardResponse struct {
	Revision       int64                  `json:"revision"`
	Assets         []domain.Asset         `json:"assets"`
	CurrentFXRate  domain.Decimal         `json:"currentFxRate"`
	CurrentTotals  domain.DashboardTotals `json:"currentTotals"`
	History        []SnapshotResponse     `json:"history"`
	SpendingLimits []domain.SpendingLimit `json:"spendingLimits"`
	Income         domain.IncomeTotals    `json:"income"`
}

type SnapshotResponse struct {
	ID          int64                  `json:"id"`
	CommittedAt string                 `json:"committedAt"`
	FXRate      domain.Decimal         `json:"fxRate"`
	Assets      []domain.Asset         `json:"assets"`
	Totals      domain.DashboardTotals `json:"totals"`
}

type DashboardRequest struct {
	Revision       int64                  `json:"revision"`
	Assets         []domain.Asset         `json:"assets"`
	FXRate         domain.Decimal         `json:"fxRate"`
	SpendingLimits []domain.SpendingLimit `json:"spendingLimits"`
	Income         domain.IncomeTotals    `json:"income"`
}

func NewDashboardResponse(dashboard domain.Dashboard) DashboardResponse {
	history := make([]SnapshotResponse, 0, len(dashboard.History))
	for _, snapshot := range dashboard.History {
		history = append(history, SnapshotResponse{
			ID:          snapshot.ID,
			CommittedAt: domain.FormatLondonTimestamp(snapshot.CommittedAt),
			FXRate:      snapshot.FXRate,
			Assets:      snapshot.Assets,
			Totals:      snapshot.Totals,
		})
	}
	return DashboardResponse{
		Revision:       dashboard.Revision,
		Assets:         dashboard.Assets,
		CurrentFXRate:  dashboard.CurrentFXRate,
		CurrentTotals:  dashboard.CurrentTotals,
		History:        history,
		SpendingLimits: dashboard.SpendingLimits,
		Income:         dashboard.Income,
	}
}

type ErrorCode string

const (
	ERROR_CODE_INVALID_JSON            ErrorCode = "invalid_json"
	ERROR_CODE_VALIDATION              ErrorCode = "validation_error"
	ERROR_CODE_INVALID_DECIMAL         ErrorCode = "invalid_decimal"
	ERROR_CODE_INVALID_CURRENCY        ErrorCode = "invalid_currency"
	ERROR_CODE_MISSING_ASSET_VALUE     ErrorCode = "missing_asset_value"
	ERROR_CODE_DUPLICATE_ID            ErrorCode = "duplicate_id"
	ERROR_CODE_INVALID_IDEMPOTENCY_KEY ErrorCode = "invalid_idempotency_key"
	ERROR_CODE_REVISION_CONFLICT       ErrorCode = "revision_conflict"
	ERROR_CODE_IDEMPOTENCY_CONFLICT    ErrorCode = "idempotency_conflict"
	ERROR_CODE_NOT_FOUND               ErrorCode = "not_found"
	ERROR_CODE_INTERNAL                ErrorCode = "internal_error"
)

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func NewError(code ErrorCode, message string) ErrorResponse {
	return ErrorResponse{Error: APIError{Code: code, Message: message}}
}

func StatusCodeForError(code ErrorCode) int {
	switch code {
	case ERROR_CODE_REVISION_CONFLICT, ERROR_CODE_IDEMPOTENCY_CONFLICT:
		return 409
	case ERROR_CODE_NOT_FOUND:
		return 404
	case ERROR_CODE_INTERNAL:
		return 500
	default:
		return 400
	}
}

func ValidateIdempotencyKey(key string) error {
	key = strings.TrimSpace(key)
	length := len(key)
	if length < MIN_IDEMPOTENCY_KEY_LENGTH || length > MAX_IDEMPOTENCY_KEY_LENGTH {
		return fmt.Errorf("idempotency key length must be between %d and %d characters", MIN_IDEMPOTENCY_KEY_LENGTH, MAX_IDEMPOTENCY_KEY_LENGTH)
	}
	return nil
}
