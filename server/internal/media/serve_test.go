// Package media tests. P4.4 — serve + visibility gate + soft-delete.
//
// Contract under test (TestServeGate):
//   - a public story's asset → 200 for an anonymous client
//   - a private story's asset → 403 for an anonymous/other user, but streams
//     for the owner and an admin
//   - a soft-deleted asset is no longer served (404)
//   - DELETE /api/media/:aid soft-deletes; only the owner/admin may delete
package media

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

// insertUser adds a users row and returns its id.
func insertUser(t *testing.T, database *sql.DB, githubLogin string, role string) int64 {
	t.Helper()
	res, err := database.Exec(
		`INSERT INTO users (github_login, role, created_at) VALUES (?, ?, datetime('now'))`,
		githubLogin, role)
	if err != nil {
		t.Fatalf("insert user %s: %v", githubLogin, err)
	}
	id, _ := res.LastInsertId()
	return id
}

// insertStory adds a stories row and returns its id.
func insertStory(t *testing.T, database *sql.DB, authorID int64, slug, visibility string) int64 {
	t.Helper()
	res, err := database.Exec(`
		INSERT INTO stories (slug, author_id, title, visibility, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'approved', datetime('now'), datetime('now'))`,
		slug, authorID, "Test "+slug, visibility)
	if err != nil {
		t.Fatalf("insert story %s: %v", slug, err)
	}
	id, _ := res.LastInsertId()
	return id
}

