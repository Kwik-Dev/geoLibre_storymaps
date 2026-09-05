package auth_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

// newTestDB opens a fresh SQLite DB in a temp dir and runs the migrations.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	cfg := &config.Config{DBPath: filepath.Join(t.TempDir(), "test.db")}
	handle, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(handle); err != nil {
		handle.Close()
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { handle.Close() })
	return handle
}

// fakeOAuthServer returns an httptest server that answers the token endpoint
// (/access_token) with a fixed access token.
func fakeOAuthServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/access_token" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "gh_test_access_token"})
	}))
}

// fakeAPIServer fake an httptest server that answers /user with a fixed profile.
func fakeAPIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": 987654, "login": "octocat"})
	}))
}

func extractState(t *testing.T, loc string) string {
	t.Helper()
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse redirect Location %q: %v", loc, err)
	}
	s := u.Query().Get("state")
	if s == "" {
		t.Fatalf("redirect Location %q had no state param", loc)
	}
	return s
}

func userCount(t *testing.T, handle *sql.DB) int {
	t.Helper()
	var n int
	if err := handle.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		t.Fatalf("count users: %v", err)
	}
	return n
}

// TestGitHubOAuth verifies the full flow: authorize stores a single-use state,
// callback upserts exactly one user row keyed on github_id and issues a
// session, a replayed state is rejected (400), and a forged state is rejected
// (400).
func TestGitHubOAuth(t *testing.T) {
	tokenSrv := fakeOAuthServer()
	defer tokenSrv.Close()

	apiSrv := fakeAPIServer()
	defer apiSrv.Close()

	cfg := auth.GitHubConfig{
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
		OAuthBase:      tokenSrv.URL,
		APIBase:        apiSrv.URL,
		FrontendOrigin: "http://localhost:5173",
		JWTSecret:      "test-jwt-secret",
	}

	handle := newTestDB(t)
	h := auth.NewGitHubHandler(cfg, handle)

	// 1. authorize → 302 with a Location carrying a fresh state.
	authReq := httptest.NewRequest(http.MethodGet, "/api/auth/github", nil)
	authRec := httptest.NewRecorder()
	h.Authorize(authRec, authReq)

	if authRec.Code != http.StatusFound {
		t.Fatalf("authorize status = %d, want 302", authRec.Code)
	}
	state := extractState(t, authRec.Header().Get("Location"))

	// 2. callback with the valid state + code → 302, session cookie set,
	//    and exactly one user row upserted.
	cbReq := httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?state="+state+"&code=abc123", nil)
	cbRec := httptest.NewRecorder()
	h.Callback(cbRec, cbReq)

	if cbRec.Code != http.StatusFound {
		t.Fatalf("callback status = %d, want 302; body=%s", cbRec.Code, cbRec.Body.String())
	}
	if n := userCount(t, handle); n != 1 {
		t.Fatalf("after callback users = %d, want exactly 1", n)
	}
	var ghID, role string
	if err := handle.QueryRow(`SELECT github_id, role FROM users LIMIT 1`).Scan(&ghID, &role); err != nil {
		t.Fatalf("read user row: %v", err)
	}
	if ghID != "987654" {
		t.Fatalf("upserted github_id = %q, want 987654", ghID)
	}
	if role != "user" {
		t.Fatalf("upserted role = %q, want user", role)
	}

	// session issued: refresh cookie present
	found := false
	for _, c := range cbRec.Result().Cookies() {
		if c.Name == "refresh" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected a refresh session cookie to be issued")
	}

	// 3. replayed state → 400.
	replayRec := httptest.NewRecorder()
	h.Callback(replayRec, httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?state="+state+"&code=abc123", nil))
	if replayRec.Code != http.StatusBadRequest {
		t.Fatalf("replay status = %d, want 400", replayRec.Code)
	}

	// 4. forged state → 400.
	forgeRec := httptest.NewRecorder()
	h.Callback(forgeRec, httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?state=forged-state-123", nil))
	if forgeRec.Code != http.StatusBadRequest {
		t.Fatalf("forged status = %d, want 400", forgeRec.Code)
	}

	// Still exactly one user after all callbacks (no duplicate rows).
	if n := userCount(t, handle); n != 1 {
		t.Fatalf("users after replay/forged = %d, want 1", n)
	}
}
