package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

// TestWhoamiRoute verifies GET /api/auth/whoami is actually served by the real
// router built in New() and sits behind RequireAuth:
//   - without a token → 401 (the endpoint requires auth)
//   - with a valid Bearer token → 200 with the user profile + admin flag
//   - the response never contains password_hash
//
// This guards against the endpoint existing only as dead code: it must be
// reachable through server.New()'s production mux.
func TestWhoamiRoute(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "whoami-route.db")
	database := openDB(t, dbPath)
	defer database.Close()
	defer os.Remove(dbPath)
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Seed a normal (non-admin) user row.
	if _, err := database.Exec(
		`INSERT INTO users (github_login, github_id, role, created_at)
		 VALUES ('octocat', '42', 'user', datetime('now'))`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	var uid int64
	if err := database.QueryRow(`SELECT id FROM users WHERE github_id='42'`).Scan(&uid); err != nil {
		t.Fatalf("read user id: %v", err)
	}

	const secret = "whoami-route-test-secret"
	cfg := &config.Config{JWTSecret: secret}
	admin := auth.NewAdminHandler(auth.GitHubConfig{JWTSecret: secret}, database)
	whoami := auth.NewWhoamiHandler(database)
	srv := New(cfg, database, admin, whoami)

	// 1. GET /api/auth/whoami without a token → 401 (behind RequireAuth).
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth/whoami", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("whoami (no token) via real router status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	// 2. With a valid Bearer token → 200, user object incl. role + admin flag,
	//    no password_hash.
	tok, err := auth.NewJWT(secret).Sign(auth.User{ID: uid, Role: "user"})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/auth/whoami", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("whoami (with token) via real server status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if strings.Contains(body, "password_hash") {
		t.Fatalf("whoami response leaked password_hash: %s", body)
	}
	var resp struct {
		ID          int64  `json:"id"`
		GithubLogin string `json:"github_login"`
		Role        string `json:"role"`
		Admin       bool   `json:"admin"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal whoami: %v", err)
	}
	if resp.ID != uid || resp.GithubLogin != "octocat" || resp.Role != "user" || resp.Admin {
		t.Fatalf("unexpected whoami body: %+v", resp)
	}
}
