// Package api implements the HTTP handlers for user-created stories. This file
// is the P3.3/P3.4 card: the DB → camelCase story JSON adapter (StoryView) and
// the legacy-JSON export endpoint (GET /api/stories/:id/export). StoryView
// converts a stories row plus its chapters into the exact legacy story JSON
// shape the frontend renderer consumes (the same shape as the embedded
// *-storymap.json files). It maps snake_case/int DB values to camelCase and
// omits empty media fields.
package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
)

// Defaults for legacy top-level fields that are not stored in the stories
// table. The DB schema keeps only the columns the API needs; the renderer
// expects the full legacy shape, so the adapter fills the rest with sensible
// defaults.
const (
	defaultStyle         = "https://tiles.openfreemap.org/styles/liberty"
	defaultTheme         = "dark"
	defaultInsetPosition = "bottom-left"
	defaultStartSlide    = "none"
	defaultEndSlide      = "none"
)

// defaultGlobalView is the fallback camera used when a story has no global_view.
var defaultGlobalView = json.RawMessage(`{"center":[0,20],"zoom":0.6,"pitch":0,"bearing":0}`)

// StoryViewDoc is the camelCase legacy story JSON shape (top level).
type StoryViewDoc struct {
	Title         string          `json:"title"`
	Subtitle      string          `json:"subtitle"`
	Byline        string          `json:"byline"`
	Footer        string          `json:"footer"`
	Theme         string          `json:"theme"`
	Style         string          `json:"style"`
	InsetWidth    int             `json:"insetWidth"`
	InsetHeight   int             `json:"insetHeight"`
	InsetPosition string          `json:"insetPosition"`
	GlobalView    json.RawMessage `json:"globalView"`
	StartSlide    string          `json:"startSlide"`
	EndSlide      string          `json:"endSlide"`
	Chapters      []ChapterView   `json:"chapters"`
}

// ChapterView is the camelCase legacy chapter shape. Media fields (image/video/
// audio) are omitted when empty, matching the legacy renderer's expectations.
type ChapterView struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Alignment       string          `json:"alignment"`
	Hidden          bool            `json:"hidden"`
	Location        *Location       `json:"location,omitempty"`
	MapAnimation    string          `json:"mapAnimation"`
	RotateAnimation bool            `json:"rotateAnimation"`
	OnChapterEnter  json.RawMessage `json:"onChapterEnter,omitempty"`
	OnChapterExit   json.RawMessage `json:"onChapterExit,omitempty"`
	Source          json.RawMessage `json:"source,omitempty"`
	Image           string          `json:"image,omitempty"`
	Video           string          `json:"video,omitempty"`
	Audio           string          `json:"audio,omitempty"`
	AutoPlayAudio   bool            `json:"autoPlayAudio"`
}

// StoryView converts a stories row and its chapters into the exact legacy
// story JSON shape the renderer consumes. It maps snake_case/int DB values to
// camelCase and omits empty media fields. This is the P3.3 HANDOFF.
func StoryView(s Story, chapters []Chapter, basePath string) any {
	doc := StoryViewDoc{
		Title:         s.Title,
		Subtitle:      s.Subtitle,
		Byline:        s.Byline,
		Footer:        "",
		Theme:         defaultTheme,
		Style:         defaultStyle,
		InsetWidth:    0,
		InsetHeight:   0,
		InsetPosition: defaultInsetPosition,
		GlobalView:    defaultGlobalView,
		StartSlide:    defaultStartSlide,
		EndSlide:      defaultEndSlide,
		Chapters:      make([]ChapterView, 0, len(chapters)),
	}
	for _, c := range chapters {
		doc.Chapters = append(doc.Chapters, chapterView(c, basePath))
	}
	return doc
}

// chapterView maps a DB chapter row to the legacy camelCase chapter shape.
func chapterView(c Chapter, basePath string) ChapterView {
	v := ChapterView{
		ID:              strconv.FormatInt(c.ID, 10),
		Title:           c.Title,
		Description:     c.DescriptionMD,
		Alignment:       c.Alignment,
		Hidden:          c.Hidden,
		MapAnimation:    c.MapAnimation,
		RotateAnimation: c.RotateAnimation,
		AutoPlayAudio:   false,
	}
	if c.Location != nil {
		v.Location = c.Location
	}
	if len(c.OnChapterEnter) > 0 {
		v.OnChapterEnter = c.OnChapterEnter
	}
	if len(c.OnChapterExit) > 0 {
		v.OnChapterExit = c.OnChapterExit
	}
	if c.Source != "" {
		v.Source = json.RawMessage(c.Source)
	}
	switch c.MediaType {
	case "image":
		v.Image = mediaURL(c, basePath)
	case "video":
		v.Video = mediaURL(c, basePath)
	case "audio":
		v.Audio = mediaURL(c, basePath)
	}
	return v
}

