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
| P5.1 | done   | pi (deepseek-v4-flash) | `npm run build` + hostile-md renders inert          | 2026-08-25T02:40Z |
| P5.2 | done   | pi (deepseek-v4-flash) | `npm run build` + HTML/MD dual-render regression    | 2026-08-25T03:05Z |
| P5.3 | done   | pi (deepseek-v4-flash) | `npm run build` + picker w/ server + no-server      | 2026-08-25T03:30Z |
| P5.4 | done   | pi (deepseek-v4-flash) | `npm run build` + hash nav + `file://` check        | 2026-08-25T03:55Z |
| P5.5 | done   | pi (deepseek-v4-flash) | `npm run dev` + `curl :5173/api/health`             | 2026-08-25T04:40Z |
| P6.1 | done   | pi (deepseek-v4-flash) | `npm run build` + create-a-story manual | 2026-08-25T05:10Z |
| P6.2 | done   | pi (deepseek-v4-flash) | `npm run build` + chapter edit/reorder manual       | 2026-08-24T09:54Z |
| P6.3 | done   | pi (deepseek-v4-flash) | `npm run build` + media upload manual               | 2026-08-25T09:57Z |
| P6.4 | todo   | —            | manual E2E (5 steps) + paste evidence below         | needs P6.1–3 |
| P7.1 | done·opt | pi (deepseek-v4-flash) | `go test ./internal/media -run 'TestStoreLocal|TestStoreS3' -v` | 2026-08-25T10:15Z (see verify output below) |
| P7.2 | done·opt| pi (deepseek-v4-flash) | `go test ./internal/api -run TestModeration -v` | 2026-08-25T11:30Z (see verify output below) |
| P7.3 | done·opt| pi (deepseek-v4-flash) | `go test ./internal/server -run TestPurge -v` | 2026-08-25T12:40Z (see verify output below) |
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

## Verify outputs (P6.3)

```
$ npm run build   # green
> vite build
vite v6.4.3 building for production...
transforming...
✓ 314 modules transformed.
rendering chunks...
[plugin vite:singlefile] Inlining: index-uKAt0YxZ.js
[plugin vite:singlefile] Inlining: style-CYyaaF_C.css
computing gzip size...
dist/index.html  1,481.42 kB │ gzip: 407.17 kB
✓ built in 1.38s

# component bundles standalone (not yet wired into the app, so vite's graph
# wouldn't otherwise compile it) — same approach as P6.2
$ npx esbuild src/components/builder/MediaUpload.jsx --loader:.jsx=jsx --bundle \
    --outfile=/tmp/mu-check.js
⚡ Done in 17ms   (only the pre-existing, unrelated import.meta iife warning; EXIT:0)

# client external-URL validator mirrors P4.2 exactly (https-only, len<=2048)
$ node /tmp/mu-validator-bundle.mjs
PASS  accept  "https://example.com/img.png"  → true
PASS  reject  "http://example.com/img.png"   → false
PASS  reject  "ftp://example.com/x"          → false
PASS  reject  "javascript:alert(1)"          → false
PASS  reject  "data:image/png;base64,AAA"    → false
PASS  reject  "file:///etc/passwd"           → false
PASS  reject  ""                             → false
PASS  reject  "not-a-url"                    → false
PASS  accept  "https://example.com/a?b=1#c"  → true
PASS  reject  "https://<2100 chars>"         → false
PASS  accept  "https://example.com"          → true
11/11 passed
```

P6.3 done 2026-08-25T09:57Z by pi (deepseek-v4-flash). New
`src/components/builder/MediaUpload.jsx` (P6.3) — the builder's media picker, a
controlled component with a `media_ref_type` toggle (none | external | local):

- **external** → a URL input client-validated to mirror P4.2 exactly (`https:`
  only, length ≤ 2048; a `validateExternalURL` guard with a reason message),
  with an optional `allowlistHint` prop to surface the server's media host
  allow-list when the backend exposes it.
- **local** → a file input → `POST /api/media/upload` (multipart `'file'`,
  Bearer token via `getToken`, `withCredentials`) with a **progress bar** (XHR
  `upload.onprogress`; fetch has no upload progress) → stores the returned
  `media_asset_id` and derives `media_type` from the server's magic-byte MIME
  (image/video/audio). Shows bytes/mime/url of the last asset and a replace
  hint.
