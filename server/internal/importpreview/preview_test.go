package importpreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

func TestPreviewCSVKeepsValidRowsAndReportsInvalidRows(t *testing.T) {
	data := []byte("Date,Description,Amount,Currency,Reference\n2026-07-01,Coffee,-12.50,GBP,a\nnot-a-date,Invalid,2,GBP,b\n2026-07-02,Salary,100,INR,c\n2026-07-03,Bad,1,USD,d\n")
	preview, err := PreviewImport(data, ImportRequest{Filename: "statement.csv", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: 2, Debit: -1, Credit: -1, Currency: 3, Reference: 4}})
	if err != nil {
		t.Fatal(err)
	}
	if preview.ValidRows != 2 || preview.InvalidCount != 2 {
		t.Fatalf("counts = %d/%d", preview.ValidRows, preview.InvalidCount)
	}
	if preview.Transactions[0].Amount.Cmp(decimal.RequireFromString("-12.50")) != 0 || preview.Transactions[0].SourceRow != 2 {
		t.Fatalf("unexpected first transaction: %+v", preview.Transactions[0])
	}
	if preview.InvalidRows[0].SourceRow != 3 || preview.InvalidRows[1].SourceRow != 5 {
		t.Fatalf("invalid rows = %+v", preview.InvalidRows)
	}
	if preview.PeriodStart.Day() != 1 || preview.PeriodEnd.Day() != 2 || preview.Checksum == "" {
		t.Fatal("preview metadata missing")
	}
}

func TestPreviewCSVNormalizesDebitAndCredit(t *testing.T) {
	data := []byte("Date,Description,Debit,Credit\n2026-07-01,Debit,4.25,\n2026-07-02,Credit,,8.50\n2026-07-03,Both,1,2\n2026-07-04,Neither,,\n")
	preview, err := PreviewImport(data, ImportRequest{Filename: "statement.csv", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: -1, Debit: 2, Credit: 3, Currency: -1, Reference: -1}})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Transactions[0].Amount.Cmp(decimal.RequireFromString("-4.25")) != 0 || preview.Transactions[1].Amount.Cmp(decimal.RequireFromString("8.50")) != 0 {
		t.Fatal("debit/credit normalization failed")
	}
	if preview.InvalidCount != 2 {
		t.Fatalf("invalid count = %d", preview.InvalidCount)
	}
}

func TestPreviewXLSX(t *testing.T) {
	file := excelize.NewFile()
	file.SetSheetRow("Sheet1", "A1", &[]interface{}{"Date", "Description", "Amount", "Currency"})
	file.SetSheetRow("Sheet1", "A2", &[]interface{}{"2026-07-04", "Train", "-3.20", "GBP"})
	file.SetSheetRow("Sheet1", "A3", &[]interface{}{"2026-07-05", "", "-1.00", "GBP"})
	file.SetSheetRow("Sheet1", "A4", &[]interface{}{"", "", "", ""})
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		t.Fatal(err)
	}
	preview, err := PreviewImport(buffer.Bytes(), ImportRequest{Filename: "statement.xlsx", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: 2, Debit: -1, Credit: -1, Currency: 3, Reference: -1}})
	if err != nil {
		t.Fatal(err)
	}
	if preview.ValidRows != 1 || preview.InvalidCount != 1 || preview.Transactions[0].Amount.Cmp(decimal.RequireFromString("-3.20")) != 0 || preview.InvalidRows[0].SourceRow != 3 {
		t.Fatalf("unexpected XLSX preview: %+v", preview)
	}
}

func TestPreviewValidatesMappingsAndMalformedFiles(t *testing.T) {
	if _, err := PreviewImport([]byte("x"), ImportRequest{Filename: "statement.csv", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: -1, Debit: -1, Credit: -1}}); err == nil {
		t.Fatal("missing amount mapping accepted")
	}
	if _, err := PreviewImport([]byte("not an xlsx"), ImportRequest{Filename: "statement.xlsx", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: 2, Debit: -1, Credit: -1}}); err == nil {
		t.Fatal("malformed XLSX accepted")
	}
	if _, err := PreviewImport([]byte("x"), ImportRequest{Filename: "statement.csv", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: -1, Debit: 2, Credit: -1}}); err == nil {
		t.Fatal("incomplete split mapping accepted")
	}
	if _, err := PreviewImport([]byte("x"), ImportRequest{Filename: "statement.csv", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: 2, Debit: 3, Credit: -1}}); err == nil {
		t.Fatal("mixed amount mapping accepted")
	}
}

func TestPreviewSkipsBlankRowsQuotedFieldsAndPreservesSourceRows(t *testing.T) {
	data := []byte("Date,Description,Amount\n2026-07-01,\"Coffee, shop\",-1.25\n,,\n2026-07-02,Train,2\n")
	preview, err := PreviewImport(data, ImportRequest{Filename: "statement.csv", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: 2, Debit: -1, Credit: -1, Currency: -1, Reference: -1}})
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Transactions) != 2 || preview.Transactions[1].SourceRow != 4 || preview.Transactions[0].Description != "Coffee, shop" {
		t.Fatalf("blank/quoted rows mishandled: %+v", preview.Transactions)
	}
	if _, err := PreviewImport([]byte("Date,Description,Amount\n2026-07-01,broken\"quote,-1\n"), ImportRequest{Filename: "statement.csv", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: 2, Debit: -1, Credit: -1}}); err == nil {
		t.Fatal("malformed CSV record accepted")
	}
}

func TestPreviewMetadataAndFingerprintCompatibility(t *testing.T) {
	data := []byte("Date,Description,Amount\n2026-07-01,Coffee,-1.25\n")
	preview, err := PreviewImport(data, ImportRequest{Filename: "statement.csv", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: 2, Debit: -1, Credit: -1, Currency: -1, Reference: -1}})
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(data)
	if preview.Checksum != hex.EncodeToString(hash[:]) || preview.PeriodStart.Location() != domain.LondonLocation() || preview.PeriodEnd.Location() != domain.LondonLocation() {
		t.Fatalf("metadata mismatch: %+v", preview)
	}
	transaction := preview.Transactions[0]
	if transaction.Fingerprint != domain.TransactionFingerprint(transaction) || domain.TransactionSourceIdentity(transaction) != "statement:2" {
		t.Fatal("transaction identity is not Phase 13 compatible")
	}
	if !transaction.Date.Equal(time.Date(2026, 7, 1, 0, 0, 0, 0, domain.LondonLocation())) {
		t.Fatalf("unexpected London date: %v", transaction.Date)
	}
	second, err := PreviewImport(data, ImportRequest{Filename: "statement.csv", AccountID: "account", StatementID: "statement", Mapping: ColumnMapping{Date: 0, Description: 1, Amount: 2, Debit: -1, Credit: -1, Currency: -1, Reference: -1}})
	if err != nil || len(second.Transactions) != 1 {
		t.Fatal("preview should have no persistence side effects")
	}
}
