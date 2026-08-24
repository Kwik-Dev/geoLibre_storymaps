package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

// newExportRouter builds a production-shaped router: RequireAuth middleware +
// the export route. It returns the router and the token-signing JWT for seeding
// authenticated requests.
func newExportRouter(t *testing.T, database *sql.DB) *chi.Mux {
	t.Helper()
	r := chi.NewRouter()
	auther := auth.NewAuthenticator(testSecret, false)
	r.Route("/api", func(api chi.Router) {
		api.Use(auther.RequireAuth)
		NewExportHandler(database, auther).Routes(api)
	})
	return r
}

// TestExport exercises the P3.4 contract:
//   - GET /api/stories/:id/export returns 200 JSON matching the golden legacy
//     shape with the right Content-Type and Content-Disposition
//   - a private story exported by a non-owner → 403
//   - a public story is exportable anonymously
func TestExport(t *testing.T) {
	dir := t.TempDir()
	database := openDB(t, filepath.Join(dir, "p34.db"))
	defer database.Close()
	defer os.RemoveAll(dir)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	router := newExportRouter(t, database)

	ownerID := seedUser(t, database, "101", "user")
	otherID := seedUser(t, database, "202", "user")
	ownerTok := tokenFor(t, ownerID, "user")
	otherTok := tokenFor(t, otherID, "user")

	// ---- a public, approved story with chapters ----
	var pubID int64
	if err := database.QueryRow(`
		INSERT INTO stories (slug, author_id, title, subtitle, byline, visibility, status, global_view, created_at, updated_at)
		VALUES ('my-story', ?, 'My Story', 'Sub', 'By', 'public', 'approved', '{"center":[0,20],"zoom":0.6,"pitch":0,"bearing":0}', datetime('now'), datetime('now'))
		RETURNING id`, ownerID).Scan(&pubID); err != nil {
		t.Fatalf("insert public story: %v", err)
	}
	insertChapter(t, database, pubID, 1, "Chapter One", "# Hello", "left", 0,
		`{"center":[-8.86,37.29],"zoom":11,"pitch":40,"bearing":0}`, "flyTo", 0,
		`[]`, `[]`, `{"id":1}`, "image", "external", "https://example.com/img.jpg", nil)
	insertChapter(t, database, pubID, 2, "Chapter Two", "Second", "center", 1,
		`{"center":[126.06,9.88],"zoom":10}`, "easeTo", 1,
		``, ``, ``, "none", "none", "", nil)

	// ---- a private story owned by ownerID ----
	var privID int64
	if err := database.QueryRow(`
		INSERT INTO stories (slug, author_id, title, subtitle, byline, visibility, status, global_view, created_at, updated_at)
		VALUES ('private-story', ?, 'Private', '', '', 'private', 'draft', NULL, datetime('now'), datetime('now'))
		RETURNING id`, ownerID).Scan(&privID); err != nil {
		t.Fatalf("insert private story: %v", err)
	}

	// ---- public story exportable anonymously: 200 + headers + golden shape ----
	rec := doReq(t, router, "GET", "/api/stories/my-story/export", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("anon export public story status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("export Content-Type = %q, want application/json", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); cd != `attachment; filename="my-story.storymap.json"` {
		t.Fatalf("export Content-Disposition = %q, want attachment; filename=\"my-story.storymap.json\"", cd)
	}

	// The body must deep-equal the golden legacy shape.
	golden, err := os.ReadFile("_test/story_view.golden.json")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var got, want interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal export body: %v", err)
	}
	if err := json.Unmarshal(golden, &want); err != nil {
		t.Fatalf("unmarshal golden: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("export body mismatch:\n got: %s\nwant: %s", rec.Body.String(), golden)
	}

	// ---- private story exported by a non-owner → 403 ----
	rec = doReq(t, router, "GET", "/api/stories/private-story/export", "", otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner export private story status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	// ---- private story exported anonymously → 403 ----
	rec = doReq(t, router, "GET", "/api/stories/private-story/export", "", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("anon export private story status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	// ---- owner can export their own private story → 200 ----
	rec = doReq(t, router, "GET", "/api/stories/private-story/export", "", ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner export private story status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if cd := rec.Header().Get("Content-Disposition"); cd != `attachment; filename="private-story.storymap.json"` {
		t.Fatalf("owner export Content-Disposition = %q, want attachment; filename=\"private-story.storymap.json\"", cd)
	}

	// ---- export by numeric id also works ----
	rec = doReq(t, router, "GET", "/api/stories/"+int64str(pubID)+"/export", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("export by id status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	// ---- missing story → 404 ----
	rec = doReq(t, router, "GET", "/api/stories/nope/export", "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("export missing story status = %d, want 404; body=%s", rec.Code, rec.Body.String())
	}
}