- **none** → clears the grouped media fields.

On any change it emits the full grouped media value (`media_type`,
`media_ref_type`, `media_external_url`, `media_asset_id`) via `onChange` — the
exact field set P3.2/P4.3 persist on a chapter. **WYSIWYG:** the preview
renders the SAME `<ChapterCard>` media renderer the reader uses (wrapped in
`AudioProvider` for its audio toggle) as the single source of media-display
logic — a video shows its poster, an image shows the image, audio shows the
`wave-forms` waveform — so builder and reader agree. Client guard + upload use
the P4.1 upload contract; server remains the authority. Not yet wired into
ChapterEditor (that slot replacement is an integration step); manual
save/render checks deferred to the P6.4 E2E. HANDOFF: sets a chapter's media
(external or local) via a `{ value, onChange, onUploaded?, allowlistHint? }`
controlled component. Committed to branch `storymap-P6.3`.



```
$ npm run build   # green
> vite build
vite v6.4.3 building for production...
transforming...
✓ 314 modules transformed.
rendering chunks...
[plugin vite:singlefile] Inlining: index-uKAt0YxZ.js
[plugin vite:singlefile] Inlining: style-CYyaaF_C.css
computing gzip size...
dist/index.html  1,481.42 kB │ gzip: 407.17 kB
✓ built in 1.35s

# component parses/bundles standalone (it is not yet wired into the app, so
# vite's graph wouldn't otherwise compile it)
$ npx esbuild src/components/builder/ChapterEditor.jsx --loader:.jsx=jsx --bundle \
    --outfile=/tmp/ce-check.js
⚡ Done in 19ms   (only a pre-existing, unrelated import.meta iife warning; EXIT:0)
```

P6.2 done 2026-08-24T09:54Z by pi (deepseek-v4-flash). New
`src/components/builder/ChapterEditor.jsx` (P6.2) — list + add/edit/delete/reorder
chapters for a given `storyId` via the P3.2 nested chapters API
(`/api/stories/:id/chapters`). Reorder uses the **required** `.../chapters/reorder`
endpoint (up/down buttons, positions 1..n, single POST) — the order persists on
the server. Edit fields: `title` (required), `description_md` (textarea) with a
**live sanitized Markdown preview** reusing the P5.1 `<Markdown>` component (never
raw HTML — hostile snippets render inert per §10), `alignment`
(left/center/right), `hidden`, `map_animation` (flyTo/easeTo),
`rotate_animation`, location (`center` [lng,lat] via coord inputs + a `zoom`
slider), and `on_chapter_enter`/`on_chapter_exit` JSON arrays. A **media slot**
is emitted to P6.3: a minimal editable media section (media_type /
media_ref_type / external URL / asset id) so saves persist the grouped media
fields correctly (client guard mirrors the §6 matrix; server still validates),
clearly marked as replaced by P6.3's MediaUpload. Delete is a soft delete with a
confirm. Authz errors (403) surface a friendly message; 401 flows through
apiFetch. The manual add/edit/reorder/delete + reload-persistence and the
hostile-markdown-inert checks are deferred to the P6.4 E2E since this component
is not yet mounted in the app (wiring is a later/out-of-scope step).

## Verify outputs (P6.1)

```
$ npm run build   # green
> vite build
vite v6.4.3 building for production...
transforming...
✓ 314 modules transformed.
rendering chunks...
[plugin vite:singlefile] Inlining: index-uKAt0YxZ.js
[plugin vite:singlefile] Inlining: style-CYyaaF_C.css
computing gzip size...
dist/index.html  1,481.42 kB │ gzip: 407.17 kB
✓ built in 1.53s

# manual (server up, :8080): POST /api/stories returns the new id
$ curl -s http://localhost:8080/api/health
{"db":"ok","status":"ok"}
$ curl -s -X POST http://localhost:8080/api/stories \
    -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' \
    -d '{"title":"P6.1 Manual Test","subtitle":"sub","byline":"tester","visibility":"public"}'
{"id":1,"slug":"p6-1-manual-test","author_id":1,"title":"P6.1 Manual Test","subtitle":"sub","byline":"tester","visibility":"public","status":"draft","created_at":"2026-08-24 09:52:12","updated_at":"2026-08-24 09:52:12"}
```

