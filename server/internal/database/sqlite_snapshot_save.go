package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

type DashboardSnapshotSave struct {
	Assets           []domain.Asset
	Snapshot         domain.Snapshot
	SpendingLimits   []domain.SpendingLimit
	Income           domain.IncomeTotals
	CurrentFXRate    domain.Decimal
	ExpectedRevision int64
	Revision         int64
	Idempotency      IdempotencyResult
}

func (s *SQLiteStore) SaveDashboardSnapshot(ctx context.Context, save DashboardSnapshotSave) (DashboardSnapshotCommit, error) {
	if err := save.Snapshot.Validate(); err != nil {
		return DashboardSnapshotCommit{}, fmt.Errorf("validate snapshot save: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DashboardSnapshotCommit{}, fmt.Errorf("begin snapshot save: %w", err)
	}
	defer tx.Rollback()
	existing, err := claimIdempotency(ctx, tx, save.Idempotency)
	if err != nil {
		return DashboardSnapshotCommit{}, err
	}
	if existing != nil {
		return DashboardSnapshotCommit{Replayed: true, Idempotency: *existing}, nil
	}
	updated, err := updateRevisionIfExpected(ctx, tx, save.ExpectedRevision, save.Revision)
	if err != nil {
		return DashboardSnapshotCommit{}, err
	}
	if !updated {
		return DashboardSnapshotCommit{}, ErrRevisionConflict
	}
	if err := deactivateAssets(ctx, tx); err != nil {
		return DashboardSnapshotCommit{}, err
	}
	if err := saveCurrentAssets(ctx, tx, save.Assets); err != nil {
		return DashboardSnapshotCommit{}, err
	}
	if err := saveSnapshot(ctx, tx, save.Snapshot); err != nil {
		return DashboardSnapshotCommit{}, err
	}
	if err := saveMutableDashboardState(ctx, tx, save); err != nil {
		return DashboardSnapshotCommit{}, err
	}
	if err := tx.Commit(); err != nil {
		return DashboardSnapshotCommit{}, fmt.Errorf("commit snapshot save: %w", err)
	}
	return DashboardSnapshotCommit{Idempotency: save.Idempotency}, nil
}

func claimIdempotency(ctx context.Context, tx queryer, result IdempotencyResult) (*IdempotencyResult, error) {
	if result.Key == "" {
		return nil, nil
	}
	if result.RequestHash == "" || result.ResponseJSON == "" || result.CreatedAt.IsZero() {
		return nil, fmt.Errorf("idempotency result is incomplete")
	}
	var existing IdempotencyResult
	var createdAt string
	err := tx.QueryRowContext(ctx, `SELECT idempotency_key, request_hash, response_json, created_at FROM idempotency_keys WHERE idempotency_key = ?`, result.Key).Scan(&existing.Key, &existing.RequestHash, &existing.ResponseJSON, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx, `INSERT INTO idempotency_keys (idempotency_key, request_hash, response_json, created_at) VALUES (?, ?, ?, ?)`, result.Key, result.RequestHash, result.ResponseJSON, result.CreatedAt.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return nil, fmt.Errorf("claim idempotency key: %w", err)
		}
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read idempotency key: %w", err)
	}
	if existing.RequestHash != result.RequestHash {
		return nil, ErrIdempotencyConflict
	}
	existing.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse idempotency timestamp: %w", err)
	}
	return &existing, nil
}

func updateRevisionIfExpected(ctx context.Context, tx queryer, expected, revision int64) (bool, error) {
	if expected < 0 || revision < 0 || revision != expected+1 {
		return false, fmt.Errorf("invalid dashboard revision transition")
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO dashboard_revision (id, revision) VALUES (1, ?) ON CONFLICT(id) DO UPDATE SET revision = excluded.revision WHERE dashboard_revision.revision = ?`, revision, expected)
	if err != nil {
		return false, fmt.Errorf("update dashboard revision: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check dashboard revision update: %w", err)
	}
	return rows == 1, nil
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
