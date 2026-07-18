package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/WorldOccupier/finny/server/internal/domain"
)

var ErrNotFound = errors.New("database record not found")
var ErrRevisionConflict = errors.New("dashboard revision conflict")
var ErrIdempotencyConflict = errors.New("idempotency key conflict")

type IdempotencyResult struct {
	Key          string
	RequestHash  string
	ResponseJSON string
	CreatedAt    time.Time
}

type DashboardSnapshotCommit struct {
	Replayed    bool
	Idempotency IdempotencyResult
}

type SQLiteStore struct{ db *sql.DB }

func NewSQLiteStore(db *sql.DB) *SQLiteStore { return &SQLiteStore{db: db} }

type queryer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func parseDecimal(value, field string) (domain.Decimal, error) {
	parsed, err := domain.NewDecimal(value)
	if err != nil {
		return domain.Decimal{}, fmt.Errorf("parse %s: %w", field, err)
	}
	return parsed, nil
}
