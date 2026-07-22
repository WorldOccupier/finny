package database

import (
	"context"
	"fmt"
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
	"github.com/shopspring/decimal"
)

type TransactionSummary struct {
	Currency domain.Currency
	Amount   domain.Decimal
}

func (s *SQLiteStore) SaveAccount(ctx context.Context, account domain.Account) error {
	if err := account.Validate(); err != nil {
		return fmt.Errorf("validate account: %w", err)
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO accounts (id, bank_source, account_label, currency, owner) VALUES (?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET bank_source=excluded.bank_source, account_label=excluded.account_label, currency=excluded.currency, owner=excluded.owner`, account.ID, account.BankSource, account.AccountLabel, account.Currency, account.Owner)
	return err
}

func (s *SQLiteStore) ListAccounts(ctx context.Context) ([]domain.Account, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, bank_source, account_label, currency, owner FROM accounts ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Account
	for rows.Next() {
		var a domain.Account
		if err := rows.Scan(&a.ID, &a.BankSource, &a.AccountLabel, &a.Currency, &a.Owner); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (s *SQLiteStore) SaveStatement(ctx context.Context, statement domain.Statement) error {
	if err := statement.Validate(); err != nil {
		return fmt.Errorf("validate statement: %w", err)
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO statements (id, account_id, imported_by, filename, format, checksum, period_start, period_end, status, imported_rows, invalid_rows, duplicate_rows) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET status=excluded.status, imported_rows=excluded.imported_rows, invalid_rows=excluded.invalid_rows, duplicate_rows=excluded.duplicate_rows`, statement.ID, statement.AccountID, statement.ImportedBy, statement.Filename, statement.Format, statement.Checksum, statement.PeriodStart.Format(timeFormat), statement.PeriodEnd.Format(timeFormat), statement.Status, statement.ImportedRows, statement.InvalidRows, statement.DuplicateRows)
	return err
}

func (s *SQLiteStore) SaveTransactions(ctx context.Context, transactions []domain.Transaction) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, t := range transactions {
		if err := t.Validate(); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO transactions (id, account_id, statement_id, transaction_date, amount, currency, description, reference, source_row, fingerprint) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET amount=excluded.amount, description=excluded.description, reference=excluded.reference, fingerprint=excluded.fingerprint`, t.ID, t.AccountID, t.StatementID, t.Date.Format(timeFormat), t.Amount.String(), t.Currency, t.Description, t.Reference, t.SourceRow, t.Fingerprint)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) ListTransactions(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, account_id, statement_id, transaction_date, amount, currency, description, reference, source_row, fingerprint FROM transactions WHERE account_id = ? ORDER BY transaction_date, source_row`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		var date, amount string
		if err := rows.Scan(&t.ID, &t.AccountID, &t.StatementID, &date, &amount, &t.Currency, &t.Description, &t.Reference, &t.SourceRow, &t.Fingerprint); err != nil {
			return nil, err
		}
		parsed, err := decimal.NewFromString(amount)
		if err != nil {
			return nil, fmt.Errorf("parse transaction amount: %w", err)
		}
		t.Amount = parsed
		t.Date, err = domain.ParseLondonTimestamp(date)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (s *SQLiteStore) SummarizeTransactions(ctx context.Context, accountID string) ([]TransactionSummary, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT currency, COALESCE(SUM(amount), 0) FROM transactions WHERE account_id = ? GROUP BY currency ORDER BY currency`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []TransactionSummary
	for rows.Next() {
		var item TransactionSummary
		var amount string
		if err := rows.Scan(&item.Currency, &amount); err != nil {
			return nil, err
		}
		item.Amount, err = domain.NewDecimal("0")
		if err != nil {
			return nil, err
		}
		item.Amount.Decimal, err = decimal.NewFromString(amount)
		if err != nil {
			return nil, fmt.Errorf("parse transaction summary: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

const timeFormat = time.RFC3339Nano
