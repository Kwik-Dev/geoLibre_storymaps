// Async story source — the P5.3 HANDOFF.
//
// getStories()  → merges the embedded (bundled) stories with the API's
//                 `GET /api/stories` listing, deduped by id (embedded wins on
//                 a collision). Graceful degradation: if the server is
//                 unreachable we fall back to the embedded stories only.
// getStory(id)  → returns { id, label, config } for a single story:
//                 embedded stories resolve synchronously from the bundle;
//                 API stories are loaded via `GET /api/stories/:id/export`
//                 (the legacy story JSON the renderer consumes). Returns null
//                 if the story can't be found / the server is down.
import { stories as embeddedStories, defaultStoryId } from './stories.js';
import { listStories, getStoryExport } from './api/client.js';

const embeddedById = new Map(embeddedStories.map((s) => [s.id, s]));

/**
 * Load the full, merged story list (embedded + API), deduped by id.
 * Never rejects — on any fetch failure it resolves to the embedded stories.
 */
export async function getStories() {
    let apiStories = [];
    try {
        apiStories = await listStories();
    } catch (_) {
        // Graceful degradation: no server / fetch error → embedded only.
        return embeddedStories;
    }

    const merged = embeddedStories.slice();
    const seen = new Set(merged.map((s) => s.id));
    for (const s of apiStories) {
        const id = String(s.id != null ? s.id : s.slug);
        if (seen.has(id)) continue;
        seen.add(id);
        // Meta-only listing → label for the picker; full config is lazy-loaded
        // by getStory(id) when the story is actually opened.
        merged.push({ id, label: s.title || id, config: null, meta: s });
    }
    return merged;
}

/**
 * Async getStory(id): embedded → bundled config; API → exported story JSON.
 * Returns null when the story is unknown or the server is unreachable.
 */
export async function getStory(id) {
    const local = embeddedById.get(id);
    if (local) return local;
    try {
        const config = await getStoryExport(id);
        return { id, label: (config && config.title) || id, config };
    } catch (_) {
        return null;
    }
}
