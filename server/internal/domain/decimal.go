package domain

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// Decimal is the domain representation for non-negative financial values.
// It embeds decimal.Decimal so later calculation packages can use its exact arithmetic.
type Decimal struct {
	decimal.Decimal
}

func NewDecimal(value string) (Decimal, error) {
	if value == "" {
		return Decimal{}, fmt.Errorf("decimal value must not be empty")
	}

	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return Decimal{}, fmt.Errorf("invalid decimal value: %w", err)
	}
	if parsed.IsNegative() {
		return Decimal{}, fmt.Errorf("decimal value must not be negative")
	}

	return Decimal{Decimal: parsed}, nil
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != '"' {
		return fmt.Errorf("decimal must be a JSON string")
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("invalid decimal JSON string: %w", err)
	}

	parsed, err := NewDecimal(value)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
