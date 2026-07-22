package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

type Store interface {
	LoadDashboard(context.Context) (domain.Dashboard, error)
	SaveDashboard(context.Context, domain.Dashboard) error
	SaveDashboardSnapshot(context.Context, DashboardSnapshotSave) (DashboardSnapshotCommit, error)
	GetRevision(context.Context) (int64, error)
	SetRevision(context.Context, int64) error
	GetCurrentFX(context.Context) (domain.Decimal, error)
	SetCurrentFX(context.Context, domain.Decimal) error
	ListAssets(context.Context) ([]domain.Asset, error)
	SaveAsset(context.Context, domain.Asset) error
	ListSnapshots(context.Context) ([]domain.Snapshot, error)
	SaveSnapshot(context.Context, domain.Snapshot) error
	ListSpendingLimits(context.Context) ([]domain.SpendingLimit, error)
	SaveSpendingLimit(context.Context, domain.SpendingLimit) error
	GetIncome(context.Context) (domain.IncomeTotals, error)
	SaveIncome(context.Context, domain.IncomeTotals) error
	GetIdempotencyResult(context.Context, string) (IdempotencyResult, error)
	SaveIdempotencyResult(context.Context, IdempotencyResult) error
	SaveAccount(context.Context, domain.Account) error
	ListAccounts(context.Context) ([]domain.Account, error)
	SaveStatement(context.Context, domain.Statement) error
	ListStatements(context.Context) ([]domain.Statement, error)
	SaveTransactions(context.Context, []domain.Transaction) error
	SaveImport(context.Context, domain.Statement, []domain.Transaction) error
	ListTransactions(context.Context, string) ([]domain.Transaction, error)
	SummarizeTransactions(context.Context, string) ([]TransactionSummary, error)
}

var _ Store = (*SQLiteStore)(nil)

func (s *SQLiteStore) LoadDashboard(ctx context.Context) (domain.Dashboard, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("begin dashboard read: %w", err)
	}
	defer tx.Rollback()
	dashboard, err := loadDashboard(ctx, tx)
	if err != nil {
		return domain.Dashboard{}, err
	}
	err = tx.Commit()
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("commit dashboard read: %w", err)
	}
	return dashboard, nil
}

func loadDashboard(ctx context.Context, q queryer) (domain.Dashboard, error) {
	revision, err := getRevision(ctx, q)
	if errors.Is(err, ErrNotFound) {
		revision = 0
	} else if err != nil {
		return domain.Dashboard{}, err
	}
	fxRate, err := getCurrentFX(ctx, q)
	if errors.Is(err, ErrNotFound) {
		fxRate, err = domain.NewDecimal("0")
	}
	if err != nil {
		return domain.Dashboard{}, err
	}
	assets, err := listAssets(ctx, q)
	if err != nil {
		return domain.Dashboard{}, err
	}
	snapshots, err := listSnapshots(ctx, q)
	if err != nil {
		return domain.Dashboard{}, err
	}
	limits, err := listSpendingLimits(ctx, q)
	if err != nil {
		return domain.Dashboard{}, err
	}
	income, err := getIncome(ctx, q)
	if errors.Is(err, ErrNotFound) {
		income = domain.IncomeTotals{}
	} else if err != nil {
		return domain.Dashboard{}, err
	}
	totals := domain.DashboardTotals{}
	if len(snapshots) > 0 {
		totals = snapshots[len(snapshots)-1].Totals
	}
	return domain.Dashboard{Revision: revision, Assets: assets, CurrentFXRate: fxRate, CurrentTotals: totals, History: snapshots, SpendingLimits: limits, Income: income}, nil
}

func (s *SQLiteStore) SaveDashboard(ctx context.Context, dashboard domain.Dashboard) error {
	err := dashboard.Validate()
	if err != nil {
		return fmt.Errorf("validate dashboard: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin dashboard save: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `UPDATE assets SET active = 0`)
	if err != nil {
		return fmt.Errorf("deactivate assets: %w", err)
	}
	for _, asset := range dashboard.Assets {
		err = saveAsset(ctx, tx, asset)
		if err != nil {
			return err
		}
	}
	for _, snapshot := range dashboard.History {
		err = saveSnapshot(ctx, tx, snapshot)
		if err != nil {
			return err
		}
	}
	err = replaceSpendingLimits(ctx, tx, dashboard.SpendingLimits)
	if err != nil {
		return err
	}
	err = saveIncome(ctx, tx, dashboard.Income)
	if err != nil {
		return err
	}
	err = setCurrentFX(ctx, tx, dashboard.CurrentFXRate)
	if err != nil {
		return err
	}
	err = setRevision(ctx, tx, dashboard.Revision)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit dashboard save: %w", err)
	}
	return nil
}
