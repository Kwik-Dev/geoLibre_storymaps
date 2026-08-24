// Package media tests. P4.1 — POST /api/media/upload (local disk).
//
// Contract under test:
//   - a file claiming `.png` whose magic bytes are HTML is rejected
//   - an over-size file → 413 with nothing written to disk
//   - a valid image is stored under a random name (never the client name)
//   - an attempted `../../../etc/passwd` filename is neutralized
//   - a rejected upload writes no file to disk
package media

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
)

const uploadTestSecret = "test-secret-for-p41"

// newUploadRouter builds a production-shaped router: RequireAuth middleware +
// the upload handler. It returns the router, the token minting JWT secret is
// fixed (uploadTestSecret), and the media dir.
func newUploadRouter(t *testing.T, database *sql.DB, mediaDir string, maxBytes int64) *chi.Mux {
	t.Helper()
	r := chi.NewRouter()
	auther := auth.NewAuthenticator(uploadTestSecret, false)
	r.Route("/api", func(api chi.Router) {
		api.Use(auther.RequireAuth)
		api.Post("/media/upload", NewUploadHandler(database, mediaDir, maxBytes).ServeHTTP)
	})
	return r
}

func openDBTest(t *testing.T, path string) *sql.DB {
	t.Helper()
	d, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open(%q): %v", path, err)
	}
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := d.Exec(pragma); err != nil {
			t.Fatalf("%s: %v", pragma, err)
		}
	}
	return d
}

func uploadToken(t *testing.T, id int64, role string) string {
	t.Helper()
	tok, err := auth.NewJWT(uploadTestSecret).Sign(auth.User{ID: id, Role: role})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return tok
}

// multipartBody builds a multipart/form-data request body with a single "file"
// field named filename and content data. It returns the body and the
// Content-Type (with boundary).
func multipartBody(filename string, data []byte) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", filename)
	_, _ = fw.Write(data)
	_ = mw.Close()
	return &buf, mw.FormDataContentType()
}

func doUpload(t *testing.T, r *chi.Mux, token, ct string, body *bytes.Buffer) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/media/upload", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// countFiles recursively counts regular files under root.
func countFiles(root string) int {
	n := 0
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			n++
		}
		return nil
	})
	return n
}

