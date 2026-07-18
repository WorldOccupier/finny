package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type migration struct {
	version int64
	name    string
	sql     string
}

func Migrate(ctx context.Context, db *sql.DB) error {
	return migrateFS(ctx, db, migrationFiles)
}

func migrateFS(ctx context.Context, db *sql.DB, files fs.FS) error {
	migrations, err := loadMigrations(files)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	for _, current := range migrations {
		var appliedName string
		err := db.QueryRowContext(ctx, `SELECT name FROM schema_migrations WHERE version = ?`, current.version).Scan(&appliedName)
		if err == nil {
			if appliedName != current.name {
				return fmt.Errorf("migration version %d has name %q, expected %q", current.version, appliedName, current.name)
			}
			continue
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration version %d: %w", current.version, err)
		}

		if err := applyMigration(ctx, db, current); err != nil {
			return err
		}
	}

	return nil
}

func loadMigrations(files fs.FS) ([]migration, error) {
	entries, err := fs.Glob(files, "migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(entries)

	result := make([]migration, 0, len(entries))
	seen := make(map[int64]struct{}, len(entries))
	for _, entry := range entries {
		base := path.Base(entry)
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 || !strings.HasSuffix(parts[1], ".sql") {
			return nil, fmt.Errorf("invalid migration filename %q", base)
		}
		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || version <= 0 {
			return nil, fmt.Errorf("invalid migration version in %q", base)
		}
		if _, exists := seen[version]; exists {
			return nil, fmt.Errorf("duplicate migration version %d", version)
		}
		contents, err := fs.ReadFile(files, entry)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", base, err)
		}
		seen[version] = struct{}{}
		result = append(result, migration{version: version, name: base, sql: string(contents)})
	}

	return result, nil
}

func applyMigration(ctx context.Context, db *sql.DB, current migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", current.version, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, current.sql); err != nil {
		return fmt.Errorf("apply migration %d: %w", current.version, err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO schema_migrations (version, name, applied_at)
		VALUES (?, ?, ?)
	`, current.version, current.name, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("record migration %d: %w", current.version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d: %w", current.version, err)
	}
	return nil
}
