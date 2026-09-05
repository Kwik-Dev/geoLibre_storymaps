package auth

import (
	"database/sql"
	"net/http"
)

// WhoamiHandler serves GET /api/auth/whoami for the authenticated user. It must
// be mounted behind RequireAuth so the request context carries the user; it
// never returns the password_hash (or any secret) in its response.
type WhoamiHandler struct {
	db *sql.DB
}

// NewWhoamiHandler builds a WhoamiHandler backed by db, used to look up the
// user's full profile (github_login, admin_email) by id.
func NewWhoamiHandler(db *sql.DB) *WhoamiHandler {
	return &WhoamiHandler{db: db}
}

// ServeHTTP returns the current user's public profile:
//
//	{"id", "github_login", "admin_email", "role", "admin": role=="admin"}
//
// If the request is not authenticated (no user in context) it returns 401. The
// id + role come straight from the authenticated session; github_login and
// admin_email are hydrated from the users row so the frontend can greet the
// user. password_hash is never read into the response.
func (h *WhoamiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	githubLogin := ""
	adminEmail := ""
	if h.db != nil {
		var gl, ae sql.NullString
		err := h.db.QueryRow(
			`SELECT github_login, admin_email FROM users WHERE id = ?`, u.ID,
		).Scan(&gl, &ae)
		if err == nil {
			githubLogin = gl.String
			adminEmail = ae.String
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           u.ID,
		"github_login": githubLogin,
		"admin_email":  adminEmail,
		"role":         u.Role,
		"admin":        u.Role == "admin",
	})
}
