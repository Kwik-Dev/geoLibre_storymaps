package server

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

// TestSeed verifies the demo seed behaviour:
//   - With SEED_DEMO=1, a demo story + chapters exist.
//   - With SEED_DEMO unset (default), nothing is inserted.
//   - Running seed twice is idempotent (no duplicate story).
func TestSeed(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "seed-test.db")

	// Open with the same pragmas Open() would set.
	database := openDB(t, dbPath)
	defer database.Close()
	defer os.Remove(dbPath)

	// Migrate so tables exist.
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	cfg := &config.Config{
		AdminEmail: "test@example.com",
	}

	t.Run("default-off", func(t *testing.T) {
		// SEED_DEMO not set — seed must be a no-op.
		if err := SeedDemo(cfg, database); err != nil {
			t.Fatalf("SeedDemo (off): %v", err)
		}
		assertStoryCount(t, database, 0, "expected no stories when SEED_DEMO is off")
	})

	t.Run("seed-on", func(t *testing.T) {
		os.Setenv("SEED_DEMO", "1")
		defer os.Unsetenv("SEED_DEMO")

		if err := SeedDemo(cfg, database); err != nil {
			t.Fatalf("SeedDemo (on): %v", err)
		}

		// One story should exist.
		assertStoryCount(t, database, 1, "expected one demo story")

		// Chapters should exist.
		var chapterCount int
		if err := database.QueryRow(`SELECT COUNT(*) FROM chapters`).Scan(&chapterCount); err != nil {
			t.Fatalf("count chapters: %v", err)
		}
		if chapterCount == 0 {
			t.Fatal("expected at least one chapter after seeding")
		}

		// Verify the story slug.
		var slug string
		if err := database.QueryRow(`SELECT slug FROM stories LIMIT 1`).Scan(&slug); err != nil {
			t.Fatalf("get story slug: %v", err)
		}
		if slug != "demo-scrollytelling" {
			t.Fatalf("expected slug 'demo-scrollytelling', got %q", slug)
		}

		// Verify visibility is public and status is approved.
		var visibility, status string
		if err := database.QueryRow(`SELECT visibility, status FROM stories WHERE slug = ?`, slug).Scan(&visibility, &status); err != nil {
			t.Fatalf("get story visibility/status: %v", err)
		}
		if visibility != "public" {
			t.Fatalf("expected visibility 'public', got %q", visibility)
		}
		if status != "approved" {
			t.Fatalf("expected status 'approved', got %q", status)
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		// SEED_DEMO still set from previous subtest.
		if err := SeedDemo(cfg, database); err != nil {
			t.Fatalf("SeedDemo (idempotent): %v", err)
		}
		assertStoryCount(t, database, 1, "expected still one story after idempotent second call")
	})
}

// openDB opens a SQLite database with the same pragmas as db.Open().
func openDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open(%q): %v", path, err)
	}
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := database.Exec(pragma); err != nil {
			database.Close()
			t.Fatalf("%s: %v", pragma, err)
		}
	}
	return database
}

// assertStoryCount checks that the stories table contains exactly want rows.
func assertStoryCount(t *testing.T, db *sql.DB, want int, msg string) {
	t.Helper()
	var got int
	if err := db.QueryRow(`SELECT COUNT(*) FROM stories`).Scan(&got); err != nil {
		t.Fatalf("count stories: %v", err)
	}
	if got != want {
		t.Fatalf("%s: got %d, want %d", msg, got, want)
	}
}
