# Handout — user-created storymaps, delegated in pieces

This is the **master brief** for implementing the user-created storymaps feature.
It exists because a single large Smithers run kept **timing out with no progress**.
The fix: break the work into **small, self-contained, independently-verifiable
"cards"** (see `CARDS.md`) that each run to green in one short session — so a
worker (a local Ollama model, a pi/claude session, or a human) can take **one card
at a time** and finish it without a long-lived run.

> **Rule of the game:** a worker reads *this file + exactly one card*, implements
> only that card's `SCOPE`, runs its `VERIFY` until it passes, then marks the
> card `closed` in `STATUS.md`. No card may open a PR or touch files outside its
> `SCOPE`. Merge/PR is a human step, done at the end.

## 0. What you must know first

- **The spec / source of truth is `feature_request_user_created_storymap.md`.**
  Read it once. Every card cites a section of it (`§4` data model, `§6` media,
  `§7` backend, `§8` frontend, `§10` security, `§11` phases). Do **not** change the
  locked decisions below; if one is wrong, stop and ask the human.
- **Locked decisions (do not contradict):**
  1. **GitHub OAuth2** sign-in for normal users; **local email/password login is
     admin-only** (seeded from env on startup). **No public self-registration.**
  2. **Per-story visibility** is `private` | `public`, chosen by the owner.
  3. **Per-chapter media** is EITHER an **external URL the user supplies**
     (must be `https://`, length-bounded, optional host allow-list) OR a file
     **uploaded to the Go server's local disk**.
  4. **Existing embedded `*-storymap.json` stories keep working with no server
     at all** (static `file://` / no-server fallback). Their descriptions stay
     **HTML** (`dangerouslySetInnerHTML`). User stories use **Markdown**
     (rendered + sanitized in React). No HTML→Markdown migration of old stories.
  5. **Backend = Go + SQLite + a "simple" REST API.**
- **Repo / branch:** `/Users/ymmtny/workspace/ocean/geoLibre_storymaps`,
  branch `gitbutler/workspace`. Use GitButler (`but`) for git writes; **do not**
  create commits/PRs unless the human says so.

## 1. Environment a worker needs (verify once per machine)

| Tool | Need | Check |
|---|---|---|
| Go 1.26+ | backend; **no cgo** (pure-Go SQLite) | `go version` |
| Node 24 + npm | frontend build | `node -v && npm -v` |
| `git` / `but` | git writes | `but --version` |
| **Ollama** (worker model) | the agent that writes code | server up + model present |
| `modernc.org/sqlite`, `github.com/go-chi/chi/v5` | backend deps | `cd server && go build ./...` fetches them |

- `golang-jwt/jwt/v5` and `golang.org/x/crypto` are already in the Go module cache.
- Frontend deps to add in `src/`: `react-markdown`, `remark-gfm`, `rehype-sanitize`.
- The four embedded stories are `freesound-ocean`, `pixabay-images`,
  `pixabay-videos`, `surfing` (static JSON in `src` / `public`). Media lives in
  `public/{audio,images,videos}` today.

## 2. Worker model — local Ollama, `qwen3.8:27b-mlx` (why the timeouts stopped)

The implementing agent runs **entirely local**, so nothing waits on an external
API. Verified present on this machine:

```
endpoint   http://127.0.0.1:11434            # Ollama server (OpenAI-compatible)
chat URL   http://127.0.0.1:11434/v1/chat/completions
models list http://127.0.0.1:11434/v1/models  →  qwen3.8:27b-mlx, gemma4:26b-mlx, nemotron-3.5-lightning:30b-a3b-mlx
api key    "ollama"            (Ollama ignores it, but an OpenAI-compatible client needs a non-empty value)
model      qwen3.8:27b-mlx     (18 GB, capabilities: completion, vision, tools, thinking)
```

