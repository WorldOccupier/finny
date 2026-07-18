package domain

import (
	"fmt"
	"time"
)

type Snapshot struct {
	ID          int64           `json:"id"`
	CommittedAt time.Time       `json:"committedAt"`
	FXRate      Decimal         `json:"fxRate"`
	AssetValues []AssetValue    `json:"assetValues"`
	Totals      DashboardTotals `json:"totals"`
}

func (s Snapshot) Validate() error {
	if s.CommittedAt.IsZero() {
		return fmt.Errorf("snapshot committed time must be set")
	}
	if s.FXRate.IsNegative() {
		return fmt.Errorf("snapshot FX rate must not be negative")
	}
	for _, value := range s.AssetValues {
		if err := value.Validate(); err != nil {
			return err
		}
	}
	return s.Totals.Validate()
}
