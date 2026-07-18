package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const SQLITE_DRIVER = "sqlite"

func Open(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open(SQLITE_DRIVER, path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}

	// A single connection keeps PRAGMA foreign_keys consistent for this POC and
	// also makes file::memory:?cache=shared databases deterministic in tests.
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("enable SQLite foreign keys: %w", err)
	}
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("ping SQLite database: %w", err)
	}

	return db, nil
}
