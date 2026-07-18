package domain

import (
	"fmt"
	"strings"
)

type SpendingLimit struct {
	Key      string   `json:"key"`
	Amount   Decimal  `json:"amount"`
	Currency Currency `json:"currency"`
}

func (s SpendingLimit) Validate() error {
	if strings.TrimSpace(s.Key) == "" {
		return fmt.Errorf("spending limit key must not be empty")
	}
	if !s.Currency.Valid() {
		return fmt.Errorf("invalid spending limit currency %q", s.Currency)
	}
	if s.Amount.IsNegative() {
		return fmt.Errorf("spending limit amount must not be negative")
	}
	return nil
}
