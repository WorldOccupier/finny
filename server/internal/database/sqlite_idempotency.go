package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *SQLiteStore) GetIdempotencyResult(ctx context.Context, key string) (IdempotencyResult, error) {
	var result IdempotencyResult
	var createdAt string
	err := s.db.QueryRowContext(ctx, `SELECT idempotency_key, request_hash, response_json, created_at FROM idempotency_keys WHERE idempotency_key = ?`, key).Scan(&result.Key, &result.RequestHash, &result.ResponseJSON, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return IdempotencyResult{}, ErrNotFound
	}
	if err != nil {
		return IdempotencyResult{}, fmt.Errorf("read idempotency result: %w", err)
	}
	result.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return IdempotencyResult{}, fmt.Errorf("parse idempotency timestamp: %w", err)
	}
	return result, nil
}

func (s *SQLiteStore) SaveIdempotencyResult(ctx context.Context, result IdempotencyResult) error {
	if result.Key == "" || result.RequestHash == "" || result.ResponseJSON == "" || result.CreatedAt.IsZero() {
		return fmt.Errorf("idempotency result is incomplete")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO idempotency_keys (idempotency_key, request_hash, response_json, created_at) VALUES (?, ?, ?, ?)`, result.Key, result.RequestHash, result.ResponseJSON, result.CreatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("write idempotency result: %w", err)
	}
	return nil
}
