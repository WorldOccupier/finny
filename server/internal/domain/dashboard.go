package domain

import "fmt"

type Dashboard struct {
	Revision       int64           `json:"revision"`
	Assets         []Asset         `json:"assets"`
	CurrentFXRate  Decimal         `json:"currentFxRate"`
	CurrentTotals  DashboardTotals `json:"currentTotals"`
	History        []Snapshot      `json:"history"`
	SpendingLimits []SpendingLimit `json:"spendingLimits"`
	Income         IncomeTotals    `json:"income"`
}

func (d Dashboard) Validate() error {
	for _, asset := range d.Assets {
		if err := asset.Validate(); err != nil {
			return err
		}
	}
	if d.CurrentFXRate.IsNegative() {
		return fmt.Errorf("current FX rate must not be negative")
	}
	if err := d.CurrentTotals.Validate(); err != nil {
		return err
	}
	for _, snapshot := range d.History {
		if err := snapshot.Validate(); err != nil {
			return err
		}
	}
	for _, limit := range d.SpendingLimits {
		if err := limit.Validate(); err != nil {
			return err
		}
	}
	return d.Income.Validate()
}
