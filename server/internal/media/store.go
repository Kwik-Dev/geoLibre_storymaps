// Package media: P7.1 — Upload store abstraction (S3 / Drive proxy).
//
// This file defines the Store interface for persisting uploaded media bytes.
// The default (STORE_KIND=local / unset) is LocalStore, which writes to the
// local filesystem under MEDIA_DIR — producing the exact same file layout as
// P4.1. When STORE_KIND=s3 is set, an S3-backed store is used instead.
//
// Upload routes to the configured store (see NewStore).  Serve / GET /media/:aid
// is NOT affected by this abstraction in P7.1 — it continues to read from the
// local filesystem (LocalStore) or, for S3, from the S3 bucket via a direct
// URL or proxy (future card).
package media

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Store is the persistence interface for uploaded media.
//
//   - Put stores the data from r and returns a reference that can later be
//     passed to Get / URL / Delete, plus the number of bytes written.
//     kind is the media type ("image", "video", "audio"); name is the
//     sanitised client-supplied filename (for metadata only).
//   - Get returns an io.ReadCloser for the object identified by ref.
//   - URL returns a direct-access URL for ref (may be empty if the store
//     does not expose direct URLs).
//   - Delete removes the object identified by ref.
type Store interface {
	Put(ctx context.Context, kind, name string, r io.Reader) (ref string, bytes int64, err error)
	Get(ctx context.Context, ref string) (io.ReadCloser, error)
	URL(ref string) string
	Delete(ctx context.Context, ref string) error
}

// StoreKind represents the configured store type.
type StoreKind string

const (
	StoreKindLocal StoreKind = "local"
	StoreKindS3    StoreKind = "s3"
)

// NewStore returns the Store implementation for the given kind.  When kind is
// empty or "local" it returns a LocalStore writing under mediaDir.  When kind
// is "s3" it returns an S3Store (the caller must configure it via the returned
// *S3Store's SetClient / SetBucket methods, or via S3Config).  Any other kind
// returns nil, false.
func NewStore(kind StoreKind, mediaDir string) (Store, bool) {
	switch kind {
	case "", StoreKindLocal:
		return &LocalStore{mediaDir: mediaDir}, true
	case StoreKindS3:
		return &S3Store{
			prefix:  "media/",
			bucket:  os.Getenv("S3_BUCKET"),
			region:  os.Getenv("S3_REGION"),
			baseURL: os.Getenv("S3_BASE_URL"),
		}, true
	}
	return nil, false
}

// ─── LocalStore ────────────────────────────────────────────────────────────

// LocalStore stores uploaded media on the local filesystem under mediaDir.
// The file layout is identical to P4.1: MEDIA_DIR/<YYYY-MM>/<random-hex>.<ext>.
type LocalStore struct {
	mediaDir string
}

// Put writes the streamed data to mediaDir/<YYYY-MM>/<random-hex>.<ext>,
// derives the extension from kind, and returns the relative path as ref.
// This produces the exact same file layout as P4.1's storeFile.
func (s *LocalStore) Put(ctx context.Context, kind, name string, r io.Reader) (ref string, bytes int64, err error) {
	ext := kindToExt(kind) // e.g. "image" → "png", "video" → "mp4", "audio" → "mp3"

	hexName, err := randomHex(16)
	if err != nil {
		return "", 0, fmt.Errorf("generate random name: %v", err)
	}
	ym := time.Now().UTC().Format("2006-01")
	rel := ym + "/" + hexName + "." + ext
	if strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
		return "", 0, fmt.Errorf("invalid storage path generated: %q", rel)
	}

	abs := filepath.Join(s.mediaDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", 0, fmt.Errorf("create media directory: %v", err)
	}

	out, err := os.OpenFile(abs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", 0, fmt.Errorf("open storage file: %v", err)
	}
	defer func() {
		_ = out.Close()
		if err != nil {
			_ = os.Remove(abs)
		}
	}()

	written, err := io.Copy(out, r)
	if err != nil {
		return "", 0, fmt.Errorf("write file: %v", err)
	}
	if written == 0 {
		return "", 0, fmt.Errorf("empty file written")
	}

	if err := out.Sync(); err != nil {
		return "", 0, fmt.Errorf("sync file: %v", err)
	}

	return rel, written, nil
}

// Get opens the file at mediaDir/ref for reading.
func (s *LocalStore) Get(ctx context.Context, ref string) (io.ReadCloser, error) {
	abs := filepath.Join(s.mediaDir, filepath.FromSlash(ref))
	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open %q: %v", ref, err)
	}
	return f, nil
}

// URL returns an empty string for LocalStore — files are served through the
// /media/:aid HTTP endpoint, not via a direct store URL.
func (s *LocalStore) URL(ref string) string {
	return ""
}

// Delete removes the file at mediaDir/ref.
func (s *LocalStore) Delete(ctx context.Context, ref string) error {
	abs := filepath.Join(s.mediaDir, filepath.FromSlash(ref))
	if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %q: %v", ref, err)
	}
	return nil
}

// ─── S3Store ───────────────────────────────────────────────────────────────

// s3Client is the minimal S3 API that S3Store requires.  In production this
// would be wired to an *aws-sdk-go-v2* S3 client; tests inject an in-memory
// implementation that satisfies the same contract.
type s3Client interface {
	PutObject(ctx context.Context, bucket, key string, body io.Reader) (int64, error)
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error
}

// S3Config holds the settings for an S3-backed store.
type S3Config struct {
	Bucket  string
	Region  string
	Prefix  string // optional object-key prefix (e.g. "media/")
	BaseURL string // optional direct-access base URL (e.g. "https://my-bucket.s3.amazonaws.com")
}

