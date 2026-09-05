# Cards — M5 Frontend cut-over (async getStory, Markdown, routing, proxy)

Read `delegation/HANDOUT.md §0.4/§5` first. Card fields as in M1. A card is
`closed` only when its `VERIFY` passes. No PR, no out-of-scope edits.
The frontend stack: React 18 + Vite 6 + MapLibre GL 5 + scrollama 3, building a
single-file `dist/index.html` (`vite-plugin-singlefile`). **No-server /
`file://` mode must keep working.**

---

## P5.1 — Markdown renderer component (`react-markdown` + GFM + sanitize)
- **DEPS:** `npm i react-markdown remark-gfm rehype-sanitize` · **core** · est ~15m
- **READ:** feature_request §5 (Markdown dual-render), §10 (XSS); HANDOUT §0.4
- **SCOPE:**
     - `package.json`: add `react-markdown`, `remark-gfm`, `rehype-sanitize`.
     - `src/components/Markdown.jsx` — `Markdown({ text })` using
        `react-markdown` + `remark-gfm` + `rehype-sanitize` (**`rehype-sanitize`
       is the XSS boundary — do not remove it**). It renders to DOM nodes, **not**
      `dangerouslySetInnerHTML`.
- **VERIFY:**
    ```
    npm i && npm run build      # green
    # add a small vitest/react test that a malicious markdown string — e.g.
      "<img src=x onerror=alert(1)>" or an inline "<script>" — renders as inert text.
     ```
   expects: build is green; the hostile-string test renders the payload as text,
   with no `onerror`/`onclick`/script execution.
- **HANDOFF:** `<Markdown text="…" />`.
- **GOTCHAS:** `rehype-sanitize` must be wired in the markdown pipeline; don't
   pass user HTML through `dangerouslySetInnerHTML` from this component.

---

## P5.2 — Dual render in `ChapterCard`
- **DEPS:** P5.1 · **core** · est ~15m
- **READ:** feature_request §5; HANDOUT §0.4
- **SCOPE:** `src/components/ChapterCard.jsx` — branch per chapter:
     - if `chapter.description` (legacy **HTML**) is present → keep the existing
       `dangerouslySetInnerHTML` path (**embedded stories unchanged**).
     - else if `chapter.description_md` is present → render `<Markdown text=… />`.
     - else render nothing.
     Keep the existing image / video(+poster) / audio(waveform) media rendering
     as-is.
- **VERIFY:**
     ```
    npm run build     # green
    # manual / regression:
   #  - load an embedded `*-storymap.json` story → its HTML chapter still renders
   #  - load a user story with `description_md` → renders **sanitized** markdown
   #    (a hostile snippet like "<script>alert(1)</script>" stays inert)
     ```
- **HANDOFF:** `ChapterCard` now supports **both** `description` (HTML) and
   `description_md` (markdown).
- **GOTCHAS:** never mix HTML and markdown on one chapter; the embedded (HTML)
   path must stay behaviorally identical.

---

## P5.3 — Async `getStory` + data-driven story picker
- **DEPS:** P5.2, P3.1 (API) · **core** · est ~20m
- **READ:** feature_request §8 (getStory), §9; HANDOUT §7
- **SCOPE:**
     - `src/api/client.js` — a `fetch` wrapper; base =
        `import.meta.env.VITE_API` or same-origin `/api`; auth via
        `withCredentials` + a `Authorization` header if a token is in memory.
     - Replace the static `getStory`/`getStories` source in `src/App.jsx`/
       `src/stories.js`: **merge** the embedded `STORIES` with the API
        `GET /api/stories` (anon → public only; a logged-in owner → also their
        own). Dedupe by `id`. Async load with loading/error/empty states.
     - **Graceful degradation:** if the server is unreachable, catch the failure
       and fall back to embedded stories only (no-server mode must not crash).
- **VERIFY:**
     ```
    npm run build   # green
    # manual:
   #  - with the Go server up, the picker shows API stories
   #  - with the server DOWN (no fetch), the picker still shows embedded-only,
   #    no crash, a friend/empty notice is shown
     ```
- **HANDOFF:** async `getStories(mode)`, `getStory(id)`; `client.js`.
- **GOTCHAS:** the offline/no-server path is a hard requirement — catch the
   fetch error and keep showing embedded stories; dedupe embedded + API by id.

---

## P5.4 — Hash routing (`#/stories/<id>`) + empty states
- **DEPS:** P5.3 · **core** · est ~15m
- **READ:** feature_request §8; HANDOUT §0.4
- **SCOPE:** `src/hashRoute.js` (+ `src/App.jsx`) — route on the **hash** so it
     works in the single-file `file://` build:
       - `#/stories/<id>` → load + render that story.
       - `#/` or no hash → the story picker/list.
       - unknown id → friendly "story not found / create one".
       - empty: no stories at all → "Create a story" CTA; no-server & none
        embedded → a clear "no embedded stories" note.
     Use `hashchange` (not history API) so `file://` deep-links keep working.
- **VERIFY:**
     ```
    npm run build   # green
    # manual:
   #  - navigate `#/stories/<id>` → opens that story
   #  - `#/` → shows the picker
   #  - bad id → empty state; deep-link refresh (F5) keeps the route
   #  - open `dist/index.html` via `file://` → routing still works
     ```
- **HANDOFF:** the hash router + empty states; `App.jsx` renders by route.
- **GOTCHAS:** **hash** routing is what keeps the `file://` single-file build
   functional — don't switch to the history API; verify the `file://` case.

---

## P5.5 — Vite dev proxy `/api` + `/media` → Go :8080
- **DEPS:** none (build config) · **core** · est ~10m
- **READ:** feature_request §9; HANDOUT §9
- **SCOPE:** `vite.config.js` — dev `server.proxy`:
      `/api` → `http://localhost:8080`, `/media` → `http://localhost:8080`
      (same path). Keep the `vite-plugin-singlefile` config intact.
- **VERIFY:**
     ```
    npm run dev      # in one terminal; Go server on :8080 in another
    curl -s http://localhost:5173/api/health   # 200 proxied to the Go server
    curl -s -o /dev/null -w '%{http_code}\n' http://localhost:5173/   # 200 (app)
     ```
- **HANDOFF:** dev proxy is in place; `npm run build` still emits a single-file
   `dist/index.html`.
- **GOTCHAS:** proxy is **dev-only**; don't disturb `vite-plugin-singlefile`.
