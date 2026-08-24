package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultOAuthBase      = "https://github.com/login/oauth"
	defaultAPIBase        = "https://api.github.com"
	defaultFrontendOrigin = "http://localhost:5173"
	stateTTL              = 10 * time.Minute
)

// GitHubConfig holds all values the GitHub OAuth2 flow needs. Values are
// populated from environment variables (see GitHubConfigFromEnv) but can also
// be set directly for tests.
type GitHubConfig struct {
	ClientID       string
	ClientSecret   string
	OAuthBase      string // token/authorize base, default https://github.com/login/oauth
	APIBase        string // default https://api.github.com
	FrontendOrigin string // where to redirect the browser after login
	JWTSecret      string
}

// GitHubConfigFromEnv builds a GitHubConfig from the environment:
//   GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, GITHUB_OAUTH_BASE,
//   GITHUB_API_BASE, FRONTEND_ORIGIN, JWT_SECRET.
func GitHubConfigFromEnv() GitHubConfig {
	return GitHubConfig{
		ClientID:       os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret:   os.Getenv("GITHUB_CLIENT_SECRET"),
		OAuthBase:      getEnv("GITHUB_OAUTH_BASE", defaultOAuthBase),
		APIBase:        getEnv("GITHUB_API_BASE", defaultAPIBase),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", defaultFrontendOrigin),
		JWTSecret:      os.Getenv("JWT_SECRET"),
	}
}

// User is the subset of the users table row the auth flow needs to expose.
type User struct {
	ID          int64
	GithubLogin string
	GithubID    string
	Role        string
}

// stateStore is a small in-memory TTL map used to track 1×-use OAuth `state`
// values. Storing state in memory (rather than in the DB) is fine because a
// server restart simply invalidates all in-flight authorizations.
type stateStore struct {
	mu    sync.Mutex
	items map[string]time.Time // state -> expiry
	ttl   time.Duration
}

func newStateStore(ttl time.Duration) *stateStore {
	return &stateStore{items: make(map[string]time.Time), ttl: ttl}
}

// put records a new state value with its expiry, opportunistically evicting
// expired entries.
func (s *stateStore) put(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[state] = time.Now().Add(s.ttl)
	now := time.Now()
	for k, exp := range s.items {
		if now.After(exp) {
			delete(s.items, k)
		}
	}
}

// consume atomically removes a state value and reports whether it was present
// and unexpired. It is the single-use enforcement point: a value can only be
// consumed once, so a replay (reusing the same `state`) is rejected.
func (s *stateStore) consume(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.items[state]
	if !ok {
		return false
	}
	delete(s.items, state)
	return time.Now().Before(exp)
}

// newRandomState returns a cryptographically-random hex state string.
func newRandomState() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// IssueSession signs an HS256 access JWT for the user (15-minute expiry) and
// sets an httpOnly, SameSite=Strict refresh cookie on w. It returns the signed
// access token. This is the session handoff used by both the GitHub callback
// and (in later cards) the admin login.
func IssueSession(w http.ResponseWriter, cfg GitHubConfig, user User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  strconv.FormatInt(user.ID, 10),
		"role": user.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return signed, nil
}

// getEnv returns the value of key, or fallback if unset/empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// writeJSON writes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

