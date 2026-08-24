package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

const testSecret = "test-secret-for-p31"

// newTestRouter builds a production-shaped router: RequireAuth middleware + the
// stories routes. It returns the router and the token-signing JWT for seeding
// authenticated requests.
func newTestRouter(t *testing.T, database *sql.DB) *chi.Mux {
	t.Helper()
	r := chi.NewRouter()
	auther := auth.NewAuthenticator(testSecret, false)
	r.Route("/api", func(api chi.Router) {
		api.Use(auther.RequireAuth)
		NewStoriesHandler(database, auther).Routes(api)
	})
	return r
}

func openDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	d, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open(%q): %v", path, err)
	}
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := d.Exec(pragma); err != nil {
			t.Fatalf("%s: %v", pragma, err)
		}
	}
	return d
}

// seedUser inserts a users row and returns its id.
func seedUser(t *testing.T, database *sql.DB, githubID, role string) int64 {
	t.Helper()
	roleCol := "'user'"
	if role == "admin" {
		roleCol = "'admin'"
	}
	if _, err := database.Exec(
		`INSERT INTO users (github_login, github_id, role, created_at) VALUES (?, ?, `+roleCol+`, datetime('now'))`,
		"login-"+githubID, githubID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	var id int64
	if err := database.QueryRow(`SELECT id FROM users WHERE github_id = ?`, githubID).Scan(&id); err != nil {
		t.Fatalf("read user id: %v", err)
	}
	return id
}

func tokenFor(t *testing.T, id int64, role string) string {
	t.Helper()
	tok, err := auth.NewJWT(testSecret).Sign(auth.User{ID: id, Role: role})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return tok
}

func doReq(t *testing.T, r *chi.Mux, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func decodeStories(t *testing.T, rec *httptest.ResponseRecorder) []Story {
	t.Helper()
	var resp struct {
		Stories []Story `json:"stories"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal list response: %v; body=%s", err, rec.Body.String())
	}
	return resp.Stories
}

// TestStoriesCRUD exercises the P3.1 contract:
//   - anon list shows only public + approved stories
//   - an owner's own draft is visible to that owner
//   - a non-owner / non-admin GET of a private story → 403
//   - an admin sees all stories
//   - a soft-deleted story disappears from lists
func TestStoriesCRUD(t *testing.T) {
	dir := t.TempDir()
	database := openDB(t, filepath.Join(dir, "p31.db"))
	defer database.Close()
	defer os.RemoveAll(dir)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	router := newTestRouter(t, database)

	ownerID := seedUser(t, database, "101", "user")
	otherID := seedUser(t, database, "202", "user")
	adminID := seedUser(t, database, "303", "admin")

	ownerTok := tokenFor(t, ownerID, "user")
	otherTok := tokenFor(t, otherID, "user")
	adminTok := tokenFor(t, adminID, "admin")

	// ---- create: owner makes a private draft (default) ----
	rec := doReq(t, router, "POST", "/api/stories",
		`{"title":"My Private Story","subtitle":"s","byline":"b"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create private story status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var priv Story
	if err := json.Unmarshal(rec.Body.Bytes(), &priv); err != nil {
		t.Fatalf("unmarshal created story: %v", err)
	}
	if priv.Visibility != "private" || priv.Status != "draft" || priv.AuthorID != ownerID {
		t.Fatalf("created story wrong: %+v", priv)
	}
	if priv.Slug == "" {
		t.Fatalf("created story missing slug: %+v", priv)
	}

	// ---- create: owner makes a public story (approved manually below) ----
	rec = doReq(t, router, "POST", "/api/stories",
		`{"title":"Public Story","visibility":"public"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create public story status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var pub Story
	if err := json.Unmarshal(rec.Body.Bytes(), &pub); err != nil {
		t.Fatalf("unmarshal created story: %v", err)
	}
	if pub.Visibility != "public" {
		t.Fatalf("public story visibility = %q, want public", pub.Visibility)
	}
	// The public story must be approved before anon can list it.
	if _, err := database.Exec(
		`UPDATE stories SET status='approved' WHERE id = ?`, pub.ID); err != nil {
		t.Fatalf("approve public story: %v", err)
	}

	// ---- title is required ----
	rec = doReq(t, router, "POST", "/api/stories", `{"title":"   "}`, ownerTok)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create with blank title status = %d, want 400", rec.Code)
	}
	// ---- bad visibility ----
	rec = doReq(t, router, "POST", "/api/stories", `{"title":"x","visibility":"secret"}`, ownerTok)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create with bad visibility status = %d, want 400", rec.Code)
	}
	// ---- create without a token → 401 ----
	rec = doReq(t, router, "POST", "/api/stories", `{"title":"anon"}`, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("create without token status = %d, want 401", rec.Code)
	}

	// ---- anon list: only public+approved ----
	rec = doReq(t, router, "GET", "/api/stories", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("anon list status = %d, want 200", rec.Code)
	}
	anonList := decodeStories(t, rec)
	if len(anonList) != 1 || anonList[0].ID != pub.ID {
		t.Fatalf("anon list = %+v, want only the approved public story (id %d)", anonList, pub.ID)
	}

	// ---- owner list: own draft (private) + the public one ----
	rec = doReq(t, router, "GET", "/api/stories", "", ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner list status = %d, want 200", rec.Code)
	}
	ownerList := decodeStories(t, rec)
	if len(ownerList) != 2 {
		t.Fatalf("owner list len = %d, want 2 (own private draft + public); %+v", len(ownerList), ownerList)
	}

	// ---- owner GET own private story → 200 ----
	rec = doReq(t, router, "GET", "/api/stories/"+int64str(priv.ID), "", ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner GET private story status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	// ---- non-owner non-admin GET private story → 403 ----
	rec = doReq(t, router, "GET", "/api/stories/"+int64str(priv.ID), "", otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner GET private story status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	// ---- non-owner PUT private story → 403 ----
	rec = doReq(t, router, "PUT", "/api/stories/"+int64str(priv.ID), `{"title":"hijack"}`, otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner PUT private story status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	// ---- non-owner DELETE private story → 403 ----
	rec = doReq(t, router, "DELETE", "/api/stories/"+int64str(priv.ID), "", otherTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner DELETE private story status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	// ---- owner partial update (visibility → public) ----
	rec = doReq(t, router, "PUT", "/api/stories/"+int64str(priv.ID), `{"visibility":"public"}`, ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner update status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var upd Story
	if err := json.Unmarshal(rec.Body.Bytes(), &upd); err != nil {
		t.Fatalf("unmarshal updated story: %v", err)
	}
	if upd.Visibility != "public" || upd.Title != "My Private Story" {
		t.Fatalf("updated story = %+v, want visibility public + unchanged title", upd)
	}
	// Make it approved too, then confirm anon can now see it.
	if _, err := database.Exec(
		`UPDATE stories SET status='approved' WHERE id = ?`, priv.ID); err != nil {
		t.Fatalf("approve updated story: %v", err)
	}
	rec = doReq(t, router, "GET", "/api/stories", "", "")
	anonAfter := decodeStories(t, rec)
	if len(anonAfter) != 2 {
		t.Fatalf("anon list after public toggle = %+v, want 2 stories", anonAfter)
	}

	// ---- admin sees all ----
	rec = doReq(t, router, "GET", "/api/stories", "", adminTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin list status = %d, want 200", rec.Code)
	}
	if got := len(decodeStories(t, rec)); got != 2 {
		t.Fatalf("admin list len = %d, want 2", got)
	}

	// ---- soft-delete removes from lists ----
	rec = doReq(t, router, "DELETE", "/api/stories/"+int64str(pub.ID), "", ownerTok)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("owner soft-delete status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	rec = doReq(t, router, "GET", "/api/stories/"+int64str(pub.ID), "", adminTok)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET soft-deleted story status = %d, want 404 (soft delete hides it)", rec.Code)
	}
	rec = doReq(t, router, "GET", "/api/stories", "", "")
	if got := len(decodeStories(t, rec)); got != 1 {
		t.Fatalf("anon list after soft-delete = %d, want 1", got)
	}
	// The row must still exist (soft delete, not hard delete).
	var cnt int
	if err := database.QueryRow(`SELECT COUNT(*) FROM stories WHERE id = ?`, pub.ID).Scan(&cnt); err != nil {
		t.Fatalf("count story row: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("soft-deleted story row count = %d, want 1 (row must remain)", cnt)
	}

	// ---- slug uniqueness (case-insensitive) ----
	rec = doReq(t, router, "POST", "/api/stories", `{"title":"My Private Story"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create duplicate-title story status = %d, want 201 (unique slug generated); body=%s", rec.Code, rec.Body.String())
	}
	var dup Story
	if err := json.Unmarshal(rec.Body.Bytes(), &dup); err != nil {
		t.Fatalf("unmarshal duplicate story: %v", err)
	}
	if strings.EqualFold(dup.Slug, priv.Slug) {
		t.Fatalf("duplicate story slug %q collided with %q (case-insensitive unique index violated)", dup.Slug, priv.Slug)
	}
}

// int64str is a tiny helper for path building.
func int64str(i int64) string {
	return strconv.FormatInt(i, 10)
}
