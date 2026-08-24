// Package api implements the HTTP handlers for user-created stories. This file
// is the P3.2 card: the nested chapters resource (CRUD + reorder) under
// /api/stories/:id/chapters. Every operation first loads the parent story and
// runs canAccess (P3.1 HANDOFF) so authorization is enforced on every route.
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
)

// Location is the validated JSONB shape of a chapter's camera position. It
// mirrors the legacy story JSON: center:[lng,lat], zoom, optional pitch/bearing.
type Location struct {
	Center  []float64 `json:"center"`
	Zoom    float64   `json:"zoom"`
	Pitch   *float64  `json:"pitch,omitempty"`
	Bearing *float64  `json:"bearing,omitempty"`
}

// Chapter is the camelCase-ish JSON shape returned for a chapters row. Media
// fields are carried through for later cards (P4.x) but not yet validated here.
type Chapter struct {
	ID               int64           `json:"id"`
	StoryID          int64           `json:"story_id"`
	Position         int             `json:"position"`
	Title            string          `json:"title"`
	DescriptionMD    string          `json:"description_md"`
	Alignment        string          `json:"alignment"`
	Hidden           bool            `json:"hidden"`
	Location         *Location       `json:"location,omitempty"`
	MapAnimation     string          `json:"map_animation"`
	RotateAnimation  bool            `json:"rotate_animation"`
	OnChapterEnter   json.RawMessage `json:"on_chapter_enter,omitempty"`
	OnChapterExit    json.RawMessage `json:"on_chapter_exit,omitempty"`
	Source           string          `json:"source"`
	MediaType        string          `json:"media_type"`
	MediaRefType     string          `json:"media_ref_type"`
	MediaExternalURL string          `json:"media_external_url,omitempty"`
	MediaAssetID     *int64          `json:"media_asset_id,omitempty"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}

// ChaptersHandler serves /api/stories/:id/chapters*. It loads the parent story
// and enforces canAccess on every operation (list, create, get, update,
// delete, reorder). It reuses the same DB handle as the stories handler.
type ChaptersHandler struct {
	db *sql.DB
}

// NewChaptersHandler builds a ChaptersHandler backed by db.
func NewChaptersHandler(db *sql.DB) *ChaptersHandler {
	return &ChaptersHandler{db: db}
}

// Routes registers the nested chapters routes on the given router. It is meant
// to be mounted inside the /api group (which already applies auth.RequireAuth).
func (h *ChaptersHandler) Routes(r chi.Router) {
	r.Route("/stories/{id}/chapters", func(cr chi.Router) {
		cr.Get("/", h.List)
		cr.Post("/", h.Create)
		cr.Post("/reorder", h.Reorder)
		cr.Get("/{cid}", h.Get)
		cr.Put("/{cid}", h.Update)
		cr.Delete("/{cid}", h.Delete)
	})
}

// loadStory parses the :id path param, loads the parent story (by id or slug),
// and applies canAccess. It writes the appropriate error response and returns
// false if the caller should stop. On success it returns the story and true.
func (h *ChaptersHandler) loadStory(w http.ResponseWriter, r *http.Request) (Story, bool) {
	user := auth.UserFrom(r.Context())
	id := chi.URLParam(r, "id")
	s, err := h.scanStory(h.db.QueryRow(`
		SELECT id, slug, author_id, title, subtitle, byline, visibility, status, created_at, updated_at
		FROM stories WHERE id = ? AND deleted_at IS NULL`, id))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "story not found"})
		return Story{}, false
	}
	if !canAccess(s, user) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return Story{}, false
	}
	return s, true
}

func (h *ChaptersHandler) scanStory(row *sql.Row) (Story, error) {
	var s Story
	err := row.Scan(&s.ID, &s.Slug, &s.AuthorID, &s.Title, &s.Subtitle, &s.Byline,
		&s.Visibility, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

// List serves GET /api/stories/:id/chapters, ordered by position then created_at.
func (h *ChaptersHandler) List(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.loadStory(w, r); !ok {
		return
	}
	storyID := chi.URLParam(r, "id")

	rows, err := h.db.Query(`
		SELECT id, story_id, position, title, description_md, alignment, hidden, location,
		       map_animation, rotate_animation, on_chapter_enter, on_chapter_exit, source,
		       media_type, media_ref_type, media_external_url, media_asset_id, created_at, updated_at
		FROM chapters
		WHERE story_id = ? AND deleted_at IS NULL
		ORDER BY position, created_at`, storyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list chapters"})
		return
	}
	defer rows.Close()

	chapters := []Chapter{}
	for rows.Next() {
		c, err := scanChapter(rows)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read chapters"})
			return
		}
		chapters = append(chapters, c)
	}
	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to iterate chapters"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"chapters": chapters})
}

// Create serves POST /api/stories/:id/chapters. position is auto-assigned as
// COALESCE(MAX(position),0)+1. title is required; description_md optional;
// alignment ∈ {left,center,right} (default center); location is validated JSONB.
func (h *ChaptersHandler) Create(w http.ResponseWriter, r *http.Request) {
	story, ok := h.loadStory(w, r)
	if !ok {
		return
	}

	var body struct {
		Title           string          `json:"title"`
		DescriptionMD   string          `json:"description_md"`
		Alignment       string          `json:"alignment"`
		Hidden          *bool           `json:"hidden"`
		Location        json.RawMessage `json:"location"`
		MapAnimation    string          `json:"map_animation"`
		RotateAnimation *bool           `json:"rotate_animation"`
		OnChapterEnter  json.RawMessage `json:"on_chapter_enter"`
		OnChapterExit   json.RawMessage `json:"on_chapter_exit"`
		Source          string          `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	alignment := strings.ToLower(strings.TrimSpace(body.Alignment))
	if alignment == "" {
		alignment = "center"
	}
	if alignment != "left" && alignment != "center" && alignment != "right" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "alignment must be 'left', 'center', or 'right'"})
		return
	}

	mapAnim := strings.ToLower(strings.TrimSpace(body.MapAnimation))
	if mapAnim == "" {
		mapAnim = "flyTo"
	}
	if mapAnim != "flyTo" && mapAnim != "easeTo" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "map_animation must be 'flyTo' or 'easeTo'"})
		return
	}

	hidden := false
	if body.Hidden != nil {
		hidden = *body.Hidden
	}
	rotate := false
	if body.RotateAnimation != nil {
		rotate = *body.RotateAnimation
	}

	// Validate location (if provided) before writing anything.
	var locJSON []byte
	if len(body.Location) > 0 && string(body.Location) != "null" {
		loc, err := validateLocation(body.Location)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		locJSON, err = json.Marshal(loc)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid location"})
			return
		}
	}

	var id int64
	err := h.db.QueryRow(`
		INSERT INTO chapters (story_id, position, title, description_md, alignment, hidden, location,
		                      map_animation, rotate_animation, on_chapter_enter, on_chapter_exit, source,
		                      created_at, updated_at)
		VALUES (?, COALESCE((SELECT MAX(position) FROM chapters WHERE story_id = ? AND deleted_at IS NULL), 0) + 1,
		        ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
		RETURNING id`,
		story.ID, story.ID,
		strings.TrimSpace(body.Title), strings.TrimSpace(body.DescriptionMD), alignment,
		boolToInt(hidden), nullableBytes(locJSON), mapAnim, boolToInt(rotate),
		nullableBytes(body.OnChapterEnter), nullableBytes(body.OnChapterExit),
		strings.TrimSpace(body.Source)).Scan(&id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create chapter"})
		return
	}

	c, err := h.loadByID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read created chapter"})
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

