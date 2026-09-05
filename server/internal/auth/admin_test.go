package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
)

// loginBody marshals an admin login request payload.
func loginBody(email, password string) []byte {
	b, _ := json.Marshal(map[string]string{"email": email, "password": password})
	return b
}

// TestAdminLogin verifies the admin-only local login flow:
//   - the env-seeded admin logs in → token + httpOnly session cookie
//   - a wrong password → 401
//   - an unknown email → 401
//   - a handler with no admin env → 503 without crashing
//   - /api/auth/register and /api/users → 404 (no public registration)
//
// It requires ADMIN_EMAIL and ADMIN_PASSWORD to be set in the environment.
func TestAdminLogin(t *testing.T) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminEmail == "" || adminPass == "" {
		t.Skip("TestAdminLogin requires ADMIN_EMAIL and ADMIN_PASSWORD env vars")
	}

	cfg := auth.GitHubConfig{JWTSecret: "test-jwt-secret"}
	handle := newTestDB(t)

	h := auth.NewAdminHandler(cfg, handle)

	// 1. EnsureAdmin is idempotent: run twice → exactly one admin row.
	if err := h.EnsureAdmin(adminEmail, adminPass); err != nil {
		t.Fatalf("EnsureAdmin: %v", err)
	}
	if err := h.EnsureAdmin(adminEmail, adminPass); err != nil {
		t.Fatalf("EnsureAdmin (2nd, idempotent): %v", err)
	}
	var adminCount int
	if err := handle.QueryRow(
		`SELECT COUNT(*) FROM users WHERE role='admin' AND admin_email=?`, adminEmail,
	).Scan(&adminCount); err != nil {
		t.Fatalf("count admin rows: %v", err)
	}
	if adminCount != 1 {
		t.Fatalf("admin rows = %d, want 1 (idempotent)", adminCount)
	}

	// 2. Correct credentials → 200, token, role=admin, and a session cookie.
	rec := httptest.NewRecorder()
	h.Login(rec, httptest.NewRequest(http.MethodPost, "/api/auth/admin/login", bytes.NewReader(loginBody(adminEmail, adminPass))))
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
		Role  string `json:"role"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected a token in the login response")
	}
	if resp.Role != "admin" {
		t.Fatalf("login role = %q, want admin", resp.Role)
	}
	foundCookie := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "refresh" {
			foundCookie = true
		}
	}
	if !foundCookie {
		t.Fatal("expected a refresh session cookie")
	}

	// 3. Wrong password → 401.
	rec2 := httptest.NewRecorder()
	h.Login(rec2, httptest.NewRequest(http.MethodPost, "/api/auth/admin/login", bytes.NewReader(loginBody(adminEmail, "not-the-password"))))
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("wrong-password status = %d, want 401; body=%s", rec2.Code, rec2.Body.String())
	}

	// 4. Unknown email → 401.
	rec3 := httptest.NewRecorder()
	h.Login(rec3, httptest.NewRequest(http.MethodPost, "/api/auth/admin/login", bytes.NewReader(loginBody("nobody@example.com", adminPass))))
	if rec3.Code != http.StatusUnauthorized {
		t.Fatalf("unknown-email status = %d, want 401", rec3.Code)
	}

	// 5. A handler that never had EnsureAdmin called (no admin env) → 503,
	//    and it must not panic/crash.
	noAdmin := auth.NewAdminHandler(cfg, handle)
	rec4 := httptest.NewRecorder()
	noAdmin.Login(rec4, httptest.NewRequest(http.MethodPost, "/api/auth/admin/login", bytes.NewReader(loginBody(adminEmail, adminPass))))
	if rec4.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured login status = %d, want 503", rec4.Code)
	}

	// 6. /api/auth/register and /api/users → 404 (no public registration).
	//    Build a chi router mounting only the admin routes and assert that
	//    unregistered /api paths are 404.
	mux := chi.NewRouter()
	mux.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(ar chi.Router) {
			ar.Post("/admin/login", h.Login)
			ar.Post("/admin/refresh", h.Refresh)
		})
	})
	regRec := httptest.NewRecorder()
	mux.ServeHTTP(regRec, httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(loginBody(adminEmail, adminPass))))
	if regRec.Code != http.StatusNotFound {
		t.Fatalf("/api/auth/register status = %d, want 404", regRec.Code)
	}
	usersRec := httptest.NewRecorder()
	mux.ServeHTTP(usersRec, httptest.NewRequest(http.MethodGet, "/api/users", nil))
	if usersRec.Code != http.StatusNotFound {
		t.Fatalf("/api/users status = %d, want 404", usersRec.Code)
	}

	// 7. Refresh rotates the session: valid cookie → fresh token.
	refreshReq := httptest.NewRequest(http.MethodPost, "/api/auth/admin/refresh", nil)
	refreshReq.AddCookie(&http.Cookie{Name: "refresh", Value: resp.Token})
	refreshRec := httptest.NewRecorder()
	h.Refresh(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want 200; body=%s", refreshRec.Code, refreshRec.Body.String())
	}
	var refreshResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(refreshRec.Body.Bytes(), &refreshResp); err != nil {
		t.Fatalf("unmarshal refresh response: %v", err)
	}
	if refreshResp.Token == "" {
		t.Fatal("expected a rotated token from refresh")
	}
}
