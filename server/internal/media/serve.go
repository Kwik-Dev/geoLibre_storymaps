// Package media: P4.4 — serve uploaded media + visibility gating + soft-delete.
//
// This file implements:
//   - GET /media/:aid        — stream a stored asset from disk (never buffered
//     whole), gated by the visibility of the story(ies) that reference it.
//   - DELETE /api/media/:aid — soft-delete an asset (sets deleted_at);
//     owner/admin only. Serving skips soft-deleted assets (P7.3 purges them
//     later).
//
// Security model (feature_request §6, §10; HANDOUT §4/§6):
//   - stored_path is always a server-generated relative path (P4.1), so we
//     resolve it strictly under mediaDir and refuse any traversal.
//   - Random ids are NOT security: an asset may be referenced by a chapter, so
//     we map asset → referencing story(ies) → visibility and enforce authz.
//     A private story's asset requires the owner or an admin (else 403); a
//     public story's asset is served to anyone.
//   - The file is streamed with io.Copy, never loaded into memory.
package media

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
)

// mediaStory is the minimal story projection the media package needs to enforce
// visibility. Because the locked media_assets schema has NO owner column
// (HANDOUT §4), ownership/visibility is derived from the stories that reference
// an asset through live (non-deleted) chapters.
type mediaStory struct {
	ID         int64
	AuthorID   int64
	Visibility string
}

// mediaCanAccess mirrors api.canAccess (P3.1 HANDOFF): a public story is open
// to everyone; otherwise the caller must be the owner or an admin. It is
// re-declared here because media cannot import api (api imports media).
func mediaCanAccess(s mediaStory, user *auth.User) bool {
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

// MediaHandler serves and soft-deletes uploaded media. It needs the DB (to map
// asset → story visibility and to record soft-deletes), the media directory
// (root under which stored_path resolves), and an optional *auth.Authenticator
// for *optional* auth on the public GET /media route (so an owner/admin can
// reach a private asset's bytes). auther may be nil, in which case every media
// request is treated as anonymous.
type MediaHandler struct {
	db       *sql.DB
	mediaDir string
	auther   *auth.Authenticator
}

// NewMediaHandler builds a MediaHandler. mediaDir is the root under which
// stored_path relative paths are resolved. auther is used for optional
// authentication on the public serve route; it may be nil.
func NewMediaHandler(db *sql.DB, mediaDir string, auther *auth.Authenticator) *MediaHandler {
	return &MediaHandler{db: db, mediaDir: mediaDir, auther: auther}
}

// Serve handles GET /media/:aid. It looks up the asset, rejects soft-deleted
// assets (404), enforces the story-visibility gate, then streams the file.
func (h *MediaHandler) Serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	aid := chi.URLParam(r, "aid")
	id, err := strconv.ParseInt(aid, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "media asset not found"})
		return
	}

	var storedPath, mime string
	var deletedAt sql.NullString
	err = h.db.QueryRow(
		`SELECT stored_path, mime, deleted_at FROM media_assets WHERE id = ?`, id).
		Scan(&storedPath, &mime, &deletedAt)
	if err == sql.ErrNoRows || (err == nil && deletedAt.Valid) {
		// Missing OR soft-deleted: never served.
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "media asset not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read media asset"})
		return
	}

	// Visibility gate (asset → referencing story → visibility).
	user := h.optionalUser(r)
	allowed, err := h.canServe(id, user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check media access"})
		return
	}
	if !allowed {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	// Resolve stored_path strictly under mediaDir; never trust the value (it is
	// server-generated, but defense in depth against any traversal bug).
	if storedPath == "" || strings.Contains(storedPath, "..") ||
		strings.HasPrefix(storedPath, "/") || filepath.IsAbs(storedPath) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "media asset not found"})
		return
	}
	abs := filepath.Join(h.mediaDir, filepath.FromSlash(storedPath))

	f, err := os.Open(abs)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "media asset not found"})
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to stat media asset"})
		return
	}

	if mime != "" {
		w.Header().Set("Content-Type", mime)
	}
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	w.WriteHeader(http.StatusOK)
	// Stream — never buffer the whole file into memory.
	_, _ = io.Copy(w, f)
}

// Delete handles DELETE /api/media/:aid. It is a soft delete (sets deleted_at);
// only an admin or the author of a referencing story may delete. The route is
// mounted behind auth.RequireAuth (so a user is always present).
func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	user := auth.UserFrom(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	aid := chi.URLParam(r, "aid")
	id, err := strconv.ParseInt(aid, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "media asset not found"})
		return
	}

	// Must exist and not already be soft-deleted.
	var exists int
	if err := h.db.QueryRow(
		`SELECT COUNT(*) FROM media_assets WHERE id = ? AND deleted_at IS NULL`, id).Scan(&exists); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check media asset"})
		return
	}
	if exists == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "media asset not found"})
		return
	}

	allowed, err := h.canDelete(id, user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check media access"})
		return
	}
	if !allowed {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	if _, err := h.db.Exec(
		`UPDATE media_assets SET deleted_at = datetime('now') WHERE id = ?`, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete media asset"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// optionalUser returns the caller's identity for a public media route, or nil
// for an anonymous caller. It uses optional auth (never writes a 401); if the
// handler has no authenticator, everyone is treated as anonymous.
func (h *MediaHandler) optionalUser(r *http.Request) *auth.User {
	if h.auther == nil {
		return nil
	}
	return h.auther.UserFromRequest(r)
}

// storyRefs returns the live (non-deleted chapter + non-deleted story)
// references to an asset. An asset referenced by no live chapter is treated as
// a just-uploaded, unassociated asset.
func (h *MediaHandler) storyRefs(assetID int64) ([]mediaStory, error) {
	rows, err := h.db.Query(`
		SELECT s.id, s.author_id, s.visibility
		FROM stories s
		JOIN chapters c ON c.story_id = s.id
		WHERE c.media_asset_id = ? AND c.deleted_at IS NULL AND s.deleted_at IS NULL`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []mediaStory
	for rows.Next() {
		var s mediaStory
		if err := rows.Scan(&s.ID, &s.AuthorID, &s.Visibility); err != nil {
			return nil, err
		}
		refs = append(refs, s)
	}
	return refs, rows.Err()
}

// canServe reports whether the asset's bytes may be served to user. An asset
// referenced by a public story is open to anyone; an asset referenced only by
// private story(ies) requires the owner or an admin; an unassociated (not yet
// referenced) asset is served to anyone (it is the uploader's own, before it
// is tied to any story's visibility).
func (h *MediaHandler) canServe(assetID int64, user *auth.User) (bool, error) {
	refs, err := h.storyRefs(assetID)
	if err != nil {
		return false, err
	}
	if len(refs) == 0 {
		return true, nil
	}
	for _, s := range refs {
		if mediaCanAccess(s, user) {
			return true, nil
		}
	}
	return false, nil
}

// canDelete reports whether user may soft-delete the asset. Owner/admin means:
// the user is an admin, or the author of at least one story that references the
// asset. An asset referenced by no story is deletable only by an admin (there
// is no per-asset owner column to fall back on).
func (h *MediaHandler) canDelete(assetID int64, user *auth.User) (bool, error) {
	refs, err := h.storyRefs(assetID)
	if err != nil {
		return false, err
	}
	if user != nil && user.Role == "admin" {
		return true, nil
	}
	for _, s := range refs {
		if user != nil && user.ID == s.AuthorID {
			return true, nil
		}
	}
	return false, nil
}

