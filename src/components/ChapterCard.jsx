import React, { useRef } from 'react';
import { useAudio } from '../audio/AudioContext.jsx';
import Markdown from './Markdown.jsx';

// Relative bar heights for the sound-wave placeholder shown on chapters
// without a photo (e.g. the freesound field recordings).
const WAVE_BARS = [
    0.35, 0.6, 0.9, 1.25, 0.7, 1.1, 0.45, 1.4, 0.85, 1.2, 0.4, 1.0,
    0.65, 1.3, 0.9, 0.5, 1.05, 0.75, 1.25, 0.55, 1.0, 0.4, 0.85, 1.15,
];

/**
 * Draggable / resizable story card: title bar (with per-chapter audio toggle),
 * image/video/waveform + description body, and a bottom-right resize handle.
 */
export default function ChapterCard({ chapter, index, theme }) {
    const cardRef = useRef(null);
    const offsetRef = useRef({ x: 0, y: 0 });
    const { state, toggle } = useAudio();

    const isThisPlaying = state.playing && state.id === chapter.id;

    const onBarPointerDown = (e) => {
        // The audio toggle is a button inside the bar — never start a drag from it.
        if (e.target.closest('button')) return;
        e.preventDefault();
        const sx = e.clientX;
        const sy = e.clientY;
        const bx = offsetRef.current.x;
        const by = offsetRef.current.y;
        const move = (ev) => {
            offsetRef.current = { x: bx + (ev.clientX - sx), y: by + (ev.clientY - sy) };
            if (cardRef.current) {
                cardRef.current.style.transform = `translate(${offsetRef.current.x}px,${offsetRef.current.y}px)`;
            }
        };
        const up = () => {
            window.removeEventListener('pointermove', move);
            window.removeEventListener('pointerup', up);
        };
        window.addEventListener('pointermove', move);
        window.addEventListener('pointerup', up);
    };

    const onBarDoubleClick = () => {
        offsetRef.current = { x: 0, y: 0 };
        if (cardRef.current) cardRef.current.style.transform = '';
    };

    const onResizePointerDown = (e) => {
        e.preventDefault();
        e.stopPropagation();
        const sx = e.clientX;
        const sy = e.clientY;
        const r = cardRef.current.getBoundingClientRect();
        const bw = r.width;
        const bh = r.height;
        const move = (ev) => {
            if (cardRef.current) {
                cardRef.current.style.width = Math.max(200, bw + (ev.clientX - sx)) + 'px';
                cardRef.current.style.height = Math.max(120, bh + (ev.clientY - sy)) + 'px';
            }
        };
        const up = () => {
            window.removeEventListener('pointermove', move);
            window.removeEventListener('pointerup', up);
        };
        window.addEventListener('pointermove', move);
        window.addEventListener('pointerup', up);
    };

    return (
        <div className={`sm-card ${theme}`} ref={cardRef}>
            <div className="sm-bar" onPointerDown={onBarPointerDown} onDoubleClick={onBarDoubleClick}>
                <span className="sm-grip">☰</span>
                <span className="sm-title">{chapter.title || `Chapter ${index + 1}`}</span>
                {chapter.audio && (
                    <button
                        type="button"
                        className="sm-audio"
                        aria-pressed={isThisPlaying}
                        aria-label={isThisPlaying ? 'Pause audio' : 'Play audio'}
                        title={isThisPlaying ? 'Pause audio' : 'Play audio'}
                        onClick={(e) => {
                            e.stopPropagation();
                            toggle(chapter);
                        }}
                    >
                        {isThisPlaying ? '🔊' : '🔈'}
                    </button>
                )}
            </div>
            <div className="sm-body">
                {chapter.video ? (
                    <video
                        src={chapter.video}
                        poster={chapter.image}
                        controls
                        loop
                        muted
                        playsInline
                        preload="metadata"
                    />
                ) : chapter.image ? (
                    <img src={chapter.image} alt={chapter.title || ''} />
                ) : (
                    <div
                        className={`sm-wave${isThisPlaying ? ' playing' : ''}`}
                        role="img"
                        aria-label={`${chapter.title || 'Chapter ' + (index + 1)} — audio recording`}
                    >
                        <svg viewBox="0 0 120 40" preserveAspectRatio="none">
                            {WAVE_BARS.map((h, i) => {
                                const height = Math.max(4, h * 20);
                                return (
                                    <rect
                                        key={i}
                                        x={i * 5}
                                        y={20 - height / 2}
                                        width={3}
                                        height={height}
                                        rx={1.5}
                                    />
                                );
                            })}
                        </svg>
                    </div>
                )}
                {chapter.description ? (
                    <p dangerouslySetInnerHTML={{ __html: chapter.description }} />
                ) : chapter.description_md ? (
                    <div className="sm-description-md">
                        <Markdown text={chapter.description_md} />
                    </div>
                ) : null}
            </div>
            <div className="sm-resize" onPointerDown={onResizePointerDown} />
        </div>
    );
}
