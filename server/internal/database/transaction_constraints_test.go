package database

import (
	"context"
	"testing"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

func TestTransactionForeignKeysAndSourceIdentityAreEnforced(t *testing.T) {
	db := openMigratedDatabase(t)
	if _, err := db.Exec(`INSERT INTO statements (id, account_id, imported_by, filename, format, checksum, period_start, period_end, status, imported_rows, invalid_rows, duplicate_rows) VALUES ('s', 'missing', 'user_one', 'x.csv', 'csv', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', '2026-07-01T00:00:00+01:00', '2026-07-01T00:00:00+01:00', 'imported', 1, 0, 0)`); err == nil {
		t.Fatal("statement with missing account was accepted")
	}
	if _, err := db.Exec(`INSERT INTO accounts (id, bank_source, account_label, currency, owner) VALUES ('a', 'Bank', 'Current', 'GBP', 'joint')`); err != nil {
		t.Fatal(err)
	}
	statement := `INSERT INTO statements (id, account_id, imported_by, filename, format, checksum, period_start, period_end, status, imported_rows, invalid_rows, duplicate_rows) VALUES ('s', 'a', 'user_one', 'x.csv', 'csv', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab', '2026-07-01T00:00:00+01:00', '2026-07-01T00:00:00+01:00', 'imported', 1, 0, 0)`
	if _, err := db.Exec(statement); err != nil {
		t.Fatal(err)
	}
	transaction := `INSERT INTO transactions (id, account_id, statement_id, transaction_date, amount, currency, description, source_row, fingerprint) VALUES ('t', 'a', 's', '2026-07-01T00:00:00+01:00', '-1.25', 'GBP', 'Coffee', 2, 'fp')`
	if _, err := db.Exec(transaction); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(transaction); err == nil {
		t.Fatal("duplicate transaction was accepted")
	}
}

func TestAccountPersistenceRetainsIndividualAndJointOwnership(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	accounts := []domain.Account{
		{ID: "one", BankSource: "Bank", AccountLabel: "One", Currency: domain.CURRENCY_GBP, Owner: domain.OWNERSHIP_USER_ONE},
		{ID: "two", BankSource: "Bank", AccountLabel: "Two", Currency: domain.CURRENCY_INR, Owner: domain.OWNERSHIP_USER_TWO},
		{ID: "joint", BankSource: "Bank", AccountLabel: "Joint", Currency: domain.CURRENCY_GBP, Owner: domain.OWNERSHIP_JOINT},
	}
	for _, account := range accounts {
		if err := store.SaveAccount(ctx, account); err != nil {
			t.Fatal(err)
		}
	}
	loaded, err := store.ListAccounts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 3 || !domain.VisibleTo(loaded[0], domain.USER_ONE) || !domain.VisibleTo(loaded[2], domain.USER_TWO) {
		t.Fatalf("accounts = %+v", loaded)
	}
}
