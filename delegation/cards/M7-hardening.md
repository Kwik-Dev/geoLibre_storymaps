# Cards — M7 Hardening (OPTIONAL / later)

Read `delegation/HANDOUT.md §14 (out of scope)` first. Card fields as in M1.
These are **not** part of the MVP — build them only after all core cards in
M1–M6 are `closed`. No PR, no out-of-scope edits.

---

## P7.1 — Upload store abstraction (S3 / Drive proxy)
- **DEPS:** P4.1 · **optional** · est ~30m
- **READ:** feature_request §6, §14 (out of scope: server-mediated upload proxy)
- **SCOPE:** `server/internal/media/store.go` — a `Store` interface
       (`Put(kind, name, r) (ref, error)`, `Get(ref) (io.ReadCloser, error)`,
        `URL(ref) string`, `Delete(ref) error`);
        `LocalStore` (current behavior) as the **default** when unconfigured; an
        `S3Store` (`aws-sdk-go-v2`) — or a server-mediated Drive upload — when
        `STORE_KIND=s3|drive` and creds are set. `Upload` routes to the
       configured store; when nothing is configured it stays **local**.
- **VERIFY:**
      ```
    cd server
    CGO_ENABLED=0 go build ./...
    CGO_ENABLED=0 go test ./internal/media -run 'TestStoreLocal|TestStoreS3' -v
      ```
   expects: a `TestStoreLocal` against `LocalStore` and a `TestStoreS3` against
   a local MinIO or an `httptest` fake proving the interface switches on `STORE_KIND`.
- **HANDOFF:** `media.Store` interface + `LocalStore`.
- **GOTCHAS:** the default must remain **local** with no behavior change; keep
   the `LocalStore` path byte-for-byte identical.

---

## P7.2 — Moderation gate before `public`
- **DEPS:** P3.1 · **optional** · est ~20m
- **READ:** feature_request §14 (out of scope: moderation)
- **SCOPE:** when a story is set to `visibility=public` and
        `MODERATION_REQUIRED=1`, move it to `status='pending'` and keep it out of
        the public list until an admin sets `status='approved'`. Add admin
        approve/reject routes. Don't hide the story from its own owner.
- **VERIFY:**
      ```
    cd server
    CGO_ENABLED=0 go test ./internal/api -run TestModeration -v
      ```
   expects: publish → `pending` → hidden from the public list → admin approve →
   visible; reject → hidden.
- **HANDOFF:** the `pending`↔`approved` status workflow.
- **GOTCHAS:** gate behind env (default **off** so M3 stays simple); never hide a
   story from its owner.

---

## P7.3 — Soft-delete purge cron
- **DEPS:** P4.4, P1.2 · **optional** · est ~20m
- **READ:** feature_request §6, §14
- **SCOPE:** `server/internal/server/purge.go` — a periodic job (ticker
        + on startup) that **hard-deletes** `media_assets`/`stories`/`chapters`
        whose `deleted_at` is set **and** older than `PURGE_TTL` (default 30d).
        Idempotent, in bounded batches.
- **VERIFY:**
      ```
    cd server
    CGO_ENABLED=0 go test ./internal/server -run TestPurge -v
      ```
   expects: with a tiny `PURGE_TTL` (e.g. 1s) and a pre-aged soft-deleted row,
   the row is purged; a freshly soft-deleted row is **not** purged.
- **HANDOFF:** the purge job (`server.StartPurge`).
- **GOTCHAS:** `PURGE_TTL` must be > 0; purge is **irreversible** (only already
   soft-deleted rows); batch the work to avoid huge transactions.

---

## P7.4 — Docs + security-checklist sign-off
- **DEPS:** all · **optional (MVP-deferable)** · est ~20m
- **READ:** feature_request §10 (security checklist); HANDOUT §6
- **SCOPE:**
      - update `README.md` to document the two run modes:
          (a) user-created path — Go server + SQLite + this builder;
          (b) the embedded/static `file://` **no-server fallback** still works.
          Also note Ollama-local agent use and the new env vars.
      - walk the §10 security checklist and record, in `delegation/STATUS.md`,
        the status of **each** item with a pointer to the enforcing
        test/route: XSS/sanitization, path traversal, CSRF `state`, authz gates,
        JWT expiry, upload size cap.
- **VERIFY:**
      - `README.md` renders with no broken relative links;
      - **every** §10 item has a recorded status + a concrete pointer to the
        test/route that enforces it.
- **HANDOFF:** docs + security sign-off.
- **GOTCHAS:** don't mark a checklist item "covered" without a concrete
   test/route to point at.