// insertAsset writes a physical file under mediaDir and records a media_assets
// row referencing it. It returns the asset id. softDelete marks the row
// deleted_at (a soft-deleted asset).
func insertAsset(t *testing.T, database *sql.DB, mediaDir string, kind string, content []byte, softDelete bool) int64 {
	t.Helper()
	rel := fmt.Sprintf("2026-08/%s-%s", kind, "asset")
	abs := filepath.Join(mediaDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir asset dir: %v", err)
	}
	if err := os.WriteFile(abs, content, 0o644); err != nil {
		t.Fatalf("write asset file: %v", err)
	}
	deletedCol := "NULL"
	if softDelete {
		deletedCol = "datetime('now')"
	}
	res, err := database.Exec(fmt.Sprintf(`
		INSERT INTO media_assets (kind, stored_path, filename, bytes, mime, created_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), %s)`, deletedCol),
		kind, rel, filepath.Base(rel), len(content), "image/png")
	if err != nil {
		t.Fatalf("insert asset: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

// insertChapter attaches a chapter to a story that references a media asset.
func insertChapter(t *testing.T, database *sql.DB, storyID int64, position int, assetID int64) {
	t.Helper()
	if _, err := database.Exec(`
		INSERT INTO chapters (story_id, position, title, media_type, media_ref_type, media_asset_id, created_at, updated_at)
		VALUES (?, ?, ?, 'image', 'local', ?, datetime('now'), datetime('now'))`,
		storyID, position, "Chapter", assetID); err != nil {
		t.Fatalf("insert chapter: %v", err)
	}
}

// newServeRouter builds a production-shaped router with both the public
// GET /media/:aid serve route and the protected DELETE /api/media/:aid route.
func newServeRouter(t *testing.T, database *sql.DB, mediaDir string) (*chi.Mux, *auth.Authenticator) {
	t.Helper()
	auther := auth.NewAuthenticator(uploadTestSecret, false)
	mh := NewMediaHandler(database, mediaDir, auther)
	r := chi.NewRouter()
	r.Get("/media/{aid}", mh.Serve)
	r.Route("/api", func(api chi.Router) {
		api.Use(auther.RequireAuth)
		api.Delete("/media/{aid}", mh.Delete)
	})
	return r, auther
}

func doServe(t *testing.T, r *chi.Mux, aid string, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/media/"+aid, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func doDelete(t *testing.T, r *chi.Mux, aid string, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, "/api/media/"+aid, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// TestServeGate exercises the P4.4 contract end-to-end through the router.
func TestServeGate(t *testing.T) {
	dir := t.TempDir()
	defer os.RemoveAll(dir)

	database := openDBTest(t, filepath.Join(dir, "p44.db"))
	defer database.Close()
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	mediaDir := filepath.Join(dir, "media")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		t.Fatalf("mkdir mediaDir: %v", err)
	}

	ownerID := insertUser(t, database, "owner", "user")
	otherID := insertUser(t, database, "other", "user")
	_ = insertUser(t, database, "admin-user", "admin") // id 3 = admin

	// Stories:
	//  story 1 (public, author owner) → asset 1
	//  story 2 (private, author owner) → asset 2
	//  story 3 (private, author other) → asset 3
	//  story 1 also references asset 4, which is soft-deleted.
	storyPublic := insertStory(t, database, ownerID, "public-story", "public")
	storyPrivateOwner := insertStory(t, database, ownerID, "private-owner", "private")
	storyPrivateOther := insertStory(t, database, otherID, "private-other", "private")

	content := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG magic
	assetPublic := insertAsset(t, database, mediaDir, "image", content, false)
	assetPrivateOwner := insertAsset(t, database, mediaDir, "image", content, false)
	assetPrivateOther := insertAsset(t, database, mediaDir, "image", content, false)
	assetSoftDeleted := insertAsset(t, database, mediaDir, "image", content, true)

	insertChapter(t, database, storyPublic, 1, assetPublic)
	insertChapter(t, database, storyPrivateOwner, 1, assetPrivateOwner)
	insertChapter(t, database, storyPrivateOther, 1, assetPrivateOther)
	insertChapter(t, database, storyPublic, 2, assetSoftDeleted)

	router, _ := newServeRouter(t, database, mediaDir)
	ownerTok := uploadToken(t, ownerID, "user")
	otherTok := uploadToken(t, otherID, "user")
	adminTok := uploadToken(t, 3, "admin")

	t.Run("public-story-asset-served-to-anonymous", func(t *testing.T) {
		rec := doServe(t, router, fmt.Sprint(assetPublic), "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if got := rec.Body.Bytes(); string(got) != string(content) {
			t.Fatalf("served bytes mismatch: got %q want %q", got, content)
		}
	})

	t.Run("private-owner-asset-denied-anonymous", func(t *testing.T) {
		rec := doServe(t, router, fmt.Sprint(assetPrivateOwner), "")
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("private-other-asset-denied-other-user", func(t *testing.T) {
		// assetPrivateOther belongs to 'other's private story; the 'other' user
		// IS the owner so must get 200, while 'owner' is a foreign user → 403.
		recOwner := doServe(t, router, fmt.Sprint(assetPrivateOther), otherTok)
		if recOwner.Code != http.StatusOK {
			t.Fatalf("owner(other) status = %d, want 200; body=%s", recOwner.Code, recOwner.Body.String())
		}
		recForeign := doServe(t, router, fmt.Sprint(assetPrivateOther), ownerTok)
		if recForeign.Code != http.StatusForbidden {
			t.Fatalf("foreign(owner) status = %d, want 403; body=%s", recForeign.Code, recForeign.Body.String())
		}
	})

	t.Run("private-asset-streams-for-owner", func(t *testing.T) {
		rec := doServe(t, router, fmt.Sprint(assetPrivateOwner), ownerTok)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if got := rec.Body.Bytes(); string(got) != string(content) {
			t.Fatalf("served bytes mismatch: got %q want %q", got, content)
		}
	})

	t.Run("private-asset-streams-for-admin", func(t *testing.T) {
		rec := doServe(t, router, fmt.Sprint(assetPrivateOwner), adminTok)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("soft-deleted-asset-not-served", func(t *testing.T) {
		rec := doServe(t, router, fmt.Sprint(assetSoftDeleted), "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("nonexistent-asset-404", func(t *testing.T) {
		rec := doServe(t, router, "999999", "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("delete-private-asset-owner-only", func(t *testing.T) {
		// anonymous delete → 401 (behind RequireAuth)
		anon := doDelete(t, router, fmt.Sprint(assetPrivateOwner), "")
		if anon.Code != http.StatusUnauthorized {
			t.Fatalf("anon delete status = %d, want 401", anon.Code)
		}
		// foreign user (other) delete → 403
		foreign := doDelete(t, router, fmt.Sprint(assetPrivateOwner), otherTok)
		if foreign.Code != http.StatusForbidden {
			t.Fatalf("foreign delete status = %d, want 403", foreign.Code)
		}
		// owner delete → 204, then asset no longer served
		del := doDelete(t, router, fmt.Sprint(assetPrivateOwner), ownerTok)
		if del.Code != http.StatusNoContent {
			t.Fatalf("owner delete status = %d, want 204; body=%s", del.Code, del.Body.String())
		}
		after := doServe(t, router, fmt.Sprint(assetPrivateOwner), ownerTok)
		if after.Code != http.StatusNotFound {
			t.Fatalf("after-delete serve status = %d, want 404", after.Code)
		}
	})

	t.Run("delete-public-asset-foreign-denied-admin-allowed", func(t *testing.T) {
		// A public story's asset is still owned by its author: foreign user 403.
		foreign := doDelete(t, router, fmt.Sprint(assetPublic), otherTok)
		if foreign.Code != http.StatusForbidden {
			t.Fatalf("foreign delete of public asset status = %d, want 403", foreign.Code)
		}
		// admin can delete it.
		adm := doDelete(t, router, fmt.Sprint(assetPublic), adminTok)
		if adm.Code != http.StatusNoContent {
			t.Fatalf("admin delete status = %d, want 204; body=%s", adm.Code, adm.Body.String())
		}
	})
}
