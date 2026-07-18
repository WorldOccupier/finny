package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func (s *SQLiteStore) GetRevision(ctx context.Context) (int64, error) { return getRevision(ctx, s.db) }

func (s *SQLiteStore) SetRevision(ctx context.Context, revision int64) error {
	return setRevision(ctx, s.db, revision)
}

func getRevision(ctx context.Context, q queryer) (int64, error) {
	var revision int64
	err := q.QueryRowContext(ctx, `SELECT revision FROM dashboard_revision WHERE id = 1`).Scan(&revision)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("read dashboard revision: %w", err)
	}
	return revision, nil
}

func setRevision(ctx context.Context, q queryer, revision int64) error {
	if revision < 0 {
		return fmt.Errorf("dashboard revision must not be negative")
	}
	_, err := q.ExecContext(ctx, `INSERT INTO dashboard_revision (id, revision) VALUES (1, ?) ON CONFLICT(id) DO UPDATE SET revision = excluded.revision`, revision)
	if err != nil {
		return fmt.Errorf("write dashboard revision: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetCurrentFX(ctx context.Context) (domain.Decimal, error) {
	return getCurrentFX(ctx, s.db)
}

func (s *SQLiteStore) SetCurrentFX(ctx context.Context, rate domain.Decimal) error {
	return setCurrentFX(ctx, s.db, rate)
}

func getCurrentFX(ctx context.Context, q queryer) (domain.Decimal, error) {
	var value string
	err := q.QueryRowContext(ctx, `SELECT CAST(fx_rate AS TEXT) FROM current_fx WHERE id = 1`).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Decimal{}, ErrNotFound
	}
	if err != nil {
		return domain.Decimal{}, fmt.Errorf("read current FX rate: %w", err)
	}
	return parseDecimal(value, "current FX rate")
}

func setCurrentFX(ctx context.Context, q queryer, rate domain.Decimal) error {
	if rate.IsNegative() {
		return fmt.Errorf("current FX rate must not be negative")
	}
	_, err := q.ExecContext(ctx, `INSERT INTO current_fx (id, fx_rate) VALUES (1, ?) ON CONFLICT(id) DO UPDATE SET fx_rate = excluded.fx_rate`, rate.String())
	if err != nil {
		return fmt.Errorf("write current FX rate: %w", err)
	}
	return nil
}
