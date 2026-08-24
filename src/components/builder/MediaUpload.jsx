import React, { useRef, useState } from 'react';
import { AudioProvider } from '../../audio/AudioContext.jsx';
import ChapterCard from '../ChapterCard.jsx';
import { getToken } from '../../api/client.js';

// P6.3 — MediaUpload.
//
// The builder's media picker: a single control that lets a chapter's media be
// one of
//   - none      → no media
//   - external  → an https:// URL the user supplies (client-validated to
//                 mirror P4.2 exactly)
//   - local     → a file uploaded to the Go server via POST /api/media/upload
//                 (multipart 'file'); the returned media_asset_id is stored.
//
// On any change it emits the full grouped media value up via `onChange`
// (media_type / media_ref_type / media_external_url / media_asset_id) — the
// exact field set P4.3/P3.2 persist on the chapter. The parent owns that state
// (controlled component) and writes it to the chapter on save.
//
// WYSIWYG: the preview renders the SAME <ChapterCard> media renderer the
// reader uses (single source of media-display logic), so what the builder
// shows for a video (poster), an image, or an audio waveform is exactly what
// a reader sees. ChapterCard's audio toggle needs the shared AudioProvider, so
// the preview is wrapped in one here.

// The grouped media value MediaUpload emits / accepts.
export const EMPTY_MEDIA = {
    media_type: 'none',
    media_ref_type: 'none',
    media_external_url: '',
    media_asset_id: '',
};

const base = (import.meta.env.VITE_API || '/api').replace(/\/+$/, '');

/**
 * Client-side external-URL validator mirroring P4.2 (server/internal/media/
 * external.go): https scheme only, length ≤ 2048. The server remains the
 * authority (host allow-list, etc.) — this is a UX guard + shape check.
 */
export function validateExternalURL(value) {
    const s = (value || '').trim();
    if (!s) return { ok: false, reason: 'A media URL is required.' };
    if (s.length > 2048) return { ok: false, reason: 'The URL must be 2048 characters or fewer.' };
    let u;
    try {
        u = new URL(s);
    } catch (_) {
        return { ok: false, reason: 'That does not look like a valid URL.' };
    }
    if (u.protocol !== 'https:') return { ok: false, reason: 'Only https:// URLs are allowed.' };
    return { ok: true, url: s };
}

function mimeToType(mime) {
    const m = (mime || '').toLowerCase();
    if (m.startsWith('image/')) return 'image';
    if (m.startsWith('video/')) return 'video';
    if (m.startsWith('audio/')) return 'audio';
    return 'none';
}

const inputStyle = {
    boxSizing: 'border-box',
    width: '100%',
    padding: '0.5rem 0.6rem',
    marginBottom: '0.5rem',
    fontSize: '1rem',
    borderRadius: 4,
    border: '1px solid currentColor',
    background: 'transparent',
    color: 'inherit',
};
const labelStyle = { display: 'block', marginBottom: '0.2rem', opacity: 0.9, fontWeight: 600 };
const fieldRow = { marginBottom: '0.35rem' };
const previewBox = {
    border: '1px dashed currentColor',
    borderRadius: 4,
    padding: '0.75rem',
    marginTop: '0.5rem',
    background: 'transparent',
};

/**
 * Props:
 *   value   — { media_type, media_ref_type, media_external_url, media_asset_id }
 *   onChange(value) — fired with a new grouped value on any media change.
 *   onUploaded(asset) — optional, fired after a local file upload succeeds
 *                       (asset = { id, url, bytes, mime }).
 *   allowlistHint — optional string hinting the server's media host allow-list
 *                   (shown under the external URL input when provided).
 */
