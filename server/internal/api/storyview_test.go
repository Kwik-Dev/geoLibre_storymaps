package api

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

// TestStoryView exercises the P3.3 contract: StoryView(story, chapters) returns
// the exact legacy story JSON shape (camelCase keys), deep-equals the golden
// fixture, and serializes location fields as JSON numbers, not strings.
func TestStoryView(t *testing.T) {
	dir := t.TempDir()
	database := openDB(t, filepath.Join(dir, "p33.db"))
	defer database.Close()
	defer os.RemoveAll(dir)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	ownerID := seedUser(t, database, "101", "user")

	var storyID int64
	if err := database.QueryRow(`
		INSERT INTO stories (slug, author_id, title, subtitle, byline, visibility, status, global_view, created_at, updated_at)
		VALUES ('my-story', ?, 'My Story', 'Sub', 'By', 'public', 'approved', '{"center":[0,20],"zoom":0.6,"pitch":0,"bearing":0}', datetime('now'), datetime('now'))
		RETURNING id`, ownerID).Scan(&storyID); err != nil {
		t.Fatalf("insert story: %v", err)
	}

	insertChapter(t, database, storyID, 1, "Chapter One", "# Hello", "left", 0,
		`{"center":[-8.86,37.29],"zoom":11,"pitch":40,"bearing":0}`, "flyTo", 0,
		`[]`, `[]`, `{"id":1}`, "image", "external", "https://example.com/img.jpg", nil)
	insertChapter(t, database, storyID, 2, "Chapter Two", "Second", "center", 1,
		`{"center":[126.06,9.88],"zoom":10}`, "easeTo", 1,
		``, ``, ``, "none", "none", "", nil)

	// Load the story + chapters the way the API would.
	sh := NewStoriesHandler(database, nil, false)
	story, err := sh.loadByID(storyID)
	if err != nil {
		t.Fatalf("load story: %v", err)
	}
	chapters := loadChapters(t, database, storyID)

	gotJSON, err := json.Marshal(StoryView(story, chapters, ""))
	if err != nil {
		t.Fatalf("marshal StoryView: %v", err)
	}

	golden, err := os.ReadFile("_test/story_view.golden.json")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(gotJSON, &got); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal(golden, &want); err != nil {
		t.Fatalf("unmarshal golden: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("StoryView mismatch:\n got: %s\nwant: %s", gotJSON, golden)
	}

	// location fields must serialize as JSON numbers, not strings.
	var doc struct {
		Chapters []struct {
			Location *struct {
				Center []float64 `json:"center"`
				Zoom   float64   `json:"zoom"`
			} `json:"location"`
		} `json:"chapters"`
	}
	if err := json.Unmarshal(gotJSON, &doc); err != nil {
		t.Fatalf("unmarshal for number check: %v", err)
	}
	if len(doc.Chapters) != 2 || doc.Chapters[0].Location == nil {
		t.Fatalf("expected 2 chapters with location; got %+v", doc.Chapters)
	}
	if doc.Chapters[0].Location.Zoom != 11 || len(doc.Chapters[0].Location.Center) != 2 {
		t.Fatalf("location zoom/center not numbers: %+v", doc.Chapters[0].Location)
	}
}

// insertChapter inserts a chapters row with the given values. Empty JSON/string
// fields are stored as NULL so the adapter's omit-empty behavior is exercised.
func insertChapter(t *testing.T, database *sql.DB, storyID int64, position int, title, desc, alignment string, hidden int, location, mapAnim string, rotate int, enter, exit, source, mediaType, mediaRefType, mediaURL string, assetID *int64) {
	t.Helper()
	var asset interface{}
	if assetID != nil {
		asset = *assetID
	}
	if _, err := database.Exec(`
		INSERT INTO chapters (story_id, position, title, description_md, alignment, hidden, location,
			map_animation, rotate_animation, on_chapter_enter, on_chapter_exit, source,
			media_type, media_ref_type, media_external_url, media_asset_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		storyID, position, title, desc, alignment, hidden, nullableStr(location), mapAnim, rotate,
		nullableStr(enter), nullableStr(exit), source, mediaType, mediaRefType, mediaURL, asset); err != nil {
		t.Fatalf("insert chapter: %v", err)
	}
}

// nullableStr returns nil for an empty string so it is stored as SQL NULL.
func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// loadChapters reads a story's chapters in render order (position, created_at).
func loadChapters(t *testing.T, database *sql.DB, storyID int64) []Chapter {
	t.Helper()
	rows, err := database.Query(`
		SELECT id, story_id, position, title, description_md, alignment, hidden, location,
			map_animation, rotate_animation, on_chapter_enter, on_chapter_exit, source,
			media_type, media_ref_type, media_external_url, media_asset_id, created_at, updated_at
		FROM chapters WHERE story_id = ? AND deleted_at IS NULL ORDER BY position, created_at`, storyID)
	if err != nil {
		t.Fatalf("query chapters: %v", err)
	}
	defer rows.Close()
	var chapters []Chapter
	for rows.Next() {
		c, err := scanChapter(rows)
		if err != nil {
			t.Fatalf("scan chapter: %v", err)
		}
		chapters = append(chapters, c)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate chapters: %v", err)
	}
	return chapters
}
