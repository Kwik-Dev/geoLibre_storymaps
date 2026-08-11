import React, { useCallback, useEffect, useRef, useState } from 'react';
import { AudioProvider, useAudio } from './audio/AudioContext.jsx';
import { stories, defaultStoryId, getStory } from './stories.js';
import MapView from './components/MapView.jsx';
import Story from './components/Story.jsx';
import NavSidebar from './components/NavSidebar.jsx';

const layerTypes = {
    fill: ['fill-opacity'],
    line: ['line-opacity'],
    circle: ['circle-opacity', 'circle-stroke-opacity'],
    symbol: ['icon-opacity', 'text-opacity'],
    raster: ['raster-opacity'],
    'fill-extrusion': ['fill-extrusion-opacity'],
    heatmap: ['heatmap-opacity'],
    hillshade: ['hillshade-exaggeration'],
};

function getLayerPaintType(map, layer) {
    const sl = map.getLayer(layer);
    return sl ? layerTypes[sl.type] : null;
}

function setLayerOpacity(map, layer) {
    if (!map.getLayer(layer.layer)) return;
    const paintProps = getLayerPaintType(map, layer.layer);
    if (!paintProps) return;
    paintProps.forEach((prop) => {
        if (layer.duration) {
            map.setPaintProperty(layer.layer, prop + '-transition', { duration: layer.duration });
        }
        map.setPaintProperty(layer.layer, prop, layer.opacity);
    });
}

function slideBg(mode, theme) {
    if (mode === 'black') return '#000000';
    if (mode === 'blank') return theme === 'light' ? '#fafafa' : '#444444';
    return null;
}

