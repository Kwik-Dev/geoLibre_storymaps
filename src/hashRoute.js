// Hash-based router (P5.4). We route on window.location.hash — NOT the history
// API — so deep links keep working in the single-file `file://` build.
//
//   #/stories/<id>   → load + render that specific story
//   #/               → the story picker / list
//   (no hash)        → the story picker / list
//
// The hash is the single source of truth for what the app shows; the picker
// navigates by writing the hash, and a `hashchange` listener re-renders.

export function parseHash() {
    const hash = (typeof window !== 'undefined' && window.location.hash) || '';
    const match = hash.match(/^#\/stories\/([^/?#]+)/);
    if (match) {
        try {
            return { type: 'story', id: decodeURIComponent(match[1]) };
        } catch (_) {
            // malformed percent-encoding → fall back to the raw segment
            return { type: 'story', id: match[1] };
        }
    }
    // '#/' (or anything that isn't a story path, incl. no hash) → the list.
    return { type: 'list' };
}

/** Navigate to a story by writing the hash (fires hashchange → re-render). */
export function navigateToStory(id) {
    if (typeof window === 'undefined') return;
    const next = `#/stories/${encodeURIComponent(id)}`;
    if (window.location.hash === next) return; // no-op on the same route
    window.location.hash = next;
}

/** Navigate back to the picker / list. */
export function navigateToHome() {
    if (typeof window === 'undefined') return;
    if (window.location.hash === '#/' || window.location.hash === '') {
        window.location.hash = '#/';
        return;
    }
    window.location.hash = '#/';
}

/** Subscribe to hash changes; returns an unsubscribe function. */
export function onHashChange(cb) {
    if (typeof window === 'undefined') return () => {};
    window.addEventListener('hashchange', cb);
    return () => window.removeEventListener('hashchange', cb);
}
