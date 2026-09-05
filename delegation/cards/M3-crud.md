# Cards — M3 Stories + Chapters CRUD + JSON adapter

Read `delegation/HANDOUT.md §0/§4` first. Card fields as in M1. A card is
`closed` only when its `VERIFY` passes. No PR, no out-of-scope edits.

---

## P3.1 — Stories resource: CRUD + visibility/authz
- **DEPS:** P2.3 · **core** · est ~30m
- **READ:** feature_request §7 (stories), §10 (authz); HANDOUT §4, §7
- **SCOPE:**
    - `server/internal/api/stories.go` + route registration on the chi mux
      (registered behind `auth.RequireAuth`, with the public routes excepted per
       P2.3).
    - `GET /api/stories` — **list**: anon → only `visibility='public'` AND
      `status='approved'`; an owner → also their own (any visibility/status);
      `role='admin'` → all. Filter **in SQL**, not in Go post-filter.
    - `POST /api/stories` — `title` (non-empty, required), `subtitle`/`byline`
       optional, `visibility` ∈ {private,public} (default private). New story →
      `status='draft'`, `author_id` = the ctx user, `slug` generated +
       unique (case-insensitive unique index).
    - `GET /api/stories/:id` — load + **authz** via `canAccess(story,user)`.
    - `PUT /api/stories/:id` — owner or admin; partial update.
    - `DELETE /api/stories/:id` — **soft-delete** (`deleted_at`), owner or admin.
    - `canAccess(story, user) bool`: `public` → true; else owner or
      `role='admin'` only. 403 otherwise.
- **VERIFY:**
    ```
    CGO_ENABLED=0 go test ./internal/api -run TestStoriesCRUD -v
     ```
   expects: create a story; anon listing shows only public+approved; an owner's
   own draft is visible to that owner; a non-owner/non-admin `GET` of a private
   story → **403**; admin sees all; a soft-deleted story disappears from lists.
- **HANDOFF:** `api.Stories` handler group; `canAccess(story, user) bool`.
- **GOTCHAS:** apply the visibility filter in SQL; enforce `slug` uniqueness at
   the DB level; soft-delete only (not hard).

---

## P3.2 — Chapters resource (nested) + reorder
- **DEPS:** P3.1 · **core** · est ~30m
- **READ:** feature_request §4 (chapters), §7; HANDOUT §4, §7
- **SCOPE:**
    - `server/internal/api/chapters.go` + routes under
      `/api/stories/:id/chapters` — **every** op first loads the parent story and
      runs `canAccess` (P3.1).
    - `GET /api/stories/:id/chapters` — ordered by `position, created_at`.
    - `POST /api/stories/:id/chapters` — create: `position = COALESCE(MAX(position),0)+1`,
      `title` non-empty, `description_md` optional, `alignment` ∈
      {left,center,right} (default center), `map_animation`/`rotate_animation`/
      `on_chapter_enter`/`on_chapter_exit`/`source` optional,
       `location` = **JSONB validated**: `center:[lng, lat]` with `lng`/`lat`
       finite & in range, `zoom` numeric, optional `pitch`/`bearing` in range.
    - `GET/PUT/DELETE /api/stories/:id/chapters/:cid` — same authz; delete is
       soft.
    - `POST /api/stories/:id/chapters/reorder` — body `[{id, position}]`,
       applied **atomically in one transaction**; reject ids that aren't this
       story's chapters.
- **VERIFY:**
    ```
    CGO_ENABLED=0 go test ./internal/api -run TestChapters -v
     ```
   expects: add 3 chapters → positions 1,2,3; a reorder call swaps order and
   persists; an invalid `location` JSON → **400** (rejects non-finite coords); a
   non-owner on a private story → **403** for every chapter op.
- **HANDOFF:** `api.Chapters` handler group + `reorder`; positions auto-assigned.
- **GOTCHAS:** `location` must reject NaN/Infinite; reorder is a single
   transaction (all-or-nothing); authz on **every** chapter route, not just
   create.

---

## P3.3 — DB → camelCase story JSON adapter (legacy shape)
- **DEPS:** P3.1 · **core** · est ~25m
- **READ:** feature_request §3/§5 (the legacy story JSON shape); HANDOUT §7
- **SCOPE:**
    - `server/internal/api/storyview.go` — `StoryView(story, chapters)` returns
       the **exact legacy JSON** the renderer consumes:
         - top level: `title`, `subtitle`, `byline`, `footer`, `theme`, `style`,
           `insetWidth`, `insetHeight`, `insetPosition`, `globalView`,
          `startSlide`, `endSlide`.
         - `chapters[]`: `id`, `title`, `description` (= `description_md` as a
           **string** of markdown for a user story), `alignment`, `hidden`,
           `location:{center,lng,lat,zoom,pitch,bearing}`, `mapAnimation`,
           `rotateAnimation`, `onChapterEnter`, `onChapterExit`, `source`,
          `media`/media fields, `autoPlayAudio`.
       Map snake_case/int → camelCase; omit empty media fields.
    - A golden fixture `server/internal/api/_test/story_view.golden.json`
      built from a known story, asserted by the test.
- **VERIFY:**
    ```
    CGO_ENABLED=0 go test ./internal/api -run TestStoryView -v
     ```
   expects: a round-tripped `StoryView` **deep-equals** the golden JSON with the
   right camelCase keys, and `location` fields are JSON **numbers**, not strings.
- **HANDOFF:** `StoryView(story, chapters) any` (or a typed struct that
   marshals to the golden shape).
- **GOTCHAS:** match key casing **exactly** or the frontend renderer breaks;
   numbers must serialize as numbers; nil media fields must be omitted, not
   null.

---

## P3.4 — Export endpoint (legacy story JSON)
- **DEPS:** P3.3 · **core** · est ~10m
- **READ:** feature_request §7; HANDOUT §7
- **SCOPE:** `GET /api/stories/:id/export` → the `StoryView` JSON with
    `Content-Type: application/json` and
     `Content-Disposition: attachment; filename="<slug>.storymap.json"`.
   Access: public if the story is public; else owner/admin (reuse `canAccess`).
- **VERIFY:**
    ```
    curl -s "http://localhost:8080/api/stories/<slug>/export" -H "Authorization: Bearer <token>" | jq .
    ```
   or `CGO_ENABLED=0 go test ./internal/api -run TestExport -v`
   expects: 200 JSON matching the golden legacy shape with the right
   `Content-Disposition`; a private story exported by a non-owner → **403**.
- **HANDOFF:** none (this endpoint is the user's "download my storymap" path).
- **GOTCHAS:** derive a sanitized filename from `slug`; JSON content type.
