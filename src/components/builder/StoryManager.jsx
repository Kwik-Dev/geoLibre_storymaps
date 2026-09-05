import React, { useCallback, useEffect, useState } from 'react';
import { listStories, deleteStory, updateStory } from '../../api/client.js';
import { useAuth } from '../../auth/AuthContext.jsx';
import AuthButtons from '../AuthButtons.jsx';

const btnStyle = {
    padding: '0.3rem 0.7rem',
    fontSize: '0.85rem',
    cursor: 'pointer',
    borderRadius: 4,
    border: '1px solid currentColor',
    background: 'transparent',
    color: 'inherit',
    textDecoration: 'none',
    fontFamily: 'inherit',
    whiteSpace: 'nowrap',
};
const inputStyle = {
    boxSizing: 'border-box',
    width: '100%',
    padding: '0.4rem 0.5rem',
    fontSize: '0.95rem',
    borderRadius: 4,
    border: '1px solid currentColor',
    background: 'transparent',
    color: 'inherit',
    fontFamily: 'inherit',
};
const rowStyle = {
    display: 'flex',
    alignItems: 'center',
    gap: '0.75rem',
    padding: '0.6rem 0',
    borderBottom: '1px solid rgba(127,127,127,0.25)',
};

// Storymap management (CRUD) page — #/manage. Lists the stories the current
// user can see (their own + public) with View / Edit / Delete actions and a
// "New story" entry point. Create/update/delete require a session; the list
// itself is readable anonymously.
export default function StoryManager() {
    const { user, loading: authLoading } = useAuth();
    const [stories, setStories] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [deleting, setDeleting] = useState(null);
    const [renaming, setRenaming] = useState(null);
    const [renameValue, setRenameValue] = useState('');
    const [savingRename, setSavingRename] = useState(false);

    const load = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const list = await listStories();
            setStories(list);
        } catch (e) {
            setError((e && e.message) || 'Failed to load stories.');
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        load();
    }, [load]);

    const handleDelete = async (s) => {
        if (!window.confirm(`Delete story "${s.title}"? This cannot be undone.`)) return;
        setDeleting(s.id);
        setError(null);
        try {
            await deleteStory(s.id);
            await load();
        } catch (e) {
            setError((e && e.message) || 'Failed to delete story.');
        } finally {
            setDeleting(null);
        }
    };

    const startRename = (s) => {
        setRenaming(s.id);
        setRenameValue(s.title || '');
        setError(null);
    };

    const cancelRename = () => {
        setRenaming(null);
        setRenameValue('');
    };

    const saveRename = async (s) => {
        const title = renameValue.trim();
        if (!title) {
            setError('Title is required.');
            return;
        }
        setSavingRename(true);
        setError(null);
        try {
            await updateStory(s.id, { title });
            await load();
            setRenaming(null);
            setRenameValue('');
        } catch (e) {
            setError((e && e.message) || 'Failed to rename story.');
        } finally {
            setSavingRename(false);
        }
    };

    return (
        <div style={{ maxWidth: 760, margin: '0 auto', padding: '2rem 1.5rem', color: 'inherit' }}>
            <div
                style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '0.5rem',
                }}
            >
                <h1 style={{ margin: 0 }}>My Storymaps</h1>
                <AuthButtons />
            </div>
            <p style={{ opacity: 0.85 }}>
                <a href="#/" style={{ textDecoration: 'underline' }}>
                    ← Back to all stories
                </a>
            </p>

            <div style={{ margin: '1rem 0' }}>
                <a href="#/create" style={{ ...btnStyle, padding: '0.5rem 1rem' }}>
                    + New story
                </a>
            </div>

            {error && <p style={{ color: '#c0392b' }}>{error}</p>}
            {loading && <p style={{ opacity: 0.85 }}>Loading stories…</p>}

            {!loading && stories.length === 0 && (
                <p style={{ opacity: 0.85 }}>
                    {authLoading
                        ? 'Checking session…'
                        : user
                          ? 'No stories yet. Create one to get started.'
                          : 'Sign in to manage your stories.'}
                </p>
            )}

            {!loading && stories.length > 0 && (
                <div>
                    {stories.map((s) => (
                        <div key={s.id} style={rowStyle}>
                            <div style={{ flex: 1, minWidth: 0 }}>
                                {renaming === s.id ? (
                                    <div>
                                        <input
                                            type="text"
                                            value={renameValue}
                                            onChange={(e) => setRenameValue(e.target.value)}
                                            onKeyDown={(e) => {
                                                if (e.key === 'Enter') saveRename(s);
                                                if (e.key === 'Escape') cancelRename();
                                            }}
                                            autoFocus
                                            maxLength={120}
                                            style={{ ...inputStyle, marginBottom: '0.25rem' }}
                                        />
                                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                                            <button
                                                type="button"
                                                style={btnStyle}
                                                disabled={savingRename}
                                                onClick={() => saveRename(s)}
                                            >
                                                {savingRename ? 'Saving…' : 'Save'}
                                            </button>
                                            <button type="button" style={btnStyle} onClick={cancelRename}>
                                                Cancel
                                            </button>
                                        </div>
                                    </div>
                                ) : (
                                    <>
                                        <div style={{ fontWeight: 600 }}>{s.title}</div>
                                        <div style={{ opacity: 0.7, fontSize: '0.85rem' }}>
                                            {s.visibility} · {s.status} · {s.created_at}
                                        </div>
                                    </>
                                )}
                            </div>
                            <a href={`#/stories/${encodeURIComponent(s.id)}`} style={btnStyle}>
                                View
                            </a>
                            <a href={`#/stories/${encodeURIComponent(s.id)}/edit`} style={btnStyle}>
                                Edit
                            </a>
                            <button type="button" style={btnStyle} onClick={() => startRename(s)}>
                                Rename
                            </button>
                            <button
                                type="button"
                                style={btnStyle}
                                disabled={deleting === s.id}
                                onClick={() => handleDelete(s)}
                            >
                                {deleting === s.id ? 'Deleting…' : 'Delete'}
                            </button>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
