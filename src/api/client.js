// Lightweight API client for the user-created storymaps backend.
//
// base = import.meta.env.VITE_API (optional, e.g. a separate origin) or the
// same-origin '/api' mount. Requests use `withCredentials` so the httpOnly
// refresh cookie flows automatically, and add an `Authorization: Bearer`
// header whenever a token is held in memory (from auth, set via setToken).
let token = null;

export function setToken(t) {
    token = t || null;
}
export function getToken() {
    return token;
}

const base = (import.meta.env.VITE_API || '/api').replace(/\/+$/, '');

/**
 * fetch wrapper: injects auth, unwraps JSON errors, throws on non-2xx.
 * Never rejects on network failure — callers handle the offline case.
 */
export async function apiFetch(path, options = {}) {
    const headers = new Headers(options.headers || {});
    if (token) headers.set('Authorization', `Bearer ${token}`);
    const resp = await fetch(`${base}${path}`, {
        ...options,
        headers,
        credentials: 'include',
    });
    if (!resp.ok) {
        let detail = '';
        try {
            const j = await resp.json();
            detail = (j && j.error) || '';
        } catch (_) {
            /* non-JSON error body */
        }
        const err = new Error(detail || `Request failed (${resp.status})`);
        err.status = resp.status;
        throw err;
    }
    return resp;
}

/**
 * GET /api/stories — the public story listing (anon → public only; a
 * logged-in owner also sees their own via the Authorization header/cookie).
 * Returns the array of story meta records.
 */
export async function listStories() {
    const resp = await apiFetch('/stories');
    const data = await resp.json();
    return Array.isArray(data) ? data : (data && data.stories) || [];
}

/**
 * GET /api/stories/:id/export — the legacy camelCase story JSON (full story
 * incl. chapters) that the existing renderer consumes unchanged.
 */
export async function getStoryExport(id) {
    const resp = await apiFetch(`/stories/${encodeURIComponent(id)}/export`);
    return resp.json();
}
