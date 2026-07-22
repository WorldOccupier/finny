package importpreview

import (
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

type ColumnMapping struct {
	Date        int
	Description int
	Amount      int
	Debit       int
	Credit      int
	Currency    int
	Reference   int
}

type ImportRequest struct {
	Filename    string
	AccountID   string
	StatementID string
	Mapping     ColumnMapping
}

type InvalidRow struct {
	SourceRow int    `json:"sourceRow"`
	Message   string `json:"message"`
}

type Preview struct {
	Checksum     string
	PeriodStart  time.Time
	PeriodEnd    time.Time
	Transactions []domain.Transaction
	InvalidRows  []InvalidRow
	ValidRows    int
	InvalidCount int
}

type inputRow struct {
	sourceRow int
	cells     []string
}
