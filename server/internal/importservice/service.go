package importservice

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/WorldOccupier/finny/server/internal/database"
	"github.com/WorldOccupier/finny/server/internal/domain"
	"github.com/WorldOccupier/finny/server/internal/importpreview"
	"github.com/shopspring/decimal"
)

var ErrPreviewNotFound = errors.New("preview not found")
var ErrPreviewUsed = errors.New("preview already confirmed")
var ErrDuplicateStatement = errors.New("statement already imported")
var ErrPreviewExpired = errors.New("preview expired")

const PREVIEW_TTL = time.Hour

type Preview struct {
	Token      string
	Result     importpreview.Preview
	AccountID  string
	ImportedBy domain.UserID
	CreatedAt  time.Time
	Filename   string
	Format     domain.StatementFormat
}
type Service struct {
	store      database.Store
	mu         sync.Mutex
	previews   map[string]*Preview
	confirming map[string]bool
}

func (s *Service) Statements(ctx context.Context) ([]domain.Statement, error) {
	return s.store.ListStatements(ctx)
}
func (s *Service) Accounts(ctx context.Context) ([]domain.Account, error) {
	return s.store.ListAccounts(ctx)
}
func (s *Service) Transactions(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	return s.store.ListTransactions(ctx, accountID)
}

func (s *Service) VisibleTransactions(ctx context.Context, user domain.UserID, accountID string) ([]domain.Transaction, error) {
	accounts, err := s.store.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	var result []domain.Transaction
	for _, account := range accounts {
		if (accountID != "" && account.ID != accountID) || !domain.VisibleTo(account, user) {
			continue
		}
		items, err := s.store.ListTransactions(ctx, account.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	return result, nil
}

func (s *Service) SummaryPeriod(ctx context.Context, user domain.UserID, accountID, period, anchor string) ([]database.TransactionSummary, error) {
	items, err := s.VisibleTransactions(ctx, user, accountID)
	if err != nil {
		return nil, err
	}
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, err
	}
	point := time.Now().In(location)
	if anchor != "" {
		point, err = time.ParseInLocation("2006-01-02", anchor, location)
		if err != nil {
			return nil, fmt.Errorf("invalid summary date")
		}
	}
	start := point
	switch period {
	case "day":
		start = time.Date(point.Year(), point.Month(), point.Day(), 0, 0, 0, 0, location)
	case "week":
		start = time.Date(point.Year(), point.Month(), point.Day(), 0, 0, 0, 0, location)
		start = start.AddDate(0, 0, -int((start.Weekday()+6)%7))
	case "month":
		start = time.Date(point.Year(), point.Month(), 1, 0, 0, 0, 0, location)
	case "year":
		start = time.Date(point.Year(), time.January, 1, 0, 0, 0, 0, location)
	default:
		return nil, fmt.Errorf("period must be day, week, month, or year")
	}
	end := start.AddDate(0, 0, 1)
	if period == "week" {
		end = start.AddDate(0, 0, 7)
	}
	if period == "month" {
		end = start.AddDate(0, 1, 0)
	}
	if period == "year" {
		end = start.AddDate(1, 0, 0)
	}
	totals := map[domain.Currency]decimal.Decimal{}
	for _, item := range items {
		if !item.Date.Before(start) && item.Date.Before(end) {
			totals[item.Currency] = totals[item.Currency].Add(item.Amount)
		}
	}
	result := make([]database.TransactionSummary, 0, len(totals))
	for currency, amount := range totals {
		result = append(result, database.TransactionSummary{Currency: currency, Amount: domain.Decimal{Decimal: amount}})
	}
	return result, nil
}
func (s *Service) Summary(ctx context.Context, accountID string) ([]database.TransactionSummary, error) {
	return s.store.SummarizeTransactions(ctx, accountID)
}

func New(store database.Store) *Service {
	return &Service{store: store, previews: make(map[string]*Preview), confirming: make(map[string]bool)}
}

func (s *Service) Preview(data []byte, request importpreview.ImportRequest, importedBy domain.UserID) (Preview, error) {
	result, err := importpreview.PreviewImport(data, request)
	if err != nil {
		return Preview{}, err
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return Preview{}, err
	}
	format := domain.STATEMENT_FORMAT_CSV
	if strings.HasSuffix(strings.ToLower(request.Filename), ".xlsx") {
		format = domain.STATEMENT_FORMAT_XLSX
	}
	p := Preview{Token: hex.EncodeToString(tokenBytes), Result: result, AccountID: request.AccountID, ImportedBy: importedBy, CreatedAt: time.Now(), Filename: request.Filename, Format: format}
	s.mu.Lock()
	s.previews[p.Token] = &p
	s.mu.Unlock()
	return p, nil
}

func (s *Service) Confirm(ctx context.Context, token string) (domain.Statement, int, error) {
	s.mu.Lock()
	p, ok := s.previews[token]
	if !ok {
		s.mu.Unlock()
		return domain.Statement{}, 0, ErrPreviewNotFound
	}
	if time.Since(p.CreatedAt) > PREVIEW_TTL {
		delete(s.previews, token)
		s.mu.Unlock()
		return domain.Statement{}, 0, ErrPreviewExpired
	}
	if s.confirming[token] {
		s.mu.Unlock()
		return domain.Statement{}, 0, ErrPreviewUsed
	}
	s.confirming[token] = true
	defer func() { s.mu.Lock(); delete(s.confirming, token); s.mu.Unlock() }()
	if p.Result.PeriodStart.IsZero() {
		s.mu.Unlock()
		return domain.Statement{}, 0, fmt.Errorf("preview has no valid period")
	}
	s.mu.Unlock()
	accounts, err := s.store.ListAccounts(ctx)
	if err != nil {
		return domain.Statement{}, 0, err
	}
	found := false
	for _, account := range accounts {
		if account.ID == p.AccountID {
			found = true
			break
		}
	}
	if !found {
		return domain.Statement{}, 0, fmt.Errorf("account not found")
	}
	statements, err := s.store.ListStatements(ctx)
	if err != nil {
		return domain.Statement{}, 0, err
	}
	for _, statement := range statements {
		if statement.Checksum == p.Result.Checksum {
			return domain.Statement{}, 0, ErrDuplicateStatement
		}
	}
	statementID := fmt.Sprintf("statement-%s", token)
	transactions := make([]domain.Transaction, len(p.Result.Transactions))
	copy(transactions, p.Result.Transactions)
	for i := range transactions {
		transactions[i].StatementID = statementID
		transactions[i].ID = fmt.Sprintf("%s:%d", statementID, transactions[i].SourceRow)
		transactions[i].Fingerprint = domain.TransactionFingerprint(transactions[i])
	}
	existing, err := s.store.ListTransactions(ctx, p.AccountID)
	if err != nil {
		return domain.Statement{}, 0, err
	}
	duplicates := 0
	filtered := transactions[:0]
	for _, item := range transactions {
		duplicate := false
		if item.Reference == "" {
			for _, old := range existing {
				if old.Reference == "" && old.Fingerprint == item.Fingerprint && old.Date.Equal(item.Date) && overlaps(item.Date, p.Result.PeriodStart, p.Result.PeriodEnd, statements, old.StatementID) {
					duplicate = true
					break
				}
			}
		}
		if duplicate {
			duplicates++
			continue
		}
		filtered = append(filtered, item)
	}
	transactions = filtered
	statement := domain.Statement{ID: statementID, AccountID: p.AccountID, ImportedBy: p.ImportedBy, Filename: p.Filename, Format: p.Format, Checksum: p.Result.Checksum, PeriodStart: p.Result.PeriodStart, PeriodEnd: p.Result.PeriodEnd, Status: domain.STATEMENT_STATUS_IMPORTED, ImportedRows: len(transactions), InvalidRows: p.Result.InvalidCount, DuplicateRows: duplicates}
	if err := s.store.SaveImport(ctx, statement, transactions); err != nil {
		return domain.Statement{}, 0, err
	}
	s.mu.Lock()
	delete(s.previews, token)
	s.mu.Unlock()
	return statement, len(transactions), nil
}

func overlaps(date, start, end time.Time, statements []domain.Statement, statementID string) bool {
	for _, statement := range statements {
		if statement.ID == statementID && !date.Before(statement.PeriodStart) && !date.After(statement.PeriodEnd) && !end.Before(statement.PeriodStart) && !start.After(statement.PeriodEnd) {
			return true
		}
	}
	return false
}