// Get serves GET /api/stories/:id/chapters/:cid.
func (h *ChaptersHandler) Get(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.loadStory(w, r); !ok {
		return
	}
	c, ok := h.loadChapter(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// Update serves PUT /api/stories/:id/chapters/:cid. It is a partial update:
// empty/omitted fields keep their current values. location, if provided, is
// validated.
func (h *ChaptersHandler) Update(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.loadStory(w, r); !ok {
		return
	}
	existing, ok := h.loadChapter(w, r)
	if !ok {
		return
	}

	var body struct {
		Title           *string         `json:"title"`
		DescriptionMD   *string         `json:"description_md"`
		Alignment       *string         `json:"alignment"`
		Hidden          *bool           `json:"hidden"`
		Location        json.RawMessage `json:"location"`
		MapAnimation    *string         `json:"map_animation"`
		RotateAnimation *bool           `json:"rotate_animation"`
		OnChapterEnter  json.RawMessage `json:"on_chapter_enter"`
		OnChapterExit   json.RawMessage `json:"on_chapter_exit"`
		Source          *string         `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	title := existing.Title
	desc := existing.DescriptionMD
	alignment := existing.Alignment
	hidden := existing.Hidden
	mapAnim := existing.MapAnimation
	rotate := existing.RotateAnimation
	source := existing.Source
	var locJSON []byte
	if existing.Location != nil {
		locJSON, _ = json.Marshal(existing.Location)
	}

	if body.Title != nil {
		if strings.TrimSpace(*body.Title) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
			return
		}
		title = strings.TrimSpace(*body.Title)
	}
	if body.DescriptionMD != nil {
		desc = strings.TrimSpace(*body.DescriptionMD)
	}
	if body.Alignment != nil {
		a := strings.ToLower(strings.TrimSpace(*body.Alignment))
		if a != "left" && a != "center" && a != "right" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "alignment must be 'left', 'center', or 'right'"})
			return
		}
		alignment = a
	}
	if body.Hidden != nil {
		hidden = *body.Hidden
	}
	if body.MapAnimation != nil {
		m := strings.ToLower(strings.TrimSpace(*body.MapAnimation))
		if m != "flyTo" && m != "easeTo" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "map_animation must be 'flyTo' or 'easeTo'"})
			return
		}
		mapAnim = m
	}
	if body.RotateAnimation != nil {
		rotate = *body.RotateAnimation
	}
	if body.Source != nil {
		source = strings.TrimSpace(*body.Source)
	}

	// location: if the field is present (even null), replace it. null clears it.
	if body.Location != nil {
		if string(body.Location) == "null" {
			locJSON = nil
		} else {
			loc, err := validateLocation(body.Location)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			locJSON, err = json.Marshal(loc)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid location"})
				return
			}
		}
	}

	// on_chapter_enter / on_chapter_exit: present (incl. null) replaces; absent keeps.
	enter := existing.OnChapterEnter
	exit := existing.OnChapterExit
	if body.OnChapterEnter != nil {
		enter = body.OnChapterEnter
	}
	if body.OnChapterExit != nil {
		exit = body.OnChapterExit
	}

	if _, err := h.db.Exec(`
		UPDATE chapters SET title = ?, description_md = ?, alignment = ?, hidden = ?, location = ?,
		       map_animation = ?, rotate_animation = ?, on_chapter_enter = ?, on_chapter_exit = ?,
		       source = ?, updated_at = datetime('now')
		WHERE id = ? AND deleted_at IS NULL`,
		title, desc, alignment, boolToInt(hidden), nullableBytes(locJSON), mapAnim,
		boolToInt(rotate), nullableBytes(enter), nullableBytes(exit), source, existing.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update chapter"})
		return
	}

	updated, err := h.loadByID(existing.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read updated chapter"})
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete serves DELETE /api/stories/:id/chapters/:cid. It is a soft delete.
func (h *ChaptersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.loadStory(w, r); !ok {
		return
	}
	c, ok := h.loadChapter(w, r)
	if !ok {
		return
	}
	if _, err := h.db.Exec(
		`UPDATE chapters SET deleted_at = datetime('now') WHERE id = ?`, c.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete chapter"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Reorder serves POST /api/stories/:id/chapters/reorder. The body is an array
// of {id, position}. It is applied atomically in a single transaction, and any
// id that is not a chapter of this story is rejected (all-or-nothing).
func (h *ChaptersHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	story, ok := h.loadStory(w, r)
	if !ok {
		return
	}

	var body []struct {
		ID       int64 `json:"id"`
		Position int   `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if len(body) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "reorder body must be a non-empty array"})
		return
	}

	// Validate every id belongs to this story before mutating anything.
	for _, item := range body {
		var n int
		if err := h.db.QueryRow(
			`SELECT COUNT(*) FROM chapters WHERE id = ? AND story_id = ? AND deleted_at IS NULL`,
			item.ID, story.ID).Scan(&n); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to validate reorder"})
			return
		}
		if n != 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "reorder contains an id that is not a chapter of this story"})
			return
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to begin reorder transaction"})
		return
	}
	defer tx.Rollback()

	for _, item := range body {
		if _, err := tx.Exec(
			`UPDATE chapters SET position = ?, updated_at = datetime('now') WHERE id = ?`,
			item.Position, item.ID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reorder chapters"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to commit reorder"})
		return
	}

	// Return the freshly ordered list.
	rows, err := h.db.Query(`
		SELECT id, story_id, position, title, description_md, alignment, hidden, location,
		       map_animation, rotate_animation, on_chapter_enter, on_chapter_exit, source,
		       media_type, media_ref_type, media_external_url, media_asset_id, created_at, updated_at
		FROM chapters
		WHERE story_id = ? AND deleted_at IS NULL
		ORDER BY position, created_at`, story.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list chapters"})
		return
	}
	defer rows.Close()

	chapters := []Chapter{}
	for rows.Next() {
		c, err := scanChapter(rows)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read chapters"})
			return
		}
		chapters = append(chapters, c)
	}
	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to iterate chapters"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"chapters": chapters})
}

