import React, { createContext, useContext, useState } from 'react';

// Singleton <audio> element shared by every chapter (feature request: one
// playback element, play on chapter enter, pause on exit).
let sharedAudio = null;
function getSharedAudio() {
    if (!sharedAudio) sharedAudio = new Audio();
    return sharedAudio;
}

function sameSrc(a, b) {
    if (!a || !b) return false;
    try {
        return new URL(a, window.location.href).href === new URL(b, window.location.href).href;
    } catch {
        return a === b;
    }
}

const AudioContext = createContext(null);

export function AudioProvider({ children }) {
    // { id: chapterId | null, playing: bool }
    const [state, setState] = useState({ id: null, playing: false });

    const playChapter = (chapter) => {
        if (!chapter?.audio) {
            pause();
            return;
        }
        const el = getSharedAudio();
        // Swap the source when the chapter's track changes; reuse it otherwise
        // so the same ocean loop keeps playing across repeated visits.
        if (!sameSrc(el.src, chapter.audio)) el.src = chapter.audio;
        const p = el.play();
        if (p && typeof p.then === 'function') {
            p.then(() => setState({ id: chapter.id, playing: true })).catch(() => {
                // Autoplay can be blocked before a user gesture; reflect the
                // real state so the card toggle still lets the listener start it.
                setState({ id: chapter.id, playing: false });
            });
        } else {
            setState({ id: chapter.id, playing: true });
        }
    };

    const pause = () => {
        const el = getSharedAudio();
        if (!el.paused) el.pause();
        setState({ id: null, playing: false });
    };

    const toggle = (chapter) => {
        if (!chapter?.audio) return;
        if (state.playing && state.id === chapter.id) pause();
        else playChapter(chapter);
    };

    return (
        <AudioContext.Provider value={{ state, playChapter, pause, toggle }}>
            {children}
        </AudioContext.Provider>
    );
}

export function useAudio() {
    return useContext(AudioContext);
}
