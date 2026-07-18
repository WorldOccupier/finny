package domain

import "fmt"

type IncomeTotals struct {
	UserOneGBP Decimal `json:"userOneGBP"`
	UserTwoGBP Decimal `json:"userTwoGBP"`
}

func (i IncomeTotals) Validate() error {
	if i.UserOneGBP.IsNegative() || i.UserTwoGBP.IsNegative() {
		return fmt.Errorf("income totals must not be negative")
	}
	return nil
}
