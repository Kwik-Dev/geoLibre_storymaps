package server

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

// TestPurge verifies the soft-delete purge cron:
//   - with a tiny PURGE_TTL (1s) and a pre-aged soft-deleted row, the row is
//     hard-deleted;
//   - a freshly soft-deleted row is NOT purged;
//   - the purge is idempotent and respects foreign keys (chapters of a purged
//     story are removed, media refs are detached).
func TestPurge(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "purge-test.db")
	database := openDB(t, dbPath)
	defer database.Close()
	defer os.Remove(dbPath)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Seed an author.
	if _, err := database.Exec(
		`INSERT INTO users (github_login, role, created_at) VALUES ('purger', 'user', datetime('now'))`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	var authorID int64
	if err := database.QueryRow(`SELECT id FROM users WHERE github_login='purger'`).Scan(&authorID); err != nil {
		t.Fatalf("read author id: %v", err)
	}

	insertStory := func(slug, deletedAt string) int64 {
		t.Helper()
		res, err := database.Exec(`
			INSERT INTO stories (slug, author_id, title, visibility, status, deleted_at)
			VALUES (?, ?, 't', 'private', 'draft', ?)`, slug, authorID, deletedAt)
		if err != nil {
			t.Fatalf("insert story %s: %v", slug, err)
		}
		id, _ := res.LastInsertId()
		return id
	}

	// A story soft-deleted long ago (older than the 1s TTL) → must be purged.
	agedID := insertStory("aged", "2000-01-01 00:00:00")
	// A story soft-deleted just now → must NOT be purged.
	freshID := insertStory("fresh", time.Now().UTC().Format("2006-01-02 15:04:05"))

	// A chapter soft-deleted long ago → must be purged.
	if _, err := database.Exec(`
		INSERT INTO chapters (story_id, position, title, deleted_at)
		VALUES (?, 0, 'old chapter', '2000-01-01 00:00:00')`, freshID); err != nil {
		t.Fatalf("insert aged chapter: %v", err)
	}
	// A chapter soft-deleted just now → must NOT be purged.
	if _, err := database.Exec(`
		INSERT INTO chapters (story_id, position, title, deleted_at)
		VALUES (?, 1, 'fresh chapter', ?)`, freshID, time.Now().UTC().Format("2006-01-02 15:04:05")); err != nil {
		t.Fatalf("insert fresh chapter: %v", err)
	}

	// A media asset soft-deleted long ago → must be purged.
	if _, err := database.Exec(`
		INSERT INTO media_assets (kind, stored_path, filename, deleted_at)
		VALUES ('image', 'x', 'x.png', '2000-01-01 00:00:00')`); err != nil {
		t.Fatalf("insert aged asset: %v", err)
	}
	// A media asset soft-deleted just now → must NOT be purged.
	if _, err := database.Exec(`
		INSERT INTO media_assets (kind, stored_path, filename, deleted_at)
		VALUES ('image', 'y', 'y.png', ?)`, time.Now().UTC().Format("2006-01-02 15:04:05")); err != nil {
		t.Fatalf("insert fresh asset: %v", err)
	}

	// Run the purge with a tiny TTL (1s).
	res, err := PurgeOnce(database, time.Second)
	if err != nil {
		t.Fatalf("PurgeOnce: %v", err)
	}

	// The aged story must be gone; the fresh one must remain.
	if rowExists(t, database, "stories", agedID) {
		t.Fatalf("aged story %d should have been purged", agedID)
	}
	if !rowExists(t, database, "stories", freshID) {
		t.Fatalf("fresh story %d should NOT have been purged", freshID)
	}
	if res.Stories != 1 {
		t.Fatalf("expected 1 story purged, got %d", res.Stories)
	}

	// The aged chapter must be gone; the fresh one must remain.
	if chapterCount(t, database, "old chapter") != 0 {
		t.Fatal("aged chapter should have been purged")
	}
	if chapterCount(t, database, "fresh chapter") != 1 {
		t.Fatal("fresh chapter should NOT have been purged")
	}
	if res.Chapters != 1 {
		t.Fatalf("expected 1 chapter purged, got %d", res.Chapters)
	}

	// The aged asset must be gone; the fresh one must remain.
	if assetCount(t, database, "x.png") != 0 {
		t.Fatal("aged asset should have been purged")
	}
	if assetCount(t, database, "y.png") != 1 {
		t.Fatal("fresh asset should NOT have been purged")
	}
	if res.MediaAssets != 1 {
		t.Fatalf("expected 1 media asset purged, got %d", res.MediaAssets)
	}

	// Idempotency: a second pass deletes nothing more.
	res2, err := PurgeOnce(database, time.Second)
	if err != nil {
		t.Fatalf("PurgeOnce (2nd): %v", err)
	}
	if res2.Stories != 0 || res2.Chapters != 0 || res2.MediaAssets != 0 {
		t.Fatalf("second purge should be a no-op, got %+v", res2)
	}
}

// TestPurgeTTLGuard verifies that a non-positive TTL is rejected.
func TestPurgeTTLGuard(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "purge-ttl.db")
	database := openDB(t, dbPath)
	defer database.Close()
	defer os.Remove(dbPath)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if _, err := PurgeOnce(database, 0); err == nil {
		t.Fatal("expected error for ttl <= 0")
	}
}

// TestStartPurge verifies the background job runs a purge on startup and can be
// stopped cleanly.
func TestStartPurge(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "purge-job.db")
	database := openDB(t, dbPath)
	defer database.Close()
	defer os.Remove(dbPath)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Seed an author + an aged soft-deleted story.
	if _, err := database.Exec(
		`INSERT INTO users (github_login, role, created_at) VALUES ('purger2', 'user', datetime('now'))`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	var authorID int64
	if err := database.QueryRow(`SELECT id FROM users WHERE github_login='purger2'`).Scan(&authorID); err != nil {
		t.Fatalf("read author id: %v", err)
	}
	if _, err := database.Exec(`
		INSERT INTO stories (slug, author_id, title, visibility, status, deleted_at)
		VALUES ('job-aged', ?, 't', 'private', 'draft', '2000-01-01 00:00:00')`, authorID); err != nil {
		t.Fatalf("insert aged story: %v", err)
	}

	job := StartPurge(database, time.Second, 50*time.Millisecond)
	// Give the startup pass a moment to run.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if storyCount(t, database, "job-aged") == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	job.Stop()

	if storyCount(t, database, "job-aged") != 0 {
		t.Fatal("StartPurge should have purged the aged story on startup")
	}
}

func rowExists(t *testing.T, db *sql.DB, table string, id int64) bool {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM `+table+` WHERE id = ?`, id).Scan(&n); err != nil {
		t.Fatalf("count %s %d: %v", table, id, err)
	}
	return n > 0
}

func chapterCount(t *testing.T, db *sql.DB, title string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM chapters WHERE title = ?`, title).Scan(&n); err != nil {
		t.Fatalf("count chapter %q: %v", title, err)
	}
	return n
}

func assetCount(t *testing.T, db *sql.DB, filename string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM media_assets WHERE filename = ?`, filename).Scan(&n); err != nil {
		t.Fatalf("count asset %q: %v", filename, err)
	}
	return n
}

func storyCount(t *testing.T, db *sql.DB, slug string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM stories WHERE slug = ?`, slug).Scan(&n); err != nil {
		t.Fatalf("count story %q: %v", slug, err)
	}
	return n
}