// loadChapter parses the :cid path param, loads the chapter, and writes a 404 if
// it is missing or soft-deleted. It reports whether the caller should continue.
func (h *ChaptersHandler) loadChapter(w http.ResponseWriter, r *http.Request) (Chapter, bool) {
	cid := chi.URLParam(r, "cid")
	c, err := h.loadByID(cid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "chapter not found"})
		return Chapter{}, false
	}
	return c, true
}

func (h *ChaptersHandler) loadByID(id interface{}) (Chapter, error) {
	row := h.db.QueryRow(`
		SELECT id, story_id, position, title, description_md, alignment, hidden, location,
		       map_animation, rotate_animation, on_chapter_enter, on_chapter_exit, source,
		       media_type, media_ref_type, media_external_url, media_asset_id, created_at, updated_at
		FROM chapters WHERE id = ? AND deleted_at IS NULL`, id)
	return scanChapter(row)
}

// scanChapter scans a chapters row (from *sql.Row or *sql.Rows) into a Chapter.
func scanChapter(row interface{ Scan(...interface{}) error }) (Chapter, error) {
	var c Chapter
	var hidden, rotate int
	var locRaw, enterRaw, exitRaw []byte
	var mediaAssetID sql.NullInt64
	err := row.Scan(&c.ID, &c.StoryID, &c.Position, &c.Title, &c.DescriptionMD, &c.Alignment,
		&hidden, &locRaw, &c.MapAnimation, &rotate, &enterRaw, &exitRaw, &c.Source,
		&c.MediaType, &c.MediaRefType, &c.MediaExternalURL, &mediaAssetID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, err
	}
	c.Hidden = hidden != 0
	c.RotateAnimation = rotate != 0
	if len(locRaw) > 0 {
		var loc Location
		if json.Unmarshal(locRaw, &loc) == nil {
			c.Location = &loc
		}
	}
	if len(enterRaw) > 0 {
		c.OnChapterEnter = json.RawMessage(enterRaw)
	}
	if len(exitRaw) > 0 {
		c.OnChapterExit = json.RawMessage(exitRaw)
	}
	if mediaAssetID.Valid {
		id := mediaAssetID.Int64
		c.MediaAssetID = &id
	}
	return c, nil
}