**Gotcha (read this):** this model **reasons before answering**. A call with a
tiny `max_tokens` spent the whole budget on hidden reasoning and returned an
*empty* answer (`finish_reason: length`). For a coding agent this means:
- give it a **generous `max_tokens` (≈ 8000–16000)**, **or** disable thinking;
- to disable thinking via the OpenAI-compatible shim, pass
  `"options": { "think": false }` (Ollama's option); otherwise keep the token
  budget high so the real answer + tool calls fit.
- it supports `tools` → you can drive it as a real agent (read file → propose
  diff → run verify), which is how a card is executed.

### 2a. How to run ONE card (three equivalent ways — pick one)

1. **Local agent loop via Ollama (most robust, zero external dep, zero timeout).**
   Point any OpenAI-compatible agent at the endpoint above with model
   `qwen3.8:27b-mlx`. Feed it: "Read `delegation/HANDOUT.md` and
   `delegation/CARDS.md` card **Pxx.y**. Implement only that card's SCOPE. Run
   its VERIFY until green. Update `delegation/STATUS.md`. Stop."
   One card per loop → short → no timeout.
2. **A short coding-agent session** (pi, claude, or a `smithers oneshot`)
   with the provider pointed at Ollama. Same prompt as above.
   - To point a Smithers agent pool here, set (and confirm with `smithers tree
     show`):
       ```
       export OPENAI_BASE_URL=http://127.0.0.1:11434/v1
       export OPENAI_API_KEY=ollama
       export SMARTS_MODEL=qwen3.8:27b-mlx
       ```
     (If your Smithers version reads `SMITHERS_PROVIDER_*` instead, use
     `SMITHERS_PROVIDER_NAME=openai-compatible SMITHERS_PROVIDER_URL=http://127.0.0.1:11434/v1
     SMITHERS_PROVIDER_API_KEY=ollama`. **Do not assume which — verify first.**)
     **Do not** use the big `smithers up .smithers/workflows/storymap-build.tsx`
     run for this; it is the thing that timed out. Use per-card sessions.
3. **A human** does the card by hand. The card is written to be doable that way.

**Do one card, verify, mark closed, then pick the next by dependency (§3).**

## 3. Delegation protocol + ordering (the "in pieces" map)

### 3a. Per-card loop (every worker, every card)

1. Read `HANDOUT.md §0` + the card in `CARDS.md`.
2. Check the card's `DEPS` are `closed` in `STATUS.md`. If not, block.
3. Implement **only** the `SCOPE`. Do not open a PR. Do not refactor elsewhere.
4. Run the card's `VERIFY` command **until it passes**; paste the real output
   into `STATUS.md`.
5. Set the card `closed` with your name, the verify command, and the timestamp in
   `STATUS.md`. Mark it `running` first so two workers don't collide.
6. **Hand off:** the `HANDOFF` line of the card is the interface the next card
   consumes. If you had to change that interface, note it in `STATUS.md` and
   tell the human.

### 3b. Ordering — what can run in parallel, what is the spine

The dependency chains (DAG). Edges mean "must finish before."

**Backend spine (sequential — these share files and must land in order):**
```
P1.1 → P1.2 → P1.3 ─┐
                 └→ P1.4 (optional, after P1.3)
P1.3 → P2.1 → P2.2 → P2.3 → P2.4
P2.3 → P3.1 → P3.2 → P3.3
P3.1, P4.1/P4.2 → P4.3 ; P4.1 → P4.4
P4.x → M5 (P5.x) → M6 (P6.x) → M7 (P7.x, optional)
```

**Frontend pieces that are API-independent → can run in PARALLEL with the
backend spine** (great for 2–3 workers):
- `P5.1` Markdown renderer, `P5.2` dual-render in `ChapterCard`, `P5.5` vite dev
  proxy — depend on **no** backend. Start these immediately.
- `P5.3` data-driven picker + `P5.4` hash routing need the M3 API + `P5.2`.
- `P6.*` need M5 + the M3/M4 API.

**Suggested 3-worker split (avoid merge conflicts by partitioning files):**
- **Worker A — backend:** the spine `P1.1 … P1.3, P2.1 … P2.4, P3.1 … P3.3,
  P4.1, P4.2, P4.3, P4.4`.
- **Worker B — frontend:** `P5.1, P5.2, P5.5` now; then `P5.3, P5.4, P6.1,
  P6.2, P6.3` after the API exists.
- **Worker C — optional/hardening:** `P1.4, P3.4, P7.1 … P7.4` (only after the
  MVP closes); or stays idle until M7 is in scope.

**Sequential single-worker is safest** if only one worker runs: just walk
`CARDS.md` top-to-bottom. Parallel only helps if you partition files so two
workers never edit the same file concurrently.

### 3c. Definition of done (feature complete, MVP)

All **core** cards `closed` and `VERIFY` green, plus a **manual end-to-end**:
1. Server up (`go run ./cmd/server`), `GET /api/health` → 200.
2. `npm run dev` + `npm run build` both green.
3. A GitHub user (or seeded admin) creates a story with ≥1 chapter that has
   either an external-URL media or an uploaded file, sets it `public`, and it
   appears in the `#/` picker and renders at `#/stories/<id>`.
4. **Regression:** an embedded `*-storymap.json` story still renders with **no
   server** (no-server / `file://` mode intact).

## 4. Shared contracts (so parallel cards don't drift)

**DB schema (from `§4`; `internal/db/schema.sql`, versioned):**
- `users(id, github_login, github_id, admin_email, password_hash, role, created_at)`
    `role ∈ {user, admin}`; `github_id`/`admin_email` nullable & unique, one is set.
- `stories(id, slug, author_id→users.id, title, subtitle, byline, visibility
    {private,public}, status {draft,pending,approved}, global_view jsonb,
    created_at, updated_at, deleted_at)`.
- `chapters(id, story_id→stories.id, position int, title, description_md,
    alignment {left,center,right}, hidden bool, location jsonb, media_type
    {image,video,audio,none}, media_ref_type {external,local,none},
    media_external_url, media_asset_id→media_assets.id, created_at, updated_at,
    deleted_at)`.
- `media_assets(id, kind {image,video,audio}, stored_path, filename, bytes,
    mime, created_at, deleted_at)`.

**REST surface (from `§7`; camelCase, matches the legacy story JSON shape):**
```
# auth
GET  /api/auth/github            start GitHub OAuth (sets 1×-use state)
GET  /api/auth/github/callback   finish, upsert user by github_id, issue JWT
POST /api/auth/admin/login       admin-only, bcrypt vs env-seeded admin
POST /api/auth/admin/refresh     rotate refresh (httpOnly cookie)
GET  /api/auth/whoami
# stories / chapters (all require a session except list-public + export-public)
GET  /api/stories                list (anon: public only; owner: +own; admin: all)
POST /api/stories                create (title, subtitle, byline, visibility)
GET  /api/stories/:id            get (authz by visibility/owner/admin)
PUT  /api/stories/:id            update
DELETE /api/stories/:id          soft-delete
GET  /api/stories/:id/chapters   list
POST /api/stories/:id/chapters   create
GET  /api/stories/:id/chapters/:cid
PUT  /api/stories/:id/chapters/:cid
DELETE /api/stories/:id/chapters/:cid
POST /api/stories/:id/chapters/reorder   [ {id, position} ]
GET  /api/stories/:id/export     legacy story JSON (Content-Disposition)
# media
POST /api/media/upload          multipart 'file'; magic-byte MIME; size cap; random basename
GET  /media/:aid                serve file; gated by owning story visibility if private
DELETE /api/media/:aid          soft-delete
# ops
GET  /api/health
```

**Media rule (from `§6`):** on a chapter, exactly one of these holds:
`media_ref_type=external` → `media_external_url` set + validated
(`https://` only, length ≤ 2048, host must be in `ALLOWED_MEDIA_HOSTS` when that
env is non-empty); `media_ref_type=local` → `media_asset_id` set (a valid
`media_assets` row the author owns or is public); `media_ref_type=none` → both
empty. `media_type` ∈ {image,video,audio}; if `none`, ref must be `none`.

**Auth rules:** every `/api/*` route except `GET /api/health`,
`GET /api/auth/github*`, `POST /api/auth/admin/*`, `GET /api/stories` (public
list) and `GET /api/stories/:id/export` (when the target story is public) requires
a valid Bearer JWT (else **401**). Authorization for private stories: owner or
`role=admin` only (else **403**).

## 5. File map (the only files any card may touch, by card)

```
server/go.mod, server/cmd/server/main.go, server/internal/config/*.go        (P1.1)
server/internal/db/{db.go, schema.sql, migrate.go}                          (P1.2, P1.4 seed)
server/internal/server/server.go, .../seed.go                               (P1.3, P1.4)
server/internal/auth/{oauth.go, github.go, admin.go, jwt.go, middleware.go} (P2.x)
server/internal/api/{stories.go, chapters.go, storyview.go}                 (P3.x)
server/internal/media/{upload.go, external.go, serve.go, store.go}          (P4.x / P7.1)
server/internal/server/moderation.go, .../purge.go                          (P7.2, P7.3)
src/components/Markdown.jsx                                                 (P5.1)
src/components/ChapterCard.jsx                                              (P5.2)
vite.config.js (dev proxy)                                                  (P5.5)
src/api/client.js, src/App.jsx, src/getStory.js, src/hashRoute.js          (P5.3, P5.4)
src/components/builder/{StoryForm,ChapterEditor,MediaUpload}.jsx           (P6.x)
README.md                                                                   (P7.4)
package.json (frontend deps)                                                (P5.1 — npm i the 3 md libs)
```
A card edits only its own rows here; that partition keeps workers from colliding.

## 6. Verification + status

- A card is `closed` only when its `VERIFY` command **passes** (paste the real
  output into `STATUS.md`). "It probably works" is not `closed`.
- **`STATUS.md`** is the live board: every card's state, owner, last verify
  output. One card may be `running` per worker; a worker claims by flipping its
  card `running` before starting.
- **No PR / no commit** by any card. Integration + commit/PR is a human step
  (`but`, per `AGENTS.md`).

## 7. Known environment facts (verified 2026-08-17)

- Go `go1.26` (darwin/arm64), `sqlite3` CLI 3.51.0. No `server/` dir yet.
- Module cache has `golang-jwt/jwt/v5@v5.3.1`, `golang.org/x/crypto@v0.53.0`;
  `github.com/go-chi/chi/v5` + `modernc.org/sqlite` **not** cached (the first
  `go get` in `P1.1`/`P1.2` fetches them via `proxy.golang.org`; reachable).
- Frontend stack: React 18 + Vite 6 + MapLibre GL 5 + scrollama 3; builds a
  single-file `dist/index.html` (`vite-plugin-singlefile`).
- Ollama server is up with `qwen3.8:27b-mlx` (see §2).

## 8. Escalation / "notify the human" triggers

Stop the card and tell the human if:
- a card changes a `HANDOFF` interface another card already consumed;
- a locked decision (§0) turns out wrong;
- a `VERIFY` can't pass after 3 attempts (post the last output in `STATUS.md`);
- a security gap appears in `§10`; or any destructive / irreversible action.
