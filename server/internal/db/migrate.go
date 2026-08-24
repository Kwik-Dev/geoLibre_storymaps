package db

import (
	"database/sql"
	"embed"
	"fmt"
)

//go:embed schema.sql
var schemaFS embed.FS

// schemaSQL reads the embedded schema.sql into a string.
func schemaSQL() (string, error) {
	b, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return "", err
	}
	return string(b), err
}

// Migrate applies any pending database schema migrations. It creates the
// schema_migrations tracking table on first run, then applies embedded
// versioned SQL in order.  The DDL in schema.sql is idempotent (all
// CREATE … IF NOT EXISTS), so running Migrate twice is safe and produces
// only one version row.
func Migrate(db *sql.DB) error {
	// Create the migration-tracking table if it doesn't exist.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Check whether version 1 has already been applied.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 1`).Scan(&count); err != nil {
		return fmt.Errorf("check migration 1: %w", err)
	}

	if count == 0 {
		sql, err := schemaSQL()
		if err != nil {
			return fmt.Errorf("read schema.sql: %w", err)
		}
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("apply migration 1: %w", err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (1)`); err != nil {
			return fmt.Errorf("record migration 1: %w", err)
		}
	}

	return nil
}
