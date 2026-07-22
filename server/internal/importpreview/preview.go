package importpreview

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

const (
	FILE_EXTENSION_CSV  = ".csv"
	FILE_EXTENSION_XLSX = ".xlsx"
	DEFAULT_CURRENCY    = domain.CURRENCY_GBP
)

func PreviewImport(data []byte, request ImportRequest) (Preview, error) {
	if err := validateRequest(request); err != nil {
		return Preview{}, err
	}
	checksum := sha256.Sum256(data)
	rows, err := readRows(data, request.Filename)
	if err != nil {
		return Preview{}, err
	}
	preview := Preview{Checksum: hex.EncodeToString(checksum[:])}
	for _, row := range rows {
		transaction, err := parseTransaction(row.cells, row.sourceRow, request)
		if err != nil {
			preview.InvalidRows = append(preview.InvalidRows, InvalidRow{SourceRow: row.sourceRow, Message: err.Error()})
			continue
		}
		preview.Transactions = append(preview.Transactions, transaction)
	}
	preview.ValidRows = len(preview.Transactions)
	preview.InvalidCount = len(preview.InvalidRows)
	if preview.ValidRows > 0 {
		preview.PeriodStart = preview.Transactions[0].Date
		preview.PeriodEnd = preview.Transactions[0].Date
		for _, transaction := range preview.Transactions[1:] {
			if transaction.Date.Before(preview.PeriodStart) {
				preview.PeriodStart = transaction.Date
			}
			if transaction.Date.After(preview.PeriodEnd) {
				preview.PeriodEnd = transaction.Date
			}
		}
	}
	return preview, nil
}

func validateRequest(request ImportRequest) error {
	if strings.TrimSpace(request.AccountID) == "" || strings.TrimSpace(request.StatementID) == "" {
		return fmt.Errorf("account and statement IDs are required")
	}
	if request.Mapping.Date < 0 || request.Mapping.Description < 0 {
		return fmt.Errorf("date and description mappings are required")
	}
	if request.Mapping.Amount >= 0 {
		if request.Mapping.Debit >= 0 || request.Mapping.Credit >= 0 {
			return fmt.Errorf("signed amount cannot be combined with debit or credit mappings")
		}
	} else if request.Mapping.Debit < 0 || request.Mapping.Credit < 0 {
		return fmt.Errorf("both debit and credit mappings are required")
	}
	ext := strings.ToLower(filepath.Ext(request.Filename))
	if ext != FILE_EXTENSION_CSV && ext != FILE_EXTENSION_XLSX {
		return fmt.Errorf("unsupported statement format %q", ext)
	}
	return nil
}

func parseTransaction(row []string, sourceRow int, request ImportRequest) (domain.Transaction, error) {
	value := func(index int) string {
		if index < 0 || index >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[index])
	}
	date, err := parseDate(value(request.Mapping.Date))
	if err != nil {
		return domain.Transaction{}, err
	}
	description := value(request.Mapping.Description)
	if description == "" {
		return domain.Transaction{}, fmt.Errorf("description is required")
	}
	amount, err := parseAmount(value(request.Mapping.Amount), value(request.Mapping.Debit), value(request.Mapping.Credit))
	if err != nil {
		return domain.Transaction{}, err
	}
	currency := domain.Currency(value(request.Mapping.Currency))
	if currency == "" {
		currency = DEFAULT_CURRENCY
	}
	if !currency.Valid() {
		return domain.Transaction{}, fmt.Errorf("unsupported currency %q", currency)
	}
	transaction := domain.Transaction{ID: fmt.Sprintf("%s:%d", request.StatementID, sourceRow), AccountID: request.AccountID, StatementID: request.StatementID, Date: date, Amount: amount, Currency: currency, Description: description, Reference: value(request.Mapping.Reference), SourceRow: sourceRow}
	transaction.Fingerprint = domain.TransactionFingerprint(transaction)
	if err := transaction.Validate(); err != nil {
		return domain.Transaction{}, err
	}
	return transaction, nil
}

func parseDate(value string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02", "02/01/2006", "01/02/2006", time.RFC3339} {
		parsed, err := time.ParseInLocation(layout, value, domain.LondonLocation())
		if err == nil {
			return parsed.In(domain.LondonLocation()), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q", value)
}