// ExportHandler serves GET /api/stories/:id/export — the user's "download my
// storymap" path. It returns the StoryView legacy JSON with a JSON content type
// and a Content-Disposition attachment filename derived from the story slug.
// Access is public when the story is public; otherwise only the owner or an
// admin (reusing canAccess). The route is allowlisted by the middleware, so the
// handler performs *optional* auth to learn the caller's identity and enforce
// the private-story check itself.
type ExportHandler struct {
	db       *sql.DB
	auther   *auth.Authenticator
	basePath string
}

// NewExportHandler builds an ExportHandler backed by db. auther is used for
// optional authentication (so an owner/admin can export a private story); it may
// be nil, in which case only public stories can be exported. basePath is the URL
// prefix the app is served under (e.g. "/maps"); "" = root.
func NewExportHandler(db *sql.DB, auther *auth.Authenticator, basePath string) *ExportHandler {
	return &ExportHandler{db: db, auther: auther, basePath: basePath}
}

// Routes registers the export route on the given router. It is meant to be
// mounted inside the /api group (which already applies auth.RequireAuth; the
// export path is allowlisted there so the handler can authorise it).
func (h *ExportHandler) Routes(r chi.Router) {
	r.Get("/stories/{id}/export", h.Export)
}

// Export serves GET /api/stories/:id/export. It loads the story (by id or
// slug), applies canAccess (public → anyone; else owner/admin), loads the
// chapters, and streams the StoryView JSON with a JSON content type and a
// Content-Disposition attachment filename derived from the slug.
func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFrom(r.Context())
	// The export route is allowlisted by the middleware, so it never attaches a
	// user. Use optional auth to learn the caller's identity when a valid token
	// is present (owner/admin can export a private story).
	if user == nil && h.auther != nil {
		user = h.auther.UserFromRequest(r)
	}

	id := chi.URLParam(r, "id")
	s, err := h.scanStory(h.db.QueryRow(`
		SELECT id, slug, author_id, title, subtitle, byline, visibility, status, created_at, updated_at
		FROM stories WHERE (id = ? OR slug = ?) AND deleted_at IS NULL`, id, id))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "story not found"})
		return
	}
	if !canAccess(s, user) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	chapters, err := h.loadChapters(s.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load chapters"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.storymap.json\"", s.Slug))
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(StoryView(s, chapters, h.basePath))
}

// scanStory scans a stories row into a Story.
func (h *ExportHandler) scanStory(row *sql.Row) (Story, error) {
	var s Story
	err := row.Scan(&s.ID, &s.Slug, &s.AuthorID, &s.Title, &s.Subtitle, &s.Byline,
		&s.Visibility, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

// loadChapters reads a story's chapters in render order (position, created_at).
func (h *ExportHandler) loadChapters(storyID int64) ([]Chapter, error) {
	rows, err := h.db.Query(`
		SELECT id, story_id, position, title, description_md, alignment, hidden, location,
		       map_animation, rotate_animation, on_chapter_enter, on_chapter_exit, source,
		       media_type, media_ref_type, media_external_url, media_asset_id, created_at, updated_at
		FROM chapters WHERE story_id = ? AND deleted_at IS NULL
		ORDER BY position, created_at`, storyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chapters := []Chapter{}
	for rows.Next() {
		c, err := scanChapter(rows)
		if err != nil {
			return nil, err
		}
		chapters = append(chapters, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chapters, nil
}

// mediaURL resolves a chapter's media to the URL the renderer loads directly.
// External refs use the stored URL; local refs point at the media serve path
// (P4.x serves /media/:aid), prefixed with the base path (if any) so the URL
// stays under the subpath the app is served from. Empty refs yield "" (omitted
// by omitempty).
func mediaURL(c Chapter, basePath string) string {
	switch c.MediaRefType {
	case "external":
		return c.MediaExternalURL
	case "local":
		if c.MediaAssetID != nil {
			return basePath + fmt.Sprintf("/media/%d", *c.MediaAssetID)
		}
	}
	return ""
}
