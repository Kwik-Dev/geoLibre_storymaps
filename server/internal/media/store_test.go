// Package media tests. P7.1 — Store abstraction (LocalStore + S3Store).
//
// Contract under test:
//   - TestStoreLocal: LocalStore Put/Get/URL/Delete round-trip through a
//     temporary directory; verifies the file layout matches P4.1 (relative
//     YYYY-MM/<hex>.<ext> under mediaDir).
//   - TestStoreS3: S3Store with an in-memory fake S3 client, verifying that
//     the interface dispatches through the s3Client abstraction (proves the
//     "interface switches on STORE_KIND" requirement).
package media

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── TestStoreLocal ────────────────────────────────────────────────────────

func TestStoreLocal(t *testing.T) {
	dir := t.TempDir()
	defer os.RemoveAll(dir)

	ctx := context.Background()
	s := &LocalStore{mediaDir: dir}

	content := []byte("hello, store! this is a test file.")
	r := bytes.NewReader(content)

	ref, n, err := s.Put(ctx, "image", "test-img.png", r)
	if err != nil {
		t.Fatalf("LocalStore.Put: %v", err)
	}
	if n != int64(len(content)) {
		t.Fatalf("bytes written = %d, want %d", n, len(content))
	}
	if ref == "" {
		t.Fatal("ref is empty")
	}
	// Ref must be a relative path under mediaDir (YYYY-MM/<hex>.<ext>).
	if filepath.IsAbs(ref) || strings.Contains(ref, "..") || strings.HasPrefix(ref, "/") {
		t.Fatalf("ref is not a safe relative path: %q", ref)
	}
	if !strings.HasPrefix(ref, "20") {
		t.Fatalf("ref does not start with YYYY-MM: %q", ref)
	}
	if !strings.Contains(ref, ".") {
		t.Fatalf("ref has no extension: %q", ref)
	}
	if strings.Contains(ref, "test-img") {
		t.Fatalf("ref leaks client filename: %q", ref)
	}

	// The physical file must exist at the resolved path.
	abs := filepath.Join(dir, filepath.FromSlash(ref))
	if fi, err := os.Stat(abs); err != nil {
		t.Fatalf("file not found: %v", err)
	} else if fi.Size() != int64(len(content)) {
		t.Fatalf("file size = %d, want %d", fi.Size(), len(content))
	}

	// Get must read back the exact content.
	rc, err := s.Get(ctx, ref)
	if err != nil {
		t.Fatalf("LocalStore.Get: %v", err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read Get result: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("Get returned %q, want %q", got, content)
	}

	// URL must be empty for LocalStore (files are served via /media/:aid).
	if url := s.URL(ref); url != "" {
		t.Fatalf("LocalStore.URL = %q, want empty", url)
	}

	// Delete must remove the file.
	if err := s.Delete(ctx, ref); err != nil {
		t.Fatalf("LocalStore.Delete: %v", err)
	}
	if _, err := os.Stat(abs); !os.IsNotExist(err) {
		t.Fatalf("file still exists after Delete: %v", err)
	}
	// Get on a deleted ref must fail.
	if _, err := s.Get(ctx, ref); err == nil {
		t.Fatal("LocalStore.Get after Delete: expected error, got nil")
	}
}

// ─── TestStoreS3 ───────────────────────────────────────────────────────────

// memS3Client is an in-memory S3 client for testing S3Store.
type memS3Client struct {
	data map[string][]byte // key → content
}

func (m *memS3Client) PutObject(_ context.Context, bucket, key string, body io.Reader) (int64, error) {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	all, err := io.ReadAll(body)
	if err != nil {
		return 0, err
	}
	fullKey := bucket + ":" + key
	m.data[fullKey] = all
	return int64(len(all)), nil
}

func (m *memS3Client) GetObject(_ context.Context, bucket, key string) (io.ReadCloser, error) {
	if m.data == nil {
		return nil, fmt.Errorf("memS3: not found %s/%s", bucket, key)
	}
	fullKey := bucket + ":" + key
	b, ok := m.data[fullKey]
	if !ok {
		return nil, fmt.Errorf("memS3: not found %s/%s", bucket, key)
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (m *memS3Client) DeleteObject(_ context.Context, bucket, key string) error {
	if m.data != nil {
		delete(m.data, bucket+":"+key)
	}
	return nil
}

func TestStoreS3(t *testing.T) {
	ctx := context.Background()
	mem := &memS3Client{}

	s := &S3Store{
		client:  mem,
		bucket:  "test-bucket",
		prefix:  "media/",
		baseURL: "https://test-bucket.s3.amazonaws.com",
	}

	content := []byte("s3 test content")
	r := bytes.NewReader(content)

	ref, n, err := s.Put(ctx, "video", "test-video.mp4", r)
	if err != nil {
		t.Fatalf("S3Store.Put: %v", err)
	}
	if n != int64(len(content)) {
		t.Fatalf("bytes written = %d, want %d", n, len(content))
	}
	if ref == "" || !strings.HasPrefix(ref, "media/") {
		t.Fatalf("ref = %q, want media/<hex>-<name>.<ext>", ref)
	}
	if !strings.Contains(ref, ".mp4") {
		t.Fatalf("ref does not preserve extension: %q", ref)
	}

	// Get must return the content.
	rc, err := s.Get(ctx, ref)
	if err != nil {
		t.Fatalf("S3Store.Get: %v", err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read S3 Get result: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("Get returned %q, want %q", got, content)
	}

	// URL must return the direct S3 URL.
	url := s.URL(ref)
	expectedURL := "https://test-bucket.s3.amazonaws.com/" + ref
	if url != expectedURL {
		t.Fatalf("URL = %q, want %q", url, expectedURL)
	}

	// Delete must remove the object.
	if err := s.Delete(ctx, ref); err != nil {
		t.Fatalf("S3Store.Delete: %v", err)
	}
	if _, err := s.Get(ctx, ref); err == nil {
		t.Fatal("S3Store.Get after Delete: expected error, got nil")
	}

	// Put via a nil client must fail.
	s2 := &S3Store{}
	if _, _, err := s2.Put(ctx, "image", "x.png", bytes.NewReader([]byte("x"))); err == nil {
		t.Fatal("S3Store.Put with nil client: expected error, got nil")
	}
}

// ─── TestNewStore ──────────────────────────────────────────────────────────

func TestNewStore(t *testing.T) {
	dir := t.TempDir()
	defer os.RemoveAll(dir)

	// Empty kind → LocalStore.
	s, ok := NewStore("", dir)
	if !ok {
		t.Fatal("NewStore(\"\", dir) returned ok=false")
	}
	if _, isLocal := s.(*LocalStore); !isLocal {
		t.Fatalf("NewStore(\"\") = %T, want *LocalStore", s)
	}

	// "local" → LocalStore.
	s, ok = NewStore(StoreKindLocal, dir)
	if !ok {
		t.Fatal("NewStore(\"local\", dir) returned ok=false")
	}
	if _, isLocal := s.(*LocalStore); !isLocal {
		t.Fatalf("NewStore(\"local\") = %T, want *LocalStore", s)
	}

	// "s3" → S3Store.
	s, ok = NewStore(StoreKindS3, dir)
	if !ok {
		t.Fatal("NewStore(\"s3\", dir) returned ok=false")
	}
	if _, isS3 := s.(*S3Store); !isS3 {
		t.Fatalf("NewStore(\"s3\") = %T, want *S3Store", s)
	}

	// Unknown kind → nil, false.
	if s, ok := NewStore("gcs", dir); ok || s != nil {
		t.Fatalf("NewStore(\"gcs\") = (%T, %t), want (nil, false)", s, ok)
	}
}

// ─── TestMemStore ──────────────────────────────────────────────────────────

// TestMemStore verifies that the in-memory store satisfies the Store contract
// for all four methods. This is the simplest way to prove the interface works.
func TestMemStore(t *testing.T) {
	ctx := context.Background()
	m := NewMemStore()

	content := []byte("memstore test content")
	r := bytes.NewReader(content)

	ref, n, err := m.Put(ctx, "audio", "test.mp3", r)
	if err != nil {
		t.Fatalf("MemStore.Put: %v", err)
	}
	if n != int64(len(content)) {
		t.Fatalf("bytes written = %d, want %d", n, len(content))
	}
	if ref == "" {
		t.Fatal("ref is empty")
	}
	if !strings.HasPrefix(ref, "mem/audio/") {
		t.Fatalf("ref = %q, want mem/audio/<hex>", ref)
	}

	// Get.
	rc, err := m.Get(ctx, ref)
	if err != nil {
		t.Fatalf("MemStore.Get: %v", err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read MemStore Get: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("Get returned %q, want %q", got, content)
	}

	// URL is empty.
	if u := m.URL(ref); u != "" {
		t.Fatalf("MemStore.URL = %q, want empty", u)
	}

	// Delete.
	if err := m.Delete(ctx, ref); err != nil {
		t.Fatalf("MemStore.Delete: %v", err)
	}
	if _, err := m.Get(ctx, ref); err == nil {
		t.Fatal("MemStore.Get after Delete: expected error, got nil")
	}
}
