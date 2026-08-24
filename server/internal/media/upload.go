// Package media implements local-disk upload of chapter media for
// user-created storymaps. It is the P4.x backend: this file is P4.1, the
// multipart upload endpoint (POST /api/media/upload).
//
// Security model (see feature_request §6, §10):
//   - MIME is detected from **magic bytes**, never from Content-Type or the
//     client filename/extension.
//   - Only image / video / audio kinds are accepted; anything else (including
//     HTML smuggled as a .png) is rejected.
//   - The request body is capped with http.MaxBytesReader (413 on over-size)
//     and the file is streamed to disk chunk-by-chunk, never buffered whole.
//   - Files are stored under MEDIA_DIR/<YYYY-MM>/ with a **crypto/rand** hex
//     basename, so client-supplied paths can never cause traversal.
//   - stored_path is always a server-generated relative path.
package media

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	// defaultMaxBytes is the upload size cap when MEDIA_MAX_BYTES is unset.
	defaultMaxBytes = 25 * 1024 * 1024 // 25 MB
	// multipartOverhead is extra headroom on the http.MaxBytesReader cap to
	// cover multipart boundaries/envelope so the *file* may reach maxBytes.
	multipartOverhead = 1 << 20 // 1 MB
	// magicReadSize is how many leading bytes we sniff to detect MIME.
	magicReadSize = 512
)

// Asset is the result of a successful upload. It is also the JSON body
// returned by POST /api/media/upload.
type Asset struct {
	ID    int64  `json:"id"`
	URL   string `json:"url"`   // "/media/<id>"
	Bytes int64  `json:"bytes"` // bytes actually written to disk
	MIME  string `json:"mime"`

	Kind       string `json:"-"`
	Filename   string `json:"-"`
	StoredPath string `json:"-"` // server-generated, always relative
}

// StatusError is an HTTP error with an explicit status code and a message
// safe to return to the client.
type StatusError struct {
	Status  int
	Message string
}

func (e *StatusError) Error() string { return e.Message }

func statusError(status int, format string, args ...interface{}) *StatusError {
	return &StatusError{Status: status, Message: fmt.Sprintf(format, args...)}
}

// errTooLarge is returned when the uploaded file (or the whole multipart body)
// exceeds the configured cap. It maps to 413.
var errTooLarge = statusError(http.StatusRequestEntityTooLarge, "upload too large")

// UploadHandler serves POST /api/media/upload. It is designed to be mounted
// behind auth.RequireAuth (which 401s anonymous callers) on the /api group.
// It uses the configured Store to persist bytes; the default is LocalStore
// (P7.1) which writes to the local filesystem.
type UploadHandler struct {
	db       *sql.DB
	store    Store
	mediaDir string
	maxBytes int64
}

// NewUploadHandler builds an UploadHandler that stores files under mediaDir
// (created on demand) and rejects uploads larger than maxBytes. If maxBytes is
// <= 0 the default (25 MB) is used.  The store parameter is the persistence
// backend; pass nil to use the default LocalStore.
func NewUploadHandler(db *sql.DB, mediaDir string, maxBytes int64, store Store) *UploadHandler {
	if maxBytes <= 0 {
		maxBytes = defaultMaxBytes
	}
	if store == nil {
		store = &LocalStore{mediaDir: mediaDir}
	}
	return &UploadHandler{db: db, mediaDir: mediaDir, maxBytes: maxBytes, store: store}
}

// ServeHTTP handles POST /api/media/upload. It writes 201 with the Asset JSON
// on success, or an appropriate error status otherwise.
func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	asset, err := h.Upload(r.Context(), w, r)
	if err != nil {
		var se *StatusError
		if errors.As(err, &se) {
			writeJSON(w, se.Status, map[string]string{"error": se.Message})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, asset)
}

// Upload is the P4.1 HANDOFF: it parses the multipart body, validates the file
// by magic bytes, streams it to disk under a random basename, records the row
// in media_assets, and returns the resulting Asset. It does not itself require
// authentication (that is enforced by the middleware that mounts it); it
// returns a *StatusError for any client-correctable problem.
func (h *UploadHandler) Upload(ctx context.Context, w http.ResponseWriter, r *http.Request) (*Asset, error) {
	// Cap the whole body before reading anything. The +overhead lets the file
	// itself reach maxBytes (multipart envelope excluded).
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes+multipartOverhead)

	mr, err := r.MultipartReader()
	if err != nil {
		return nil, statusError(http.StatusBadRequest, "expected multipart/form-data body")
	}

	var filePart *multipart.Part
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			if isTooLarge(err) {
				return nil, errTooLarge
			}
			return nil, statusError(http.StatusBadRequest, "could not parse multipart body")
		}
		if part.FormName() == "file" {
			filePart = part
			break
		}
		// Discard any other form fields.
		part.Close()
	}
	if filePart == nil {
		return nil, statusError(http.StatusBadRequest, "multipart field 'file' is required")
	}
	defer filePart.Close()

	// Neutralize any path the client supplied: we only ever keep the basename,
	// and the physical path is server-generated (see storeFile).
	filename := sanitizeFilename(filePart.FileName())
	return h.storeFile(ctx, filePart, filename)
}