// validPNG is a minimal real PNG (8-byte signature + IHDR chunk).
var validPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
	0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R', // IHDR chunk header
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
	0x08, 0x06, 0x00, 0x00, 0x00, // bit depth 8, color type 6
	0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00, 0x00, // CRC + data
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// TestUpload exercises the P4.1 contract.
func TestUpload(t *testing.T) {
	dir := t.TempDir()
	defer os.RemoveAll(dir)

	database := openDBTest(t, filepath.Join(dir, "p41.db"))
	defer database.Close()
	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	mediaDir := filepath.Join(dir, "media")
	ownerID := int64(1)
	router := newUploadRouter(t, database, mediaDir, 64*1024) // 64KB cap
	tok := uploadToken(t, ownerID, "user")

	t.Run("html-disguised-as-png-rejected", func(t *testing.T) {
		html := []byte("<!DOCTYPE html><html><body>not an image</body></html>")
		body, ct := multipartBody("photo.png", html)
		rec := doUpload(t, router, tok, ct, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "unsupported") {
			t.Fatalf("expected 'unsupported file type' error, got %s", rec.Body.String())
		}
		if n := countFiles(mediaDir); n != 0 {
			t.Fatalf("rejected upload wrote %d files to disk, want 0", n)
		}
	})

	t.Run("oversize-rejected-with-nothing-written", func(t *testing.T) {
		// 128KB of zero bytes > 64KB cap.
		big := bytes.Repeat([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, 16*1024)
		body, ct := multipartBody("big.png", big)
		rec := doUpload(t, router, tok, ct, body)
		if rec.Code != http.StatusRequestEntityTooLarge && rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 413 (or 400); body=%s", rec.Code, rec.Body.String())
		}
		if n := countFiles(mediaDir); n != 0 {
			t.Fatalf("oversize upload wrote %d files to disk, want 0", n)
		}
	})

	t.Run("valid-image-stores-random-name", func(t *testing.T) {
		body, ct := multipartBody("original-name.png", validPNG)
		rec := doUpload(t, router, tok, ct, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
		}
		var asset Asset
		if err := json.Unmarshal(rec.Body.Bytes(), &asset); err != nil {
			t.Fatalf("unmarshal asset: %v", err)
		}
		if asset.ID == 0 {
			t.Fatalf("asset id not set: %+v", asset)
		}
		if asset.URL != "/media/1" {
			t.Fatalf("asset url = %q, want /media/1", asset.URL)
		}
		if asset.MIME != "image/png" {
			t.Fatalf("asset mime = %q, want image/png", asset.MIME)
		}
		if asset.Bytes != int64(len(validPNG)) {
			t.Fatalf("asset bytes = %d, want %d", asset.Bytes, len(validPNG))
		}

		// The stored row must have a server-generated relative path and keep
		// the original filename for the filename column.
		var storedPath, filename, kind string
		var bytes int64
		if err := database.QueryRow(
			`SELECT stored_path, filename, kind, bytes FROM media_assets WHERE id = ?`,
			asset.ID).Scan(&storedPath, &filename, &kind, &bytes); err != nil {
			t.Fatalf("read asset row: %v", err)
		}
		if kind != "image" {
			t.Fatalf("kind = %q, want image", kind)
		}
		if filename != "original-name.png" {
			t.Fatalf("filename = %q, want original-name.png", filename)
		}
		// Random basename: path must be YYYY-MM/<hex>.png and not the client name.
		if strings.Contains(storedPath, "original-name") {
			t.Fatalf("stored_path leaks client name: %q", storedPath)
		}
		if !strings.HasPrefix(storedPath, "20") || !strings.HasSuffix(storedPath, ".png") {
			t.Fatalf("stored_path not a YYYY-MM/random.png relative path: %q", storedPath)
		}
		if filepath.IsAbs(storedPath) {
			t.Fatalf("stored_path is absolute: %q", storedPath)
		}
		if strings.Contains(storedPath, "..") {
			t.Fatalf("stored_path contains traversal: %q", storedPath)
		}

		// The physical file exists at the resolved relative path.
		abs := filepath.Join(mediaDir, filepath.FromSlash(storedPath))
		if fi, err := os.Stat(abs); err != nil || fi.Size() != int64(len(validPNG)) {
			t.Fatalf("stored file missing/wrong size: %v", err)
		}
	})

	t.Run("traversal-neutralized", func(t *testing.T) {
		body, ct := multipartBody("../../../etc/passwd", validPNG)
		rec := doUpload(t, router, tok, ct, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
		}
		var asset Asset
		if err := json.Unmarshal(rec.Body.Bytes(), &asset); err != nil {
			t.Fatalf("unmarshal asset: %v", err)
		}
		var storedPath, filename string
		if err := database.QueryRow(
			`SELECT stored_path, filename FROM media_assets WHERE id = ?`,
			asset.ID).Scan(&storedPath, &filename); err != nil {
			t.Fatalf("read asset row: %v", err)
		}
		// The physical path must be server-generated and safe.
		if strings.Contains(storedPath, "..") || filepath.IsAbs(storedPath) || strings.HasPrefix(storedPath, "/") {
			t.Fatalf("stored_path not neutralized: %q", storedPath)
		}
		// The filename column is the basename of what the client sent.
		if filename != "passwd" {
			t.Fatalf("filename = %q, want 'passwd' (basename of ../../../etc/passwd)", filename)
		}
		// No traversal escaped into the real filesystem outside mediaDir.
		if n := countFiles(filepath.Join(dir, "etc")); n != 0 {
			t.Fatalf("traversal wrote outside mediaDir: %d files", n)
		}
	})

	t.Run("anonymous-upload-401", func(t *testing.T) {
		before := countFiles(mediaDir)
		body, ct := multipartBody("anon.png", validPNG)
		req := httptest.NewRequest(http.MethodPost, "/api/media/upload", body)
		req.Header.Set("Content-Type", ct)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
		}
		if after := countFiles(mediaDir); after != before {
			t.Fatalf("anonymous upload wrote files: before=%d after=%d", before, after)
		}
	})
}
