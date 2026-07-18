package domain

import (
	"fmt"
	"time"
)

type Snapshot struct {
	ID          int64           `json:"id"`
	CommittedAt time.Time       `json:"committedAt"`
	FXRate      Decimal         `json:"fxRate"`
	Assets      []Asset         `json:"assets"`
	Totals      DashboardTotals `json:"totals"`
}

func (s Snapshot) Validate() error {
	if s.CommittedAt.IsZero() {
		return fmt.Errorf("snapshot committed time must be set")
	}
	if s.FXRate.IsNegative() {
		return fmt.Errorf("snapshot FX rate must not be negative")
	}
	for _, asset := range s.Assets {
		if err := asset.Validate(); err != nil {
			return err
		}
	}
	return s.Totals.Validate()
}
