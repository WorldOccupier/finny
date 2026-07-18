package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func (s *SQLiteStore) ListAssets(ctx context.Context) ([]domain.Asset, error) {
	return listAssets(ctx, s.db)
}

func listAssets(ctx context.Context, q queryer) ([]domain.Asset, error) {
	rows, err := q.QueryContext(ctx, `SELECT id, name FROM assets WHERE active = 1 ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()
	assets := make([]domain.Asset, 0)
	for rows.Next() {
		var asset domain.Asset
		err = rows.Scan(&asset.ID, &asset.Name)
		if err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		assets = append(assets, asset)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assets: %w", err)
	}
	if len(assets) == 0 {
		return assets, nil
	}
	latestID, err := latestSnapshotID(ctx, q)
	if errors.Is(err, ErrNotFound) {
		return assets, nil
	}
	if err != nil {
		return nil, err
	}
	values, err := assetValuesForSnapshot(ctx, q, latestID)
	if err != nil {
		return nil, err
	}
	for index := range assets {
		assets[index].Values = values[assets[index].ID]
	}
	return assets, nil
}

func (s *SQLiteStore) SaveAsset(ctx context.Context, asset domain.Asset) error {
	err := asset.Validate()
	if err != nil {
		return fmt.Errorf("validate asset: %w", err)
	}
	return saveAsset(ctx, s.db, asset)
}

func saveAsset(ctx context.Context, q queryer, asset domain.Asset) error {
	_, err := q.ExecContext(ctx, `INSERT INTO assets (id, name, active) VALUES (?, ?, 1) ON CONFLICT(id) DO UPDATE SET name = excluded.name, active = 1`, asset.ID, asset.Name)
	if err != nil {
		return fmt.Errorf("write asset %d: %w", asset.ID, err)
	}
	return nil
}

func latestSnapshotID(ctx context.Context, q queryer) (int64, error) {
	var id int64
	err := q.QueryRowContext(ctx, `SELECT id FROM snapshots ORDER BY committed_at DESC LIMIT 1`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("read latest snapshot: %w", err)
	}
	return id, nil
}

func assetValuesForSnapshot(ctx context.Context, q queryer, snapshotID int64) (map[int64][]domain.AssetValue, error) {
	rows, err := q.QueryContext(ctx, `SELECT asset_id, value_type, CAST(value AS TEXT) FROM snapshot_asset_values WHERE snapshot_id = ? ORDER BY asset_id, value_type`, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("read snapshot asset values: %w", err)
	}
	defer rows.Close()
	values := make(map[int64][]domain.AssetValue)
	for rows.Next() {
		var assetID int64
		var valueType, rawValue string
		err = rows.Scan(&assetID, &valueType, &rawValue)
		if err != nil {
			return nil, fmt.Errorf("scan snapshot asset value: %w", err)
		}
		value, err := parseDecimal(rawValue, "asset value")
		if err != nil {
			return nil, err
		}
		values[assetID] = append(values[assetID], domain.AssetValue{AssetID: assetID, Type: domain.ValueType(valueType), Value: value})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate snapshot asset values: %w", err)
	}
	return values, nil
}
