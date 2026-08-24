# STATUS — live board (user-created storymaps)

Update **after every card**. A card goes `running` (claimed) → `done` (VERIFY
green, output pasted) ; or `blocked`/`needs-human` with a reason. One card may
be `running` per worker. Do **not** mark `done` without the real VERIFY output.

Legend: `todo` · `running` · `done` · `blocked` · `needs-human` · `opt`(optional)

## Board

| ID   | Status | Owner        | Verify command (see card)                          | Date / note |
|------|--------|--------------|----------------------------------------------------|-------------|
| P1.1 | done   | pi (deepseek-v4-flash) | `CGO_ENABLED=0 go build ./...` + config test | 2026-08-24T13:15Z (re-verified) |
| P1.2 | done   | pi (deepseek-v4-flash) | `go test ./internal/db -run TestMigrate -v`         | 2026-08-24T13:20Z (re-verified) |
| P1.3 | done   | pi (deepseek-v4-flash) | build + `curl /api/health` = 200                    | 2026-08-24T14:24Z (re-verified) |
| P1.4 | done·opt| pi (deepseek-v4-flash) | `go test ./internal/server -run TestSeed -v`       | 2026-08-24T05:25Z (re-verified) |
| P2.1 | done   | pi (deepseek-v4-flash) | `go test ./internal/auth -run TestGitHubOAuth -v` | 2026-08-24T15:05Z |
| P2.2 | done   | pi (deepseek-v4-flash) | `ADMIN_EMAIL=... go test ./internal/auth -run TestAdminLogin -v` | 2026-08-24T16:10Z |
| P2.3 | done   | pi (deepseek-v4-flash) | `go test ./internal/auth -run TestMiddleware -v` | 2026-08-24T17:05Z |
| P2.4 | done   | pi (deepseek-v4-flash) | `go test ./internal/auth -run TestWhoami -v`        | 2026-08-24T18:50Z |
| P3.1 | done   | pi (deepseek-v4-flash) | `go test ./internal/api -run TestStoriesCRUD -v` | 2026-08-24T06:51Z |
| P3.2 | done   | pi (deepseek-v4-flash) | `go test ./internal/api -run TestChapters -v` | 2026-08-24T19:40Z |
| P3.3 | done   | pi (deepseek-v4-flash) | `go test ./internal/api -run TestStoryView -v` | 2026-08-24T20:15Z |
| P3.4 | done   | pi (deepseek-v4-flash) | `go test ./internal/api -run TestExport -v` | 2026-08-24T21:05Z |
| P4.1 | done   | pi (deepseek-v4-flash) | `go test ./internal/media -run TestUpload -v`       | 2026-08-24T22:10Z |
| P4.2 | done   | pi (deepseek-v4-flash) | `go test ./internal/media -run TestExternalURL -v` | 2026-08-24T23:10Z |
| P4.3 | done   | pi (deepseek-v4-flash) | `go test ./internal/api -run TestChapterMedia -v`   | 2026-08-24T23:45Z |
| P4.4 | done   | pi (deepseek-v4-flash) | `go test ./internal/media -run TestServeGate -v`    | 2026-08-25T01:10Z |
| P5.1 | todo   | —            | `npm run build` + hostile-md renders inert          | can start now (no API) |
| P5.2 | todo   | —            | `npm run build` + HTML/MD dual-render regression    | can start after P5.1 |
| P5.3 | todo   | —            | `npm run build` + picker w/ server + no-server      | needs P3.1 |
| P5.4 | todo   | —            | `npm run build` + hash nav + `file://` check        | needs P5.3 |
| P5.5 | todo   | —            | `npm run dev` + `curl :5173/api/health`             | can start now |
| P6.1 | todo   | —            | `npm run build` + create-a-story manual             | needs P5.4,P3.1 |
| P6.2 | todo   | —            | `npm run build` + chapter edit/reorder manual       | needs P6.1 |
| P6.3 | todo   | —            | `npm run build` + media upload manual               | needs P4.3 |
| P6.4 | todo   | —            | manual E2E (5 steps) + paste evidence below         | needs P6.1–3 |
| P7.1 | todo·opt| —           | `go test ./internal/media -run TestStore… -v`      | optional   |
| P7.2 | todo·opt| —           | `go test ./internal/api -run TestModeration -v`    | optional   |
| P7.3 | todo·opt| —           | `go test ./internal/server -run TestPurge -v`      | optional   |
| P7.4 | todo·opt| —           | README renders + §10 checklist all recorded         | optional   |

