import React from 'react';

/**
 * Fixed chapter sidebar: click a row to jump to its scroll step.
 */
export default function NavSidebar({ config, activeId, hidden, onNavigate }) {
    return (
        <nav id="nav" className={`${config.theme}${hidden ? ' hidden-nav' : ''}`}>
            {config.chapters.map((chapter, idx) => (
                <div
                    key={chapter.id}
                    className={`nav-item${activeId === chapter.id ? ' active' : ''}`}
                    data-id={chapter.id}
                    onClick={() => onNavigate(chapter.id)}
                >
                    <span className="nav-num">{idx + 1}</span>
                    <span className="nav-title">{chapter.title || `Chapter ${idx + 1}`}</span>
                </div>
            ))}
        </nav>
    );
}
