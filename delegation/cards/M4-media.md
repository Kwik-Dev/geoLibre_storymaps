# Cards — M4 Media (upload, external-URL, wiring, serving)

Read `delegation/HANDOUT.md §0.3/§4` first. Card fields as in M1. A card is
`closed` only when its `VERIFY` passes. No PR, no out-of-scope edits.

The media rule (from HANDOUT §6): on a chapter, exactly one of these holds —
`media_ref_type=external` ⇒ `media_external_url` set + validated;
`media_ref_type=local` ⇒ `media_asset_id` set (a valid `media_assets` row the
author owns or that is public); `media_ref_type=none` ⇒ both empty.
`media_type` ∈ {image,video,audio}; if `none`, ref must be `none`.

---

## P4.1 — Upload endpoint (local disk)
- **DEPS:** P3.1 · **core** · est ~25m
- **READ:** feature_request §6 (media), §10 (traversal, size cap); HANDOUT §6
- **SCOPE:**
    - `server/internal/media/upload.go` + `POST /api/media/upload`
      (auth required, multipart field `file`).
    - **Magic-byte** MIME detection (don't trust `Content-Type`/extension). Use
       `github.com/h2non/filetype` (or `golang.org/x/image` signatures) and allow
      only image/video/audio; reject anything else.
    - Enforce `MEDIA_MAX_BYTES` (default 25MB) **before writing** via
       `http.MaxBytesReader` (413 on over-size).
    - Store under `MEDIA_DIR/<YYYY-MM>/` with a **random** basename
       (`crypto/rand` hex). Set `media_assets.filename` = the original name,
       `stored_path` = a **relative** path (never absolute, never client-supplied).
      Insert the `media_assets` row (`kind`, `stored_path`, `filename`, `bytes`,
       `mime`).
    - Reject path traversal: random basenames mean `..` can't appear; also reject
       any client-supplied path/folder.
    - Return `{"id","url":"/media/<id>","bytes","mime"}`.
- **VERIFY:**
    ```
    CGO_ENABLED=0 go test ./internal/media -run TestUpload -v
     ```
   expects: a file claiming `.png` whose **magic bytes are HTML** is rejected;
   an over-size file → 413/400 with nothing written; a valid image stores a
   random name; an attempted `../../../etc/passwd` is neutralized; a rejected
   upload writes **no file to disk**.
- **HANDOFF:** `media.Upload(ctx, w, r) (*Asset, error)`; `asset.URL`.
- **GOTCHAS:** never trust the client filename; cap the size **before** the
   first write (stream, don't buffer the whole body); random basenames; MIME by
   magic bytes, not headers.

---

## P4.2 — External-URL validation
- **DEPS:** none (pure) · **core** · est ~10m
- **READ:** feature_request §6, §10; HANDOUT §6
- **SCOPE:**
    - `server/internal/media/external.go`:
       `ValidateExternalURL(s string, allowedHosts []string) error`:
         - must `url.Parse` cleanly; `Scheme == "https"` (reject `http`,
          `ftp`, `javascript:`, `data:`, `file:`, etc.); host non-empty;
          `len(s) <= 2048` (cap before anything else).
         - if `allowedHosts` is **non-empty**, the host (optionally a
          host+path prefix) must be in it; if **empty**, any `https` host is
          allowed.
    - A `media.RefType` enum {external, local, none} and a combine-check helper
       used by P4.3.
- **VERIFY:**
    ```
    CGO_ENABLED=0 go test ./internal/media -run TestExternalURL -v
     ```
   expects: rejects `http://`, `ftp://`, `javascript:alert(1)`, overlength, and
   empty strings; with an allow-list set, rejects a disallowed host and accepts
   an allowed one; with it empty, accepts any well-formed `https` URL.
- **HANDOFF:** `media.ValidateExternalURL(s string, allowedHosts []string) error`,
   `media.RefType`.
- **GOTCHAS:** default-**allow** when the allow-list is empty (document it);
   cap length first; **do NOT fetch/SSRF** the URL server-side — just validate
   it. (The user's external URL is trusted input, not a server-side fetch.)

---

## P4.3 — Wire `media_ref_type` into chapters
- **DEPS:** P3.2, P4.1, P4.2 · **core** · est ~20m
- **READ:** feature_request §6; HANDOUT §6
- **SCOPE:**
    - Extend chapter create/update (P3.2) to accept `media_type` ∈
      {image,video,audio,none}, `media_ref_type` ∈ {external,local,none}, and
       `media_external_url` / `media_asset_id`.
    - Enforce the matrix: `none` ⇒ both empty; `external` ⇒ `media_external_url`
       set **and** passes `ValidateExternalURL` with the configured allow-list;
       `local` ⇒ `media_asset_id` set **and** the row exists **and** is owned by
       the author or is a public asset. Any inconsistent combo → **400** with a
       specific code; a foreign private asset → **403**.
    - Persist exactly the chosen fields on the `chapters` row.
- **VERIFY:**
    ```
    CGO_ENABLED=0 go test ./internal/api -run TestChapterMedia -v
     ```
   expects: every valid combo stores; each invalid combo → 400 with a clear
   code; a `local` ref pointing at another user's **private** asset →
   403/400.
- **HANDOFF:** the final chapter payload + the enforced media matrix.
- **GOTCHAS:** be explicit per row; reject inconsistent combos; always check
   asset ownership; reuse `canAccess` for the story-level gate.

---

## P4.4 — Serve files + visibility gating + soft-delete
- **DEPS:** P4.1 · **core** · est ~20m
- **READ:** feature_request §6, §10 (authz); HANDOUT §6
- **SCOPE:**
    - `server/internal/media/serve.go` + `GET /media/:aid` — stream the file
       from `stored_path` (do **not** buffer a large file into memory).
    - **Gate by the owning story's visibility**: an asset may be referenced by
       a chapter; find the related story and if it is `private`, require
       `canAccess` (owner/admin) else **403**; a `public` story's asset is
       served to anyone. (Map asset → story → visibility; never leak a private
       asset by guessable id — but ids are random and you still must check.)
    - `DELETE /api/media/:aid` — **soft-delete** (`deleted_at`), owner/admin;
       serve skips soft-deleted assets (P7.3 purges them later).
- **VERIFY:**
    ```
    CGO_ENABLED=0 go test ./internal/media -run TestServeGate -v
     ```
   expects: a **public** story's asset → 200 for an anonymous client; a
   **private** story's asset → **403** for anon/other-user but streams for the
   owner/admin; a soft-deleted asset is no longer served.
- **HANDOFF:** `GET /media/:aid` route + the visibility gate.
- **GOTCHAS:** map asset→story→visibility correctly; random ids are not security
   (still authz); stream, don't buffer; soft-delete now, purge later.
