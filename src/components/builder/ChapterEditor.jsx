import React, { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../../api/client.js';
import Markdown from '../Markdown.jsx';

// P6.2 — ChapterEditor.
//
// Lists / adds / edits / deletes / reorders the chapters of a given story,
// all via the nested chapters API (P3.2) under /api/stories/:id/chapters.
// The editor reuses the sanitized <Markdown> component (P5.1) for a *live*
// preview of description_md — user HTML is never injected raw.
//
// Media is a *slot* handed to P6.3 (MediaUpload): a minimal editable media
// section is provided so saves persist the correct media fields, but P6.3
// replaces it with the drag-drop / external-URL MediaUpload component.
//
// Props:
//   storyId  — numeric story id the editor operates on.

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

const newDraft = () => ({
    id: null, // null → create; a number → update
    title: '',
    description_md: '',
    alignment: 'center',
    hidden: false,
    lng: 0,
    lat: 0,
    zoom: 2,
    map_animation: 'flyTo',
    rotate_animation: false,
    on_chapter_enter: '[]',
    on_chapter_exit: '[]',
    media_type: 'none',
    media_ref_type: 'none',
    media_external_url: '',
    media_asset_id: '',
});

// Map a chapter from the API into the editor's flat draft shape.
function chapterToDraft(c) {
    const d = newDraft();
    d.id = c.id;
    d.title = c.title || '';
    d.description_md = c.description_md || '';
    d.alignment = c.alignment || 'center';
    d.hidden = Boolean(c.hidden);
    d.map_animation = c.map_animation || 'flyTo';
    d.rotate_animation = Boolean(c.rotate_animation);
    if (c.location && Array.isArray(c.location.center) && c.location.center.length === 2) {
        d.lng = c.location.center[0];
        d.lat = c.location.center[1];
        d.zoom = typeof c.location.zoom === 'number' ? c.location.zoom : 2;
    }
    if (c.on_chapter_enter) d.on_chapter_enter = JSON.stringify(c.on_chapter_enter);
    if (c.on_chapter_exit) d.on_chapter_exit = JSON.stringify(c.on_chapter_exit);
    d.media_type = c.media_type || 'none';
    d.media_ref_type = c.media_ref_type || 'none';
    d.media_external_url = c.media_external_url || '';
    d.media_asset_id = c.media_asset_id != null ? String(c.media_asset_id) : '';
    return d;
}

// Build the payload for create/update from the draft. The location and the
// on_enter/on_exit arrays are only included when they parse; media fields are
// sent as a grouped set (see P3.2/P4.3).
function draftToPayload(d) {
    const payload = {
        title: (d.title || '').trim(),
        description_md: (d.description_md || '').trim(),
        alignment: d.alignment,
        hidden: d.hidden,
        map_animation: d.map_animation,
        rotate_animation: d.rotate_animation,
    };

    const lng = Number(d.lng);
    const lat = Number(d.lat);
    const zoom = Number(d.zoom);
    if (Number.isFinite(lng) && Number.isFinite(lat) && Number.isFinite(zoom)) {
        payload.location = { center: [lng, lat], zoom };
    }

    const enter = parseJsonArray(d.on_chapter_enter);
    if (enter) payload.on_chapter_enter = enter;
    const exit = parseJsonArray(d.on_chapter_exit);
    if (exit) payload.on_chapter_exit = exit;

    // Media slot (P6.3 will replace). Send the grouped media fields.
    payload.media_type = d.media_type;
    payload.media_ref_type = d.media_ref_type;
    payload.media_external_url = (d.media_external_url || '').trim();
    const asset = (d.media_asset_id || '').trim();
    if (asset === '') {
        payload.media_asset_id = null;
    } else {
        const n = Number(asset);
        payload.media_asset_id = Number.isFinite(n) ? n : null;
    }
    return payload;
}

function parseJsonArray(s) {
    const t = (s || '').trim();
    if (t === '') return [];
    try {
        const v = JSON.parse(t);
        return Array.isArray(v) ? v : null;
    } catch (_) {
        return null;
    }
}

export default function ChapterEditor({ storyId }) {
    const [chapters, setChapters] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [notice, setNotice] = useState(null);
    const [editing, setEditing] = useState(false); // true when the form is open
    const [draft, setDraft] = useState(newDraft());
    const [saving, setSaving] = useState(false);
    const [formError, setFormError] = useState(null);

    const load = useCallback(async () => {
        if (!storyId) return;
        setLoading(true);
        setError(null);
        try {
            const resp = await apiFetch(`/stories/${encodeURIComponent(storyId)}/chapters`);
            const data = await resp.json();
            setChapters(Array.isArray(data) ? data : (data && data.chapters) || []);
        } catch (e) {
            setError((e && e.message) || 'Failed to load chapters.');
        } finally {
            setLoading(false);
        }
    }, [storyId]);

    useEffect(() => {
        load();
    }, [load]);

    const flash = (msg) => {
        setNotice(msg);
        window.setTimeout(() => setNotice(null), 4000);
    };

    const startCreate = () => {
        setDraft(newDraft());
        setFormError(null);
        setEditing(true);
    };

    const startEdit = (c) => {
        setDraft(chapterToDraft(c));
        setFormError(null);
        setEditing(true);
    };

    const cancelEdit = () => {
        setEditing(false);
        setFormError(null);
    };

    const setField = (key, value) => setDraft((d) => ({ ...d, [key]: value }));

    const save = async (e) => {
        e.preventDefault();
        if (!(draft.title || '').trim()) {
            setFormError('Title is required.');
            return;
        }
        // Client-side guard mirroring the backend media matrix (§6): an
        // external ref needs a URL; a local ref needs an asset id.
        if (draft.media_ref_type === 'external' && !(draft.media_external_url || '').trim()) {
            setFormError('An external media URL is required when media_ref_type is "external".');
            return;
        }
        if (draft.media_ref_type === 'local' && !(draft.media_asset_id || '').trim()) {
            setFormError('A media asset id is required when media_ref_type is "local".');
            return;
        }
        setFormError(null);
        setSaving(true);
        try {
            const payload = draftToPayload(draft);
            const isNew = draft.id == null;
            const path = isNew
                ? `/stories/${encodeURIComponent(storyId)}/chapters`
                : `/stories/${encodeURIComponent(storyId)}/chapters/${draft.id}`;
            const resp = await apiFetch(path, {
                method: isNew ? 'POST' : 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });
            if (!resp.ok && !resp.status) {
                // non-2xx already thrown by apiFetch; this is defensive.
            }
            flash(isNew ? 'Chapter added.' : 'Chapter saved.');
            setEditing(false);
            await load();
        } catch (err) {
            const msg = (err && err.message) || 'Failed to save the chapter.';
            setFormError(err && err.status === 403 ? 'You do not have permission to edit this story.' : msg);
        } finally {
            setSaving(false);
        }
    };

    const remove = async (c) => {
        if (!window.confirm(`Delete chapter "${c.title || c.id}"?`)) return;
        try {
            await apiFetch(`/stories/${encodeURIComponent(storyId)}/chapters/${c.id}`, { method: 'DELETE' });
            flash('Chapter deleted.');
            await load();
        } catch (err) {
            setError((err && err.message) || 'Failed to delete the chapter.');
        }
    };

    // Move a chapter up/down then persist the new order via the reorder
    // endpoint (required by P6.2 GOTCHAS — reorder MUST hit the API).
    const move = async (index, dir) => {
        const target = index + dir;
        if (target < 0 || target >= chapters.length) return;
        const next = chapters.slice();
        const [c] = next.splice(index, 1);
        next.splice(target, 0, c);
        await persistOrder(next);
    };

    const persistOrder = async (ordered) => {
        const body = ordered.map((c, i) => ({ id: c.id, position: i + 1 }));
        try {
            const resp = await apiFetch(`/stories/${encodeURIComponent(storyId)}/chapters/reorder`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
            });
            const data = await resp.json();
            setChapters(Array.isArray(data) ? data : (data && data.chapters) || ordered);
            flash('Chapter order saved.');
        } catch (err) {
            setError((err && err.message) || 'Failed to reorder chapters.');
        }
    };

    const preview = draft.description_md || '';

    return (
        <div style={{ maxWidth: 760, margin: '0 auto', padding: '2rem 1.5rem', color: 'inherit' }}>
            <h1>Chapters</h1>
            <p style={{ opacity: 0.85 }}>
                Add, edit, reorder, or remove the chapters of this story. Use the buttons
                to the left of each chapter to change its order.
            </p>

            {notice && <p style={{ color: '#27ae60', fontStyle: 'italic' }}>{notice}</p>}
            {error && <p style={{ color: '#c0392b' }}>{error}</p>}

            {loading ? (
                <p style={{ opacity: 0.8 }}>Loading chapters…</p>
            ) : chapters.length === 0 ? (
                <p style={{ opacity: 0.85 }}>No chapters yet — add one to get started.</p>
            ) : (
                <ul style={{ listStyle: 'none', padding: 0 }}>
                    {chapters.map((c, i) => (
                        <li
                            key={c.id}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '0.5rem',
                                padding: '0.5rem 0',
                                borderBottom: '1px solid currentColor',
                            }}
                        >
                            <span style={{ display: 'inline-flex', flexDirection: 'column' }}>
                                <button
                                    type="button"
                                    disabled={i === 0}
                                    onClick={() => move(i, -1)}
                                    title="Move up"
                                    style={miniBtn}
                                >
                                    ▲
                                </button>
                                <button
                                    type="button"
                                    disabled={i === chapters.length - 1}
                                    onClick={() => move(i, 1)}
                                    title="Move down"
                                    style={miniBtn}
                                >
                                    ▼
                                </button>
                            </span>
                            <span style={{ flex: 1 }}>
                                <strong>{c.title || `Chapter ${i + 1}`}</strong>
                                <span style={{ opacity: 0.7 }}> · pos {c.position}</span>
                            </span>
                            <button type="button" style={linkBtn} onClick={() => startEdit(c)}>
                                Edit
                            </button>
                            <button type="button" style={{ ...linkBtn, color: '#c0392b' }} onClick={() => remove(c)}>
                                Delete
                            </button>
                        </li>
                    ))}
                </ul>
            )}

            <div style={{ marginTop: '1.5rem' }}>
                {!editing ? (
                    <button type="button" style={primaryBtn} onClick={startCreate}>
                        + Add chapter
                    </button>
                ) : (
                    <form onSubmit={save} noValidate style={{ marginTop: '1rem' }}>
                        <h2 style={{ margin: '0 0 0.75rem' }}>{draft.id == null ? 'Add chapter' : `Edit: ${draft.title || draft.id}`}</h2>

                        <div style={fieldRow}>
                            <label htmlFor="ce-title" style={labelStyle}>Title *</label>
                            <input
                                id="ce-title"
                                type="text"
                                value={draft.title}
                                onChange={(e) => setField('title', e.target.value)}
                                placeholder="e.g. The Pacific"
                                style={inputStyle}
                                maxLength={200}
                                required
                            />
                        </div>

                        <div style={fieldRow}>
                            <label htmlFor="ce-desc" style={labelStyle}>Description (Markdown)</label>
                            <textarea
                                id="ce-desc"
                                value={draft.description_md}
                                onChange={(e) => setField('description_md', e.target.value)}
                                rows={4}
                                style={inputStyle}
                                placeholder={'Supports **bold**, *italics*, ## headings, - lists, [links](https://…), etc.'}
                            />
                        </div>

                        {preview ? (
                            <div style={previewBox}>
                                <div style={labelStyle}>Live preview</div>
                                {/* Sanitized — never raw HTML (P5.1 / §10). */}
                                <Markdown text={preview} />
                            </div>
                        ) : null}

                        <div style={fieldRow}>
                            <label htmlFor="ce-align" style={labelStyle}>Alignment</label>
                            <select
                                id="ce-align"
                                value={draft.alignment}
                                onChange={(e) => setField('alignment', e.target.value)}
                                style={inputStyle}
                            >
                                <option value="left">Left</option>
                                <option value="center">Center</option>
                                <option value="right">Right</option>
                            </select>
                        </div>

                        <fieldset style={{ border: '1px solid currentColor', borderRadius: 4, padding: '0.75rem', marginBottom: '0.75rem' }}>
                            <legend style={{ opacity: 0.9 }}>Location (camera)</legend>
                            <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                                <div style={{ flex: 1, minWidth: 120 }}>
                                    <label htmlFor="ce-lng" style={labelStyle}>Longitude</label>
                                    <input
                                        id="ce-lng"
                                        type="number"
                                        step="any"
                                        value={draft.lng}
                                        onChange={(e) => setField('lng', e.target.value)}
                                        style={inputStyle}
                                    />
                                </div>
                                <div style={{ flex: 1, minWidth: 120 }}>
                                    <label htmlFor="ce-lat" style={labelStyle}>Latitude</label>
                                    <input
                                        id="ce-lat"
                                        type="number"
                                        step="any"
                                        value={draft.lat}
                                        onChange={(e) => setField('lat', e.target.value)}
                                        style={inputStyle}
                                    />
                                </div>
                                <div style={{ flex: 1, minWidth: 160 }}>
                                    <label htmlFor="ce-zoom" style={labelStyle}>
                                        Zoom: {Number(draft.zoom)}
                                    </label>
                                    <input
                                        id="ce-zoom"
                                        type="range"
                                        min="0"
                                        max="20"
                                        step="0.1"
                                        value={draft.zoom}
                                        onChange={(e) => setField('zoom', e.target.value)}
                                        style={{ ...inputStyle, padding: 0 }}
                                    />
                                </div>
                            </div>
                            <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
                                <label>
                                    <input
                                        type="checkbox"
                                        checked={draft.rotate_animation}
                                        onChange={(e) => setField('rotate_animation', e.target.checked)}
                                    />{' '}
                                    Rotate animation
                                </label>
                            </div>
                        </fieldset>

                        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
                            <div style={fieldRow}>
                                <label htmlFor="ce-mapanim" style={labelStyle}>Map animation</label>
                                <select
                                    id="ce-mapanim"
                                    value={draft.map_animation}
                                    onChange={(e) => setField('map_animation', e.target.value)}
                                    style={inputStyle}
                                >
                                    <option value="flyTo">flyTo</option>
                                    <option value="easeTo">easeTo</option>
                                </select>
                            </div>
                            <label style={{ alignSelf: 'center', marginBottom: '0.75rem' }}>
                                <input
                                    type="checkbox"
                                    checked={draft.hidden}
                                    onChange={(e) => setField('hidden', e.target.checked)}
                                />{' '}
                                Hidden
                            </label>
                        </div>

                        <fieldset style={{ border: '1px solid currentColor', borderRadius: 4, padding: '0.75rem', marginBottom: '0.75rem' }}>
                            <legend style={{ opacity: 0.9 }}>Layer opacity on enter / exit (JSON arrays)</legend>
                            <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                                <div style={{ flex: 1, minWidth: 200 }}>
                                    <label htmlFor="ce-enter" style={labelStyle}>on_chapter_enter</label>
                                    <textarea
                                        id="ce-enter"
                                        rows={2}
                                        value={draft.on_chapter_enter}
                                        onChange={(e) => setField('on_chapter_enter', e.target.value)}
                                        style={{ ...inputStyle, fontFamily: 'monospace', fontSize: '0.85rem' }}
                                        placeholder='[]'
                                    />
                                </div>
                                <div style={{ flex: 1, minWidth: 200 }}>
                                    <label htmlFor="ce-exit" style={labelStyle}>on_chapter_exit</label>
                                    <textarea
                                        id="ce-exit"
                                        rows={2}
                                        value={draft.on_chapter_exit}
                                        onChange={(e) => setField('on_chapter_exit', e.target.value)}
                                        style={{ ...inputStyle, fontFamily: 'monospace', fontSize: '0.85rem' }}
                                        placeholder='[]'
                                    />
                                </div>
                            </div>
                        </fieldset>

                        <fieldset style={{ border: '1px solid currentColor', borderRadius: 4, padding: '0.75rem', marginBottom: '0.75rem' }}>
                            <legend style={{ opacity: 0.9 }}>Media (slot — P6.3 MediaUpload replaces this)</legend>
                            <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                                <div>
                                    <label htmlFor="ce-mt" style={labelStyle}>Type</label>
                                    <select
                                        id="ce-mt"
                                        value={draft.media_type}
                                        onChange={(e) => setField('media_type', e.target.value)}
                                        style={inputStyle}
                                    >
                                        <option value="none">none</option>
                                        <option value="image">image</option>
                                        <option value="video">video</option>
                                        <option value="audio">audio</option>
                                    </select>
                                </div>
                                <div>
                                    <label htmlFor="ce-mrt" style={labelStyle}>Reference</label>
                                    <select
                                        id="ce-mrt"
                                        value={draft.media_ref_type}
                                        onChange={(e) => setField('media_ref_type', e.target.value)}
                                        style={inputStyle}
                                    >
                                        <option value="none">none</option>
                                        <option value="external">external</option>
                                        <option value="local">local</option>
                                    </select>
                                </div>
                                {draft.media_ref_type === 'external' && (
                                    <div style={{ flex: 1, minWidth: 220 }}>
                                        <label htmlFor="ce-murl" style={labelStyle}>External URL (https://)</label>
                                        <input
                                            id="ce-murl"
                                            type="text"
                                            value={draft.media_external_url}
                                            onChange={(e) => setField('media_external_url', e.target.value)}
                                            style={inputStyle}
                                            placeholder="https://…"
                                            maxLength={2048}
                                        />
                                    </div>
                                )}
                                {draft.media_ref_type === 'local' && (
                                    <div style={{ flex: 1, minWidth: 160 }}>
                                        <label htmlFor="ce-maid" style={labelStyle}>Media asset id</label>
                                        <input
                                            id="ce-maid"
                                            type="text"
                                            value={draft.media_asset_id}
                                            onChange={(e) => setField('media_asset_id', e.target.value)}
                                            style={inputStyle}
                                            placeholder="id from upload"
                                        />
                                    </div>
                                )}
                            </div>
                            <p style={{ fontSize: '0.85rem', opacity: 0.7, margin: '0.25rem 0 0' }}>
                                The drag-drop / external-URL media picker (P6.3) replaces this basic slot.
                            </p>
                        </fieldset>

                        {formError && <p style={{ color: '#c0392b' }}>{formError}</p>}

                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                            <button type="submit" disabled={saving} style={primaryBtn}>
                                {saving ? 'Saving…' : draft.id == null ? 'Add chapter' : 'Save chapter'}
                            </button>
                            <button type="button" style={linkBtn} onClick={cancelEdit}>
                                Cancel
                            </button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    );
}

const primaryBtn = {
    padding: '0.6rem 1.2rem',
    fontSize: '1rem',
    cursor: 'pointer',
    borderRadius: 4,
    border: '1px solid currentColor',
    background: 'transparent',
    color: 'inherit',
};
const linkBtn = {
    padding: '0.3rem 0.6rem',
    fontSize: '0.9rem',
    cursor: 'pointer',
    borderRadius: 4,
    border: '1px solid currentColor',
    background: 'transparent',
    color: 'inherit',
};
const miniBtn = {
    padding: '0',
    width: 22,
    height: 18,
    lineHeight: 1,
    fontSize: '0.7rem',
    cursor: 'pointer',
    border: '1px solid currentColor',
    background: 'transparent',
    color: 'inherit',
};
const previewBox = {
    border: '1px dashed currentColor',
    borderRadius: 4,
    padding: '0.75rem',
    marginBottom: '0.75rem',
    background: 'transparent',
};
