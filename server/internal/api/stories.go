// Package api implements the HTTP handlers for user-created stories. It is the
// P3.x backend: stories CRUD (+ visibility/authorization) lives here, chapters
// and the legacy story-JSON adapter follow in later cards.
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"unicode"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
)

// Story is the camelCase-ish JSON shape returned for a stories row. It mirrors
// the relevant columns of the stories table; media/chapters are handled by
// later cards.
type Story struct {
	ID         int64  `json:"id"`
	Slug       string `json:"slug"`
	AuthorID   int64  `json:"author_id"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Byline     string `json:"byline"`
	Visibility string `json:"visibility"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// StoriesHandler serves /api/stories*. It reads authorization off the request
// context (attached by auth.RequireAuth) and enforces visibility via
// canAccess. The public GET listing is allowlisted by the middleware (anon
// users can list public+approved stories); everything else requires a session.
type StoriesHandler struct {
	db     *sql.DB
	auther *auth.Authenticator
}

// NewStoriesHandler builds a StoriesHandler backed by db. auther is used for
// optional authentication on the public list route (so an owner can see their
// own private stories); it may be nil, in which case the list is anonymous-only.
func NewStoriesHandler(db *sql.DB, auther *auth.Authenticator) *StoriesHandler {
	return &StoriesHandler{db: db, auther: auther}
}

// Routes registers the stories routes on the given router. It is meant to be
// mounted inside the /api group, which already applies auth.RequireAuth (the
// public GET /api/stories listing is allowlisted there).
func (h *StoriesHandler) Routes(r chi.Router) {
	r.Get("/stories", h.List)
	r.Post("/stories", h.Create)
	r.Get("/stories/{id}", h.Get)
	r.Put("/stories/{id}", h.Update)
	r.Delete("/stories/{id}", h.Delete)
}

// canAccess reports whether user (nil = anonymous) may view/modify the given
// story: public stories are open to everyone; otherwise only the owner or an
// admin. This is the P3.1 HANDOFF used by the chapters + export cards too.
func canAccess(s Story, user *auth.User) bool {
	if s.Visibility == "public" {
		return true
	}
	if user == nil {
		return false
	}
	if user.Role == "admin" {
		return true
	}
	return user.ID == s.AuthorID
}

// List serves GET /api/stories. It filters in SQL (never a Go post-filter):
//
//	anon      → visibility='public' AND status='approved'
//	owner     → their own stories (any visibility/status) plus the public ones
//	admin     → all non-deleted stories
func (h *StoriesHandler) List(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFrom(r.Context())
	// The public list route is allowlisted by the middleware, so it never
	// attaches a user. Use optional auth to learn the caller's identity when a
	// valid token is present (owner sees their own private stories too).
	if user == nil && h.auther != nil {
		user = h.auther.UserFromRequest(r)
	}

	query := `
		SELECT id, slug, author_id, title, subtitle, byline, visibility, status, created_at, updated_at
		FROM stories
		WHERE deleted_at IS NULL`
	var args []interface{}

	switch {
	case user == nil:
		query += ` AND visibility = 'public' AND status = 'approved'`
	case user.Role == "admin":
		// admin sees everything
	default:
		query += ` AND (visibility = 'public' AND status = 'approved' OR author_id = ?)`
		args = append(args, user.ID)
	}
	query += ` ORDER BY id DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list stories"})
		return
	}
	defer rows.Close()

	stories := []Story{}
	for rows.Next() {
		var s Story
		if err := rows.Scan(&s.ID, &s.Slug, &s.AuthorID, &s.Title, &s.Subtitle, &s.Byline,
			&s.Visibility, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read stories"})
			return
		}
		stories = append(stories, s)
	}
	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to iterate stories"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"stories": stories})
}

// Create serves POST /api/stories. It requires an authenticated session (the
// middleware 401s otherwise). title is required; subtitle/byline optional;
// visibility ∈ {private, public} (default private). A new story is created
// with status='draft' and author_id = the requesting user. A unique slug is
// generated from the title and guaranteed unique (case-insensitively).
func (h *StoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFrom(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		Title      string `json:"title"`
		Subtitle   string `json:"subtitle"`
		Byline     string `json:"byline"`
		Visibility string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	vis := strings.ToLower(strings.TrimSpace(body.Visibility))
	if vis == "" {
		vis = "private"
	}
	if vis != "private" && vis != "public" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "visibility must be 'private' or 'public'"})
		return
	}

	slug, err := h.uniqueSlug(slugify(body.Title))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate unique slug"})
		return
	}

	var id int64
	err = h.db.QueryRow(`
		INSERT INTO stories (slug, author_id, title, subtitle, byline, visibility, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'draft', datetime('now'), datetime('now'))
		RETURNING id`,
		slug, user.ID, strings.TrimSpace(body.Title), strings.TrimSpace(body.Subtitle),
		strings.TrimSpace(body.Byline), vis).Scan(&id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create story"})
		return
	}

	s, err := h.loadByID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read created story"})
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

// Get serves GET /api/stories/:id. It loads the story and applies canAccess:
// public → anyone; otherwise owner/admin only (403 for others).
func (h *StoriesHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFrom(r.Context())
	s, ok := h.load(w, r)
	if !ok {
		return
	}
	if !canAccess(s, user) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// Update serves PUT /api/stories/:id. Only the owner or an admin may update.
// It is a partial update: empty/omitted fields keep their current values.
func (h *StoriesHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFrom(r.Context())
	s, ok := h.load(w, r)
	if !ok {
		return
	}
	if !canAccess(s, user) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	var body struct {
		Title      *string `json:"title"`
		Subtitle   *string `json:"subtitle"`
		Byline     *string `json:"byline"`
		Visibility *string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	title := s.Title
	subtitle := s.Subtitle
	byline := s.Byline
	vis := s.Visibility
	if body.Title != nil {
		if strings.TrimSpace(*body.Title) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
			return
		}
		title = strings.TrimSpace(*body.Title)
	}
	if body.Subtitle != nil {
		subtitle = strings.TrimSpace(*body.Subtitle)
	}
	if body.Byline != nil {
		byline = strings.TrimSpace(*body.Byline)
	}
	if body.Visibility != nil {
		v := strings.ToLower(strings.TrimSpace(*body.Visibility))
		if v != "private" && v != "public" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "visibility must be 'private' or 'public'"})
			return
		}
		vis = v
	}

	if _, err := h.db.Exec(`
		UPDATE stories SET title = ?, subtitle = ?, byline = ?, visibility = ?, updated_at = datetime('now')
		WHERE id = ? AND deleted_at IS NULL`,
		title, subtitle, byline, vis, s.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update story"})
		return
	}

	updated, err := h.loadByID(s.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read updated story"})
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete serves DELETE /api/stories/:id. Only the owner or an admin may delete;
// it is a soft delete (sets deleted_at) — rows are never removed.
func (h *StoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFrom(r.Context())
	s, ok := h.load(w, r)
	if !ok {
		return
	}
	if !canAccess(s, user) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	if _, err := h.db.Exec(
		`UPDATE stories SET deleted_at = datetime('now') WHERE id = ?`, s.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete story"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// load parses the :id path param, loads the story, and writes a 404 if it is
// missing or soft-deleted. It reports whether the caller should continue.
func (h *StoriesHandler) load(w http.ResponseWriter, r *http.Request) (Story, bool) {
	id := chi.URLParam(r, "id")
	s, err := h.loadBySlugOrID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "story not found"})
		return Story{}, false
	}
	return s, true
}

func (h *StoriesHandler) loadByID(id int64) (Story, error) {
	return h.scanStory(h.db.QueryRow(`
		SELECT id, slug, author_id, title, subtitle, byline, visibility, status, created_at, updated_at
		FROM stories WHERE id = ? AND deleted_at IS NULL`, id))
}

func (h *StoriesHandler) loadBySlugOrID(id string) (Story, error) {
	return h.scanStory(h.db.QueryRow(`
		SELECT id, slug, author_id, title, subtitle, byline, visibility, status, created_at, updated_at
		FROM stories WHERE id = ? AND deleted_at IS NULL`, id))
}

func (h *StoriesHandler) scanStory(row *sql.Row) (Story, error) {
	var s Story
	err := row.Scan(&s.ID, &s.Slug, &s.AuthorID, &s.Title, &s.Subtitle, &s.Byline,
		&s.Visibility, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

// uniqueSlug returns base if it is unused (case-insensitively), otherwise
// appends a random numeric suffix until a free slug is found.
func (h *StoriesHandler) uniqueSlug(base string) (string, error) {
	for attempt := 0; attempt < 8; attempt++ {
		candidate := base
		if attempt > 0 {
			candidate = fmt.Sprintf("%s-%d", base, 100000+rand.Intn(900000))
		}
		var n int
		if err := h.db.QueryRow(
			`SELECT COUNT(*) FROM stories WHERE lower(slug) = lower(?)`, candidate).Scan(&n); err != nil {
			return "", err
		}
		if n == 0 {
			return candidate, nil
		}
	}
	return "", errors.New("could not generate a unique slug")
}

// slugify converts a title into a URL-safe lowercase slug: letters/digits are
// kept, runs of other characters become a single dash, and the result is
// trimmed. An all-punctuation title falls back to "story".
func slugify(s string) string {
	var b strings.Builder
	prevSep := true
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSep = false
		} else if !prevSep {
			b.WriteRune('-')
			prevSep = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "story"
	}
	return out
}

// writeJSON marshals v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
