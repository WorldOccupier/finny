package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestAccountVisibility(t *testing.T) {
	for _, test := range []struct {
		owner OwnershipScope
		user  UserID
		want  bool
	}{
		{OWNERSHIP_USER_ONE, USER_ONE, true},
		{OWNERSHIP_USER_ONE, USER_TWO, false},
		{OWNERSHIP_USER_TWO, USER_ONE, false},
		{OWNERSHIP_USER_TWO, USER_TWO, true},
		{OWNERSHIP_JOINT, USER_ONE, true},
		{OWNERSHIP_JOINT, USER_TWO, true},
	} {
		account := Account{ID: "account", BankSource: "Bank", AccountLabel: "Current", Currency: CURRENCY_GBP, Owner: test.owner}
		if got := VisibleTo(account, test.user); got != test.want {
			t.Errorf("VisibleTo(%q, %q) = %t, want %t", test.owner, test.user, got, test.want)
		}
	}
}

func TestTransactionValidation(t *testing.T) {
	amount, err := decimal.NewFromString("-12.50")
	if err != nil {
		t.Fatal(err)
	}
	transaction := Transaction{ID: "transaction", AccountID: "account", StatementID: "statement", Date: time.Date(2026, 7, 22, 12, 0, 0, 0, LondonLocation()), Amount: amount, Currency: CURRENCY_GBP, Description: "Coffee", SourceRow: 2}
	if err := transaction.Validate(); err != nil {
		t.Fatal(err)
	}
	transaction.Currency = Currency("USD")
	if err := transaction.Validate(); err == nil {
		t.Fatal("unsupported transaction currency was accepted")
	}
	transaction.Currency = CURRENCY_GBP
	transaction.Amount = decimal.Decimal{}
	if err := transaction.Validate(); err != nil {
		t.Fatal(err)
	}
	transaction.Date = time.Time{}
	if err := transaction.Validate(); err == nil {
		t.Fatal("zero transaction date was accepted")
	}
	transaction.Date = time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	if err := transaction.Validate(); err == nil {
		t.Fatal("non-London transaction date was accepted")
	}
}

func TestDomainMetadataValidation(t *testing.T) {
	if err := (User{ID: UserID("unknown")}).Validate(); err == nil {
		t.Fatal("invalid user was accepted")
	}
	if err := (Account{ID: "account", BankSource: "Bank", AccountLabel: "Current", Currency: CURRENCY_INR, Owner: OWNERSHIP_JOINT}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Account{ID: "account", BankSource: "", AccountLabel: "Current", Currency: CURRENCY_GBP, Owner: OWNERSHIP_USER_ONE}).Validate(); err == nil {
		t.Fatal("missing bank source was accepted")
	}
	statement := Statement{ID: "statement", AccountID: "account", ImportedBy: USER_ONE, Filename: "statement.csv", Format: STATEMENT_FORMAT_CSV, Checksum: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", PeriodStart: time.Date(2026, 7, 1, 0, 0, 0, 0, LondonLocation()), PeriodEnd: time.Date(2026, 7, 31, 0, 0, 0, 0, LondonLocation()), Status: STATEMENT_STATUS_IMPORTED}
	if err := statement.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []Statement{
		statement,
		{ID: statement.ID, AccountID: statement.AccountID, ImportedBy: statement.ImportedBy, Filename: statement.Filename, Format: StatementFormat("pdf"), Checksum: statement.Checksum, PeriodStart: statement.PeriodStart, PeriodEnd: statement.PeriodEnd, Status: statement.Status},
		{ID: statement.ID, AccountID: statement.AccountID, ImportedBy: statement.ImportedBy, Filename: statement.Filename, Format: statement.Format, Checksum: "bad", PeriodStart: statement.PeriodStart, PeriodEnd: statement.PeriodEnd, Status: statement.Status},
		{ID: statement.ID, AccountID: statement.AccountID, ImportedBy: statement.ImportedBy, Filename: statement.Filename, Format: statement.Format, Checksum: statement.Checksum, PeriodStart: statement.PeriodStart, PeriodEnd: statement.PeriodEnd, Status: statement.Status, InvalidRows: -1},
	} {
		if invalid == statement {
			continue
		}
		if err := invalid.Validate(); err == nil {
			t.Errorf("invalid statement metadata was accepted: %+v", invalid)
		}
	}
	statement.PeriodStart = statement.PeriodStart.UTC()
	if err := statement.Validate(); err == nil {
		t.Fatal("non-London statement period was accepted")
	}
}

func TestTransactionFingerprintNormalizesTextAndExcludesSourceIdentity(t *testing.T) {
	amount, err := decimal.NewFromString("12.5")
	if err != nil {
		t.Fatal(err)
	}
	first := Transaction{AccountID: "account", StatementID: "statement-a", Date: time.Date(2026, 7, 22, 23, 0, 0, 0, time.UTC), Amount: amount, Currency: CURRENCY_GBP, Description: "  Coffee\tShop ", Reference: " Ref-1 "}
	first.Date = time.Date(2026, 7, 22, 23, 0, 0, 0, LondonLocation())
	second := first
	second.StatementID = "statement-b"
	second.SourceRow = 99
	second.Description = "coffee shop"
	second.Reference = "ref-1"
	if TransactionFingerprint(first) != TransactionFingerprint(second) {
		t.Fatal("equivalent transactions received different fingerprints")
	}
	second.Amount = decimal.NewFromInt(13)
	if TransactionFingerprint(first) == TransactionFingerprint(second) {
		t.Fatal("different transactions received the same fingerprint")
	}
	if TransactionSourceIdentity(first) == TransactionSourceIdentity(second) {
		t.Fatal("source-row identity ignored statement and row")
	}
	withDelimiter := first
	withDelimiter.Description = "a|b"
	withDelimiter.Reference = "c"
	withoutDelimiter := first
	withoutDelimiter.Description = "a"
	withoutDelimiter.Reference = "b|c"
	if TransactionFingerprint(withDelimiter) == TransactionFingerprint(withoutDelimiter) {
		t.Fatal("delimiter-containing fields collided")
	}
}