// storeFile detects the file's real type, streams it through the configured
// Store, and inserts the media_assets row.  The store produces a ref that is
// stored in stored_path; for LocalStore this ref is identical to the relative
// path produced by P4.1 (MEDIA_DIR/<YYYY-MM>/<random-hex>.<ext>).
func (h *UploadHandler) storeFile(ctx context.Context, part *multipart.Part, filename string) (*Asset, error) {
	// 1. Sniff magic bytes from the leading chunk.
	header := make([]byte, magicReadSize)
	n, err := io.ReadFull(part, header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		if isTooLarge(err) {
			return nil, errTooLarge
		}
		return nil, statusError(http.StatusBadRequest, "failed to read upload")
	}
	header = header[:n]
	if len(header) == 0 {
		return nil, statusError(http.StatusBadRequest, "empty file")
	}

	kind, mime, _ := detectMagic(header)
	if kind == "" {
		return nil, statusError(http.StatusBadRequest,
			"unsupported file type: only image, video, audio are allowed")
	}

	// 2. Stream through the store.  Prepend the header bytes so the store
	//    receives the complete file (header + rest).
	reader := io.MultiReader(bytes.NewReader(header), part)
	ref, total, err := h.store.Put(ctx, kind, filename, reader)
	if err != nil {
		return nil, statusError(http.StatusInternalServerError, "failed to store file: %s", err.Error())
	}
	if total > h.maxBytes {
		// Store accepted the data but it exceeds our cap; remove it and 413.
		_ = h.store.Delete(ctx, ref)
		return nil, errTooLarge
	}

	// 3. Record the row.  On any DB failure, remove the stored object.
	cleanup := func() { _ = h.store.Delete(ctx, ref) }
	res, err := h.db.ExecContext(ctx, `
		INSERT INTO media_assets (kind, stored_path, filename, bytes, mime, created_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))`,
		kind, ref, filename, total, mime)
	if err != nil {
		cleanup()
		return nil, statusError(http.StatusInternalServerError, "failed to record media asset")
	}
	id, err := res.LastInsertId()
	if err != nil {
		cleanup()
		return nil, statusError(http.StatusInternalServerError, "failed to record media asset")
	}

	return &Asset{
		ID:         id,
		URL:        fmt.Sprintf("/media/%d", id),
		Bytes:      total,
		MIME:       mime,
		Kind:       kind,
		Filename:   filename,
		StoredPath: ref,
	}, nil
}

// sanitizeFilename strips any path components from a client-supplied filename
// and falls back to "upload" if nothing usable remains. It never trusts the
// original value for anything but the media_assets.filename column.
func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "" || base == "." || base == "/" || strings.Contains(base, "..") {
		return "upload"
	}
	return base
}

// detectMagic returns the media kind, MIME type, and file extension determined
// from magic bytes. It returns ("", "", "") when the bytes match no allowed
// type (image/video/audio). This is the only authority on file type — the
// client filename and Content-Type are ignored.
func detectMagic(b []byte) (kind, mime, ext string) {
	if len(b) < 8 {
		return "", "", ""
	}
	switch {
	case b[0] == 0x89 && b[1] == 'P' && b[2] == 'N' && b[3] == 'G':
		return "image", "image/png", "png"
	case b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return "image", "image/jpeg", "jpg"
	case b[0] == 'G' && b[1] == 'I' && b[2] == 'F' && b[3] == '8':
		return "image", "image/gif", "gif"
	case b[0] == 'R' && b[1] == 'I' && b[2] == 'F' && b[3] == 'F':
		// RIFF container: WEBP (image) or WAVE (audio); anything else unknown.
		switch {
		case b[8] == 'W' && b[9] == 'E' && b[10] == 'B' && b[11] == 'P':
			return "image", "image/webp", "webp"
		case b[8] == 'W' && b[9] == 'A' && b[10] == 'V' && b[11] == 'E':
			return "audio", "audio/wav", "wav"
		}
		return "", "", ""
	case b[0] == 'O' && b[1] == 'g' && b[2] == 'g' && b[3] == 'S':
		return "audio", "audio/ogg", "ogg"
	case b[0] == 0x1A && b[1] == 0x45 && b[2] == 0xDF && b[3] == 0xA3:
		// EBML — Matroska/WebM.
		return "video", "video/webm", "webm"
	case b[0] == 'I' && b[1] == 'D' && b[2] == '3':
		return "audio", "audio/mpeg", "mp3"
	case b[0] == 0xFF && (b[1]&0xE0) == 0xE0:
		// MP3 frame sync.
		return "audio", "audio/mpeg", "mp3"
	case len(b) >= 12 && b[4] == 'f' && b[5] == 't' && b[6] == 'y' && b[7] == 'p':
		// ISO-BMFF (MP4/M4A). Distinguish audio brands from video.
		switch string(b[8:12]) {
		case "M4A ", "M4B ", "M4P ":
			return "audio", "audio/mp4", "m4a"
		default:
			return "video", "video/mp4", "mp4"
		}
	}
	return "", "", ""
}

// isTooLarge reports whether err came from http.MaxBytesReader.
func isTooLarge(err error) bool {
	var mbe *http.MaxBytesError
	return errors.As(err, &mbe)
}

// writeJSON marshals v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
