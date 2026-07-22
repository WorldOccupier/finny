package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID          string          `json:"id"`
	AccountID   string          `json:"accountId"`
	StatementID string          `json:"statementId"`
	Date        time.Time       `json:"date"`
	Amount      decimal.Decimal `json:"amount"`
	Currency    Currency        `json:"currency"`
	Description string          `json:"description"`
	Reference   string          `json:"reference,omitempty"`
	SourceRow   int             `json:"sourceRow"`
	Fingerprint string          `json:"fingerprint"`
}

func (t Transaction) Validate() error {
	if strings.TrimSpace(t.ID) == "" || strings.TrimSpace(t.AccountID) == "" || strings.TrimSpace(t.StatementID) == "" || t.SourceRow < 1 {
		return fmt.Errorf("transaction identifiers and source row must be valid")
	}
	if !isLondonTime(t.Date) {
		return fmt.Errorf("transaction date must be valid")
	}
	if !t.Currency.Valid() || strings.TrimSpace(t.Description) == "" {
		return fmt.Errorf("transaction currency and description must be valid")
	}
	return nil
}

func TransactionFingerprint(t Transaction) string {
	value, _ := json.Marshal([]string{t.AccountID, t.Date.Format("2006-01-02"), t.Amount.String(), string(t.Currency), normalize(t.Description), normalize(t.Reference)})
	hash := sha256.Sum256(value)
	return hex.EncodeToString(hash[:])
}

func TransactionSourceIdentity(t Transaction) string {
	return fmt.Sprintf("%s:%d", t.StatementID, t.SourceRow)
}

func normalize(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}
