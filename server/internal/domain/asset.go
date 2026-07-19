package domain

import (
	"fmt"
	"strings"
)

type Asset struct {
	ID         int64        `json:"id"`
	Name       string       `json:"name"`
	Values     []AssetValue `json:"values"`
	ValueTypes []ValueType  `json:"valueTypes,omitempty"`
}

func (a Asset) Validate() error {
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("asset name must not be empty")
	}
	for _, value := range a.Values {
		if err := value.Validate(); err != nil {
			return err
		}
	}
	return nil
}