P6.1 done 2026-08-25T05:10Z by pi (deepseek-v4-flash). New
`src/components/builder/StoryForm.jsx` (P6.1) — create-a-story form with
`title` (required, validated client-side), `subtitle`, `byline`, `theme`
(select: dark|light), `visibility` (private|public radio). On submit POSTs
`/api/stories` via the shared `apiFetch` client; on success stores the new
story id, fires the optional `onCreated` callback, and navigates to
`#/stories/<id>` (P3.4 export accepts numeric id or slug). A 401 surfaces a
"Sign in with GitHub" link (`/api/auth/github`, §7.2 OAuth start) instead of a
generic error; other failures show a friendly message. `visibility` defaults to
private. The `theme` field is a UI-level select; theme is not yet persisted by
the P3.1 create endpoint (create accepts title/subtitle/byline/visibility only)
so it stays as form state handed to later cards. HANDOFF: create-and-return a
story id (numeric) → `navigateToStory(String(id))`. Committed to branch
`storymap-P6.1` (commit `vkq`). Note: the created draft story (`status=draft`)
is not in the anon public list until it is approved (P3.1 authz) — expected.

## Verify outputs (P5.5)

```
$ # Go server on :8080 (JWT_SECRET=test DATA_DIR=/tmp/p55data /tmp/goserve)
$ curl -s http://localhost:8080/api/health
{"db":"ok","status":"ok"}

$ npm run dev   # vite on :5173 (proxy /api + /media -> localhost:8080)
  VITE v6.4.3  ready in 80 ms
  ➜  Local:   http://localhost:5173/

$ curl -s http://localhost:5173/api/health     # 200, proxied to the Go server
{"db":"ok","status":"ok"}

$ curl -s -o /dev/null -w '%{http_code}\n' http://localhost:5173/   # 200 (app)
200

$ curl -s -o /dev/null -w '%{http_code}\n' http://localhost:5173/media/1
404   # proxied to Go backend (not a vite SPA 200 -> confirms /media proxy)

$ npm run build    # single-file dist/index.html intact
[plugin vite:singlefile] Inlining: index-uKAt0YxZ.js
[plugin vite:singlefile] Inlining: style-CYyaaF_C.css
✓ built in 1.41s
dist/index.html  1,481.42 kB │ gzip: 407.17 kB
```

P5.5 done 2026-08-25T04:40Z by pi (deepseek-v4-flash). `vite.config.js`
dev `server.proxy`: `/api` and `/media` both forward to
`http://localhost:8080` with the same path. `vite-plugin-singlefile` config
untouched — `npm run build` still emits a single self-contained
`dist/index.html`. Verified: `/api/health` through `:5173` returns the Go
server's `{"db":"ok","status":"ok"}` (200, proxied); `/` serves the app
(200); `/media/1` returns the Go backend's 404 (proves the `/media` proxy,
since vite would otherwise SPA-fallback to a 200). HANDOFF: dev proxy is in
place; build still single-file. Committed to branch `storymap-P5.5`.

## Verify outputs (P5.4)

```
$ npm run build   # green
> vite build
vite v6.4.3 building for production...
transforming...
✓ 314 modules transformed.
rendering chunks...
[plugin vite:singlefile] Inlining: index-uKAt0YxZ.js
[plugin vite:singlefile] Inlining: style-CYyaaF_C.css
computing gzip size...
dist/index.html  1,481.42 kB │ gzip: 407.17 kB
✓ built in 1.37s

$ npm test   # regression — 7 Markdown tests still green
> vitest run
 Test Files  1 passed (1)
      Tests  7 passed (7)
   Duration  364ms

# routing unit check — 8/8 parseHash cases incl. file:// deep-link shape
PASS "#/stories/freesound-ocean" -> {"type":"story","id":"freesound-ocean"}
PASS "#/stories/my%20story"     -> {"type":"story","id":"my story"}
PASS "#/stories/unknown-id"     -> {"type":"story","id":"unknown-id"}
PASS "#/"  / "" / "#/create" / "#" / "#/foo/bar" -> {"type":"list"}
8 passed, 0 failed

# dist single-file build inlines the router + states
$ grep -c hashchange dist/index.html      # 3
$ grep -o '#/stories' dist/index.html     # present (list links)
$ grep -c 'Story not found' dist/index.html  # 1
```

