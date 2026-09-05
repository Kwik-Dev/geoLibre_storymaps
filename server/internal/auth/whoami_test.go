package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
)

// TestWhoami verifies GET /api/auth/whoami (P2.4):
//   - with a valid token → 200 with the user object incl. role and admin flag
//   - without a token → 401
//   - the response never contains password_hash (secrets/hashes redacted)
func TestWhoami(t *testing.T) {
	handle := newTestDB(t)

	// seed a normal user row
	if _, err := handle.Exec(
		`INSERT INTO users (github_login, github_id, role, created_at)
		 VALUES ('octocat', '42', 'user', datetime('now'))`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	var uid int64
	if err := handle.QueryRow(`SELECT id FROM users WHERE github_id='42'`).Scan(&uid); err != nil {
		t.Fatalf("read user id: %v", err)
	}

	auther := auth.NewAuthenticator(signTestSecret, false)
	who := auth.NewWhoamiHandler(handle)
	mux := chi.NewRouter()
	mux.Route("/api", func(r chi.Router) {
		r.Use(auther.RequireAuth)
		r.Get("/auth/whoami", who.ServeHTTP)
	})

	// 1. without a token → 401
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth/whoami", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("whoami without token status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	// 2. with a valid token → 200, user object incl. role + admin, no secrets
	token := signToken(strconv.FormatInt(uid, 10), "user", time.Now().Add(15*time.Minute))
	req := httptest.NewRequest(http.MethodGet, "/api/auth/whoami", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("whoami with token status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if strings.Contains(body, "password_hash") || strings.Contains(body, "PasswordHash") {
		t.Fatalf("whoami response leaked password_hash: %s", body)
	}

	var resp struct {
		ID          int64  `json:"id"`
		GithubLogin string `json:"github_login"`
		AdminEmail  string `json:"admin_email"`
		Role        string `json:"role"`
		Admin       bool   `json:"admin"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal whoami response: %v", err)
	}
	if resp.ID != uid {
		t.Fatalf("whoami id = %d, want %d", resp.ID, uid)
	}
	if resp.GithubLogin != "octocat" {
		t.Fatalf("whoami github_login = %q, want octocat", resp.GithubLogin)
	}
	if resp.Role != "user" {
		t.Fatalf("whoami role = %q, want user", resp.Role)
	}
	if resp.Admin {
		t.Fatalf("whoami admin = true, want false for a 'user' role")
	}
}

// TestWhoamiAdminFlag verifies the admin flag derives from role (admin → true).
func TestWhoamiAdminFlag(t *testing.T) {
	handle := newTestDB(t)
	if _, err := handle.Exec(
		`INSERT INTO users (admin_email, role, created_at) VALUES ('admin@example.com', 'admin', datetime('now'))`); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	var uid int64
	if err := handle.QueryRow(`SELECT id FROM users WHERE admin_email='admin@example.com'`).Scan(&uid); err != nil {
		t.Fatalf("read admin id: %v", err)
	}

	auther := auth.NewAuthenticator(signTestSecret, false)
	who := auth.NewWhoamiHandler(handle)
	mux := chi.NewRouter()
	mux.Route("/api", func(r chi.Router) {
		r.Use(auther.RequireAuth)
		r.Get("/auth/whoami", who.ServeHTTP)
	})

	token := signToken(strconv.FormatInt(uid, 10), "admin", time.Now().Add(15*time.Minute))
	req := httptest.NewRequest(http.MethodGet, "/api/auth/whoami", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("whoami admin status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		AdminEmail string `json:"admin_email"`
		Role       string `json:"role"`
		Admin      bool   `json:"admin"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal admin whoami: %v", err)
	}
	if resp.Role != "admin" {
		t.Fatalf("admin whoami role = %q, want admin", resp.Role)
	}
	if !resp.Admin {
		t.Fatalf("admin whoami admin = false, want true (flag derived from role)")
	}
	if resp.AdminEmail != "admin@example.com" {
		t.Fatalf("admin whoami admin_email = %q, want admin@example.com", resp.AdminEmail)
	}
}
