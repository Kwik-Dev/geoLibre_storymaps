package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is an unexported key type for request-scoped context values so no
// other package can accidentally collide with the middleware's value.
type contextKey int

const (
	ctxUser contextKey = iota
)

// UserFrom returns the authenticated user attached to the request context by
// RequireAuth, or nil if the request was not authenticated. Handlers use it to
// learn who is making the request.
//
//	func handle(w http.ResponseWriter, r *http.Request) {
//		u := auth.UserFrom(r.Context())
//		// u != nil ⇒ the request carried a valid token
//	}
func UserFrom(ctx context.Context) *User {
	u, _ := ctx.Value(ctxUser).(*User)
	return u
}

// Authenticator verifies bearer JWTs and the refresh cookie, and applies the
// public-route allowlist. Build one with NewAuthenticator.
type Authenticator struct {
	jwt    *JWT
	secure bool // whether the refresh cookie is marked Secure (production)
}

// NewAuthenticator builds an Authenticator that verifies HS256 tokens signed
// with the JWT secret. secure controls whether a refresh cookie it issues is
// marked Secure; in development over http://localhost it stays false, but in
// production (https) it must be true.
func NewAuthenticator(secret string, secure bool) *Authenticator {
	return &Authenticator{jwt: NewJWT(secret), secure: secure}
}

// RequireAuth is the authz middleware for /api/*. It lets the public allowlist
// through untouched and requires a valid token for everything else:
//
//  1. read `Authorization: Bearer <access>` first;
//  2. else fall back to a valid refresh token in the httpOnly `refresh` cookie;
//  3. verify the token; on failure return 401 {"error":"unauthorized"};
//  4. on success attach the user to the request context (UserFrom).
//
// The refresh cookie is never treated differently from a bearer token — both
// carry the same compact JWT, so the same verifier applies.
func (a *Authenticator) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.isPublic(r) {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := bearerToken(r)
		if tokenString == "" {
			// Fall back to the httpOnly refresh cookie.
			if c, err := r.Cookie("refresh"); err == nil {
				tokenString = c.Value
			}
		}
		if tokenString == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		claims, err := a.jwt.Parse(tokenString)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		user, err := UserFromClaims(claims)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		ctx := context.WithValue(r.Context(), ctxUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// isPublic reports whether the request hits an allowlisted /api path that is
// served without a token. The refresh-cookie routes are admin-only but still
// must be reachable before a session exists, so they are allowlisted too. The
// public story *listing* and *export* paths are method-aware so their protected
// write counterparts (POST /api/stories) still require a token.
func (a *Authenticator) isPublic(r *http.Request) bool {
	method := r.Method
	path := r.URL.Path

	switch {
	case method == http.MethodGet && path == "/api/health":
		return true
	case method == http.MethodGet && path == "/api/auth/github":
		return true
	case method == http.MethodGet && path == "/api/auth/github/callback":
		return true
	case method == http.MethodPost && path == "/api/auth/admin/login":
		return true
	case method == http.MethodPost && path == "/api/auth/admin/refresh":
		return true
	case method == http.MethodGet && path == "/api/stories":
		// Public listing — anonymous users see public stories only.
		return true
	case method == http.MethodGet && strings.HasPrefix(path, "/api/stories/") &&
		strings.HasSuffix(path, "/export"):
		// GET /api/stories/:id/export is public only when the target story is
		// public; the export handler (P3.4) enforces that check. The middleware
		// simply lets the route through so the handler can authorise it.
		return true
	}
	return false
}

// bearerToken extracts the access token from an Authorization: Bearer header.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && h[:len(prefix)] == prefix {
		return h[len(prefix):]
	}
	return ""
}
