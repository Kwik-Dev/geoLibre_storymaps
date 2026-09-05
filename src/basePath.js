// Base-path helper for subpath deployments (e.g. storyboard.ink/maps).
//
// VITE_BASE_PATH is the URL prefix the app is served under (e.g. "/maps"),
// matching the server's BASE_PATH. It is used to build absolute URLs for the
// API, media, and OAuth start paths so they stay under the same subpath.
// Defaults to "" (root deployment).
const raw = (import.meta.env.VITE_BASE_PATH || '').trim();

function normalize(p) {
    if (!p || p === '/') return '';
    if (!p.startsWith('/')) p = '/' + p;
    return p.replace(/\/+$/, '');
}

export const basePath = normalize(raw);

/** Join the base path with a leading-slash path (e.g. "/api/health"). */
export function withBase(path) {
    return basePath + path;
}
