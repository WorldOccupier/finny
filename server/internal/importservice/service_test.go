package importservice

import (
	"context"
	"testing"

	"github.com/WorldOccupier/finny/server/internal/database"
	"github.com/WorldOccupier/finny/server/internal/domain"
	"github.com/WorldOccupier/finny/server/internal/importpreview"
)

func TestConfirmPersistsOnceAndRejectsChecksumRetry(t *testing.T) {
	db, err := database.Open(context.Background(), "file:import-service-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	store := database.NewSQLiteStore(db)
	if err := store.SaveAccount(context.Background(), domain.Account{ID: "checking", BankSource: "Bank", AccountLabel: "Current", Currency: domain.CURRENCY_GBP, Owner: domain.OWNERSHIP_JOINT}); err != nil {
		t.Fatal(err)
	}
	service := New(store)
	request := importpreview.ImportRequest{Filename: "statement.csv", AccountID: "checking", StatementID: "preview-statement", Mapping: importpreview.ColumnMapping{Date: 0, Description: 1, Amount: 2, Currency: -1, Debit: -1, Credit: -1}}
	preview, err := service.Preview([]byte("date,description,amount\n2026-07-01,Coffee,-1.25\n"), request, domain.USER_ONE)
	if err != nil {
		t.Fatal(err)
	}
	statement, count, err := service.Confirm(context.Background(), preview.Token)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || statement.ImportedBy != domain.USER_ONE {
		t.Fatalf("statement=%+v count=%d", statement, count)
	}
	if _, _, err := service.Confirm(context.Background(), preview.Token); err != ErrPreviewNotFound {
		t.Fatalf("second confirmation error=%v", err)
	}
	second, err := service.Preview([]byte("date,description,amount\n2026-07-01,Coffee,-1.25\n"), request, domain.USER_ONE)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.Confirm(context.Background(), second.Token); err != ErrDuplicateStatement {
		t.Fatalf("checksum retry error=%v", err)
	}
}
