package database

import (
	"context"
	"fmt"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func (s *SQLiteStore) GetIncome(ctx context.Context) (domain.IncomeTotals, error) {
	return getIncome(ctx, s.db)
}

func getIncome(ctx context.Context, q queryer) (domain.IncomeTotals, error) {
	rows, err := q.QueryContext(ctx, `SELECT user_key, CAST(amount AS TEXT) FROM income_totals ORDER BY user_key`)
	if err != nil {
		return domain.IncomeTotals{}, fmt.Errorf("read income totals: %w", err)
	}
	defer rows.Close()
	var income domain.IncomeTotals
	found := false
	for rows.Next() {
		var userKey, amount string
		err = rows.Scan(&userKey, &amount)
		if err != nil {
			return domain.IncomeTotals{}, fmt.Errorf("scan income total: %w", err)
		}
		value, err := parseDecimal(amount, "income total")
		if err != nil {
			return domain.IncomeTotals{}, err
		}
		switch userKey {
		case "user_one":
			income.UserOneGBP = value
			found = true
		case "user_two":
			income.UserTwoGBP = value
			found = true
		}
	}
	if err = rows.Err(); err != nil {
		return domain.IncomeTotals{}, fmt.Errorf("iterate income totals: %w", err)
	}
	if !found {
		return domain.IncomeTotals{}, ErrNotFound
	}
	return income, nil
}

func (s *SQLiteStore) SaveIncome(ctx context.Context, income domain.IncomeTotals) error {
	err := income.Validate()
	if err != nil {
		return fmt.Errorf("validate income: %w", err)
	}
	return saveIncome(ctx, s.db, income)
}

func saveIncome(ctx context.Context, q queryer, income domain.IncomeTotals) error {
	_, err := q.ExecContext(ctx, `DELETE FROM income_totals`)
	if err != nil {
		return fmt.Errorf("replace income totals: %w", err)
	}
	for _, record := range []struct {
		key   string
		value domain.Decimal
	}{{"user_one", income.UserOneGBP}, {"user_two", income.UserTwoGBP}} {
		_, err = q.ExecContext(ctx, `INSERT INTO income_totals (user_key, amount, currency) VALUES (?, ?, 'GBP')`, record.key, record.value.String())
		if err != nil {
			return fmt.Errorf("write income total %s: %w", record.key, err)
		}
	}
	return nil
}
