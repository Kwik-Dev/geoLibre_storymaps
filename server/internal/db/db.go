package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config"
)

// Open opens a SQLite database at the path specified in cfg.DBPath and
// configures WAL journal mode, busy timeout, and foreign key enforcement.
// Callers should call Migrate on the returned handle before use.
func Open(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// WAL mode for concurrent-read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	// Busy timeout so concurrent writers wait instead of failing immediately.
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}

	// Foreign key enforcement (off by default in SQLite).
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}

	return db, nil
}
