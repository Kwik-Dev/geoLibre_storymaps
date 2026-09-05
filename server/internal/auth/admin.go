package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the cost factor used to hash the seeded admin password.
// It must be >= 10 per OWASP guidance.
const bcryptCost = 12

// ErrInvalidCredentials is returned when an admin email/password pair does not
// match the seeded admin row.
var ErrInvalidCredentials = errors.New("invalid admin credentials")

// AdminHandler implements the admin-only local (email/password) login flow.
// The admin user is seeded idempotently from ADMIN_EMAIL/ADMIN_PASSWORD at
// startup and is matched by admin_email. There is deliberately NO public
// self-registration: the only way to get an admin account is the env-seeded
// row (or a GitHub OAuth "user" account, which is never an admin).
type AdminHandler struct {
	cfg     GitHubConfig
	db      *sql.DB
	enabled bool // true once EnsureAdmin has seeded an admin from env
}

// NewAdminHandler builds an AdminHandler backed by db.
func NewAdminHandler(cfg GitHubConfig, db *sql.DB) *AdminHandler {
	return &AdminHandler{cfg: cfg, db: db}
}

// EnsureAdmin idempotently upserts the admin user from the environment. It
// bcrypt-hashes ADMIN_PASSWORD and upserts a users row matched by admin_email
// with role='admin', setting password_hash. If either ADMIN_EMAIL or
// ADMIN_PASSWORD is empty it is a no-op, so a pure GitHub-auth server stays
// fine. Safe to call repeatedly (restart-safe / idempotent). It never logs the
// password or its hash.
func (h *AdminHandler) EnsureAdmin(adminEmail, adminPassword string) error {
	if adminEmail == "" || adminPassword == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("bcrypt hash admin password: %w", err)
	}
	_, err = h.db.Exec(`
		INSERT INTO users (admin_email, password_hash, role, created_at)
		VALUES (?, ?, 'admin', datetime('now'))
		ON CONFLICT(admin_email) DO UPDATE SET
			password_hash = excluded.password_hash,
			role = 'admin'
	`, adminEmail, string(hash))
	if err != nil {
		return fmt.Errorf("upsert admin user: %w", err)
	}
	h.enabled = true
	return nil
}

// LoginAdmin verifies an admin email/password pair against the seeded row and
// returns the matching user. It is the HANDOFF for P2.2. Wrong credentials
// yield ErrInvalidCredentials (never a nil-user + nil-err pair).
func (h *AdminHandler) LoginAdmin(email, password string) (*User, error) {
	var u User
	var hash string
	err := h.db.QueryRow(`
		SELECT id, admin_email, role, password_hash FROM users
		WHERE admin_email = ? AND role = 'admin'
	`, email).Scan(&u.ID, &u.AdminEmail, &u.Role, &hash)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("lookup admin user: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return &u, nil
}

// Login serves POST /api/auth/admin/login. It parses {"email","password"},
// verifies the bcrypt hash, and issues an admin session (JWT + httpOnly
// cookie). Wrong credentials → 401; admin login not configured → 503 (no
// crash).
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.enabled {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "admin login not configured"})
		return
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Email == "" || body.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	user, err := h.LoginAdmin(body.Email, body.Password)
	if err == ErrInvalidCredentials {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "login failed"})
		return
	}

	token, err := IssueSession(w, h.cfg, *user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue session"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"role":  user.Role,
	})
}

// Refresh rotates the admin session: it validates the current refresh cookie
// and issues a fresh one. POST /api/auth/admin/refresh.
func (h *AdminHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if !h.enabled {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "admin login not configured"})
		return
	}

	cookie, err := r.Cookie("refresh")
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "no session"})
		return
	}

	userID, err := h.verifyAndUserID(cookie.Value)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var u User
	if err := h.db.QueryRow(`
		SELECT id, admin_email, role FROM users WHERE id = ? AND role = 'admin'
	`, userID).Scan(&u.ID, &u.AdminEmail, &u.Role); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	token, err := IssueSession(w, h.cfg, u)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to rotate session"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// verifyAndUserID validates an HS256 access token signed with the JWT secret
// and returns the sub (user id) claim. It is a lightweight inline verifier used
// by Refresh until the dedicated middleware (P2.3) lands.
func (h *AdminHandler) verifyAndUserID(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return 0, fmt.Errorf("missing sub claim")
	}
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid sub claim")
	}
	return id, nil
}
