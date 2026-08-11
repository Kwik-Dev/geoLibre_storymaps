import React, { useEffect, useRef } from 'react';
import scrollama from 'scrollama';
import ChapterCard from './ChapterCard.jsx';

const alignments = { left: 'lefty', center: 'centered', right: 'righty', full: 'fully' };

/**
 * Scroll-space content: optional start/end slide targets plus one .step per
 * chapter. Scrollama fires onStepEnter/onStepExit (delegated to App) when a
 * step crosses the 50% viewport line — scroll position drives the camera.
 */
export default function Story({ config, activeId, ready, onStepEnter, onStepExit }) {
    const handlerRef = useRef({ onStepEnter, onStepExit });
    handlerRef.current = { onStepEnter, onStepExit };

    useEffect(() => {
        if (!ready) return;
        const scroller = scrollama();
        scroller
            .setup({ step: '.step', offset: 0.5 })
            .onStepEnter((response) => handlerRef.current.onStepEnter(response))
            .onStepExit((response) => handlerRef.current.onStepExit(response));

        const onResize = () => scroller.resize();
        window.addEventListener('resize', onResize);
        return () => {
            window.removeEventListener('resize', onResize);
            if (scroller.destroy) scroller.destroy();
        };
    }, [ready]);

    const { chapters, startSlide, endSlide, startStepId, endStepId, theme } = config;

    return (
        <div id="features">
            {startSlide && startSlide !== 'none' && (
                <div className="step sm-slide-step" id={startStepId} />
            )}
            {chapters.map((chapter, idx) => {
                const alignment = alignments[chapter.alignment] || 'centered';
                const active = activeId === chapter.id;
                return (
                    <div
                        key={chapter.id}
                        id={chapter.id}
                        className={`step ${alignment}${active ? ' active' : ''}${chapter.hidden ? ' hidden' : ''}`}
                    >
                        <ChapterCard chapter={chapter} index={idx} theme={theme} />
                    </div>
                );
            })}
            {endSlide && endSlide !== 'none' && (
                <div className="step sm-slide-step" id={endStepId} />
            )}
        </div>
    );
}
