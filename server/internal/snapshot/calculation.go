package snapshot

import (
	"fmt"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

const CALCULATION_SCALE int32 = 18

func CalculateTotals(assets []domain.Asset, fxRate domain.Decimal) (domain.DashboardTotals, error) {
	if fxRate.IsZero() || fxRate.IsNegative() {
		return domain.DashboardTotals{}, fmt.Errorf("snapshot FX rate must be greater than zero")
	}
	zero, err := domain.NewDecimal("0")
	if err != nil {
		return domain.DashboardTotals{}, err
	}
	ukTotal := zero.Decimal
	indiaTotal := zero.Decimal
	for _, asset := range assets {
		for _, value := range asset.Values {
			switch value.Type {
			case domain.UK_GBP:
				ukTotal = ukTotal.Add(value.Value.Decimal)
			case domain.INDIA_INR:
				indiaTotal = indiaTotal.Add(value.Value.Decimal)
			default:
				return domain.DashboardTotals{}, fmt.Errorf("invalid value type %q", value.Type)
			}
		}
	}
	combinedGBP := ukTotal.Add(indiaTotal.DivRound(fxRate.Decimal, CALCULATION_SCALE))
	combinedINR := ukTotal.Mul(fxRate.Decimal).Add(indiaTotal)
	return domain.DashboardTotals{
		Country: []domain.TotalValue{
			{Type: domain.UK_GBP, Value: domain.Decimal{Decimal: ukTotal}},
			{Type: domain.INDIA_INR, Value: domain.Decimal{Decimal: indiaTotal}},
		},
		Combined: []domain.CombinedTotal{
			{Currency: domain.CURRENCY_GBP, Value: domain.Decimal{Decimal: combinedGBP}},
			{Currency: domain.CURRENCY_INR, Value: domain.Decimal{Decimal: combinedINR}},
		},
	}, nil
}
