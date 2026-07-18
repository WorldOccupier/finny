package domain

import "fmt"

type TotalValue struct {
	Type  ValueType `json:"type"`
	Value Decimal   `json:"value"`
}

type Totals struct {
	Values []TotalValue `json:"values"`
}

type CombinedTotal struct {
	Currency Currency `json:"currency"`
	Value    Decimal  `json:"value"`
}

type DashboardTotals struct {
	Country  []TotalValue    `json:"country"`
	Combined []CombinedTotal `json:"combined"`
}

func (t TotalValue) Validate() error {
	if !t.Type.Valid() {
		return fmt.Errorf("invalid total value type %q", t.Type)
	}
	if t.Value.IsNegative() {
		return fmt.Errorf("total value must not be negative")
	}
	return nil
}

func (t CombinedTotal) Validate() error {
	if !t.Currency.Valid() {
		return fmt.Errorf("invalid combined total currency %q", t.Currency)
	}
	if t.Value.IsNegative() {
		return fmt.Errorf("combined total must not be negative")
	}
	return nil
}

func (t DashboardTotals) Validate() error {
	for _, value := range t.Country {
		if err := value.Validate(); err != nil {
			return err
		}
	}
	for _, value := range t.Combined {
		if err := value.Validate(); err != nil {
			return err
		}
	}
	return nil
}