function StoryMap({ story }) {
    const config = story;
    const [ready, setReady] = useState(false);
    const [activeId, setActiveId] = useState(() =>
        config.startSlide && config.startSlide !== 'none' ? null : config.chapters[0].id
    );
    const [chromeHidden, setChromeHidden] = useState(
        Boolean(config.startSlide && config.startSlide !== 'none')
    );
    const [slideCover, setSlideCover] = useState(null);
    const [navHidden, setNavHidden] = useState(Boolean(config.hideChapterNav));

    const mapRef = useRef(null);
    const insetMapRef = useRef(null);
    const markerRef = useRef(null);
    const insetMarkerRef = useRef(null);
    const cameraTokenRef = useRef(0);
    const startSlideInitializedRef = useRef(false);

    const audio = useAudio();
    const audioRef = useRef(audio);
    audioRef.current = audio;

    // Apply the start slide's initial state during first render (before map
    // load) so a slow style/tile load can't flash the story chrome (#998).
    useEffect(() => {
        if (config.startSlide && config.startSlide !== 'none') {
            setChromeHidden(true);
            setSlideCover(slideBg(config.startSlide, config.theme) || null);
        }
    }, [config]);

    const setMarkersVisible = useCallback((visible) => {
        const value = visible ? '' : 'hidden';
        if (markerRef.current) markerRef.current.getElement().style.visibility = value;
        if (insetMarkerRef.current) insetMarkerRef.current.getElement().style.visibility = value;
    }, []);

    // Drive a start/closing slide (#998): blank/black paint a solid cover over
    // the map; global zooms out; "adjacent" previews the first (start) or last
    // (end) chapter with the text hidden.
    const enterSlide = useCallback(
        (mode, isStart) => {
            setActiveId(null);
            setChromeHidden(true);
            const bg = slideBg(mode, config.theme);
            if (bg) {
                setSlideCover(bg);
                return;
            }
            setSlideCover(null);
            const map = mapRef.current;
            if (!map) return;
            map.stop();
            ++cameraTokenRef.current;
            const adjacent = isStart ? config.chapters[0] : config.chapters[config.chapters.length - 1];
            const loc = mode === 'global' ? config.globalView : adjacent?.location || config.globalView;
            map.flyTo(loc);
            if (mode === 'global') {
                setMarkersVisible(false);
            } else {
                setMarkersVisible(true);
                if (config.showMarkers && markerRef.current) markerRef.current.setLngLat(loc.center);
                if (insetMapRef.current && insetMarkerRef.current) {
                    insetMapRef.current.setCenter(loc.center);
                    insetMarkerRef.current.setLngLat(loc.center);
                }
            }
        },
        [setMarkersVisible]
    );

    const handleStepEnter = useCallback(
        (response) => {
            const id = response.element.id;
            setActiveId(id);
            if (id === config.startStepId || id === config.endStepId) {
                // Skip the redundant first enter for the start slide that was
                // already initialized synchronously on map load (#998).
                if (id === config.startStepId && startSlideInitializedRef.current) {
                    startSlideInitializedRef.current = false;
                    return;
                }
                enterSlide(id === config.startStepId ? config.startSlide : config.endSlide, id === config.startStepId);
                return;
            }
            setSlideCover(null);
            setChromeHidden(false);
            const chapter = config.chapters.find((c) => c.id === id);
            if (!chapter) return;
            const map = mapRef.current;
            if (!map) return;

            // Cancel any in-progress move (e.g. a prior chapter's rotation) and
            // bump the token so its pending moveend handler is ignored.
            map.stop();
            const token = ++cameraTokenRef.current;
            map[chapter.mapAnimation || 'flyTo'](chapter.location);

            // Re-show the marker in case a preceding global slide hid it.
            setMarkersVisible(true);
            if (config.showMarkers && markerRef.current) markerRef.current.setLngLat(chapter.location.center);
            if (insetMapRef.current && insetMarkerRef.current) {
                insetMapRef.current.setCenter(chapter.location.center);
                insetMarkerRef.current.setLngLat(chapter.location.center);
            }
            if (chapter.onChapterEnter?.length > 0) chapter.onChapterEnter.forEach((l) => setLayerOpacity(map, l));

            // Audio: start this chapter's track (single shared element).
            if (chapter.audio && chapter.autoPlayAudio) audioRef.current.playChapter(chapter);

            if (chapter.rotateAnimation) {
                map.once('moveend', () => {
                    if (token !== cameraTokenRef.current) return;
                    const bearing = map.getBearing();
                    map.rotateTo(bearing + 180, { duration: 30000, easing: (t) => t });
                });
            }
        },
        [enterSlide, setMarkersVisible]
    );

    const handleStepExit = useCallback((response) => {
        const id = response.element.id;
        setActiveId((prev) => (prev === id ? null : prev));
        const chapter = config.chapters.find((c) => c.id === id);
        if (!chapter) return;
        if (chapter.onChapterExit?.length > 0) {
            chapter.onChapterExit.forEach((l) => setLayerOpacity(mapRef.current, l));
        }
        // Audio: pause when leaving the chapter whose track is playing.
        const { state } = audioRef.current;
        if (state.id === id) audioRef.current.pause();
    }, []);

    const onMapReady = useCallback(() => {
        setReady(true);
        // Set the start slide's camera deterministically on load, mirroring the
        // in-app presenter's first step (#998). Runs before Scrollama's first
        // (async) enter for the in-view start slide, so the flag above skips it.
        if (config.startSlide && config.startSlide !== 'none') {
            enterSlide(config.startSlide, true);
            startSlideInitializedRef.current = true;
        }
    }, [enterSlide]);

    const handleNavigate = useCallback((id) => {
        const el = document.getElementById(id);
        const card = el && el.querySelector('.sm-card');
        (card || el).scrollIntoView({ block: 'center' });
    }, []);

    return (
        <>
            <MapView
                config={config}
                mapRef={mapRef}
                insetMapRef={insetMapRef}
                markerRef={markerRef}
                insetMarkerRef={insetMarkerRef}
                onReady={onMapReady}
            />
            {slideCover && <div id="slide-cover" style={{ background: slideCover, display: 'block' }} />}
            <div id="story">
                <header id="header" className={`${config.theme}${chromeHidden ? ' hidden' : ''}`}>
                    {config.title && <h1>{config.title}</h1>}
                    {config.subtitle && <h2>{config.subtitle}</h2>}
                    {config.byline && <p>{config.byline}</p>}
                </header>
                <Story
                    config={config}
                    activeId={activeId}
                    ready={ready}
                    onStepEnter={handleStepEnter}
                    onStepExit={handleStepExit}
                />
                {config.footer && (
                    <footer
                        id="footer"
                        className={`${config.theme}${chromeHidden ? ' hidden' : ''}`}
                        dangerouslySetInnerHTML={{ __html: `<p>${config.footer}</p>` }}
                    />
                )}
            </div>
            <NavSidebar config={config} activeId={activeId} hidden={navHidden} onNavigate={handleNavigate} />
            <button
                id="nav-toggle"
                className={config.theme}
                type="button"
                title={config.navToggleLabel}
                aria-label={config.navToggleLabel}
                aria-pressed={navHidden ? 'false' : 'true'}
                onClick={() => setNavHidden((h) => !h)}
            >
                ☰
            </button>
        </>
    );
}

// Top-right control: switch between the storymap JSONs at runtime. Lives
// outside the keyed <StoryMap> so it survives story remounts; pauses audio
// before swapping so the previous chapter's track doesn't keep playing.
function StoryPicker({ config, storyId, onSelect }) {
    const audio = useAudio();
    return (
        <div id="story-picker" className={config.theme}>
            <label htmlFor="story-select">Story</label>
            <select
                id="story-select"
                value={storyId}
                onChange={(e) => {
                    audio.pause();
                    onSelect(e.target.value);
                }}
            >
                {stories.map((s) => (
                    <option key={s.id} value={s.id}>
                        {s.label}
                    </option>
                ))}
            </select>
        </div>
    );
}

export default function App() {
    const [selectedId, setSelectedId] = useState(defaultStoryId);
    const story = getStory(selectedId);

    // Keep the browser tab title in sync with the active story.
    useEffect(() => {
        document.title = story.config.title;
    }, [story]);

    const handleSelect = useCallback((id) => {
        // Jump to the top so the new story's first step (or start slide) is
        // in view; Scrollama re-scans on remount.
        window.scrollTo(0, 0);
        setSelectedId(id);
    }, []);

    return (
        <AudioProvider>
            <StoryPicker config={story.config} storyId={story.id} onSelect={handleSelect} />
            {/* key forces a full remount (map, scrollama, state) per story */}
            <StoryMap key={story.id} story={story.config} />
        </AudioProvider>
    );
}