// S3Store stores uploaded media in an S3-compatible bucket.
// The client field is pluggable so tests can inject a memory-backed
// implementation without the real AWS SDK.
type S3Store struct {
	client  s3Client
	bucket  string
	region  string
	prefix  string
	baseURL string
}

// SetClient sets the S3 client implementation.  Must be called before Put/Get/
// Delete — NewStore returns a bare S3Store that a test or the real main() can
// wire up.
func (s *S3Store) SetClient(c s3Client) { s.client = c }

// SetBucket sets the S3 bucket name.
func (s *S3Store) SetBucket(b string) { s.bucket = b }

// SetPrefix sets the object-key prefix (e.g. "uploads/media/").
func (s *S3Store) SetPrefix(p string) { s.prefix = p }

// SetBaseURL sets the direct-access base URL for the URL() method.
func (s *S3Store) SetBaseURL(u string) { s.baseURL = u }

// Put streams the data to s3://<bucket>/<prefix><random-hex>-<sanitised-name>.
// ref is the full object key.
func (s *S3Store) Put(ctx context.Context, kind, name string, r io.Reader) (ref string, bytes int64, err error) {
	if s.client == nil {
		return "", 0, fmt.Errorf("S3Store: client not set; call SetClient")
	}
	if s.bucket == "" {
		return "", 0, fmt.Errorf("S3Store: bucket not set; call SetBucket or set S3_BUCKET")
	}

	ext := kindToExt(kind)
	hexName, err := randomHex(16)
	if err != nil {
		return "", 0, fmt.Errorf("generate random name: %v", err)
	}

	// Sanitise name to a safe object-key suffix.
	safeName := sanitiseForS3(name)
	key := s.prefix + hexName + "-" + safeName + "." + ext

	n, err := s.client.PutObject(ctx, s.bucket, key, r)
	if err != nil {
		return "", 0, fmt.Errorf("S3 PutObject: %v", err)
	}

	return key, n, nil
}

// Get retrieves the object at the given key from S3.
func (s *S3Store) Get(ctx context.Context, ref string) (io.ReadCloser, error) {
	if s.client == nil {
		return nil, fmt.Errorf("S3Store: client not set")
	}
	if s.bucket == "" {
		return nil, fmt.Errorf("S3Store: bucket not set")
	}
	return s.client.GetObject(ctx, s.bucket, ref)
}

// URL returns the direct S3 object URL if baseURL is configured, or an empty
// string otherwise.
func (s *S3Store) URL(ref string) string {
	if s.baseURL != "" {
		return strings.TrimRight(s.baseURL, "/") + "/" + strings.TrimLeft(ref, "/")
	}
	return ""
}

// Delete removes the object at the given key from S3.
func (s *S3Store) Delete(ctx context.Context, ref string) error {
	if s.client == nil {
		return fmt.Errorf("S3Store: client not set")
	}
	if s.bucket == "" {
		return fmt.Errorf("S3Store: bucket not set")
	}
	return s.client.DeleteObject(ctx, s.bucket, ref)
}

// ─── Helpers ───────────────────────────────────────────────────────────────

// kindToExt maps media kind ("image", "video", "audio") to a default file
// extension.  The extension is only used for local file naming; the actual
// MIME is determined by magic bytes in the upload handler.
func kindToExt(kind string) string {
	switch kind {
	case "image":
		return "png"
	case "video":
		return "mp4"
	case "audio":
		return "mp3"
	default:
		return "bin"
	}
}

// sanitiseForS3 makes a filename safe for use in an S3 object key.
func sanitiseForS3(name string) string {
	// Strip path components and non-printable characters.
	base := filepath.Base(name)
	if base == "" || base == "." || base == "/" || strings.Contains(base, "..") {
		return "upload"
	}
	// Replace characters that are problematic in S3 keys.
	r := strings.NewReplacer(
		"/", "_", "\\", "_", "..", "_",
		"\x00", "", "\n", "", "\r", "",
	)
	return r.Replace(base)
}

// randomHex returns n random bytes encoded as lowercase hex.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ─── MemStore (in-memory store for testing) ────────────────────────────────

// MemStore is an in-memory Store implementation used for testing the Store
// contract and for S3Store client mocking.  Not for production use.
type MemStore struct {
	data map[string][]byte
}

// NewMemStore creates an empty in-memory store.
func NewMemStore() *MemStore {
	return &MemStore{data: make(map[string][]byte)}
}

// Put stores the bytes in memory under a ref of the form "mem/<kind>/<hex>".
func (m *MemStore) Put(ctx context.Context, kind, name string, r io.Reader) (ref string, bytes int64, err error) {
	all, err := io.ReadAll(r)
	if err != nil {
		return "", 0, err
	}
	hexName, err := randomHex(8)
	if err != nil {
		return "", 0, err
	}
	ref = "mem/" + kind + "/" + hexName
	m.data[ref] = all
	return ref, int64(len(all)), nil
}

// Get returns the bytes for ref.
func (m *MemStore) Get(ctx context.Context, ref string) (io.ReadCloser, error) {
	b, ok := m.data[ref]
	if !ok {
		return nil, fmt.Errorf("memstore: ref %q not found", ref)
	}
	return io.NopCloser(strings.NewReader(string(b))), nil
}

// URL always returns "" for MemStore.
func (m *MemStore) URL(ref string) string { return "" }

// Delete removes the entry for ref.
func (m *MemStore) Delete(ctx context.Context, ref string) error {
	delete(m.data, ref)
	return nil
}
