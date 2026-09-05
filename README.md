# GeoLibre Storymaps — React + Audio

React (Vite) refactor of [`a-tour-of-five-cities.html`](a-tour-of-five-cities.html) — a
MapLibre GL scrollytelling story — with per-chapter audio playback.

## Run

```bash
npm install
npm run dev      # dev server → http://localhost:5173
npm run build    # production build → dist/
npm run preview  # serve the production build
```

The build is a **single self-contained HTML file** (`dist/index.html` — JS and CSS
inlined via `vite-plugin-singlefile`), so you can also just open it directly from
disk (double-click / `open dist/index.html`). Media (audio/images/videos) sits in
folders next to it and is referenced with relative paths from the story JSONs.

## Structure

```
public/audio/                 # freesound ocean tracks (copied from audio-data-colleciton)
public/images/                # ocean chapter images (pixabay, copied from audio-data-colleciton)
public/videos/                # ocean chapter videos (pixabay, copied from audio-data-colleciton)
src/
  main.jsx                    # entry (React root, maplibre css, styles)
  App.jsx                     # orchestrator: camera flyTo, scroll handlers, slides, chrome, nav, audio, story picker
  stories.js                  # story registry — every JSON below is a selectable story
  styles.css                  # ported storymap styles + audio toggle styles
  audio/AudioContext.jsx      # singleton <audio> element + play/pause/toggle API
  components/
    MapView.jsx               # fixed MapLibre map + inset map + markers (imperative refs)
    Story.jsx                 # Scrollama .step scroller (onStepEnter/onStepExit)
    ChapterCard.jsx           # draggable/resizable card; image or video; audio toggle in the title bar
    NavSidebar.jsx            # clickable chapter list
freesound-ocean-storymap.json # from audio-data-colleciton/freesound_audio_ocean/metadata.json — 3 geotagged recordings
pixabay-image-storymap.json   # from audio-data-colleciton/pixabay_media_images/metadata.json — 2 photos
pixabay-video-storymap.json   # from audio-data-colleciton/pixabay_media_videos/metadata.json — 2 films (chapter.video)
a-tour-of-five-cities.json    # original five-cities story data (reverted, not used by the app)
```

## Stories

The app ships with three storymap JSONs, switchable via the **Story** dropdown in
the top-right corner (pauses audio, remounts the map + scroller):

| Story | JSON | Source collection | Chapters |
|---|---|---|---|
| Freesound · Ocean recordings | `freesound-ocean-storymap.json` | `freesound_audio_ocean/metadata.json` | 3 (geotagged, with audio) |
| Pixabay · Ocean images | `pixabay-image-storymap.json` | `pixabay_media_images/metadata.json` | 2 (photos) |
| Pixabay · Ocean videos | `pixabay-video-storymap.json` | `pixabay_media_videos/metadata.json` | 2 (films, `chapter.video`) |

Add a new story: drop a JSON in the project root and register it in
`src/stories.js`. Chapters support an optional `video` field (local mp4, rendered
as an inline `<video>` with the `image` as poster).

## Audio feature

Each chapter supports two new config fields:

| Field | Type | Description |
|---|---|---|
| `audio` | string (URL) | Track played while the chapter is active |
| `autoPlayAudio` | boolean | Start the track automatically on chapter enter |

Behavior:

1. A **single shared `HTMLAudioElement`** (module-level singleton in `AudioContext.jsx`)
   is reused by every chapter — one track at a time.
2. `onStepEnter` → if the chapter has `audio` + `autoPlayAudio`, the shared element's
   `src` is swapped and playback starts. `onStepExit` → the track pauses.
   (`play()` rejections from autoplay policy are caught; the toggle still lets the
   listener start audio with an explicit gesture.)
3. Each card's title bar shows a **speaker toggle** (🔊/🔈) to start/stop that
   chapter's track manually.

Tracks: `public/audio/161697_ocean-1.mp3`, `161698_ocean-and-plane.mp3`,
`174763_pacific-ocean.mp3` (CC-BY freesound.org previews, from
`audio-data-colleciton/freesound_audio_ocean/`).

## Credits

This project is derived from [GeoLibre](https://github.com/opengeos/GeoLibre)
and [MapLibre Storytelling](https://github.com/opengeos/maplibre-gl-storymaps),
both © 2026 Qiusheng Wu, licensed under the MIT License. See [`NOTICE`](NOTICE)
and [`LICENSE`](LICENSE) for full attribution.
