package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ghUser is the GitHub API /user response subset we care about.
type ghUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

// GitHubHandler serves the GitHub OAuth2 authorize + callback endpoints.
// It owns a single-use state store so each `state` value can only be consumed
// once (this is the CSRF defense: a replayed or forged `state` is rejected).
type GitHubHandler struct {
	cfg    GitHubConfig
	db     *sql.DB
	states *stateStore
	client *http.Client
}

// NewGitHubHandler builds a GitHubHandler backed by db for the user upsert.
func NewGitHubHandler(cfg GitHubConfig, db *sql.DB) *GitHubHandler {
	return &GitHubHandler{
		cfg:    cfg,
		db:     db,
		states: newStateStore(stateTTL),
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// Authorize starts the OAuth flow: it generates a fresh random `state`, stores
// it as single-use, and redirects the browser to GitHub's authorize endpoint.
//
//	GET /api/auth/github
func (h *GitHubHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	state := newRandomState()
	h.states.put(state)

	redirectURI := h.redirectURI()
	authURL := fmt.Sprintf(
		"%s/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		h.cfg.OAuthBase,
		url.QueryEscape(h.cfg.ClientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape("read:user"),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// Callback completes the OAuth flow. It:
//  1. verifies + consumes the single-use `state` (replay/forged → 400),
//  2. exchanges the `code` for an access token at the token endpoint,
//  3. fetches the GitHub user profile,
//  4. upserts the users row by `github_id` (role=user, admin_email left null),
//  5. issues a session (JWT + httpOnly cookie) and redirects to the frontend.
//
//	GET /api/auth/github/callback?code=...&state=...
func (h *GitHubHandler) Callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	state := q.Get("state")
	code := q.Get("code")

	// CSRF gate: consume the state. A replayed or forged value is rejected.
	if !h.states.consume(state) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired state"})
		return
	}

	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code"})
		return
	}

	accessToken, err := h.exchangeCode(code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token exchange failed"})
		return
	}

	ghUser, err := h.fetchUser(accessToken)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch GitHub user"})
		return
	}

	user, err := h.upsertUser(ghUser)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to upsert user"})
		return
	}

	if _, err := IssueSession(w, h.cfg, user); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue session"})
		return
	}

	// Redirect to the frontend origin, carrying the user id in the hash.
	dest := strings.TrimRight(h.cfg.FrontendOrigin, "/") + "/#/"
	if user.ID > 0 {
		dest += fmt.Sprintf("user/%d", user.ID)
	}
	http.Redirect(w, r, dest, http.StatusFound)
}

// redirectURI is the callback URL GitHub redirects the browser back to.
func (h *GitHubHandler) redirectURI() string {
	return strings.TrimRight(h.cfg.FrontendOrigin, "/") + "/api/auth/github/callback"
}

// exchangeCode POSTs the authorization code to the GitHub token endpoint and
// returns the access token. The client secret is sent server-side only and is
// never logged.
func (h *GitHubHandler) exchangeCode(code string) (string, error) {
	form := url.Values{}
	form.Set("client_id", h.cfg.ClientID)
	form.Set("client_secret", h.cfg.ClientSecret)
	form.Set("code", code)

	req, err := http.NewRequest(http.MethodPost, h.cfg.OAuthBase+"/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	var tr struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", err
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("no access_token in token response")
	}
	return tr.AccessToken, nil
}

// fetchUser retrieves the authenticated user's GitHub profile.
func (h *GitHubHandler) fetchUser(accessToken string) (ghUser, error) {
	req, err := http.NewRequest(http.MethodGet, h.cfg.APIBase+"/user", nil)
	if err != nil {
		return ghUser{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return ghUser{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return ghUser{}, err
	}

	var u ghUser
	if err := json.Unmarshal(body, &u); err != nil {
		return ghUser{}, err
	}
	return u, nil
}

// upsertUser creates or updates the users row keyed on the GitHub numeric id.
// A GitHub account always gets role='user' and admin_email stays NULL.
func (h *GitHubHandler) upsertUser(gu ghUser) (User, error) {
	githubID := strconv.FormatInt(gu.ID, 10)

	var id int64
	err := h.db.QueryRow(`
		INSERT INTO users (github_login, github_id, role, created_at)
		VALUES (?, ?, 'user', datetime('now'))
		ON CONFLICT(github_id) DO UPDATE SET
			github_login = excluded.github_login
		RETURNING id
	`, gu.Login, githubID).Scan(&id)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:          id,
		GithubLogin: gu.Login,
		GithubID:    githubID,
		Role:        "user",
	}, nil
}
