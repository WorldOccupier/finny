package domain

import "fmt"

type AssetValue struct {
	Type  ValueType `json:"type"`
	Value Decimal   `json:"value"`
}

func (v AssetValue) Validate() error {
	if !v.Type.Valid() {
		return fmt.Errorf("invalid value type %q", v.Type)
	}
	if v.Value.IsNegative() {
		return fmt.Errorf("asset value must not be negative")
	}
	return nil
}
