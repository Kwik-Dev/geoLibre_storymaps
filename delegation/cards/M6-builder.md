# Cards — M6 Builder UI (StoryForm, ChapterEditor, MediaUpload, E2E)

Read `delegation/HANDOUT.md §0/§8` first. Card fields as in M1. A card is
`closed` only when its `VERIFY` passes. No PR, no out-of-scope edits.
The end result: a user creates a story, adds chapters, attaches media, sets
visibility, and opens it at `#/stories/<id>`.

---

## P6.1 — StoryForm (create a story)
- **DEPS:** P5.4, P3.1 · **core** · est ~25m
- **READ:** feature_request §8; HANDOUT §7
- **SCOPE:** `src/components/builder/StoryForm.jsx` — fields:
      `title` (required), `subtitle`, `byline`, `theme` (select from the existing
       themes), `visibility` (private|public). On submit → `POST /api/stories`
      (via `client.js`); on success store the new id, let the user set/publish
       visibility, then navigate to `#/stories/<id>`.
- **VERIFY:**
     ```
    npm run build   # green
    # manual (server up):
   #  - create a story → id returned → it appears in the `#/` picker → opens.
     ```
- **HANDOFF:** create-and-return a story id.
- **GOTCHAS:** on a 401 from the API surface a "Sign in with GitHub" link (the
   OAuth flow from §7); validate `title` is non-empty client-side.

---

## P6.2 — ChapterEditor (add / edit / reorder / preview)
- **DEPS:** P6.1, P3.2, P3.3 · **core** · est ~35m (biggest card)
- **READ:** feature_request §8; HANDOUT §7
- **SCOPE:** `src/components/builder/ChapterEditor.jsx` for a given story id:
      - list existing chapters (via API); add / edit / delete.
      - **reorder** (drag or up/down buttons) → `POST .../chapters/reorder`.
      - edit `title`, `description_md` (textarea) with a **live Markdown preview**
        that reuses the P5.1 `<Markdown>` component (sanitized — never inject raw
       user HTML into the preview).
      - location pick: `center` (`[lng, lat]`) via coord inputs or a small
        MapLibre picker + a `zoom` slider; plus `map_animation`,
        `rotate_animation`, `on_chapter_enter`, `on_chapter_exit`,
        `alignment`, `hidden`.
      - a media slot handed to P6.3; on save, update the chapter.
- **VERIFY:**
     ```
    npm run build   # green
    # manual (server up):
   #  - add / edit / reorder / delete a chapter; reload → they persist in order.
   #  - the live preview shows **sanitized** markdown (a hostile snippet is inert).
     ```
- **HANDOFF:** edits a story's chapters via the API; emits the media slot to P6.3.
- **GOTCHAS:** reorder **must** call the reorder endpoint; the preview reuses
   P5.1 (sanitized), no raw-HTML injection; authz errors surface a friendly
    message.

---

## P6.3 — MediaUpload (external URL or local file)
- **DEPS:** P4.1, P4.2, P4.3, P6.2 · **core** · est ~25m
- **READ:** feature_request §6, §8; HANDOUT §6
- **SCOPE:** `src/components/builder/MediaUpload.jsx` — a toggle for
      `media_ref_type`:
       - **external** → a URL input, client-validated to mirror P4.2
          (`https` only, length ≤ 2048, hint the allow-list if exposed).
       - **local** → a file input → `POST /api/media/upload` (multipart) → store
          the returned `media_asset_id` + `url`.
       - **none**.
       On save, the chosen ref writes to the chapter. Display matches the reader:
       a `<video poster>` for video, and the existing `wave-forms` waveform +
       `<audio>` for audio — **reuse the existing `ChapterCard` media renderer as
       the single source of media-display logic** so builder and reader agree
       (WYSIWYG).
- **VERIFY:**
     ```
    npm run build   # green
    # manual (server up):
   #  - external URL → saves and renders
   #  - file upload → 200 + url, and in the reader the video shows its poster,
   #    audio shows a waveform   (proves the media path end-to-end).
     ```
- **HANDOFF:** sets a chapter's media (external or local).
- **GOTCHAS:** large files need a progress UI; the client URL check must match
   P4.2 exactly; don't duplicate media render logic — reuse `ChapterCard`.

---

## P6.4 — End-to-end integration sign-off (M6 gate)
- **DEPS:** P6.1, P6.2, P6.3 (all `closed`) · **core (verification)** · est ~15m
- **READ:** HANDOUT §3c; the whole `feature_request_user_created_storymap.md`
- **SCOPE:** no code — run the full happy path against a **running server** and
   record the evidence (screenshot path / notes) in `delegation/STATUS.md`:
     1. open the builder → create "My Story" (draft)
     2. add ≥ 1 chapter: a title, a location, a markdown description, and either
        an external-URL media or an uploaded file
     3. set `visibility=public`
     4. it appears in the `#/` picker
     5. open `#/stories/<id>` → it renders (markdown + map + media)
- **VERIFY:** all five steps succeed **by hand**; paste the evidence into
   `delegation/STATUS.md`.
- **HANDOFF:** M6 acceptance passed.
- **GOTCHAS:** do **not** mark this `closed` unless P6.1–P6.3 are `closed` **and**
   the server is actually running; "I think it works" is not `closed`.
