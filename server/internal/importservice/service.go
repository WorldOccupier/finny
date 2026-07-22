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
)

var ErrPreviewNotFound = errors.New("preview not found")
var ErrPreviewUsed = errors.New("preview already confirmed")
var ErrDuplicateStatement = errors.New("statement already imported")

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
	store    database.Store
	mu       sync.Mutex
	previews map[string]*Preview
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
func (s *Service) Summary(ctx context.Context, accountID string) ([]database.TransactionSummary, error) {
	return s.store.SummarizeTransactions(ctx, accountID)
}

func New(store database.Store) *Service {
	return &Service{store: store, previews: make(map[string]*Preview)}
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
	if p.Result.PeriodStart.IsZero() {
		s.mu.Unlock()
		return domain.Statement{}, 0, fmt.Errorf("preview has no valid period")
	}
	delete(s.previews, token)
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
	statementID := fmt.Sprintf("statement-%d", len(statements)+1)
	for _, t := range p.Result.Transactions {
		t.StatementID = statementID
		t.ID = fmt.Sprintf("%s:%d", statementID, t.SourceRow)
		t.Fingerprint = domain.TransactionFingerprint(t)
	}
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
				if old.Reference == "" && old.Fingerprint == item.Fingerprint && old.Date.Equal(item.Date) {
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
	return statement, len(transactions), nil
}
