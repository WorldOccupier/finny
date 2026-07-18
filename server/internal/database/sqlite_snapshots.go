package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func (s *SQLiteStore) ListSnapshots(ctx context.Context) ([]domain.Snapshot, error) {
	return listSnapshots(ctx, s.db)
}

func listSnapshots(ctx context.Context, q queryer) ([]domain.Snapshot, error) {
	rows, err := q.QueryContext(ctx, `SELECT id, committed_at, CAST(fx_rate AS TEXT) FROM snapshots ORDER BY committed_at, id`)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	defer rows.Close()
	snapshots := make([]domain.Snapshot, 0)
	for rows.Next() {
		var snapshot domain.Snapshot
		var committedAt, fxRate string
		err = rows.Scan(&snapshot.ID, &committedAt, &fxRate)
		if err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		snapshot.CommittedAt, err = domain.ParseLondonTimestamp(committedAt)
		if err != nil {
			return nil, fmt.Errorf("parse snapshot %d timestamp: %w", snapshot.ID, err)
		}
		snapshot.FXRate, err = parseDecimal(fxRate, "snapshot FX rate")
		if err != nil {
			return nil, err
		}
		values, err := assetValuesForSnapshot(ctx, q, snapshot.ID)
		if err != nil {
			return nil, err
		}
		for _, assetValues := range values {
			snapshot.AssetValues = append(snapshot.AssetValues, assetValues...)
		}
		sort.Slice(snapshot.AssetValues, func(left, right int) bool {
			if snapshot.AssetValues[left].AssetID == snapshot.AssetValues[right].AssetID {
				return snapshot.AssetValues[left].Type < snapshot.AssetValues[right].Type
			}
			return snapshot.AssetValues[left].AssetID < snapshot.AssetValues[right].AssetID
		})
		snapshot.Totals, err = totalsForSnapshot(ctx, q, snapshot.ID)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate snapshots: %w", err)
	}
	return snapshots, nil
}

func (s *SQLiteStore) SaveSnapshot(ctx context.Context, snapshot domain.Snapshot) error {
	err := snapshot.Validate()
	if err != nil {
		return fmt.Errorf("validate snapshot: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin snapshot save: %w", err)
	}
	defer tx.Rollback()
	err = saveSnapshot(ctx, tx, snapshot)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit snapshot save: %w", err)
	}
	return nil
}

func saveSnapshot(ctx context.Context, q queryer, snapshot domain.Snapshot) error {
	_, err := q.ExecContext(ctx, `INSERT INTO snapshots (id, committed_at, fx_rate) VALUES (?, ?, ?)`, snapshot.ID, domain.FormatLondonTimestamp(snapshot.CommittedAt), snapshot.FXRate.String())
	if err != nil {
		return fmt.Errorf("write snapshot %d: %w", snapshot.ID, err)
	}
	for _, value := range snapshot.AssetValues {
		_, err = q.ExecContext(ctx, `INSERT INTO snapshot_asset_values (snapshot_id, asset_id, value_type, value) VALUES (?, ?, ?, ?)`, snapshot.ID, value.AssetID, value.Type, value.Value.String())
		if err != nil {
			return fmt.Errorf("write snapshot %d asset value: %w", snapshot.ID, err)
		}
	}
	return saveTotals(ctx, q, snapshot.ID, snapshot.Totals)
}

func saveTotals(ctx context.Context, q queryer, snapshotID int64, totals domain.DashboardTotals) error {
	for _, total := range totals.Country {
		_, err := q.ExecContext(ctx, `INSERT INTO snapshot_totals (snapshot_id, scope, value_type, value) VALUES (?, 'country', ?, ?)`, snapshotID, total.Type, total.Value.String())
		if err != nil {
			return fmt.Errorf("write country total for snapshot %d: %w", snapshotID, err)
		}
	}
	for _, total := range totals.Combined {
		_, err := q.ExecContext(ctx, `INSERT INTO snapshot_totals (snapshot_id, scope, currency, value) VALUES (?, 'combined', ?, ?)`, snapshotID, total.Currency, total.Value.String())
		if err != nil {
			return fmt.Errorf("write combined total for snapshot %d: %w", snapshotID, err)
		}
	}
	return nil
}

func totalsForSnapshot(ctx context.Context, q queryer, snapshotID int64) (domain.DashboardTotals, error) {
	rows, err := q.QueryContext(ctx, `SELECT scope, value_type, currency, CAST(value AS TEXT) FROM snapshot_totals WHERE snapshot_id = ? ORDER BY id`, snapshotID)
	if err != nil {
		return domain.DashboardTotals{}, fmt.Errorf("read snapshot totals: %w", err)
	}
	defer rows.Close()
	var totals domain.DashboardTotals
	for rows.Next() {
		var scope, rawValue string
		var valueType, currency sql.NullString
		err = rows.Scan(&scope, &valueType, &currency, &rawValue)
		if err != nil {
			return domain.DashboardTotals{}, fmt.Errorf("scan snapshot total: %w", err)
		}
		value, err := parseDecimal(rawValue, "snapshot total")
		if err != nil {
			return domain.DashboardTotals{}, err
		}
		if scope == "country" {
			totals.Country = append(totals.Country, domain.TotalValue{Type: domain.ValueType(valueType.String), Value: value})
		} else {
			totals.Combined = append(totals.Combined, domain.CombinedTotal{Currency: domain.Currency(currency.String), Value: value})
		}
	}
	if err = rows.Err(); err != nil {
		return domain.DashboardTotals{}, fmt.Errorf("iterate snapshot totals: %w", err)
	}
	return totals, nil
}