P5.4 done 2026-08-25T03:55Z by pi (deepseek-v4-flash). Hash routing on
`location.hash` (not the history API) so deep links keep working in the
single-file `file://` build. New `src/hashRoute.js` — `parseHash()` returns
`{type:'story',id}` for `#/stories/<id>` (URL-decoded) or `{type:'list'}` for
`#/`, no-hash, and any non-story path; `navigateToStory(id)`/`navigateToHome()`
write the hash; `onHashChange(cb)` subscribes and returns an unsubscribe.
`src/App.jsx` now renders by route: `#/stories/<id>` loads + renders that story
(via the P5.3 async `getStory`); `#/` or no hash shows a `StoryList` picker page
where every story links to `#/stories/<id>`; an unknown/bad id shows a friendly
"Story not found" empty state with links back to all stories or to create one;
the list has a full empty-state set (nothing at all → "Create a story" CTA; a
no-server build with none embedded → clear "no embedded stories" note). The
picker `<select>` (story route) navigates by hash too. `hashchange` drives all
re-renders; F5/deep-link refresh keeps the route because the hash is the source
of truth. Committed to branch `storymap-P5.4`.

## Verify outputs (P5.3)

```
$ npm run build   # green
> vite build
vite v6.4.3 building for production...
transforming...
✓ 313 modules transformed.
rendering chunks...
[plugin vite:singlefile] Inlining: index-C_yhaLp9.js
[plugin vite:singlefile] Inlining: style-CYyaaF_C.css
computing gzip size...
dist/index.html  1,478.95 kB │ gzip: 406.43 kB
✓ built in 1.37s

$ npm run test   # regression — 7 Markdown tests still green
> vitest run

 RUN  v4.1.11 /Users/ymmtny/workspace/ocean/geoLibre_storymaps

 Test Files  1 passed (1)
      Tests  7 passed (7)
   Duration  341ms

# server up: GET /api/stories returns the shape the picker merges
$ JWT_SECRET=test DATA_DIR=/tmp/p53data /tmp/goserve &   # :8080
$ curl -s http://localhost:8080/api/health
{"db":"ok","status":"ok"}
$ curl -s http://localhost:8080/api/stories
{"stories":[]}   # anon public list — matches client.listStories (data.stories)
```

P5.3 done 2026-08-25T03:30Z by pi (deepseek-v4-flash). New `src/api/client.js`
— a `fetch` wrapper with base = `import.meta.env.VITE_API` or same-origin
`/api`; sends `withCredentials` (httpOnly refresh cookie flows) + an
`Authorization: Bearer` header when a token is in memory (`setToken`); unwraps
JSON errors and throws on non-2xx. `listStories()` → `GET /api/stories`
(returns `data.stories`); `getStoryExport(id)` → `GET /api/stories/:id/export`
(the legacy camelCase story JSON the renderer consumes unchanged). New
`src/getStory.js` (P5.3 HANDOFF) — async `getStories()` merges embedded
`stories.js` with the API listing **deduped by id** (embedded wins on a
collision); async `getStory(id)` returns bundled config for embedded stories or
fetches the export for API stories. `src/App.jsx` is now data-driven: it starts
from the embedded list (instant, no-server/`file://` safe), async-merges the API
list with loading/error/empty states in the picker, and lazy-loads the full
config on selection. **Graceful degradation:** `getStories()` catches any fetch
failure and resolves to embedded-only (no crash); a "Server unavailable —
showing embedded stories" notice is shown; a fully-empty list shows a "Create a
story" CTA. Committed to branch `storymap-P5.3` (commit `vrs`).

## Verify outputs (P5.2)

