package auth

import "net/http"

// LogoutHandler clears the httpOnly refresh cookie, ending the session. It is
// mounted as a public route (POST /api/auth/logout) so it always succeeds even
// when the cookie is stale or absent — clearing a non-existent session is a
// harmless no-op.
type LogoutHandler struct {
	secure bool // whether the refresh cookie was marked Secure (production)
}

// NewLogoutHandler builds a LogoutHandler. secure must match the Secure flag
// used when the cookie was issued (IssueSession) so the clear is effective.
func NewLogoutHandler(secure bool) *LogoutHandler {
	return &LogoutHandler{secure: secure}
}

// ServeHTTP clears the refresh cookie by setting MaxAge=-1 with the same
// attributes used at issue time, then returns 200.
func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   h.secure,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
