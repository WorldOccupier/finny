package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type StatementFormat string

const (
	STATEMENT_FORMAT_CSV  StatementFormat = "csv"
	STATEMENT_FORMAT_XLSX StatementFormat = "xlsx"
)

type StatementStatus string

const (
	STATEMENT_STATUS_IMPORTED StatementStatus = "imported"
)

type Statement struct {
	ID            string          `json:"id"`
	AccountID     string          `json:"accountId"`
	ImportedBy    UserID          `json:"importedBy"`
	Filename      string          `json:"filename"`
	Format        StatementFormat `json:"format"`
	Checksum      string          `json:"checksum"`
	PeriodStart   time.Time       `json:"periodStart"`
	PeriodEnd     time.Time       `json:"periodEnd"`
	Status        StatementStatus `json:"status"`
	ImportedRows  int             `json:"importedRows"`
	InvalidRows   int             `json:"invalidRows"`
	DuplicateRows int             `json:"duplicateRows"`
}

func (s Statement) Validate() error {
	if strings.TrimSpace(s.ID) == "" || strings.TrimSpace(s.AccountID) == "" || strings.TrimSpace(s.Filename) == "" || strings.TrimSpace(s.Checksum) == "" {
		return fmt.Errorf("statement identifiers, filename, and checksum must not be empty")
	}
	if err := (User{ID: s.ImportedBy}).Validate(); err != nil {
		return fmt.Errorf("invalid importing user: %w", err)
	}
	if s.Format != STATEMENT_FORMAT_CSV && s.Format != STATEMENT_FORMAT_XLSX {
		return fmt.Errorf("invalid statement format %q", s.Format)
	}
	if s.Status != STATEMENT_STATUS_IMPORTED {
		return fmt.Errorf("invalid statement status %q", s.Status)
	}
	if _, err := hex.DecodeString(s.Checksum); err != nil || len(s.Checksum) != sha256.Size*2 {
		return fmt.Errorf("checksum must be a SHA-256 hex string")
	}
	if !isLondonTime(s.PeriodStart) || !isLondonTime(s.PeriodEnd) || s.PeriodEnd.Before(s.PeriodStart) {
		return fmt.Errorf("statement period is invalid")
	}
	if s.ImportedRows < 0 || s.InvalidRows < 0 || s.DuplicateRows < 0 {
		return fmt.Errorf("statement row counts must not be negative")
	}
	return nil
}
