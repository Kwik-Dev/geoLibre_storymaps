# AGENTS.md — GeoLibre Storymaps

**Stack**: React 18 + Vite 6 + MapLibre GL 5 + Scrollama 3 + `vite-plugin-singlefile`

A scrollytelling map app: scroll position drives camera position on a MapLibre map. Refactored from a vanilla HTML/JS file (`a-tour-of-five-cities.html`) into React components. Per-chapter audio playback using a shared singleton `HTMLAudioElement`.

---

## Commands

| Command | What it does |
|---|---|
| `npm run dev` | Vite dev server → `http://localhost:5173` |
| `npm run build` | Production build → `dist/index.html` (single self-contained HTML — JS/CSS inlined) |
| `npm run preview` | Serve the production build |
| `npm run convert` | Generate storymap JSONs from `audio-data-colleciton/` metadata |

No test framework exists. No linting configured.

---

## Project Structure

```
├── src/
│   ├── main.jsx                  # Entry: React root + maplibre-gl CSS import + styles.css
│   ├── App.jsx                   # Orchestrator — StoryMap + StoryPicker components
│   ├── stories.js                # Story registry — import JSONs, export array + helpers
│   ├── styles.css                # All CSS in one file (ported from the original HTML)
│   ├── audio/
│   │   └── AudioContext.jsx      # React Context wrapping a module-level singleton <audio>
│   └── components/
│       ├── MapView.jsx           # Fixed full-viewport MapLibre map + optional inset map
│       ├── Story.jsx             # Scrollama .step scroller
│       ├── ChapterCard.jsx       # Draggable/resizable card: image/video/waveform + description
│       └── NavSidebar.jsx        # Clickable chapter list (left sidebar)
├── scripts/
│   └── convert-storymaps.mjs    # Converts metadata collections to storymap JSONs
├── public/audio/                 # Audio files (mp3)
├── public/images/                # Images
├── public/videos/                # Videos (mp4)
├── audio-data-colleciton/        # Raw metadata collections (input to convert-storymaps)
├── *.json                        # Storymap JSON files — registered in src/stories.js
└── *.md                          # Feature requests + architectural doc
```

---

## Architecture & Data Flow

```
App
├── AudioProvider (React Context — shared audio state)
│   ├── StoryPicker (top-right <select> — switches stories, pauses audio before swap)
│   └── StoryMap (keyed on story.id — full remount per switch)
│       ├── MapView (imperative MapLibre API via refs, no state)
│       ├── Story (Scrollama .step container)
│       │   └── ChapterCard (per-chapter: image/video/waveform + description + audio toggle)
│       ├── NavSidebar (left chapter list)
│       ├── header (title, subtitle, byline)
│       ├── footer
│       └── slide-cover (solid overlay for start/end slide modes)
```

**Scroll → Camera flow**:
1. User scrolls → Scrollama detects `.step` at 50% viewport
2. `onStepEnter` fires in `handleStepEnter` (App.jsx)
3. `map.stop()` cancels any in-progress animation, bump `cameraTokenRef`
4. `map.flyTo(chapter.location)` — camera moves to chapter coordinates
5. Marker + inset map update to new center
6. If chapter has `rotateAnimation: true`, a moveend listener rotates 180° over 30s
7. If chapter has `audio` + `autoPlayAudio`, the shared audio element starts the track
8. On exit: audio pauses, `onChapterExit` layer opacity rules run

