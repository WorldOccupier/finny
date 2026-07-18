package database

import (
	"context"
	"fmt"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func (s *SQLiteStore) ListSpendingLimits(ctx context.Context) ([]domain.SpendingLimit, error) {
	return listSpendingLimits(ctx, s.db)
}

func listSpendingLimits(ctx context.Context, q queryer) ([]domain.SpendingLimit, error) {
	rows, err := q.QueryContext(ctx, `SELECT limit_key, CAST(amount AS TEXT), currency FROM spending_limits ORDER BY limit_key`)
	if err != nil {
		return nil, fmt.Errorf("list spending limits: %w", err)
	}
	defer rows.Close()
	limits := make([]domain.SpendingLimit, 0)
	for rows.Next() {
		var limit domain.SpendingLimit
		var amount, currency string
		err = rows.Scan(&limit.Key, &amount, &currency)
		if err != nil {
			return nil, fmt.Errorf("scan spending limit: %w", err)
		}
		limit.Amount, err = parseDecimal(amount, "spending limit")
		if err != nil {
			return nil, err
		}
		limit.Currency = domain.Currency(currency)
		limits = append(limits, limit)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate spending limits: %w", err)
	}
	return limits, nil
}

func (s *SQLiteStore) SaveSpendingLimit(ctx context.Context, limit domain.SpendingLimit) error {
	err := limit.Validate()
	if err != nil {
		return fmt.Errorf("validate spending limit: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO spending_limits (limit_key, amount, currency) VALUES (?, ?, ?) ON CONFLICT(limit_key) DO UPDATE SET amount = excluded.amount, currency = excluded.currency`, limit.Key, limit.Amount.String(), limit.Currency)
	if err != nil {
		return fmt.Errorf("write spending limit: %w", err)
	}
	return nil
}

func replaceSpendingLimits(ctx context.Context, q queryer, limits []domain.SpendingLimit) error {
	_, err := q.ExecContext(ctx, `DELETE FROM spending_limits`)
	if err != nil {
		return fmt.Errorf("replace spending limits: %w", err)
	}
	for _, limit := range limits {
		err = limit.Validate()
		if err != nil {
			return fmt.Errorf("validate spending limit: %w", err)
		}
		_, err = q.ExecContext(ctx, `INSERT INTO spending_limits (limit_key, amount, currency) VALUES (?, ?, ?)`, limit.Key, limit.Amount.String(), limit.Currency)
		if err != nil {
			return fmt.Errorf("write spending limit: %w", err)
		}
	}
	return nil
}