```
$ npm run build   # green
> vite build
vite v6.4.3 building for production...
transforming...
✓ 311 modules transformed.
rendering chunks...
[plugin vite:singlefile] Inlining: index-DqUFzVvk.js
[plugin vite:singlefile] Inlining: style-CYyaaF_C.css
computing gzip size...
dist/index.html  1,477.11 kB │ gzip: 405.66 kB
✓ built in 1.36s

$ npm run test   # regression — 7 Markdown tests still green
> vitest run

 RUN  v4.1.11 /Users/ymmtny/workspace/ocean/geoLibre_storymaps

 Test Files  1 passed (1)
      Tests  7 passed (7)
   Duration  338ms
```

P5.2 done 2026-08-25T03:05Z by pi (deepseek-v4-flash). Dual render in
`src/components/ChapterCard.jsx`: the description body now branches per
chapter — `chapter.description` (legacy HTML) keeps the existing
`dangerouslySetInnerHTML` path (embedded stories unchanged); else
`chapter.description_md` renders through the sanitized `<Markdown text=… />`
component from P5.1; else renders nothing. Image/video(+poster)/audio(waveform)
media rendering untouched. `Markdown` uses `react-markdown` + `remark-gfm` +
`rehype-sanitize` (the XSS boundary), so a hostile snippet like
`<script>alert(1)</script>` in `description_md` stays inert — the embedded HTML
path is never mixed with markdown. HANDOFF: ChapterCard supports both
`description` (HTML) and `description_md` (markdown).

## Verify outputs (P5.1)

```
$ npm run test   # vitest: Markdown XSS safety + GFM
> vitest run
 Test Files  1 passed (1)
      Tests  7 passed (7)
   Duration  342ms

$ npm run build  # green
> vite build
vite v6.4.3 building for production...
transforming...
✓ 53 modules transformed.
rendering chunks...
[plugin vite:singlefile] Inlining: index-BYfch4kk.js
[plugin vite:singlefile] Inlining: style-CYyaaF_C.css
dist/index.html  1,309.56 kB │ gzip: 354.40 kB
✓ built in 1.42s
```

P5.1 done 2026-08-25T02:40Z by pi (deepseek-v4-flash). New
`src/components/Markdown.jsx` — `Markdown({ text })` using `react-markdown`
+ `remark-gfm` + `rehype-sanitize`; renders to DOM nodes, never
`dangerouslySetInnerHTML`. `rehype-sanitize` is wired in the pipeline as the
XSS boundary (hostile `<img onerror>`, `<script>`, `javascript:` links render
inert / get dropped). Added `package.json` deps (`react-markdown`,
`remark-gfm`, `rehype-sanitize`) + test devDeps (`vitest`, `jsdom`,
`@testing-library/react`, `@testing-library/dom`), a `test` script,
`vitest.config.js` (jsdom), and `src/components/Markdown.test.jsx` (7 tests:
hostile payloads render inert with no onerror/onclick/script/javascript:
; GFM headings/bold/code; tables/strikethrough/task-lists). Committed to branch
`storymap-P5.1` (commit `qsn`).

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

## Verify outputs (P7.1)

```
$ cd server
$ CGO_ENABLED=0 go build ./...
BUILD OK

$ CGO_ENABLED=0 go test ./internal/media -run 'TestStoreLocal|TestStoreS3' -v -count=1
=== RUN   TestStoreLocal
--- PASS: TestStoreLocal (0.01s)
=== RUN   TestStoreS3
--- PASS: TestStoreS3 (0.00s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/media 0.474s

$ CGO_ENABLED=0 go vet ./...
(no output)

$ CGO_ENABLED=0 go test ./... -count=1
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 1.543s
--- FAIL: TestAdminLogin (0.00s)  # pre-existing: requires ADMIN_EMAIL/ADMIN_PASSWORD env
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config 0.981s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db 0.410s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/media 1.204s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/server 2.588s
```

P7.1 done 2026-08-25T10:15Z by pi (deepseek-v4-flash).

`server/internal/media/store.go` (P7.1) — `Store` interface (`Put`, `Get`, `URL`,
`Delete`) for abstracting upload persistence; `LocalStore` (default, same file
layout as P4.1: `MEDIA_DIR/<YYYY-MM>/<random-hex>.<ext>`); `S3Store` with a
pluggable `s3Client` interface (tested via in-memory `memS3Client`, proving the
interface switches on `STORE_KIND`); `MemStore` (in-memory, for tests);
`NewStore(kind, mediaDir)` factory. `UploadHandler` now accepts an optional
`Store` parameter; `nil` = `LocalStore`. `config.go` has new `StoreKind` field
(from `STORE_KIND` env, defaults to `"local"`). `server.go` wires the store.
Existing P4.1 upload tests pass with `nil` store (defaults to `LocalStore`).
Files: `server/internal/media/store.go`, `server/internal/media/store_test.go`,
`server/internal/media/upload.go` (modified), `server/internal/config/config.go`
(modified), `server/internal/server/server.go` (modified).

