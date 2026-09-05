package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

// newTestRouterChaptersMedia builds a production-shaped router with the nested
// chapters routes and a configured media external-URL allow-list (P4.3).
func newTestRouterChaptersMedia(t *testing.T, database *sql.DB, allowedHosts []string) *chi.Mux {
	t.Helper()
	r := chi.NewRouter()
	auther := auth.NewAuthenticator(testSecret, false)
	r.Route("/api", func(api chi.Router) {
		api.Use(auther.RequireAuth)
		NewStoriesHandler(database, auther, false).Routes(api)
		NewChaptersHandler(database).SetAllowedMediaHosts(allowedHosts).Routes(api)
	})
	return r
}

// seedAsset inserts a media_assets row and returns its id.
func seedAsset(t *testing.T, database *sql.DB, kind string) int64 {
	t.Helper()
	res, err := database.Exec(`
		INSERT INTO media_assets (kind, stored_path, filename, bytes, mime, created_at)
		VALUES (?, '2026-01/a.png', 'a.png', 10, 'image/png', datetime('now'))`, kind)
	if err != nil {
		t.Fatalf("seed asset: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("seed asset id: %v", err)
	}
	return id
}

// TestChapterMedia exercises the P4.3 contract:
//   - every valid media combo stores and round-trips
//   - each invalid combo → 400 with a clear code
//   - a local ref pointing at another user's private asset → 403
//   - making the owning story public makes the asset a public asset (usable)
//   - partial media updates re-derive + validate the full combo
func TestChapterMedia(t *testing.T) {
	dir := t.TempDir()
	database := openDB(t, filepath.Join(dir, "p43.db"))
	defer database.Close()
	defer os.RemoveAll(dir)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	router := newTestRouterChaptersMedia(t, database, []string{"cdn.example.com"})

	ownerID := seedUser(t, database, "101", "user")
	otherID := seedUser(t, database, "202", "user")
	ownerTok := tokenFor(t, ownerID, "user")
	otherTok := tokenFor(t, otherID, "user")

	createStory := func(tok string) int64 {
		rec := doReq(t, router, "POST", "/api/stories", `{"title":"Media Story"}`, tok)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create story status = %d, want 201; body=%s", rec.Code, rec.Body.String())
		}
		var st Story
		if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
			t.Fatalf("unmarshal story: %v", err)
		}
		return st.ID
	}
	createChapter := func(storyID int64, body string, tok string) *httptest.ResponseRecorder {
		return doReq(t, router, "POST", "/api/stories/"+int64str(storyID)+"/chapters", body, tok)
	}

	ownerStory := createStory(ownerTok)

	// ---- valid: media_ref_type=none (default) with media_type=none ----
	rec := createChapter(ownerStory, `{"title":"none","media_type":"none","media_ref_type":"none"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("none/none status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	// ---- valid: external URL (allowed host) ----
	rec = createChapter(ownerStory, `{"title":"ext","media_type":"image","media_ref_type":"external","media_external_url":"https://cdn.example.com/a.png"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("external allowed status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var ext Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &ext); err != nil {
		t.Fatalf("unmarshal external chapter: %v", err)
	}
	if ext.MediaType != "image" || ext.MediaRefType != "external" || ext.MediaExternalURL != "https://cdn.example.com/a.png" {
		t.Fatalf("external chapter round-trip wrong: %+v", ext)
	}

	// ---- valid: local asset (unassociated → usable), persisted ----
	assetA := seedAsset(t, database, "image")
	rec = createChapter(ownerStory, `{"title":"local","media_type":"image","media_ref_type":"local","media_asset_id":`+int64str(assetA)+`}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("local owner status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var local Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &local); err != nil {
		t.Fatalf("unmarshal local chapter: %v", err)
	}
	if local.MediaRefType != "local" || local.MediaAssetID == nil || *local.MediaAssetID != assetA {
		t.Fatalf("local chapter round-trip wrong: %+v", local)
	}

	// ---- invalid combos → 400 with a clear code ----
	invalid := []struct {
		name string
		body string
	}{
		{"external with no url", `{"media_type":"image","media_ref_type":"external"}`},
		{"external with http url", `{"media_type":"image","media_ref_type":"external","media_external_url":"http://cdn.example.com/a.png"}`},
		{"external disallowed host", `{"media_type":"image","media_ref_type":"external","media_external_url":"https://evil.com/a.png"}`},
		{"external with media_type none", `{"media_type":"none","media_ref_type":"external","media_external_url":"https://cdn.example.com/a.png"}`},
		{"local with media_type none", `{"media_type":"none","media_ref_type":"local","media_asset_id":` + int64str(assetA) + `}`},
		{"local with no asset id", `{"media_type":"image","media_ref_type":"local"}`},
		{"local with nonexistent asset", `{"media_type":"image","media_ref_type":"local","media_asset_id":99999}`},
		{"local with external url too", `{"media_type":"image","media_ref_type":"local","media_external_url":"https://cdn.example.com/a.png","media_asset_id":` + int64str(assetA) + `}`},
		{"none with external url", `{"media_type":"none","media_ref_type":"none","media_external_url":"https://cdn.example.com/a.png"}`},
		{"none with asset id", `{"media_type":"none","media_ref_type":"none","media_asset_id":` + int64str(assetA) + `}`},
		{"external with asset id", `{"media_type":"image","media_ref_type":"external","media_external_url":"https://cdn.example.com/a.png","media_asset_id":` + int64str(assetA) + `}`},
		{"unknown media_type", `{"media_type":"gif","media_ref_type":"none"}`},
		{"unknown ref_type", `{"media_type":"image","media_ref_type":"weird"}`},
	}
	for _, tc := range invalid {
		rec := createChapter(ownerStory, tc.body, ownerTok)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("invalid combo %q status = %d, want 400; body=%s", tc.name, rec.Code, rec.Body.String())
		}
	}

	// ---- other user references the owner's private asset → 403 ----
	// assetA is now referenced by owner's private story; other has no access.
	otherStory := createStory(otherTok)
	rec = createChapter(otherStory, `{"title":"x","media_type":"image","media_ref_type":"local","media_asset_id":`+int64str(assetA)+`}`, otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("other user referencing private asset status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	// ---- owner can still reference their own story's asset ----
	rec = createChapter(ownerStory, `{"title":"again","media_type":"image","media_ref_type":"local","media_asset_id":`+int64str(assetA)+`}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("owner re-reference status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	// ---- making the owning story public makes the asset a public asset ----
	rec = doReq(t, router, "PUT", "/api/stories/"+int64str(ownerStory), `{"visibility":"public"}`, ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("make public status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	rec = createChapter(otherStory, `{"title":"pub","media_type":"image","media_ref_type":"local","media_asset_id":`+int64str(assetA)+`}`, otherTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("other user referencing public asset status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	// ---- update media (grouped partial) ----
	// switch ext chapter to a video external URL
	rec = doReq(t, router, "PUT", "/api/stories/"+int64str(ownerStory)+"/chapters/"+int64str(ext.ID),
		`{"media_type":"video","media_ref_type":"external","media_external_url":"https://cdn.example.com/b.mp4"}`, ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("update external media status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var upd Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &upd); err != nil {
		t.Fatalf("unmarshal updated chapter: %v", err)
	}
	if upd.MediaType != "video" || upd.MediaExternalURL != "https://cdn.example.com/b.mp4" {
		t.Fatalf("update media round-trip wrong: %+v", upd)
	}

	// switch to a local ref (clear the external URL, point at owned asset)
	rec = doReq(t, router, "PUT", "/api/stories/"+int64str(ownerStory)+"/chapters/"+int64str(ext.ID),
		`{"media_ref_type":"local","media_external_url":"","media_asset_id":`+int64str(assetA)+`}`, ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("update to local status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	// switch to none clears external url + asset id
	rec = doReq(t, router, "PUT", "/api/stories/"+int64str(ownerStory)+"/chapters/"+int64str(ext.ID),
		`{"media_type":"none","media_ref_type":"none","media_asset_id":null}`, ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("update to none status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var none Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &none); err != nil {
		t.Fatalf("unmarshal none chapter: %v", err)
	}
	if none.MediaRefType != "none" || none.MediaAssetID != nil || none.MediaExternalURL != "" {
		t.Fatalf("update to none round-trip wrong: %+v", none)
	}
}
