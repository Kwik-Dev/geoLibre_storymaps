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

// newTestRouterWithChapters builds a production-shaped router that also mounts
// the nested chapters routes (P3.2).
func newTestRouterWithChapters(t *testing.T, database *sql.DB) *chi.Mux {
	t.Helper()
	r := chi.NewRouter()
	auther := auth.NewAuthenticator(testSecret, false)
	r.Route("/api", func(api chi.Router) {
		api.Use(auther.RequireAuth)
		NewStoriesHandler(database, auther, false).Routes(api)
		NewChaptersHandler(database).Routes(api)
	})
	return r
}

func decodeChapters(t *testing.T, rec *httptest.ResponseRecorder) []Chapter {
	t.Helper()
	var resp struct {
		Chapters []Chapter `json:"chapters"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal chapters response: %v; body=%s", err, rec.Body.String())
	}
	return resp.Chapters
}

// TestChapters exercises the P3.2 contract:
//   - adding 3 chapters auto-assigns positions 1,2,3
//   - a reorder call swaps order and persists
//   - an invalid location JSON → 400 (rejects non-finite / out-of-range coords)
//   - a non-owner on a private story → 403 for every chapter op
func TestChapters(t *testing.T) {
	dir := t.TempDir()
	database := openDB(t, filepath.Join(dir, "p32.db"))
	defer database.Close()
	defer os.RemoveAll(dir)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	router := newTestRouterWithChapters(t, database)

	ownerID := seedUser(t, database, "101", "user")
	otherID := seedUser(t, database, "202", "user")

	ownerTok := tokenFor(t, ownerID, "user")
	otherTok := tokenFor(t, otherID, "user")

	// ---- owner creates a private story ----
	rec := doReq(t, router, "POST", "/api/stories", `{"title":"Chapter Story"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create story status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var story Story
	if err := json.Unmarshal(rec.Body.Bytes(), &story); err != nil {
		t.Fatalf("unmarshal story: %v", err)
	}
	storyPath := "/api/stories/" + int64str(story.ID)

	// ---- non-owner on a private story → 403 for every chapter op ----
	// list
	rec = doReq(t, router, "GET", storyPath+"/chapters", "", otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner list chapters status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	// create
	rec = doReq(t, router, "POST", storyPath+"/chapters", `{"title":"x"}`, otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner create chapter status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	// reorder
	rec = doReq(t, router, "POST", storyPath+"/chapters/reorder", `[{"id":1,"position":1}]`, otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner reorder status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	// get/update/delete on a specific chapter (id 1 may not exist yet, but authz
	// must fire before the 404 lookup)
	rec = doReq(t, router, "GET", storyPath+"/chapters/1", "", otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner get chapter status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	rec = doReq(t, router, "PUT", storyPath+"/chapters/1", `{"title":"x"}`, otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner update chapter status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	rec = doReq(t, router, "DELETE", storyPath+"/chapters/1", "", otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner delete chapter status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	// ---- add 3 chapters → positions 1,2,3 ----
	validLoc := `{"center":[2.35,48.85],"zoom":12,"pitch":45,"bearing":30}`
	rec = doReq(t, router, "POST", storyPath+"/chapters",
		`{"title":"One","description_md":"first","alignment":"left","location":`+validLoc+`}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create chapter 1 status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var c1 Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &c1); err != nil {
		t.Fatalf("unmarshal chapter 1: %v", err)
	}
	if c1.Position != 1 {
		t.Fatalf("chapter 1 position = %d, want 1", c1.Position)
	}
	if c1.Alignment != "left" {
		t.Fatalf("chapter 1 alignment = %q, want left", c1.Alignment)
	}
	if c1.Location == nil || len(c1.Location.Center) != 2 || c1.Location.Center[0] != 2.35 {
		t.Fatalf("chapter 1 location not round-tripped: %+v", c1.Location)
	}

	rec = doReq(t, router, "POST", storyPath+"/chapters", `{"title":"Two"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create chapter 2 status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var c2 Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &c2); err != nil {
		t.Fatalf("unmarshal chapter 2: %v", err)
	}
	if c2.Position != 2 {
		t.Fatalf("chapter 2 position = %d, want 2", c2.Position)
	}
	if c2.Alignment != "center" {
		t.Fatalf("chapter 2 alignment = %q, want default center", c2.Alignment)
	}

	rec = doReq(t, router, "POST", storyPath+"/chapters", `{"title":"Three"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create chapter 3 status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var c3 Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &c3); err != nil {
		t.Fatalf("unmarshal chapter 3: %v", err)
	}
	if c3.Position != 3 {
		t.Fatalf("chapter 3 position = %d, want 3", c3.Position)
	}

	// ---- list is ordered by position ----
	rec = doReq(t, router, "GET", storyPath+"/chapters", "", ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("list chapters status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	list := decodeChapters(t, rec)
	if len(list) != 3 {
		t.Fatalf("list len = %d, want 3; %+v", len(list), list)
	}
	if list[0].ID != c1.ID || list[1].ID != c2.ID || list[2].ID != c3.ID {
		t.Fatalf("list order wrong: %+v", list)
	}

	// ---- reorder: swap 1 and 3, persist ----
	rec = doReq(t, router, "POST", storyPath+"/chapters/reorder",
		`[{"id":`+int64str(c3.ID)+`,"position":1},{"id":`+int64str(c2.ID)+`,"position":2},{"id":`+int64str(c1.ID)+`,"position":3}]`,
		ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("reorder status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	reordered := decodeChapters(t, rec)
	if len(reordered) != 3 {
		t.Fatalf("reorder list len = %d, want 3", len(reordered))
	}
	if reordered[0].ID != c3.ID || reordered[1].ID != c2.ID || reordered[2].ID != c1.ID {
		t.Fatalf("reorder did not swap order: %+v", reordered)
	}

	// ---- reorder persists across a fresh list ----
	rec = doReq(t, router, "GET", storyPath+"/chapters", "", ownerTok)
	persisted := decodeChapters(t, rec)
	if persisted[0].ID != c3.ID || persisted[2].ID != c1.ID {
		t.Fatalf("reorder did not persist: %+v", persisted)
	}

	// ---- reorder rejects an id that is not this story's chapter ----
	rec = doReq(t, router, "POST", storyPath+"/chapters/reorder",
		`[{"id":99999,"position":1}]`, ownerTok)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("reorder with foreign id status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	// ---- invalid location → 400 (non-finite / out-of-range) ----
	badLocs := []string{
		`{"center":[1e400,0],"zoom":1}`,   // lng non-finite (overflows to +Inf)
		`{"center":[0,1e400],"zoom":1}`,   // lat non-finite
		`{"center":[181,0],"zoom":1}`,     // lng out of range
		`{"center":[0,91],"zoom":1}`,      // lat out of range
		`{"center":[0,0],"zoom":1e400}`,   // zoom non-finite
		`{"center":[0,0],"zoom":1,"pitch":90}`,   // pitch out of range
		`{"center":[0,0],"zoom":1,"bearing":361}`, // bearing out of range
		`{"center":[0],"zoom":1}`,         // center not length 2
		`{"center":"nope","zoom":1}`,       // center not an array
	}
	for i, bad := range badLocs {
		rec = doReq(t, router, "POST", storyPath+"/chapters",
			`{"title":"bad","location":`+bad+`}`, ownerTok)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("bad location #%d (%s) status = %d, want 400; body=%s", i, bad, rec.Code, rec.Body.String())
		}
	}

	// ---- title is required ----
	rec = doReq(t, router, "POST", storyPath+"/chapters", `{"title":"   "}`, ownerTok)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create with blank title status = %d, want 400", rec.Code)
	}

	// ---- update a chapter (partial) ----
	rec = doReq(t, router, "PUT", storyPath+"/chapters/"+int64str(c2.ID), `{"title":"Two Updated"}`, ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("update chapter status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var upd Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &upd); err != nil {
		t.Fatalf("unmarshal updated chapter: %v", err)
	}
	if upd.Title != "Two Updated" || upd.Position != 2 {
		t.Fatalf("updated chapter = %+v, want title changed + position kept", upd)
	}

	// ---- soft-delete a chapter ----
	rec = doReq(t, router, "DELETE", storyPath+"/chapters/"+int64str(c1.ID), "", ownerTok)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete chapter status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	rec = doReq(t, router, "GET", storyPath+"/chapters/"+int64str(c1.ID), "", ownerTok)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET soft-deleted chapter status = %d, want 404", rec.Code)
	}
	rec = doReq(t, router, "GET", storyPath+"/chapters", "", ownerTok)
	if got := len(decodeChapters(t, rec)); got != 2 {
		t.Fatalf("list after soft-delete = %d, want 2", got)
	}
	// row must still exist (soft delete)
	var cnt int
	if err := database.QueryRow(`SELECT COUNT(*) FROM chapters WHERE id = ?`, c1.ID).Scan(&cnt); err != nil {
		t.Fatalf("count chapter row: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("soft-deleted chapter row count = %d, want 1", cnt)
	}
}
