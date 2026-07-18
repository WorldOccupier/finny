package database

import (
	"context"
	"fmt"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

type DashboardSnapshotSave struct {
	Assets         []domain.Asset
	Snapshot       domain.Snapshot
	SpendingLimits []domain.SpendingLimit
	Income         domain.IncomeTotals
	CurrentFXRate  domain.Decimal
	Revision       int64
}

func (s *SQLiteStore) SaveDashboardSnapshot(ctx context.Context, save DashboardSnapshotSave) error {
	if err := save.Snapshot.Validate(); err != nil {
		return fmt.Errorf("validate snapshot save: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin snapshot save: %w", err)
	}
	defer tx.Rollback()
	if err := deactivateAssets(ctx, tx); err != nil {
		return err
	}
	if err := saveCurrentAssets(ctx, tx, save.Assets); err != nil {
		return err
	}
	if err := saveSnapshot(ctx, tx, save.Snapshot); err != nil {
		return err
	}
	if err := saveMutableDashboardState(ctx, tx, save); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit snapshot save: %w", err)
	}
	return nil
}

func deactivateAssets(ctx context.Context, tx queryer) error {
	if _, err := tx.ExecContext(ctx, `UPDATE assets SET active = 0`); err != nil {
		return fmt.Errorf("deactivate assets: %w", err)
	}
	return nil
}

func saveCurrentAssets(ctx context.Context, tx queryer, assets []domain.Asset) error {
	for _, asset := range assets {
		if err := saveAsset(ctx, tx, asset); err != nil {
			return fmt.Errorf("save asset %d: %w", asset.ID, err)
		}
	}
	return nil
}

func saveMutableDashboardState(ctx context.Context, tx queryer, save DashboardSnapshotSave) error {
	if err := replaceSpendingLimits(ctx, tx, save.SpendingLimits); err != nil {
		return fmt.Errorf("save spending limits: %w", err)
	}
	if err := saveIncome(ctx, tx, save.Income); err != nil {
		return fmt.Errorf("save income: %w", err)
	}
	if err := setCurrentFX(ctx, tx, save.CurrentFXRate); err != nil {
		return fmt.Errorf("save current FX: %w", err)
	}
	if err := setRevision(ctx, tx, save.Revision); err != nil {
		return fmt.Errorf("save revision: %w", err)
	}
	return nil
}