**Key patterns**:
- **Imperative map API** — MapLibre map refs are passed up, all camera control is imperative (`map.flyTo`, `map.rotateTo`). No React state drives the map.
- **Token-based animation safety** — `cameraTokenRef` is bumped before each flyTo; stale moveend callbacks check the token and bail if another navigation happened (#998).
- **Ref forwarding for event handlers** — Scrollama is set up once (when `ready` becomes true). Callbacks use a ref (`handlerRef.current = { onStepEnter, onStepExit }`) so they always call the latest version without re-registering Scrollama.
- **Story switching via `key` remount** — `<StoryMap key={story.id}>` forces a full remount (new map, new Scrollama, fresh state) when the user picks a different story.

---

## Storymap JSON Schema

Each JSON file at the project root follows this structure:

**Global fields**:
`style`, `projection` (default `"globe"`), `showMarkers`, `markerColor`, `inset`, `insetPosition`, `insetStyle`, `insetZoom`, `theme` (`"dark"`/`"light"`), `auto`, `hideChapterNav`, `startSlide`/`endSlide` (`"none"`/`"blank"`/`"black"`/`"global"`/`"adjacent"`), `globalView`, `startStepId`/`endStepId`, `navToggleLabel`, `title`, `subtitle`, `byline`, `footer`

**Chapter fields**:
`id`, `title`, `image` (URL), `description` (HTML), `alignment` (`"left"`/`"right"`/`"center"`/`"full"`), `hidden` (boolean), `location` (`{ center: [lng, lat], zoom, pitch, bearing }`), `mapAnimation` (`"flyTo"`/`"easeTo"`), `rotateAnimation` (boolean), `onChapterEnter`/`onChapterExit` (layer opacity arrays), `audio` (URL), `autoPlayAudio` (boolean), `video` (URL — local mp4, `image` used as poster)

---

## Audio System

- Single shared `HTMLAudioElement` at module level in `AudioContext.jsx` (not per-chapter)
- React Context provides `{ state, playChapter, pause, toggle }` to all components
- `state = { id: chapterId | null, playing: boolean }`
- `sameSrc()` utility resolves relative URLs to prevent unnecessary `src` swaps
- `play()` rejections (autoplay policy) are silently caught — the toggle button still works
- On story switch, `StoryPicker` calls `audio.pause()` before changing the story

---

## Adding a New Story

1. Drop a storymap JSON file in the project root
2. Import it in `src/stories.js` and add to the `stories` array with a unique `id` and `label`
3. The JSON is auto-discovered in the top-right dropdown

Or use the converter:
1. Place metadata in `audio-data-colleciton/<collection>/`
2. Run `npm run convert` — generates `<name>-storymap.json` and copies media to `public/`
3. Register in `src/stories.js`

---

## Conventions & Gotchas

- **No tests exist** — don't look for a test runner or CI
- **No linter/formatter** — follow existing code style (4-space indent, single quotes for JSX strings, trailing `;`)
- **`.gitignore` excludes `/public`** — media files in `public/` are not tracked in git. If you add media, the `audio-data-colleciton/` folder (tracked) is the canonical source
- **Build is a single HTML file** — `vite-plugin-singlefile` inlines all JS + CSS into `dist/index.html` so it can be opened from disk (`file://`). Media stays as separate files referenced by relative paths
- **Original HTML still exists** — `a-tour-of-five-cities.html` is the pre-React reference; its CSS classes and DOM structure were preserved in the React port
- **Feature requests as markdown** — see `feature_request_audio_playback.md` and `feature_request_user_created_storymap.md` for pending feature specs
- **Architecture doc exists** — `storymap_architecure.md` has a sequence diagram and detailed chapter schema
- **`audio-data-colleciton` spelling** — the typo ("colleciton") is intentional and consistent in code
- **`convert-storymaps.mjs` won't overwrite** — if a JSON file already exists in the project root, the converter skips it (hand-tuned files are safe)
- **Map is non-interactive** — `interactive: false` on both maps; user only scrolls through content
- **Start/end slide modes** — handled by `slideBg()` and `enterSlide()` with the `#slide-cover` DOM element. Mode `"blank"`/`"black"` paints a solid overlay; `"global"` zooms out; `"adjacent"` shows first/last chapter without text. The `startSlideInitializedRef` pattern prevents redundant camera sets.
- **Waveform SVG** — audio-only chapters (no image/video) render a 24-bar animated SVG waveform in `ChapterCard.jsx`
- **Cards are draggable+resizable** — via pointer events on the title bar and bottom-right resize handle
- **LSP misidentifies JSX as Lua** — diagnostics show thousands of false-positive Lua errors on `.jsx`, `.html`, `.json`, `.css`, and `.md` files. Ignore them.

<!-- smithers:prefer-workflows START -->
## Smithers workflows

Use your best judgment, weighing speed, quality, and token usage, to decide
whether a request should run as a [smithers.sh](https://smithers.sh) workflow
or with regular subagents. Prefer a smithers workflow for multi-step plans and
for work that benefits from retries, approvals, review, or replay; reach for
plain subagents when a request is a quick one-off.

The `smithers` skill is installed: run `smithers workflow list` to see the
available workflows and `smithers workflow run <id>` to launch one.

When a session ends successfully and the work could have been a smithers
workflow, offer to turn the session into a reusable smithers workflow for next
time.
<!-- smithers:prefer-workflows END -->
