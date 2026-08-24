package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

// newTestRouterWithModeration builds a production-shaped router where
// moderationRequired is true: when a story is set to visibility=public, it
// is moved to status=pending instead of approved immediately.
func newTestRouterWithModeration(t *testing.T, database *sql.DB) *chi.Mux {
	t.Helper()
	r := chi.NewRouter()
	auther := auth.NewAuthenticator(testSecret, false)
	r.Route("/api", func(api chi.Router) {
		api.Use(auther.RequireAuth)
		NewStoriesHandler(database, auther, true).Routes(api)
	})
	return r
}

// TestModeration verifies the P7.2 moderation gate:
//   - publish → pending (hidden from the public list)
//   - admin approve → visible in the public list
//   - reject → hidden (reverts to draft)
//   - owner can still see their own pending story
func TestModeration(t *testing.T) {
	dir := t.TempDir()
	database := openDB(t, filepath.Join(dir, "mod.db"))
	defer database.Close()
	defer os.RemoveAll(dir)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	router := newTestRouterWithModeration(t, database)

	ownerID := seedUser(t, database, "owner1", "user")
	adminID := seedUser(t, database, "admin1", "admin")

	ownerTok := tokenFor(t, ownerID, "user")
	adminTok := tokenFor(t, adminID, "admin")

	// ---- owner creates a public story: should be pending (not approved) ----
	rec := doReq(t, router, "POST", "/api/stories",
		`{"title":"Needs Moderation","visibility":"public"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create public story status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var s Story
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("unmarshal created story: %v", err)
	}
	if s.Visibility != "public" {
		t.Fatalf("story visibility = %q, want public", s.Visibility)
	}
	if s.Status != "pending" {
		t.Fatalf("story status = %q, want 'pending' (moderation required); body=%s", s.Status, rec.Body.String())
	}

	// ---- anon public list: pending story must be hidden ----
	rec = doReq(t, router, "GET", "/api/stories", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("anon list status = %d, want 200", rec.Code)
	}
	anonList := decodeStories(t, rec)
	if len(anonList) != 0 {
		t.Fatalf("anon list = %+v, want empty (no approved stories yet)", anonList)
	}

	// ---- owner can still see their own pending story ----
	rec = doReq(t, router, "GET", "/api/stories", "", ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner list status = %d, want 200", rec.Code)
	}
	ownerList := decodeStories(t, rec)
	if len(ownerList) != 1 {
		t.Fatalf("owner list len = %d, want 1 (owner sees their pending story); %+v", len(ownerList), ownerList)
	}
	if ownerList[0].ID != s.ID || ownerList[0].Status != "pending" {
		t.Fatalf("owner list = %+v, want pending story id %d", ownerList, s.ID)
	}

	// ---- admin approves the story ----
	rec = doReq(t, router, "POST", "/api/stories/"+int64str(s.ID)+"/approve", "", adminTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var approved Story
	if err := json.Unmarshal(rec.Body.Bytes(), &approved); err != nil {
		t.Fatalf("unmarshal approved story: %v", err)
	}
	if approved.Status != "approved" {
		t.Fatalf("approved story status = %q, want 'approved'", approved.Status)
	}

	// ---- anon public list: approved story is now visible ----
	rec = doReq(t, router, "GET", "/api/stories", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("anon list after approve status = %d, want 200", rec.Code)
	}
	anonList = decodeStories(t, rec)
	if len(anonList) != 1 || anonList[0].ID != s.ID {
		t.Fatalf("anon list after approve = %+v, want story id %d visible", anonList, s.ID)
	}

	// ---- non-admin cannot approve ----
	rec = doReq(t, router, "POST", "/api/stories/"+int64str(s.ID)+"/approve", "", ownerTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-admin approve status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	// ---- create another pending story, then reject it ----
	rec = doReq(t, router, "POST", "/api/stories",
		`{"title":"Will Be Rejected","visibility":"public"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create second story status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var s2 Story
	if err := json.Unmarshal(rec.Body.Bytes(), &s2); err != nil {
		t.Fatalf("unmarshal second story: %v", err)
	}
	if s2.Status != "pending" {
		t.Fatalf("second story status = %q, want 'pending'", s2.Status)
	}

	// ---- admin rejects it ----
	rec = doReq(t, router, "POST", "/api/stories/"+int64str(s2.ID)+"/reject", "", adminTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("reject status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var rejected Story
	if err := json.Unmarshal(rec.Body.Bytes(), &rejected); err != nil {
		t.Fatalf("unmarshal rejected story: %v", err)
	}
	if rejected.Status != "draft" {
		t.Fatalf("rejected story status = %q, want 'draft'", rejected.Status)
	}

	// ---- anon public list: only the approved story is visible ----
	rec = doReq(t, router, "GET", "/api/stories", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("anon list after reject status = %d, want 200", rec.Code)
	}
	anonList = decodeStories(t, rec)
	if len(anonList) != 1 || anonList[0].ID != s.ID {
		t.Fatalf("anon list after reject = %+v, want only the approved story id %d", anonList, s.ID)
	}

	// ---- owner can still see their rejected story ----
	rec = doReq(t, router, "GET", "/api/stories", "", ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner list after reject status = %d, want 200", rec.Code)
	}
	ownerList = decodeStories(t, rec)
	if len(ownerList) != 2 {
		t.Fatalf("owner list after reject len = %d, want 2 (both stories visible to owner); %+v", len(ownerList), ownerList)
	}

	// ---- non-admin cannot reject ----
	rec = doReq(t, router, "POST", "/api/stories/"+int64str(s2.ID)+"/reject", "", ownerTok)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-admin reject status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	// ---- rejecting an already-rejected story (draft status) fails ----
	rec = doReq(t, router, "POST", "/api/stories/"+int64str(s2.ID)+"/reject", "", adminTok)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("reject non-pending story status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	// ---- update: changing a private story to public with moderation on should move to pending ----
	rec = doReq(t, router, "POST", "/api/stories",
		`{"title":"Private First","visibility":"private"}`, ownerTok)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create private story status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var priv Story
	if err := json.Unmarshal(rec.Body.Bytes(), &priv); err != nil {
		t.Fatalf("unmarshal private story: %v", err)
	}
	if priv.Status != "draft" {
		t.Fatalf("private story status = %q, want 'draft'", priv.Status)
	}

	// Change visibility to public — should become pending.
	rec = doReq(t, router, "PUT", "/api/stories/"+int64str(priv.ID),
		`{"visibility":"public"}`, ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("update private→public status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var updated Story
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unmarshal updated story: %v", err)
	}
	if updated.Visibility != "public" {
		t.Fatalf("updated visibility = %q, want 'public'", updated.Visibility)
	}
	if updated.Status != "pending" {
		t.Fatalf("updated story status = %q, want 'pending' after moderation gate", updated.Status)
	}

	// ---- update: changing a public story back to private does not affect moderation status ----
	rec = doReq(t, router, "PUT", "/api/stories/"+int64str(priv.ID),
		`{"visibility":"private"}`, ownerTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("update public→private status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	// Status stays pending (we don't auto-revert on private).
}
