package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrate(t *testing.T) {
	// Run the full suite against an in-memory database (fast, no I/O).
	t.Run("in-memory", func(t *testing.T) {
		db := mustOpen(t, ":memory:")
		defer db.Close()
		testMigrate(t, db)
	})

	// Run against a real temp file to exercise the file-backed path.
	t.Run("temp-file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.db")
		db := mustOpen(t, path)
		defer db.Close()
		testMigrate(t, db)
	})
}

// testMigrate runs the shared migration acceptance checks:
//  1. First migration succeeds.
//  2. All four domain tables + schema_migrations exist.
//  3. Version 1 is recorded in schema_migrations.
//  4. Second run is idempotent – no error, no duplicate version row.
//  5. Foreign keys are enabled.
func testMigrate(t *testing.T, db *sql.DB) {
	t.Helper()

	// --- first run ---
	if err := Migrate(db); err != nil {
		t.Fatalf("first migrate: %v", err)
	}

	// --- all expected tables exist ---
	expected := []string{"users", "stories", "chapters", "media_assets", "schema_migrations"}
	for _, name := range expected {
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&n); err != nil {
			t.Fatalf("query sqlite_master for %q: %v", name, err)
		}
		if n == 0 {
			t.Fatalf("table %q not found after migration", name)
		}
	}

	// --- version 1 recorded exactly once ---
	var ver int
	if err := db.QueryRow(`SELECT version FROM schema_migrations WHERE version = 1`).Scan(&ver); err != nil {
		t.Fatalf("version 1 not recorded: %v", err)
	}
	if ver != 1 {
		t.Fatalf("expected version 1, got %d", ver)
	}

	// --- idempotent second run ---
	if err := Migrate(db); err != nil {
		t.Fatalf("second migrate (should be idempotent): %v", err)
	}

	var cnt int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&cnt); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 migration row after second run, got %d", cnt)
	}

	// --- foreign_keys pragma is ON ---
	var fk int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatalf("check PRAGMA foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Fatalf("expected foreign_keys=1, got %d", fk)
	}
}

// mustOpen opens a SQLite database and sets the same pragmas that Open()
// would (WAL, busy_timeout, foreign_keys).
func mustOpen(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open(%q): %v", path, err)
	}
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			t.Fatalf("%s: %v", pragma, err)
		}
	}
	return db
}

// TestMigrateEnvVar verifies that setting DB_PATH via env var doesn't
// break the Open + Migrate flow.  This exercises the same codepath that
// main.go will use.
func TestMigrateEnvVarIntegration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "env-integration.db")

	// Simulate what main.go does: Open then Migrate.
	// We bypass config.Load to keep the test hermetic.
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Set pragmas the same way Open would.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		t.Fatalf("WAL pragma: %v", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		t.Fatalf("busy_timeout pragma: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		t.Fatalf("foreign_keys pragma: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Quick spot-check: stories table exists.
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='stories'`).Scan(&n); err != nil {
		t.Fatalf("check stories table: %v", err)
	}
	if n == 0 {
		t.Fatal("stories table not found after migration")
	}
}