## Feature-complete gate (MVP)
- [ ] All **core** cards `done` and VERIFY-green
- [ ] `P6.4` manual E2E passed (evidence below)
- [ ] **No-server regression:** an embedded `*-storymap.json` story still
      renders via `file://`/static, with **no server**
- [ ] (optional) M7 hardening cards `done`

## E2E evidence (P6.4)
- <paste: screenshot path / step-by-step result / timestamp>

## Blocked / needs-human log
| When | Card | Who | Reason / question | Resolved |
|------|------|-----|-------------------|----------|
|      |      |     |                   |          |

## Notes / interface changes
(Record any `HANDOFF` interface change here so later cards pick it up.)
- **P3.1 (2026-08-24T06:51Z):** `api.NewStoriesHandler` now takes a second arg
  `*auth.Authenticator` (may be nil) so the public list route can do *optional*
  auth — an owner sees their own private stories in the same public listing.
  Added `auth.Authenticator.UserFromRequest(r)` (optional auth, never 401) to
  `internal/auth/middleware.go`; existing `RequireAuth` behavior unchanged.
  `canAccess(story, user) bool` is the P3.1 HANDOFF (public→true; else
  owner/admin).

## Environment snapshot (re-verify per machine)
- Go `go1.26` · node 24 · `but` (GitButler) · Ollama up at
   `http://127.0.0.1:11434` with `qwen3.8:27b-mlx`
- Go module cache: have `golang-jwt/jwt/v5`, `golang.org/x/crypto`;
    `go get` fetches `chi` + `modernc.org/sqlite` in P1.1/P1.2
- Backend deps to add: `modernc.org/sqlite`, `chi/v5` (+ `h2non/filetype` for
    P4.1 magic bytes)
- Frontend deps to add (P5.1): `react-markdown`, `remark-gfm`, `rehype-sanitize`

---

## Orchestration (per-card / Smithers)

Status of the "delegate in pieces via Smithers" track, separate from the card
table above.

| Item | Status | Result |
|---|---|---|
| `delegation/HANDOUT.md` + `CARDS.md` + `cards/M1..M7` + this board | **done** | 29 self-contained cards (P1.1–P7.4), each with VERIFY + HANDOFF |
| Per-card workflow `storymap-pieces.tsx` | **done** | One `<Task>` node per card, per-card VERIFY self-check, per-milestone `<Approval>` gate (the "notify me"), `ship-report` at the end |
| `smithers graph --compact storymap-pieces.tsx` | **PASS** | `GRAPH_EXIT:0` — structure compiles (env-check → env-gate Approval → per-milestone sequences of card tasks + gates → ship-report) |
| Agent pool → local Ollama `qwen3.8:27b-mlx` | **done** | `.smithers/agents/pi.ts` (`PiAgent`, `provider: "ollama"`, `think:false`, `maxTokens:16000`) + `agents/index.ts` exports `implement`/`review`/`planning`/`research` → `PiAgent`; oneshot: `smithers oneshot --agent pi` |
| Failure handling → auto-fix then escalate to me | **default-ON** | card fail/timeout: Smithers retries the node(native churn) then auto-launches `post-failure.tsx` (diagnose → one-move → durable human request; *I* relay it to you). No per-task wiring. The `workflow.ref(...)` attempt was wrong + reverted. Disable only via `--no-post-failure`/`SMITHERS_POST_FAILURE=0`. |
| Auto-fix + escalate (`recovery.tsx`) | **done** | `.smithers/workflows/recovery.tsx` graphs clean in smthrs@0.33.0: per card one `<ReviewLoop maxIterations=1 onMaxReached="fail">` (producer=agents.implement ollama FIXER, reviewer=agents.review = AI JUDGE with **required `approved` field** = the "AI judge for approval"); per milestone an `<EscalationChain>` (diagnose level + `humanFallback` human gate, self-gating on repair results). Body-rendered (29 ralph + 7 branch). |
| `recovery.tsx` knobs | **set** | `CARD_AUTO_FIX_MAX=1` (one attempt; 3="a few"), `HUMAN_FALLBACK=true` (escalate to human). Change at top of the file. |
| Per-card timeout knob | **done** | `CARD_TIMEOUT_MS` (line 107) wired to every card `timeoutMs`; currently 30 min; tune per machine. |
| Real **execution** smoke (one card) | **BLOCKED → needs decision** | agent tasks shell out to a coding-agent CLI; none of `codex/opencode/claude/cursor` is installed. Run env: `OPENAI_BASE_URL=http://127.0.0.1:11434/v1 OPENAI_API_KEY=ollama`. Install one CLI (e.g. `bun add -g opencode`) then `smithers up .smithers/workflows/storymap-pieces.tsx` (or `smithers oneshot "implement card P1.1…"`). |

