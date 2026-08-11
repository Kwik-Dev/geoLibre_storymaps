#!/usr/bin/env node
/**
 * convert-storymaps.mjs
 *
 * Generates storymap JSONs from the audio-data-colleciton folder.
 *
 *   npm run convert            # scan every collection, copy media, write JSONs
 *   npm run convert -- --dry-run   # preview what would be generated
 *
 * For each subfolder with a metadata file (metadata.json or <name>_metadata.json) it:
 *   1. detects the media type (freesound audio / pixabay photo / pixabay film)
 *   2. builds one chapter per search hit (title, description, geolocation,
 *      source credits, audio/image/video fields)
 *   3. copies the downloaded media into public/{audio,images,videos}
 *   4. writes <collection-slug>-storymap.json to the project root
 *
 * Output files are named after the metadata file (surfing_metadata.json →
 * surfing-storymap.json; plain metadata.json falls back to the collection folder
 * name) and never touch hand-tuned files.
 * Register new stories in src/stories.js to make them selectable in the app.
 */
import {
    readdirSync,
    readFileSync,
    writeFileSync,
    mkdirSync,
    copyFileSync,
    existsSync,
} from 'node:fs';
import { join, basename, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = join(__dirname, '..');
const COLLECTIONS = join(ROOT, 'audio-data-colleciton');
const PUBLIC = join(ROOT, 'public');

const DRY_RUN = process.argv.includes('--dry-run');

// Global storymap config shared by every generated story (matches the app's
// expectations: MapView, Story, NavSidebar, App all read these fields).
const GLOBAL = {
    style: 'https://tiles.openfreemap.org/styles/liberty',
    projection: 'globe',
    showMarkers: true,
    markerColor: '#3fb1ce',
    inset: true,
    insetPosition: 'bottom-left',
    insetStyle: 'https://basemaps.cartocdn.com/gl/positron-gl-style/style.json',
    insetZoom: 1,
    theme: 'dark',
    auto: false,
    hideChapterNav: false,
    startSlide: 'none',
    endSlide: 'none',
    globalView: { center: [0, 20], zoom: 0.6, pitch: 0, bearing: 0 },
    startStepId: '__story_start__',
    endStepId: '__story_end__',
    navToggleLabel: 'Toggle chapter list',
};

// Used when a hit has no geolocation; the description notes it is approximate.
const FALLBACK_LOCATION = { center: [0, 20], zoom: 2, pitch: 0, bearing: 0 };

const slugify = (s) =>
    String(s)
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');
const capitalize = (s) => (s ? s.charAt(0).toUpperCase() + s.slice(1) : s);
const esc = (s) =>
    String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

/** Find the collection's metadata file: metadata.json or <name>_metadata.json. */
function findMetadataFile(dir) {
    const files = readdirSync(dir, { withFileTypes: true })
        .filter((d) => d.isFile() && /metadata\.json$/i.test(d.name))
        .map((d) => d.name)
        .sort();
    if (!files.length) return null;
    return files.includes('metadata.json') ? 'metadata.json' : files[0];
}

/** Derive the story slug from the metadata filename: surfing_metadata.json → surfing;
 *  plain metadata.json falls back to the collection folder name. */
function slugFromMetadata(metaFile, folderName) {
    const base = metaFile.replace(/metadata\.json$/i, '').replace(/_+$/, '');
    return slugify(base || folderName);
}

/** freesound hits have previewURL on freesound.org; pixabay hits are photo/film. */
function detectType(metadata) {
    const hit = metadata.search?.hits?.[0] ?? metadata.media?.[0];
    if (!hit) return 'audio';
    if (hit.type === 'image' || hit.type === 'photo') return 'image';
    if (hit.type === 'video' || hit.type === 'film') return 'video';
    if (String(hit.previewURL || '').includes('freesound')) return 'audio';
    if (String(hit.url || '').includes('pixabay')) return 'video';
    return 'audio';
}

/** Pull the freesound username out of a page URL like
 *  https://freesound.org/people/xserra/sounds/161697/ */
function freesoundUserFrom(pageURL) {
    const m = String(pageURL || '').match(/freesound\.org\/people\/([^/]+)\//);
    return m ? m[1] : null;
}

/** Normalize either metadata format into { hits, downloads, geolocation }.
 *  Unified format: { media: [ { id, name?, tags, username?, user_id?, userURL?,
 *    userImageURL?, pageURL, url, downloadURL?, localPath, size, hasAudio,
 *    type?, duration?, created?, license?, width?, height?, views?, downloads?,
 *    likes?, comments?, is_geotagged?, geolocation? } ] }
 *  Legacy format:  { search: { hits }, downloads: [], geolocation: [] } */
function normalizeMetadata(metadata) {
    if (Array.isArray(metadata.media)) {
        const hits = metadata.media.map((m) => {
            const tags = Array.isArray(m.tags)
                ? m.tags
                : String(m.tags || '')
                      .split(',')
                      .map((t) => t.trim())
                      .filter(Boolean);
            const isFreesound = /freesound/i.test(String(m.pageURL || m.url || ''));
            const isPixabay = /pixabay/i.test(String(m.pageURL || m.url || ''));
            const rawType = String(m.type || '').toLowerCase();
            return {
                id: m.id,
                name: m.name || (tags[0] ? capitalize(tags[0]) : null),
                tags,
                username:
                    m.username || m.user || (isFreesound ? freesoundUserFrom(m.pageURL) : null),
                user_id: m.user_id ?? null,
                userURL: m.userURL || null,
                userImageURL: m.userImageURL || null,
                license: m.license || null,
                duration: m.duration ?? null,
                type:
                    rawType === 'photo' ? 'image'
                    : rawType === 'film' ? 'video'
                    : isFreesound ? 'audio'
                    : m.hasAudio ? 'video'
                    : 'image',
                created: m.created || null,
                pageURL: m.pageURL,
                url: m.url,
                previewURL: m.previewURL || m.url,
                downloadURL: m.downloadURL || m.url,
                thumbnail:
                    m.thumbnail ||
                    m.geolocation?.localPath ||
                    (isPixabay && m.url ? String(m.url).replace(/\.mp4$/i, '.jpg') : null),
                latitude: m.geolocation?.latitude ?? null,
                longitude: m.geolocation?.longitude ?? null,
                width: m.width ?? null,
                height: m.height ?? null,
                views: m.views ?? null,
                downloads: m.downloads ?? null,
                likes: m.likes ?? null,
                comments: m.comments ?? null,
                is_geotagged: m.is_geotagged ?? null,
            };
        });
        const downloads = metadata.media
            .filter((m) => m.localPath)
            .map((m) => ({ id: m.id, localPath: m.localPath, url: m.url, size: m.size }));
        const geolocation = metadata.media
            .filter((m) => m.geolocation)
            .map((m) => ({ id: m.id, ...m.geolocation }));
        return { hits, downloads, geolocation };
    }
    return {
        hits: (metadata.search?.hits || []).map((h) => ({
            ...h,
            username: h.username || h.user || null,
            tags: Array.isArray(h.tags)
                ? h.tags
                : String(h.tags || '')
                      .split(',')
                      .map((t) => t.trim())
                      .filter(Boolean),
        })),
        downloads: metadata.downloads || [],
        geolocation: metadata.geolocation || [],
    };
}

/** Prefer the vision-model geolocation array (pixabay), else the hit's own lat/lon (freesound). */
function geoFor(metadata, hit) {
    const geo = (metadata.geolocation || []).find((g) => String(g.id) === String(hit.id));
    if (geo && geo.latitude != null && geo.longitude != null) {
        const place = [geo.placeName, geo.country].filter(Boolean).join(', ');
        return {
            center: [geo.longitude, geo.latitude],
            place: geo.placeName,
            country: geo.country,
            confidence: geo.confidence,
            note: `Geolocated${place ? ' to ' + place : ''} (${geo.latitude.toFixed(2)}°, ${geo.longitude.toFixed(2)}°)`,
        };
    }
    if (hit.latitude != null && hit.longitude != null) {
        return {
            center: [hit.longitude, hit.latitude],
            place: null,
            country: null,
            confidence: 'high',
            note: `Geotagged at (${hit.latitude.toFixed(2)}°, ${hit.longitude.toFixed(2)}°)`,
        };
    }
    return null;
}

function pickZoom(geo) {
    if (!geo) return 2;
    if (geo.confidence === 'high' && geo.place) return 11;
    if (geo.confidence === 'high') return 9;
    if (geo.confidence === 'medium') return 8;
    if (geo.confidence === 'low') return 5;
    return 9;
}

function sourceLink(hit) {
    if (hit.pageURL) return hit.pageURL;
    if (String(hit.previewURL || '').includes('freesound')) return `https://freesound.org/s/${hit.id}/`;
    return null;
}

function buildDescription(hit, geo) {
    const parts = [];
    const who = hit.username || 'Pixabay contributor';
    parts.push(
        `Recorded by <b>${esc(who)}</b>${hit.created ? ' on ' + String(hit.created).slice(0, 10) : ''}.`
    );
    if (Array.isArray(hit.tags) && hit.tags.length) {
        parts.push(`Tags: ${hit.tags.map(esc).join(', ')}.`);
    }
    const meta = [];
    if (hit.duration) meta.push(`Duration: ${Math.round(hit.duration * 10) / 10} s`);
    if (geo?.note) meta.push(geo.note);
    if (!geo) meta.push('Not geotagged — shown at a default position');
    const link = sourceLink(hit);
    if (link) meta.push(`<a href="${link}" target="_blank" rel="noopener noreferrer">Source #${hit.id}</a>`);
    if (meta.length) parts.push('<br><br>' + meta.join(' · '));
    return parts.join(' ');
}

/** Copy a downloaded media file into public/<subdir>; returns the relative URL.
 *  metadata.downloads[].localPath is relative to the audio-data-colleciton root. */
function copyMedia(localPath, subdir) {
    if (!localPath) return null;
    const src = join(COLLECTIONS, localPath);
    if (!existsSync(src)) return null;
    const name = basename(localPath);
    const destDir = join(PUBLIC, subdir);
    if (!DRY_RUN) {
        mkdirSync(destDir, { recursive: true });
        copyFileSync(src, join(destDir, name));
    }
    return `${subdir}/${name}`;
}

function buildChapter({ hit, download, type, slug, idx, geo, collectionDir }) {
    const id = `${slug}-${hit.id}`;
    const title = String(hit.name || `Chapter ${idx + 1}`).replace(/\.[a-z0-9]+$/i, '');
    const chapter = {
        id,
        title,
        description: buildDescription(hit, geo),
        alignment: idx % 2 === 0 ? 'left' : 'right',
        hidden: false,
        location: geo
            ? { center: geo.center, zoom: pickZoom(geo), pitch: 40, bearing: 0 }
            : FALLBACK_LOCATION,
        mapAnimation: 'flyTo',
        rotateAnimation: false,
        onChapterEnter: [],
        onChapterExit: [],
        source: {
            id: hit.id,
            username: hit.username,
            user_id: hit.user_id ?? null,
            userURL: hit.userURL || null,
            userImageURL: hit.userImageURL || null,
            license: hit.license || null,
            durationSec: hit.duration || null,
            created: hit.created || null,
            width: hit.width ?? null,
            height: hit.height ?? null,
            views: hit.views ?? null,
            downloads: hit.downloads ?? null,
            likes: hit.likes ?? null,
            comments: hit.comments ?? null,
            url: sourceLink(hit),
        },
    };
    if (type === 'audio') {
        chapter.audio = copyMedia(download?.localPath, 'audio') || hit.previewURL;
        chapter.autoPlayAudio = true;
    } else if (type === 'image') {
        chapter.image = copyMedia(download?.localPath, 'images') || hit.downloadURL;
    } else if (type === 'video') {
        chapter.video = copyMedia(download?.localPath, 'videos') || hit.downloadURL;
        if (hit.thumbnail) chapter.image = copyMedia(hit.thumbnail, 'images') || hit.thumbnail;
    }
    return chapter;
}

function buildGlobal(metadata, type, slug, hits) {
    const query = metadata.userInput?.query || slug;
    const source = type === 'audio' ? 'freesound.org' : 'Pixabay';
    const sourceUrl = type === 'audio' ? 'https://freesound.org/' : 'https://pixabay.com/';
    const noun = type === 'audio' ? 'Recordings' : type === 'image' ? 'Photos' : 'Videos';
    const title = `${capitalize(query)} ${type === 'audio' ? 'Soundscapes' : type === 'image' ? 'Images' : 'Videos'}`;
    const subtitle = `A scrollytelling tour of ${query} ${
        type === 'audio' ? 'field recordings' : type === 'image' ? 'photography' : 'film'
    } from ${source}`;
    const byline = `${noun}: ${source} contributors`;
    const users = [...new Set(hits.map((h) => h.username).filter(Boolean))];
    const links = users.length
        ? users.map((u) => {
              const hit = hits.find((h) => h.username === u);
              if (hit?.userURL) {
                  return `<a href="${hit.userURL}" target="_blank" rel="noopener noreferrer">${esc(u)}</a>`;
              }
              if (hit?.user_id) {
                  return `<a href="https://pixabay.com/users/${u}-${hit.user_id}/" target="_blank" rel="noopener noreferrer">${esc(u)}</a>`;
              }
              return `<a href="https://freesound.org/people/${u}/" target="_blank" rel="noopener noreferrer">${esc(u)}</a>`;
          })
        : [`${source} contributors`];
    const license = type === 'audio' ? 'CC BY 4.0' : 'Pixabay Content License';
    const footer = `${noun}: ${links.join(' &amp; ')} via <a href="${sourceUrl}" target="_blank" rel="noopener noreferrer">${source}</a> (${license}). Built with <a href="https://github.com/opengeos/GeoLibre" target="_blank" rel="noopener noreferrer">GeoLibre</a>.`;
    return { ...GLOBAL, title, subtitle, byline, footer };
}

function main() {
    if (!existsSync(COLLECTIONS)) {
        console.error(`audio-data-colleciton not found at ${COLLECTIONS}`);
        process.exit(1);
    }
    const dirs = readdirSync(COLLECTIONS, { withFileTypes: true })
        .filter((d) => d.isDirectory() && !d.name.startsWith('.'))
        .map((d) => d.name);

    let generated = 0;
    for (const name of dirs) {
        const dir = join(COLLECTIONS, name);
        const metaFile = findMetadataFile(dir);
        if (!metaFile) {
            console.log(`skip ${name}: no *_metadata.json`);
            continue;
        }
        const metaPath = join(dir, metaFile);
        let metadata;
        try {
            metadata = JSON.parse(readFileSync(metaPath, 'utf8'));
        } catch (e) {
            console.log(`skip ${name}: invalid ${metaFile} (${e.message})`);
            continue;
        }
        const { hits, downloads, geolocation } = normalizeMetadata(metadata);
        if (!hits.length) {
            console.log(`skip ${name}: no hits in ${metaFile}`);
            continue;
        }
        const type = detectType({ search: { hits } });
        const slug = slugFromMetadata(metaFile, name);
        const downloadsById = Object.fromEntries(downloads.map((d) => [String(d.id), d]));
        const chapters = hits.map((hit, idx) =>
            buildChapter({
                hit,
                download: downloadsById[String(hit.id)],
                type,
                slug,
                idx,
                geo: geoFor({ geolocation }, hit),
            })
        );
        const story = { ...buildGlobal(metadata, type, slug, hits), chapters };
        const out = join(ROOT, `${slug}-storymap.json`);
        if (!DRY_RUN) writeFileSync(out, JSON.stringify(story, null, 2) + '\n');
        console.log(
            `${DRY_RUN ? '[dry] ' : ''}✓ ${name} → ${basename(out)} (${chapters.length} chapters, ${type})`
        );
        generated++;
    }
    console.log(
        DRY_RUN
            ? `\n(dry run — nothing written) ${generated} collection(s) would be converted`
            : `\n${generated} storymap(s) written. Register new ones in src/stories.js to make them selectable.`
    );
}

main();
