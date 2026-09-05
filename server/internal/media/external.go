// Package media: P4.2 — external-URL validation.
//
// This file is pure (no DB, no I/O). It validates the *user-supplied external
// media URL* a chapter references when media_ref_type = external, and defines
// the media RefType enum plus a combine-check helper used by P4.3.
//
// Security model (feature_request §6, §10; HANDOUT §6):
//   - https-only; http/ftp/javascript:/data:/file: etc. are rejected.
//   - length-bounded (<= 2048) BEFORE parsing anything.
//   - the host is validated against an optional allow-list. When the list is
//     EMPTY the policy is DEFAULT-ALLOW (any well-formed https host passes).
//   - We do NOT fetch or SSRF the URL server-side — the user's external URL is
//     trusted input, we only validate its shape. This is deliberate and locked.
package media

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// maxExternalURLLen is the length cap for a user-supplied external media URL.
// It is enforced before any parsing. (feature_request §6: bounded length.)
const maxExternalURLLen = 2048

// ErrInvalidExternalURL wraps a human-readable reason describing why a
// candidate external media URL was rejected.
type ErrInvalidExternalURL struct {
	Reason string
}

func (e *ErrInvalidExternalURL) Error() string {
	return "invalid external media URL: " + e.Reason
}

func invalidURL(reason string) error {
	return &ErrInvalidExternalURL{Reason: reason}
}

// RefType enumerates how a chapter's media is referenced: by an external URL,
// by a locally-uploaded asset, or not at all. It is the media_ref_type column
// value and the HANDOFF to P4.3.
type RefType string

const (
	RefTypeExternal RefType = "external"
	RefTypeLocal    RefType = "local"
	RefTypeNone     RefType = "none"
)

// Valid reports whether r is one of the three known RefType values.
func (r RefType) Valid() bool {
	switch r {
	case RefTypeExternal, RefTypeLocal, RefTypeNone:
		return true
	}
	return false
}

// allowedHostPrefix reports whether u's host (optionally with a path prefix)
// matches an entry in allowedHosts. An entry may be a bare host ("cdn.example")
// or a host+path prefix ("cdn.example/videos"). The match is case-insensitive
// on the host and exact on the path prefix.
func allowedHostPrefix(u *url.URL, allowedHosts []string) bool {
	host := strings.ToLower(u.Hostname())
	// u.EscapedPath keeps the raw path; Path is decoded. For a prefix match we
	// use the decoded path and trim a leading slash.
	path := strings.TrimPrefix(u.Path, "/")

	for _, allow := range allowedHosts {
		a := strings.ToLower(strings.TrimSpace(allow))
		if a == "" {
			continue
		}
		// Split host from optional path prefix. Bare host = any path allowed.
		aHost, aPath, _ := strings.Cut(a, "/")
		if host != strings.ToLower(aHost) {
			continue
		}
		if aPath == "" {
			return true
		}
		if strings.HasPrefix(path, aPath) {
			return true
		}
	}
	return false
}

// ValidateExternalURL validates a user-supplied external media URL against the
// media policy:
//
//   - the string must be non-empty and len(s) <= 2048 (capped first);
//   - it must url.Parse cleanly;
//   - Scheme must be exactly "https" (rejects http, ftp, javascript:, data:,
//     file:, etc.);
//   - the host must be non-empty;
//   - if allowedHosts is NON-EMPTY, the URL's host (optionally a host+path
//     prefix) must appear in it;
//   - if allowedHosts is EMPTY, the policy is DEFAULT-ALLOW: any well-formed
//     https URL with a non-empty host is accepted.
//
// It returns nil when the URL is acceptable, or an *ErrInvalidExternalURL
// otherwise. It never fetches or resolves the URL (no SSRF — the URL is trusted
// user input, not a server-side fetch).
func ValidateExternalURL(s string, allowedHosts []string) error {
	if s == "" {
		return invalidURL("URL is empty")
	}
	if len(s) > maxExternalURLLen {
		return invalidURL(fmt.Sprintf("URL length %d exceeds max %d", len(s), maxExternalURLLen))
	}

	u, err := url.Parse(s)
	if err != nil {
		return invalidURL(fmt.Sprintf("could not parse URL: %v", err))
	}
	if u.Scheme != "https" {
		return invalidURL(fmt.Sprintf("scheme %q is not allowed; only https is accepted", u.Scheme))
	}
	if u.Hostname() == "" {
		return invalidURL("URL has no host")
	}
	if u.Hostname() != strings.ToLower(u.Hostname()) {
		// Lowercasing hosts is normal for the allow-list; accept mixed case.
	}

	// Allow-list is optional. When empty, default-allow any well-formed https
	// URL. When non-empty, the host (or host+path prefix) must match.
	if len(allowedHosts) > 0 && !allowedHostPrefix(u, allowedHosts) {
		return invalidURL(fmt.Sprintf("host %q is not in the media allow-list", u.Hostname()))
	}

	return nil
}

// CheckMediaRef is the combine-check helper used by P4.3. It validates that a
// chapter's (mediaType, refType, externalURL, allowList) combination is
// consistent, mirroring the matrix from HANDOUT §6:
//
//	mediaType=none ⇒ refType=none (both external URL and asset id empty)
//	refType=external ⇒ mediaType != none, externalURL set + passes ValidateExternalURL
//	refType=local ⇒ mediaType != none, (asset id checked separately by the caller)
//	refType=none ⇒ mediaType=none
//
// It returns an error describing the inconsistency, or nil when the combo is
// structurally valid. The caller (P4.3) is responsible for the asset-id
// existence/ownership check for the local case.
func CheckMediaRef(mediaType string, refType RefType, externalURL string, allowedHosts []string) error {
	if !refType.Valid() {
		return invalidURL(fmt.Sprintf("unknown media_ref_type %q", refType))
	}

	noneType := mediaType == "" || mediaType == "none"
	switch refType {
	case RefTypeNone:
		if !noneType {
			return errors.New("media_ref_type=none requires media_type=none")
		}
		if externalURL != "" {
			return errors.New("media_ref_type=none requires an empty media_external_url")
		}
	case RefTypeExternal:
		if noneType {
			return errors.New("media_ref_type=external requires a concrete media_type (image, video, or audio)")
		}
		if err := ValidateExternalURL(externalURL, allowedHosts); err != nil {
			return err
		}
	case RefTypeLocal:
		if noneType {
			return errors.New("media_ref_type=local requires a concrete media_type (image, video, or audio)")
		}
		if externalURL != "" {
			return errors.New("media_ref_type=local must not set media_external_url")
		}
	}
	return nil
}
