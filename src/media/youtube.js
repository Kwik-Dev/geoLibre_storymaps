// YouTube URL → embed URL helper.
//
// The chapter renderer plays video with a native <video> tag, which needs a
// direct media file. YouTube watch/shorts/live URLs are HTML pages, so we
// detect them and convert to an embeddable https://www.youtube.com/embed/<id>
// URL that ChapterCard renders as an <iframe> instead.
//
// Returns the embed URL, or null when the input is not a recognized YouTube
// URL (in which case the caller falls back to a plain <video>).

const YT_HOSTS = new Set(['youtube.com', 'youtu.be']);

export function youtubeEmbedUrl(url) {
    if (!url) return null;
    let u;
    try {
        u = new URL(url);
    } catch (_) {
        return null;
    }
    const host = u.hostname.replace(/^(www|m)\./, '').toLowerCase();
    if (!YT_HOSTS.has(host)) return null;

    // youtu.be/<id>
    if (host === 'youtu.be') {
        const id = u.pathname.split('/').filter(Boolean)[0];
        return id ? `https://www.youtube.com/embed/${id}` : null;
    }

    // youtube.com/watch?v=<id>
    if (u.pathname === '/watch') {
        const id = u.searchParams.get('v');
        return id ? `https://www.youtube.com/embed/${id}` : null;
    }

    // youtube.com/embed/<id>
    if (u.pathname.startsWith('/embed/')) {
        const id = u.pathname.split('/')[2];
        return id ? `https://www.youtube.com/embed/${id}` : null;
    }

    // youtube.com/shorts/<id> and youtube.com/live/<id>
    const m = u.pathname.match(/^\/(shorts|live)\/([^/]+)/);
    if (m && m[2]) return `https://www.youtube.com/embed/${m[2]}`;

    return null;
}