export default function MediaUpload({ value = EMPTY_MEDIA, onChange, onUploaded, allowlistHint }) {
    const ref = value.media_ref_type || 'none';
    const type = value.media_type || 'none';

    // Transient upload state (the uploaded asset is persisted via onChange).
    const [uploading, setUploading] = useState(false);
    const [progress, setProgress] = useState(0);
    const [uploadError, setUploadError] = useState(null);
    const [lastAsset, setLastAsset] = useState(null);
    const fileRef = useRef(null);

    const extVal = validateExternalURL(value.media_external_url);

    const setRefType = (nextRef) => {
        if (nextRef === ref) return;
        if (nextRef === 'none') {
            onChange({ ...EMPTY_MEDIA });
            return;
        }
        // Keep the existing type/url/asset when switching ref modes; external
        // keeps the url, local keeps the asset id.
        onChange({ ...value, media_ref_type: nextRef });
    };

    const setType = (nextType) => {
        onChange({ ...value, media_type: nextType });
    };

    const setExternalUrl = (raw) => {
        onChange({ ...value, media_external_url: raw });
    };

    // Upload a local file (multipart 'file') with a progress bar via XHR (fetch
    // has no upload progress). On success store the returned asset id + derive
    // the media type from the server's magic-byte MIME (P4.1).
    const doUpload = (file) => {
        if (!file) return;
        setUploading(true);
        setProgress(0);
        setUploadError(null);
        setLastAsset(null);

        const fd = new FormData();
        fd.append('file', file);

        const xhr = new XMLHttpRequest();
        xhr.open('POST', `${base}/media/upload`);
        const t = getToken();
        if (t) xhr.setRequestHeader('Authorization', `Bearer ${t}`);
        xhr.withCredentials = true;

        xhr.upload.onprogress = (e) => {
            if (e.lengthComputable) setProgress(e.loaded / e.total);
        };
        xhr.onload = () => {
            setUploading(false);
            if (xhr.status >= 200 && xhr.status < 300) {
                let asset;
                try {
                    asset = JSON.parse(xhr.responseText);
                } catch (_) {
                    setUploadError('The server returned an unreadable upload response.');
                    return;
                }
                setLastAsset(asset);
                const mtype = mimeToType(asset.mime);
                onChange({
                    media_type: mtype === 'none' ? type : mtype,
                    media_ref_type: 'local',
                    media_external_url: '',
                    media_asset_id: String(asset.id),
                });
                if (onUploaded) onUploaded(asset);
            } else {
                let detail = '';
                try {
                    const j = JSON.parse(xhr.responseText);
                    detail = (j && j.error) || '';
                } catch (_) {
                    /* non-JSON error body */
                }
                setUploadError(detail || `Upload failed (${xhr.status}).`);
            }
        };
        xhr.onerror = () => {
            setUploading(false);
            setUploadError('Network error during upload — is the server reachable?');
        };
        xhr.send(fd);
    };

    // Resolve the media URL shown by the preview, mirroring the server's
    // mediaURL (P3.3): external → the URL; local → /media/<asset_id>.
    let previewSrc = '';
    let previewPoster = '';
    if (ref === 'external') {
        previewSrc = value.media_external_url || '';
        previewPoster = previewSrc;
    } else if (ref === 'local' && value.media_asset_id) {
        previewSrc = `/media/${value.media_asset_id}`;
        previewPoster = previewSrc;
    }

    // Build a synthetic legacy chapter for the ChapterCard preview.
    const previewChapter = {
        id: 'media-upload-preview',
        title: 'Media preview',
        description: '',
    };
    if (type === 'image' && previewSrc) previewChapter.image = previewSrc;
    else if (type === 'video' && previewSrc) {
        previewChapter.video = previewSrc;
        previewChapter.image = previewPoster; // poster
    } else if (type === 'audio' && previewSrc) previewChapter.audio = previewSrc;

    const showPreview = type !== 'none' && previewSrc;

    return (
        <div style={{ width: '100%' }}>
            <div style={{ marginBottom: '0.35rem' }}>
                <label htmlFor="mu-reftype" style={labelStyle}>Media reference</label>
                <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
                    <label>
                        <input
                            type="radio"
                            name="mu-reftype"
                            checked={ref === 'none'}
                            onChange={() => setRefType('none')}
                        />{' '}
                        None
                    </label>
                    <label>
                        <input
                            type="radio"
                            name="mu-reftype"
                            checked={ref === 'external'}
                            onChange={() => setRefType('external')}
                        />{' '}
                        External URL
                    </label>
                    <label>
                        <input
                            type="radio"
                            name="mu-reftype"
                            checked={ref === 'local'}
                            onChange={() => setRefType('local')}
                        />{' '}
                        Upload a file
                    </label>
                </div>
            </div>

            {ref !== 'none' && (
                <div style={fieldRow}>
                    <label htmlFor="mu-type" style={labelStyle}>Media type</label>
                    <select
                        id="mu-type"
                        value={type}
                        onChange={(e) => setType(e.target.value)}
                        style={inputStyle}
                    >
                        <option value="image">Image</option>
                        <option value="video">Video</option>
                        <option value="audio">Audio</option>
                    </select>
                </div>
            )}

            {ref === 'external' && (
                <div style={fieldRow}>
                    <label htmlFor="mu-url" style={labelStyle}>External URL (https://)</label>
                    <input
                        id="mu-url"
                        type="text"
                        value={value.media_external_url || ''}
                        onChange={(e) => setExternalUrl(e.target.value)}
                        style={inputStyle}
                        placeholder="https://example.com/media/clip.mp4"
                        maxLength={2048}
                    />
                    {value.media_external_url && !extVal.ok && (
                        <p style={{ color: '#c0392b', fontSize: '0.85rem', margin: '0 0 0.35rem' }}>{extVal.reason}</p>
                    )}
                    {allowlistHint ? (
                        <p style={{ fontSize: '0.85rem', opacity: 0.75, margin: '0 0 0.35rem' }}>
                            Allowed hosts: {allowlistHint}
                        </p>
                    ) : (
                        <p style={{ fontSize: '0.85rem', opacity: 0.75, margin: '0 0 0.35rem' }}>
                            Only secure <code>https://</code> URLs (max 2048 chars) are accepted.
                        </p>
                    )}
                </div>
            )}

            {ref === 'local' && (
                <div style={fieldRow}>
                    <label htmlFor="mu-file" style={labelStyle}>Choose a file</label>
                    <input
                        id="mu-file"
                        type="file"
                        ref={fileRef}
                        accept="image/*,video/*,audio/*"
                        disabled={uploading}
                        onChange={(e) => doUpload(e.target.files && e.target.files[0])}
                        style={inputStyle}
                    />
                    {uploading && (
                        <div style={{ marginBottom: '0.5rem' }}>
                            <div
                                style={{
                                    height: 8,
                                    borderRadius: 4,
                                    border: '1px solid currentColor',
                                    overflow: 'hidden',
                                }}
                            >
                                <div
                                    style={{
                                        width: `${Math.round((progress || 0) * 100)}%`,
                                        height: '100%',
                                        background: 'currentColor',
                                        opacity: 0.7,
                                    }}
                                />
                            </div>
                            <p style={{ fontSize: '0.85rem', opacity: 0.8, margin: '0.2rem 0 0' }}>
                                Uploading… {Math.round((progress || 0) * 100)}%
                            </p>
                        </div>
                    )}
                    {uploadError && <p style={{ color: '#c0392b', fontSize: '0.85rem', margin: '0 0 0.35rem' }}>{uploadError}</p>}
                    {lastAsset && (
                        <p style={{ fontSize: '0.85rem', opacity: 0.85, margin: '0 0 0.35rem' }}>
                            Uploaded {lastAsset.bytes} bytes ({lastAsset.mime}) → <code>{lastAsset.url}</code>
                        </p>
                    )}
                    {!uploading && ref === 'local' && value.media_asset_id && (
                        <p style={{ fontSize: '0.85rem', opacity: 0.85, margin: '0 0 0.35rem' }}>
                            Media asset id: <code>{value.media_asset_id}</code> (served at{' '}
                            <code>/media/{value.media_asset_id}</code>). Choose a file again to replace it.
                        </p>
                    )}
                </div>
            )}

            {showPreview && (
                <div style={previewBox}>
                    <div style={labelStyle}>Preview (as a reader sees it)</div>
                    <AudioProvider>
                        <ChapterCard chapter={previewChapter} index={0} theme="dark" />
                    </AudioProvider>
                </div>
            )}
        </div>
    );
}