// validateLocation validates a JSONB location object. It requires
// center:[lng,lat] with finite lng ∈ [-180,180] and lat ∈ [-90,90], a finite
// numeric zoom, and optional pitch ∈ [0,85] / bearing ∈ [0,360]. It rejects
// NaN/Infinite coordinates (which JSON cannot even encode, but a hand-built
// body could smuggle via a large exponent) and out-of-range values.
func validateLocation(raw json.RawMessage) (*Location, error) {
	var loc Location
	if err := json.Unmarshal(raw, &loc); err != nil {
		return nil, errors.New("invalid location: must be a JSON object")
	}
	if len(loc.Center) != 2 {
		return nil, errors.New("invalid location: center must be [lng, lat]")
	}
	lng, lat := loc.Center[0], loc.Center[1]
	if !isFinite(lng) || !isFinite(lat) {
		return nil, errors.New("invalid location: center coordinates must be finite")
	}
	if lng < -180 || lng > 180 {
		return nil, errors.New("invalid location: lng out of range [-180, 180]")
	}
	if lat < -90 || lat > 90 {
		return nil, errors.New("invalid location: lat out of range [-90, 90]")
	}
	if !isFinite(loc.Zoom) {
		return nil, errors.New("invalid location: zoom must be a finite number")
	}
	if loc.Pitch != nil && (!isFinite(*loc.Pitch) || *loc.Pitch < 0 || *loc.Pitch > 85) {
		return nil, errors.New("invalid location: pitch out of range [0, 85]")
	}
	if loc.Bearing != nil && (!isFinite(*loc.Bearing) || *loc.Bearing < 0 || *loc.Bearing > 360) {
		return nil, errors.New("invalid location: bearing out of range [0, 360]")
	}
	return &loc, nil
}

func isFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// nullableBytes returns a nil interface for an empty byte slice so it is stored
// as SQL NULL rather than an empty string.
func nullableBytes(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}