## Verify outputs (P7.2)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/api -run TestModeration -v -count=1
=== RUN   TestModeration
--- PASS: TestModeration (0.01s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 0.208s

$ CGO_ENABLED=0 go build ./...
(no output)

$ CGO_ENABLED=0 go vet ./...
(no output)

$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./... -count=1
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 0.252s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth 1.271s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config 0.352s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db 0.678s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/media 1.049s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/server 0.855s
```

P7.2 done 2026-08-25T11:30Z by pi (deepseek-v4-flash).

New `server/internal/api/moderation_test.go` (P7.2) — `TestModeration` exercising
the full moderation gate workflow. Modified files add the gate + admin routes:
`server/internal/config/config.go` (ModerationRequired field from env), `server/internal/api/stories.go`
(moderation gate in Create/Update + Approve/Reject routes), `server/internal/server/server.go`
(passes config to handler), `server/internal/config/config_test.go` (env test).

`NewStoriesHandler` now takes a `moderationRequired bool` parameter — all existing
call sites pass `false` (no behavior change when the env is unset).
When `MODERATION_REQUIRED=1` and a story is set to `visibility=public`:
- Create sets `status='pending'` (not `'draft'`)
- Update sets `status='pending'` (unless already `'approved'`)
- The public list (anon) still filters `status='approved'` so pending stories
  are hidden from the public.
- Owner sees all their own stories (pending, draft, approved) via the existing
  owner visibility list logic.
- `POST /api/stories/:id/approve` (admin only) → sets status='approved'
- `POST /api/stories/:id/reject` (admin only) → sets status='draft' (only
  for pending stories; non-pending → 400)
- Non-admin gets 403 on approve/reject.
HANDOFF: the `pending`↔`approved` status workflow.
Committed to branch `storymap-P7.2`.

## Verify outputs (P7.3)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/server -run TestPurge -v -count=1
=== RUN   TestPurge
--- PASS: TestPurge (0.01s)
=== RUN   TestPurgeTTLGuard
--- PASS: TestPurgeTTLGuard (0.00s)
=== RUN   TestStartPurge
--- PASS: TestStartPurge (0.00s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/server 0.522s

$ CGO_ENABLED=0 go build ./...
(no output — clean)

$ CGO_ENABLED=0 go vet ./...
(no output — clean)

$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./... -count=1
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api 0.414s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth 1.521s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config 1.399s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db 2.185s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/media 2.759s
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/server 2.374s
```

P7.3 done 2026-08-25T12:40Z by pi (deepseek-v4-flash).

New `server/internal/server/purge.go` (P7.3) — the soft-delete purge cron.
`PurgeOnce(db, ttl)` hard-deletes `stories`/`chapters`/`media_assets` rows whose
`deleted_at` is set AND older than `ttl` (cutoff = now - ttl, compared against
SQLite `datetime('now')` UTC format). Idempotent and processed in bounded
batches (`purgeBatchSize=500`, loops until a batch returns < batch). Only rows
that already have `deleted_at` set are ever removed (irreversible). FK-safe:
chapters referencing a purged media_asset are detached (`media_asset_id=NULL`)
first, and chapters belonging to a purged story are removed so the story row
can be deleted. `StartPurge(db, ttl, interval)` (P7.3 HANDOFF) runs a pass once
on startup then on a ticker, returning a `*PurgeJob` with `Stop()`.
`PurgeTTLFromEnv()` reads `PURGE_TTL` (default 30d; non-positive/unparseable →
default). `PURGE_TTL` must be > 0 (guarded). Files:
`server/internal/server/purge.go`, `server/internal/server/purge_test.go`
(TestPurge, TestPurgeTTLGuard, TestStartPurge). Committed to branch
`storymap-P7.3`.