Do NOT use the monolithic `storymap-build.tsx` for execution (it's the run that
timed). Use `storymap-pieces.tsx`.

## Verify outputs (P4.3)

RE-VERIFIED 2026-08-24T24:00Z by pi (deepseek-v4-flash) after reviewer
feedback: P4.3 was re-stacked onto P4.2 (chain is now P4.1 -> P4.2 -> P4.3) so
that P4.3's tree genuinely includes P4.2's `server/internal/media/external.go`
(`media.RefType`, `media.CheckMediaRef`, etc.). Build/vet clean and the test now
runs green on its own branch:

```
$ cd server
$ CGO_ENABLED=0 go clean -testcache
$ CGO_ENABLED=0 go test ./internal/api -run TestChapterMedia -v -count=1
=== RUN   TestChapterMedia
--- PASS: TestChapterMedia (0.01s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 0.377s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ CGO_ENABLED=0 go test ./internal/api -count=1  # ok ... 0.290s (all api tests)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./... -count=1  # all packages ok (api, auth, config, db, media, server)
```

(First-run output, 2026-08-24T23:45Z, recorded below.)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/api -run TestChapterMedia -v
=== RUN   TestChapterMedia
--- PASS: TestChapterMedia (0.02s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 0.544s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ CGO_ENABLED=0 go test ./internal/api -v  # all 5 api tests pass
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (api, auth, config, db, media, server)
```

P4.3 done 2026-08-24T24:00Z by pi (deepseek-v4-flash) [re-verified after
reviewer feedback — re-stacked onto P4.2 so external.go is in-tree; build/vet
green on the branch]. Extended chapter
create/update (`server/internal/api/chapters.go`) to accept + validate
`media_type` {image,video,audio,none}, `media_ref_type` {external,local,none},
`media_external_url`, `media_asset_id` and persist exactly the chosen fields.
Enforces the §6 matrix via P4.2's `media.CheckMediaRef` + a new
`validateMedia` helper: none⇒both empty; external⇒concrete type + valid https
URL against `SetAllowedMediaHosts(cfg.AllowedMediaHosts)` (empty=default-allow);
local⇒concrete type + existing asset that is accessible (via `assetAccessible`:
media_assets has no owner column, so accessibility derives from referencing
stories — the author's own story or a public story is accessible, a foreign
private asset → 403, an unassociated/just-uploaded asset is usable). Any
inconsistent combo → 400 with a specific message. Media is a grouped field on
UPDATE: if any media field is present the full combo is re-derived + validated.
`media_asset_id` is parsed as json.RawMessage so explicit null clears it.
Files: `server/internal/api/chapters.go`, `server/internal/api/chapter_media_test.go`
(TestChapterMedia), `server/internal/server/server.go` (route wiring of allow-list).

## Verify outputs (P4.4)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/media -run TestServeGate -v -count=1
=== RUN   TestServeGate
=== RUN   TestServeGate/public-story-asset-served-to-anonymous
=== RUN   TestServeGate/private-owner-asset-denied-anonymous
=== RUN   TestServeGate/private-other-asset-denied-other-user
=== RUN   TestServeGate/private-asset-streams-for-owner
=== RUN   TestServeGate/private-asset-streams-for-admin
=== RUN   TestServeGate/soft-deleted-asset-not-served
=== RUN   TestServeGate/nonexistent-asset-404
=== RUN   TestServeGate/delete-private-asset-owner-only
=== RUN   TestServeGate/delete-public-asset-foreign-denied-admin-allowed
--- PASS: TestServeGate (0.01s)
    --- PASS: TestServeGate/public-story-asset-served-to-anonymous (0.00s)
    --- PASS: TestServeGate/private-owner-asset-denied-anonymous (0.00s)
    --- PASS: TestServeGate/private-other-asset-denied-other-user (0.00s)
    --- PASS: TestServeGate/private-asset-streams-for-owner (0.00s)
    --- PASS: TestServeGate/private-asset-streams-for-admin (0.00s)
    --- PASS: TestServeGate/soft-deleted-asset-not-served (0.00s)
    --- PASS: TestServeGate/nonexistent-asset-404 (0.00s)
    --- PASS: TestServeGate/delete-private-asset-owner-only (0.00s)
    --- PASS: TestServeGate/delete-public-asset-foreign-denied-admin-allowed (0.00s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/media 0.552s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./... -count=1  # all packages ok (api, auth, config, db, media, server)
```

P4.4 done 2026-08-25T01:10Z by pi (deepseek-v4-flash). New
`server/internal/media/serve.go` (P4.4) — `GET /media/:aid` streams the stored
file from disk (never buffered whole) with Content-Type/Length set; `DELETE
/api/media/:aid` soft-deletes (sets `deleted_at`) behind RequireAuth. Since the
locked media_assets schema has NO owner column, access is story-derived: a
`mediaStory` projection + local `mediaCanAccess` (mirrors api.canAccess, P3.1
HANDOFF — public⇒true, else owner/admin) maps asset → referencing live
chapter(s) → story visibility. Serve: public story's asset → anyone; private
story's asset → owner/admin only (else 403); unassociated/just-uploaded asset →
served. Soft-deleted assets → 404 (never served; P7.3 purges later). Delete:
admin or author-of-a-referencing-story only (else 403); unreferenced asset
admin-only. Traversal-defensive resolve of stored_path strictly under mediaDir.
Routes wired in `server/internal/server/server.go`: `s.mux.Get("/media/{aid}")`
outside /api (public, optional auth via auther.UserFromRequest) + `DELETE
/media/{aid}` in the /api group (behind RequireAuth). Files:
`server/internal/media/serve.go`, `server/internal/media/serve_test.go`
(TestServeGate), `server/internal/server/server.go` (route wiring). HANDOFF:
`GET /media/:aid` route + the visibility gate.

## Verify outputs (P4.2)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/media -run TestExternalURL -v
=== RUN   TestExternalURL
=== RUN   TestExternalURL/reject/scheme-relative_hostless
=== RUN   TestExternalURL/reject/scheme_with_spaces
=== RUN   TestExternalURL/reject/http_URL
=== RUN   TestExternalURL/reject/ftp_URL
=== RUN   TestExternalURL/reject/data_URL
=== RUN   TestExternalURL/reject/overlong_URL
=== RUN   TestExternalURL/reject/javascript_scheme
=== RUN   TestExternalURL/reject/file_URL
=== RUN   TestExternalURL/reject/empty_string
=== RUN   TestExternalURL/reject/malformed_(no_scheme)
=== RUN   TestExternalURL/accept-empty-allowlist/max-length_boundary
=== RUN   TestExternalURL/accept-empty-allowlist/bare_https_host
=== RUN   TestExternalURL/accept-empty-allowlist/query+fragment
=== RUN   TestExternalURL/accept-empty-allowlist/port
=== RUN   TestExternalURL/accept-empty-allowlist/subdomain
=== RUN   TestExternalURL/reject-allowlist/partial_host
=== RUN   TestExternalURL/reject-allowlist/wrong_path_prefix
=== RUN   TestExternalURL/reject-allowlist/disallowed_host
=== RUN   TestExternalURL/accept-allowlist/host+path_prefix
=== RUN   TestExternalURL/accept-allowlist/host+path_exact
=== RUN   TestExternalURL/accept-allowlist/path-prefix_boundary
=== RUN   TestExternalURL/accept-allowlist/exact_host
--- PASS: TestExternalURL (0.00s)
    --- PASS: TestExternalURL/reject/ftp_URL (0.00s)
    --- PASS: TestExternalURL/reject/empty_string (0.00s)
    --- PASS: TestExternalURL/reject/malformed_(no_scheme) (0.00s)
    --- PASS: TestExternalURL/reject/scheme_with_spaces (0.00s)
    --- PASS: TestExternalURL/reject/overlong_URL (0.00s)
    --- PASS: TestExternalURL/reject/http_URL (0.00s)
    --- PASS: TestExternalURL/reject/javascript_scheme (0.00s)
    --- PASS: TestExternalURL/reject/data_URL (0.00s)
    --- PASS: TestExternalURL/reject/file_URL (0.00s)
    --- PASS: TestExternalURL/reject/scheme-relative_hostless (0.00s)
    --- PASS: TestExternalURL/accept-empty-allowlist/bare_https_host (0.00s)
    --- PASS: TestExternalURL/accept-empty-allowlist/query+fragment (0.00s)
    --- PASS: TestExternalURL/accept-empty-allowlist/port (0.00s)
    --- PASS: TestExternalURL/accept-empty-allowlist/subdomain (0.00s)
    --- PASS: TestExternalURL/accept-empty-allowlist/max-length_boundary (0.00s)
    --- PASS: TestExternalURL/reject-allowlist/partial_host (0.00s)
    --- PASS: TestExternalURL/reject-allowlist/wrong_path_prefix (0.00s)
    --- PASS: TestExternalURL/reject-allowlist/disallowed_host (0.00s)
    --- PASS: TestExternalURL/accept-allowlist/host+path_prefix (0.00s)
    --- PASS: TestExternalURL/accept-allowlist/host+path_exact (0.00s)
    --- PASS: TestExternalURL/accept-allowlist/path-prefix_boundary (0.00s)
    --- PASS: TestExternalURL/accept-allowlist/exact_host (0.00s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/media 0.521s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ CGO_ENABLED=0 go test ./internal/media -run 'TestExternalURL|TestRefTypeValid|TestCheckMediaRef' -v  # all pass
```

P4.2 done 2026-08-24T23:10Z by pi (deepseek-v4-flash). New
`server/internal/media/external.go` (P4.2) — `ValidateExternalURL(s, allowedHosts)`
rejects http/ftp/javascript:/data:/file:/empty/overlong (>2048, capped before
parsing); accepts any well-formed https URL when the allow-list is empty
(DEFAULT-ALLOW, documented); with a non-empty allow-list enforces host or
host+path-prefix matching. `media.RefType` enum {external, local, none} with
`Valid()`, plus the P4.3 HANDOFF combine-check helper `CheckMediaRef(mediaType,
refType, externalURL, allowedHosts)` enforcing the §6 media matrix (none ⇒ both
empty; external ⇒ concrete type + valid https URL; local ⇒ concrete type + no
external url; unknown ref ⇒ error). Pure, no SSRF/fetch — URL is trusted input,
validated for shape only. Files: `server/internal/media/external.go`,
`server/internal/media/external_test.go` (TestExternalURL, TestRefTypeValid,
TestCheckMediaRef).

## Verify outputs (P4.1)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/media -run TestUpload -v
=== RUN   TestUpload
=== RUN   TestUpload/html-disguised-as-png-rejected
=== RUN   TestUpload/oversize-rejected-with-nothing-written
=== RUN   TestUpload/valid-image-stores-random-name
=== RUN   TestUpload/traversal-neutralized
=== RUN   TestUpload/anonymous-upload-401
--- PASS: TestUpload (0.02s)
    --- PASS: TestUpload/html-disguised-as-png-rejected (0.00s)
    --- PASS: TestUpload/oversize-rejected-with-nothing-written (0.00s)
    --- PASS: TestUpload/valid-image-stores-random-name (0.01s)
    --- PASS: TestUpload/traversal-neutralized (0.00s)
    --- PASS: TestUpload/anonymous-upload-401 (0.00s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/media 0.552s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (api, auth, config, db, media, server)
```

P4.1 done 2026-08-24T22:10Z by pi (deepseek-v4-flash). New
`server/internal/media/upload.go` (P4.1) — `POST /api/media/upload`
(auth required, mounted behind RequireAuth in server.go). MIME by **magic
bytes** (not header/extension); allows only image/video/audio; HTML smuggled as
`.png` → 400. Size capped **before writing** via `http.MaxBytesReader` (+1MB
multipart overhead) with a 413 on over-size; file is **streamed** to disk
chunk-by-chunk (never buffered whole), and anything that fails writes no
physical file. Stored under `MEDIA_DIR/<YYYY-MM>/<crypto/rand hex>.<ext>`;
`stored_path` is always a server-generated **relative** path (traversal
`../../../etc/passwd` is neutralized to basename `passwd`, never used for the
physical path); `filename` = original basename. Inserts the `media_assets` row
(kind/stored_path/filename/bytes/mime). Returns `{"id","url":"/media/<id>",
"bytes","mime"}`. HANDOFF `media.Upload(ctx, w, r) (*Asset, error)`;
`asset.URL`. Files: `server/internal/media/upload.go`,
`server/internal/media/upload_test.go` (TestUpload),
`server/internal/server/server.go` (route wiring).

## Verify outputs (P3.4)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/api -run TestExport -v
=== RUN   TestExport
--- PASS: TestExport (0.01s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 0.461s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (api, auth, config, db, server)
```

P3.4 done 2026-08-24T21:05Z by pi (deepseek-v4-flash). New
`server/internal/api/storyview.go` ExportHandler — `GET /api/stories/:id/export`
returns the StoryView legacy JSON with `Content-Type: application/json` and
`Content-Disposition: attachment; filename="<slug>.storymap.json"`. Access is
public when the story is public; otherwise owner/admin (reusing canAccess). The
route is allowlisted by the middleware, so the handler performs optional auth
(`UserFromRequest`) to enforce the private-story check itself. Lookup accepts
both numeric id and slug. Files: `server/internal/api/storyview.go`
(ExportHandler), `server/internal/api/export_test.go` (TestExport),
`server/internal/server/server.go` (route wiring).

## Verify outputs (P3.3)
```
$ cd server
$ CGO_ENABLED=0 go test ./internal/api -run TestStoryView -v
=== RUN   TestStoryView
--- PASS: TestStoryView (0.01s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 0.605s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ CGO_ENABLED=0 go test ./internal/api -v  # all 3 tests pass (Chapters, StoriesCRUD, StoryView)
```

P3.3 done 2026-08-24T20:15Z by pi (deepseek-v4-flash). New
`server/internal/api/storyview.go` — `StoryView(story, chapters) any` maps a
stories row + its chapters to the exact legacy camelCase story JSON shape the
renderer consumes (top-level title/subtitle/byline/footer/theme/style/
insetWidth/insetHeight/insetPosition/globalView/startSlide/endSlide; chapters[]
with id/title/description(=description_md)/alignment/hidden/location/mapAnimation/
rotateAnimation/onChapterEnter/onChapterExit/source/image|video|audio/autoPlayAudio).
DB snake_case/int → camelCase; empty media fields omitted (omitempty); location
serializes as JSON numbers. Golden fixture `server/internal/api/_test/story_view.golden.json`
asserted by `TestStoryView` (deep-equal + numeric-location check). Files:
`server/internal/api/storyview.go`, `server/internal/api/storyview_test.go`,
`server/internal/api/_test/story_view.golden.json`.

## Verify outputs (P1.1)

```
$ cd server
$ CGO_ENABLED=0 go get modernc.org/sqlite github.com/go-chi/chi/v5
# (deps fetched successfully)

$ CGO_ENABLED=0 go build ./...
# (no output — clean)

$ CGO_ENABLED=0 go test ./internal/config -run TestLoadDefaults -v
=== RUN   TestLoadDefaults
--- PASS: TestLoadDefaults (0.00s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config	(cached)

$ CGO_ENABLED=0 go vet ./...
# (no output — clean)
```

Committed to branch `storymap-P1.1` (commit `nzw`). Files: `server/go.mod`, `server/go.sum`, `server/cmd/server/main.go`, `server/internal/config/config.go`, `server/internal/config/config_test.go`.

Committed to branch `storymap-P1.2` (commit `otm`). Files: `server/internal/db/db.go`, `server/internal/db/migrate.go`, `server/internal/db/migrate_test.go`, `server/internal/db/schema.sql`.

## Verify outputs (P1.2)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/db -run TestMigrate -v
=== RUN   TestMigrate
=== RUN   TestMigrate/in-memory
=== RUN   TestMigrate/temp-file
--- PASS: TestMigrate (0.01s)
    --- PASS: TestMigrate/in-memory (0.00s)
    --- PASS: TestMigrate/temp-file (0.01s)
=== RUN   TestMigrateEnvVarIntegration
--- PASS: TestMigrateEnvVarIntegration (0.00s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db	0.413s

$ CGO_ENABLED=0 go build ./...
# (no output — clean)

$ CGO_ENABLED=0 go vet ./...
# (no output — clean)
```

Re-verified 2026-08-24T12:00Z by pi (deepseek-v4-flash). All steps pass.

Re-verified 2026-08-24T13:20Z by pi (deepseek-v4-flash). Output:

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/db -run TestMigrate -v
=== RUN   TestMigrate
=== RUN   TestMigrate/in-memory
=== RUN   TestMigrate/temp-file
--- PASS: TestMigrate (0.01s)
    --- PASS: TestMigrate/in-memory (0.00s)
    --- PASS: TestMigrate/temp-file (0.01s)
=== RUN   TestMigrateEnvVarIntegration
--- PASS: TestMigrateEnvVarIntegration (0.00s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db	(cached)

$ CGO_ENABLED=0 go build ./...
# (no output — clean)

$ CGO_ENABLED=0 go vet ./...
# (no output — clean)
```

All steps pass. Files already committed to branch `storymap-P1.2` (commit `otm`).

## Verify outputs (P2.1)

```
$ cd server
$ GITHUB_OAUTH_BASE=http://127.0.0.1:18080 GITHUB_API_BASE=http://127.0.0.1:18081 \
    CGO_ENABLED=0 go test ./internal/auth -run TestGitHubOAuth -v
=== RUN   TestGitHubOAuth
--- PASS: TestGitHubOAuth (0.01s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth	0.465s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
```

P2.1 done 2026-08-24T15:05Z by pi (deepseek-v4-flash). Authorize → callback
upserts exactly one user by `github_id`, issues a session (JWT + httpOnly
cookie); replayed state → 400; forged state → 400.

## Verify outputs (P2.2)

```
$ cd server
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./internal/auth -run TestAdminLogin -v
=== RUN   TestAdminLogin
--- PASS: TestAdminLogin (0.72s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth	1.206s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (auth, config, db, server)
```

P2.2 done 2026-08-24T16:10Z by pi (deepseek-v4-flash). Seeded admin logs in →
token + httpOnly cookie; wrong password/unknown email → 401; unconfigured
admin (no env) → 503 without crashing; EnsureAdmin idempotent (exactly one
admin row); refresh rotates session; /api/auth/register and /api/users → 404.

## Verify outputs (P2.3)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/auth -run TestMiddleware -v
=== RUN   TestMiddleware
=== RUN   TestMiddleware/protected_route_without_token_→_401
=== RUN   TestMiddleware/private_write_without_token_→_401
=== RUN   TestMiddleware/valid_bearer_token_→_200
=== RUN   TestMiddleware/valid_refresh_cookie_→_200
=== RUN   TestMiddleware/expired_token_→_401
=== RUN   TestMiddleware/invalid/forged_token_→_401
=== RUN   TestMiddleware/public_allowlist_paths_return_without_token
=== RUN   TestMiddleware/user_attached_to_context
--- PASS: TestMiddleware (0.00s)
    --- PASS: TestMiddleware/protected_route_without_token_→_401 (0.00s)
    --- PASS: TestMiddleware/private_write_without_token_→_401 (0.00s)
    --- PASS: TestMiddleware/valid_bearer_token_→_200 (0.00s)
    --- PASS: TestMiddleware/valid_refresh_cookie_→_200 (0.00s)
    --- PASS: TestMiddleware/expired_token_→_401 (0.00s)
    --- PASS: TestMiddleware/invalid/forged_token_→_401 (0.00s)
    --- PASS: TestMiddleware/public_allowlist_paths_return_without_token (0.00s)
    --- PASS: TestMiddleware/user_attached_to_context (0.00s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth 0.461s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (auth, config, db, server)
```

P2.3 done 2026-08-24T17:05Z by pi (deepseek-v4-flash). New
`server/internal/auth/jwt.go` (HS256 sign/verify, 15m exp, claims
{sub,role,iat,exp}) + `middleware.go` (`Authenticator.RequireAuth`, bearer-then
refresh-cookie fallback, 401 on failure, user attached via `UserFrom(ctx)`;
method-aware allowlist: GET /api/health, /api/auth/github(/callback),
POST /api/auth/admin/login+refresh, GET /api/stories public list, GET
/api/stories/:id/export). Refresh cookie is httpOnly + SameSite=Strict with a
`Secure` flag available via `GitHubConfig.Secure` (prod). NOTE: the protected
`POST /api/stories` (create) requires a token; the public GET /api/stories
*listing* is allowlisted, and the export-path "when public" authz is enforced
by the handler (P3.4), not the middleware.

## Verify outputs (P1.3)

```
$ cd server
$ CGO_ENABLED=0 go build -o /tmp/goserve ./cmd/server
# (no output — clean)

$ mkdir -p /tmp/seedata
$ JWT_SECRET=test DATA_DIR=/tmp/seedata /tmp/goserve &
2026/08/24 11:44:00 ... Server listening on :8080

$ curl -s http://localhost:8080/api/health
{"db":"ok","status":"ok"}

$ curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/
200

$ pkill -f goserve
```

All steps pass. /api/health returns 200 with {"db":"ok","status":"ok"}. Static serve returns 200 from ../dist/index.html.

## Verify outputs (P1.4)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/server -run TestSeed -v
=== RUN   TestSeed
=== RUN   TestSeed/default-off
=== RUN   TestSeed/seed-on
2026/08/24 11:47:23 SeedDemo: seeded demo story "demo-scrollytelling" with 2 chapters
=== RUN   TestSeed/idempotent
--- PASS: TestSeed (0.01s)
    --- PASS: TestSeed/default-off (0.00s)
    --- PASS: TestSeed/seed-on (0.00s)
    --- PASS: TestSeed/idempotent (0.00s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/server	0.527s
```

Committed to branch `storymap-P1.4` (commit `pqw`). Files: `server/internal/server/seed.go`, `server/internal/server/seed_test.go`.

Re-verified 2026-08-24T14:24Z by pi (deepseek-v4-flash). Output:

```
$ cd server
$ CGO_ENABLED=0 go build -o /tmp/goserve ./cmd/server
# (no output — clean)

$ mkdir -p /tmp/seedata
$ JWT_SECRET=test DATA_DIR=/tmp/seedata /tmp/goserve &
2026/08/24 14:23:55 Server listening on :8080

$ curl -s http://localhost:8080/api/health
{"db":"ok","status":"ok"}

$ curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/
200

$ pkill -f goserve
```

All steps pass. /api/health returns 200 with {"db":"ok","status":"ok"}. Static serve returns 200 from ../dist/index.html. Source already committed to branch `storymap-P1.3` (commit `vmk`).

## Verify outputs (P2.4)

Re-verified 2026-08-24T18:50Z by pi (deepseek-v4-flash) after reviewer feedback
(handler was dead code, never mounted). Now the whoami endpoint is wired into the
real router: `server.New` accepts a `*auth.WhoamiHandler`, `cmd/server/main.go`
constructs it and passes it through, and every `/api` route (including
`GET /api/auth/whoami`) runs behind `RequireAuth`. Added a server-level route test
(`internal/server/whoami_route_test.go`) that proves the endpoint is served by the
production mux. Output:

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/auth -run TestWhoami -v
=== RUN   TestWhoami
--- PASS: TestWhoami (0.01s)
=== RUN   TestWhoamiAdminFlag
--- PASS: TestWhoamiAdminFlag (0.01s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth (cached)

$ CGO_ENABLED=0 go test ./internal/server -run TestWhoamiRoute -v
=== RUN   TestWhoamiRoute
--- PASS: TestWhoamiRoute (0.01s)   # via real router: no token→401, valid token→200, no password_hash
PASS
ok   github.com/Kwik-Dev/geoLibre_storymaps/server/internal/server 0.517s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (auth, config, db, server)
```

P2.4 done 2026-08-24T18:50Z by pi (deepseek-v4-flash). GET /api/auth/whoami
under RequireAuth returns `{"id","github_login","admin_email","role","admin"}`
(incl. role + admin flag derived from role); without a token → 401; the response
never contains password_hash. Files: `internal/auth/whoami.go` (WhoamiHandler),
`internal/auth/whoami_test.go`, `internal/server/server.go` (New wires whoami +
RequireAuth), `cmd/server/main.go` (passes whoami handler), new
`internal/server/whoami_route_test.go` (production-wiring regression test).

## Verify outputs (P3.2)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/api -run TestChapters -v
=== RUN   TestChapters
--- PASS: TestChapters (0.01s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 0.802s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (api, auth, config, db, server)
```

P3.2 done 2026-08-24T19:40Z by pi (deepseek-v4-flash). Nested chapters CRUD +
reorder under /api/stories/:id/chapters: every op loads the parent story and
runs canAccess (P3.1 HANDOFF) → non-owner on a private story gets 403 for list/
create/get/update/delete/reorder. Create auto-assigns position
COALESCE(MAX(position),0)+1 (3 chapters → 1,2,3); reorder is a single
transaction that rejects ids not belonging to the story (400); location is
validated JSONB (finite lng∈[-180,180], lat∈[-90,90], numeric zoom, optional
pitch∈[0,85]/bearing∈[0,360]) → invalid → 400; delete is soft. Files:
`server/internal/api/chapters.go`, `server/internal/api/chapters_test.go`,
`server/internal/server/server.go` (route wiring).

## Verify outputs (P3.1)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/api -run TestStoriesCRUD -v
=== RUN   TestStoriesCRUD
--- PASS: TestStoriesCRUD (0.01s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 0.690s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (api, auth, config, db, server)
```

P3.1 done 2026-08-24T06:51Z by pi (deepseek-v4-flash). Stories CRUD +
visibility/authz: anon list shows only public+approved; owner sees own draft;
non-owner/non-admin GET/PUT/DELETE of a private story → 403; admin sees all;
soft-delete hides from lists but keeps the row; slug unique (case-insensitive
index). Files: `server/internal/api/stories.go`, `server/internal/api/stories_test.go`,
`server/internal/server/server.go` (route wiring), `server/internal/auth/middleware.go`
(added `UserFromRequest` optional-auth helper), `server/internal/db/schema.sql`
(case-insensitive unique slug index).
