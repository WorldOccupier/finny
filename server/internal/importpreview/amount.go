package importpreview

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

func parseAmount(amountValue, debitValue, creditValue string) (decimal.Decimal, error) {
	if amountValue != "" && (debitValue != "" || creditValue != "") {
		return decimal.Decimal{}, fmt.Errorf("signed amount cannot be combined with debit or credit")
	}
	if debitValue != "" && creditValue != "" {
		return decimal.Decimal{}, fmt.Errorf("debit and credit cannot both be populated")
	}
	value := amountValue
	sign := decimal.NewFromInt(1)
	if debitValue != "" {
		value, sign = debitValue, decimal.NewFromInt(-1)
	}
	if creditValue != "" {
		value = creditValue
	}
	if strings.TrimSpace(value) == "" {
		return decimal.Decimal{}, fmt.Errorf("a usable amount is required")
	}
	parsed, err := decimal.NewFromString(strings.ReplaceAll(strings.TrimSpace(value), ",", ""))
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("invalid amount %q", value)
	}
	return parsed.Mul(sign), nil
}
