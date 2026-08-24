package auth_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
)

const signTestSecret = "test-jwt-secret"

// signToken issues an HS256 token for sub/role with the given expiry, signed
// with signTest. It lets the test craft valid, expired, and invalid tokens.
func signToken(sub, role string, exp time.Time) string {
	claims := jwt.MapClaims{
		"sub":  sub,
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(signTestSecret))
	return s
}

// newMiddlewareRouter builds a chi router with RequireAuth mounted on /api and
// both public (allowlisted) and protected routes. It returns the router plus
// the Authenticator so the test can mint tokens.
func newMiddlewareRouter(t *testing.T) (*chi.Mux, *auth.Authenticator) {
	t.Helper()
	auther := auth.NewAuthenticator(signTestSecret, false)

	// protected handler that echoes the authenticated user id.
	protected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := auth.UserFrom(r.Context())
		if u == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(u.Role))
	})

	r := chi.NewRouter()
	r.Route("/api", func(api chi.Router) {
		api.Use(auther.RequireAuth)

		// public allowlisted paths
		api.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		api.Get("/auth/github", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		api.Get("/auth/github/callback", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		api.Post("/auth/admin/login", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		api.Post("/auth/admin/refresh", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		// public listing (GET)
		api.Get("/stories", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		// public export (GET) — authorisation to "when public" is the handler's job
		api.Get("/stories/{id}/export", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

		// protected write routes
		api.Post("/stories", protected)               // create a story
		api.Get("/stories/{id}", protected)           // get one story (authz by owner)
		api.Get("/auth/whoami", protected)            // current user
		api.Post("/private/write", protected)         // arbitrary protected write
	})
	return r, auther
}

// TestMiddleware verifies the auth middleware (P2.3):
//   - a protected route without a token → 401
//   - with a valid bearer token → 200
//   - a valid token in the refresh cookie → 200
//   - an expired token → 401
//   - an invalid/forged token → 401
//   - the public allowlist paths return without a token
//   - a private write route without a token → 401
func TestMiddleware(t *testing.T) {
	r, _ := newMiddlewareRouter(t)
	valid := signToken("42", "user", time.Now().Add(15*time.Minute))

	t.Run("protected route without token → 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stories/1", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("GET /api/stories/1 no-token status = %d, want 401", rec.Code)
		}
	})

	t.Run("private write without token → 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/stories", bytes.NewBufferString(`{"title":"x"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("POST /api/stories no-token status = %d, want 401", rec.Code)
		}
	})

	t.Run("valid bearer token → 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/whoami", nil)
		req.Header.Set("Authorization", "Bearer "+valid)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("whoami with valid token status = %d, want 200", rec.Code)
		}
	})

	t.Run("valid refresh cookie → 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/whoami", nil)
		req.AddCookie(&http.Cookie{Name: "refresh", Value: valid})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("whoami with refresh cookie status = %d, want 200", rec.Code)
		}
	})

	t.Run("expired token → 401", func(t *testing.T) {
		expired := signToken("good", "user", time.Now().Add(-1*time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/api/auth/whoami", nil)
		req.Header.Set("Authorization", "Bearer "+expired)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("whoami with expired token status = %d, want 401", rec.Code)
		}
	})

	t.Run("invalid/forged token → 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/whoami", nil)
		req.Header.Set("Authorization", "Bearer not.a.valid.token")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("whoami with forged token status = %d, want 401", rec.Code)
		}
	})

	t.Run("public allowlist paths return without token", func(t *testing.T) {
		cases := []struct {
			method string
			path   string
		}{
			{http.MethodGet, "/api/health"},
			{http.MethodGet, "/api/auth/github"},
			{http.MethodGet, "/api/auth/github/callback"},
			{http.MethodPost, "/api/auth/admin/login"},
			{http.MethodPost, "/api/auth/admin/refresh"},
			{http.MethodGet, "/api/stories"},            // public listing
			{http.MethodGet, "/api/stories/9/export"},   // public export path
		}
		for _, c := range cases {
			req := httptest.NewRequest(c.method, c.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("%s %s without token status = %d, want 200", c.method, c.path, rec.Code)
			}
		}
	})

	t.Run("user attached to context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/private/write", nil)
		req.Header.Set("Authorization", "Bearer "+valid)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("private write with token status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if rec.Body.String() != "user" {
			t.Fatalf("handler saw role = %q, want user", rec.Body.String())
		}
	})
}
