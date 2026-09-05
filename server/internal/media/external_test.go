// Package media tests. P4.2 — external-URL validation.
//
// Contract under test (see card):
//   - rejects http://, ftp://, javascript:alert(1), overlength, and empty strings
//   - with an allow-list set, rejects a disallowed host and accepts an allowed one
//   - with the allow-list empty, accepts any well-formed https URL
//   - RefType enum + CheckMediaRef combine-check helper
package media

import (
	"strings"
	"testing"
)

func TestExternalURL(t *testing.T) {
	long := "https://cdn.example.com/" + strings.Repeat("a", 3000)

	reject := map[string]string{
		"http URL":                  "http://cdn.example.com/img.png",
		"ftp URL":                   "ftp://cdn.example.com/img.png",
		"javascript scheme":         "javascript:alert(1)",
		"data URL":                  "data:image/png;base64,AAAA",
		"file URL":                  "file:///etc/passwd",
		"empty string":              "",
		"overlong URL":              long,
		"malformed (no scheme)":     "cdn.example.com/img.png",
		"scheme-relative hostless":  "https://",
		"scheme with spaces":        "ht tps://cdn.example.com/a",
	}
	for name, u := range reject {
		t.Run("reject/"+name, func(t *testing.T) {
			if err := ValidateExternalURL(u, nil); err == nil {
				t.Fatalf("ValidateExternalURL(%q) = nil, want error", u)
			}
		})
	}

	// Allow-list empty ⇒ default-allow any well-formed https URL.
	accept := map[string]string{
		"bare https host":      "https://cdn.example.com/img.png",
		"query+fragment":       "https://drive.google.com/file/d/ABC?usp=sharing#x",
		"port":                 "https://cdn.example.com:8443/v.mp4",
		"subdomain":            "https://a.b.c.example.net/x",
		"max-length boundary":  "https://x.example/" + strings.Repeat("a", 2029), // len == 2048
	}
	for name, u := range accept {
		t.Run("accept-empty-allowlist/"+name, func(t *testing.T) {
			if err := ValidateExternalURL(u, nil); err != nil {
				t.Fatalf("ValidateExternalURL(%q) = %v, want nil (default-allow)", u, err)
			}
		})
	}

	// Allow-list set ⇒ only listed hosts (optionally host+path prefix) pass.
	allowList := []string{"cdn.example.com", "drive.google.com/shared"}

	for name, u := range map[string]string{
		"disallowed host":   "https://evil.example.com/x.png",
		"partial host":      "https://cdn.example.co.uk/x.png",
		"wrong path prefix": "https://drive.google.com/other/x.png",
	} {
		t.Run("reject-allowlist/"+name, func(t *testing.T) {
			if err := ValidateExternalURL(u, allowList); err == nil {
				t.Fatalf("ValidateExternalURL(%q, %v) = nil, want error", u, allowList)
			}
		})
	}

	for name, u := range map[string]string{
		"exact host":              "https://cdn.example.com/img.png",
		"host+path prefix":        "https://drive.google.com/shared/video.mp4",
		"host+path exact":         "https://drive.google.com/shared",
		"path-prefix boundary":    "https://drive.google.com/shared",
	} {
		t.Run("accept-allowlist/"+name, func(t *testing.T) {
			if err := ValidateExternalURL(u, allowList); err != nil {
				t.Fatalf("ValidateExternalURL(%q, %v) = %v, want nil", u, allowList, err)
			}
		})
	}
}

func TestRefTypeValid(t *testing.T) {
	for _, rt := range []RefType{RefTypeExternal, RefTypeLocal, RefTypeNone} {
		if !rt.Valid() {
			t.Fatalf("%q should be a valid RefType", rt)
		}
	}
	for _, rt := range []RefType{"", "url", "external ", "LOCAL"} {
		if rt.Valid() {
			t.Fatalf("%q should be an invalid RefType", rt)
		}
	}
}

func TestCheckMediaRef(t *testing.T) {
	// Valid combinations.
	valid := []struct {
		mediaType string
		refType   RefType
		url       string
	}{
		{"image", RefTypeExternal, "https://cdn.example.com/a.png"},
		{"video", RefTypeExternal, "https://drive.google.com/shared/v.mp4"},
		{"audio", RefTypeLocal, ""},
		{"none", RefTypeNone, ""},
		{"", RefTypeNone, ""},
	}
	for i, c := range valid {
		t.Run("valid/"+t.Name()+string(rune('a'+i)), func(t *testing.T) {
			if err := CheckMediaRef(c.mediaType, c.refType, c.url, []string{"cdn.example.com", "drive.google.com/shared"}); err != nil {
				t.Fatalf("CheckMediaRef(%q,%q,%q) = %v, want nil", c.mediaType, c.refType, c.url, err)
			}
		})
	}

	// Invalid combinations.
	invalid := []struct {
		mediaType string
		refType   RefType
		url       string
	}{
		{"none", RefTypeExternal, "https://cdn.example.com/a.png"}, // none type w/ external ref
		{"none", RefTypeLocal, ""},                                 // none type w/ local ref
		{"image", RefTypeNone, "https://cdn.example.com/a.png"},    // none ref w/ concrete type + url
		{"image", RefTypeExternal, "http://cdn.example.com/a.png"}, // external w/ http
		{"image", RefTypeExternal, "https://evil.example.com/a.png"}, // external w/ disallowed host
		{"image", RefTypeLocal, "https://cdn.example.com/a.png"},   // local w/ external url
		{"image", RefType("bogus"), ""},                            // unknown ref type
	}
	for i, c := range invalid {
		t.Run("invalid/"+string(rune('a'+i)), func(t *testing.T) {
			if err := CheckMediaRef(c.mediaType, c.refType, c.url, []string{"cdn.example.com"}); err == nil {
				t.Fatalf("CheckMediaRef(%q,%q,%q) = nil, want error", c.mediaType, c.refType, c.url)
			}
		})
	}
}
