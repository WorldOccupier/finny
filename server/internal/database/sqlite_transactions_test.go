package database

import (
	"context"
	"testing"
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
	"github.com/shopspring/decimal"
)

func TestTransactionPersistenceAndSummary(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	if err := store.SaveAccount(ctx, domain.Account{ID: "checking", BankSource: "Bank", AccountLabel: "Current", Currency: domain.CURRENCY_GBP, Owner: domain.OWNERSHIP_JOINT}); err != nil {
		t.Fatal(err)
	}
	statement := domain.Statement{ID: "statement", AccountID: "checking", ImportedBy: domain.USER_ONE, Filename: "statement.csv", Format: domain.STATEMENT_FORMAT_CSV, Checksum: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", PeriodStart: time.Date(2026, 7, 1, 0, 0, 0, 0, domain.LondonLocation()), PeriodEnd: time.Date(2026, 7, 31, 0, 0, 0, 0, domain.LondonLocation()), Status: domain.STATEMENT_STATUS_IMPORTED, ImportedRows: 2}
	if err := store.SaveStatement(ctx, statement); err != nil {
		t.Fatal(err)
	}
	transactions := []domain.Transaction{
		{ID: "statement:2", AccountID: "checking", StatementID: "statement", Date: time.Date(2026, 7, 2, 0, 0, 0, 0, domain.LondonLocation()), Amount: decimal.RequireFromString("-12.50"), Currency: domain.CURRENCY_GBP, Description: "Coffee", SourceRow: 2, Fingerprint: "fp1"},
		{ID: "statement:3", AccountID: "checking", StatementID: "statement", Date: time.Date(2026, 7, 3, 0, 0, 0, 0, domain.LondonLocation()), Amount: decimal.RequireFromString("20"), Currency: domain.CURRENCY_GBP, Description: "Refund", SourceRow: 3, Fingerprint: "fp2"},
	}
	if err := store.SaveTransactions(ctx, transactions); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.ListTransactions(ctx, "checking")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 || loaded[0].Amount.String() != "-12.5" {
		t.Fatalf("transactions = %+v", loaded)
	}
	summary, err := store.SummarizeTransactions(ctx, "checking")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary) != 1 || summary[0].Amount.String() != "7.5" {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestSaveImportRollsBackStatementAndEarlierTransactionsOnFailure(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	if err := store.SaveAccount(ctx, domain.Account{ID: "checking", BankSource: "Bank", AccountLabel: "Current", Currency: domain.CURRENCY_GBP, Owner: domain.OWNERSHIP_JOINT}); err != nil {
		t.Fatal(err)
	}
	statement := domain.Statement{ID: "rollback", AccountID: "checking", ImportedBy: domain.USER_ONE, Filename: "statement.csv", Format: domain.STATEMENT_FORMAT_CSV, Checksum: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", PeriodStart: time.Date(2026, 7, 1, 0, 0, 0, 0, domain.LondonLocation()), PeriodEnd: time.Date(2026, 7, 31, 0, 0, 0, 0, domain.LondonLocation()), Status: domain.STATEMENT_STATUS_IMPORTED, ImportedRows: 2}
	transactions := []domain.Transaction{
		{ID: "rollback:2", AccountID: "checking", StatementID: "rollback", Date: statement.PeriodStart, Amount: decimal.RequireFromString("1"), Currency: domain.CURRENCY_GBP, Description: "valid", SourceRow: 2, Fingerprint: "valid"},
		{ID: "rollback:3", AccountID: "checking", StatementID: "rollback", Date: statement.PeriodStart, Amount: decimal.RequireFromString("1"), Currency: domain.Currency("USD"), Description: "invalid", SourceRow: 3, Fingerprint: "invalid"},
	}
	if err := store.SaveImport(ctx, statement, transactions); err == nil {
		t.Fatal("invalid import was accepted")
	}
	statements, err := store.ListStatements(ctx)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.ListTransactions(ctx, "checking")
	if err != nil {
		t.Fatal(err)
	}
	if len(statements) != 0 || len(loaded) != 0 {
		t.Fatalf("rollback left statements=%+v transactions=%+v", statements, loaded)
	}
}
