import React, { useCallback, useEffect, useState } from 'react';
import { listStories, deleteStory, updateStory, approveStory, rejectStory } from '../../api/client.js';
import { useAuth } from '../../auth/AuthContext.jsx';
import AuthButtons from '../AuthButtons.jsx';
import AdminLogin from './AdminLogin.jsx';

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

// Status badge colors: draft = neutral, pending = amber (awaiting review),
// approved = green (live in the public list).
const STATUS_BADGE = {
    draft: { background: 'rgba(127,127,127,0.2)', color: 'inherit' },
    pending: { background: 'rgba(230,126,34,0.25)', color: '#e67e22' },
    approved: { background: 'rgba(46,204,113,0.2)', color: '#27ae60' },
};

function StatusBadge({ status }) {
    const style = STATUS_BADGE[status] || STATUS_BADGE.draft;
    return (
        <span
            style={{
                display: 'inline-block',
                padding: '0.05rem 0.45rem',
                borderRadius: 999,
                fontSize: '0.75rem',
                fontWeight: 600,
                ...style,
            }}
        >
            {status}
        </span>
    );
}

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
    const [moderating, setModerating] = useState(null);
    const [showAdminLogin, setShowAdminLogin] = useState(false);

    // Admins can approve/reject stories (the moderation gate, P7.2). The
    // whoami response carries `admin` (bool) and `role`; either is sufficient.
    const isAdmin = Boolean(user && (user.admin || user.role === 'admin'));
    // The moderation queue: public stories awaiting an admin decision.
    const pendingStories = stories.filter((s) => s.status === 'pending');

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

    const handleApprove = async (s) => {
        setModerating(s.id);
        setError(null);
        try {
            await approveStory(s.id);
            await load();
        } catch (e) {
            setError((e && e.message) || 'Failed to approve story.');
        } finally {
            setModerating(null);
        }
    };

    const handleReject = async (s) => {
        setModerating(s.id);
        setError(null);
        try {
            await rejectStory(s.id);
            await load();
        } catch (e) {
            setError((e && e.message) || 'Failed to reject story.');
        } finally {
            setModerating(null);
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

            {isAdmin && pendingStories.length > 0 && (
                <div
                    style={{
                        margin: '1rem 0',
                        padding: '0.75rem 1rem',
                        border: '1px solid rgba(230,126,34,0.5)',
                        borderRadius: 6,
                    }}
                >
                    <h2 style={{ margin: '0 0 0.5rem', fontSize: '1.1rem' }}>
                        Moderation queue ({pendingStories.length})
                    </h2>
                    {pendingStories.map((s) => (
                        <div
                            key={s.id}
                            style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '0.4rem 0' }}
                        >
                            <div style={{ flex: 1, minWidth: 0 }}>
                                <div style={{ fontWeight: 600 }}>{s.title}</div>
                                <div style={{ opacity: 0.7, fontSize: '0.85rem' }}>{s.created_at}</div>
                            </div>
                            <a href={`#/stories/${encodeURIComponent(s.id)}`} style={btnStyle}>
                                View
                            </a>
                            <button
                                type="button"
                                style={btnStyle}
                                disabled={moderating === s.id}
                                onClick={() => handleApprove(s)}
                            >
                                {moderating === s.id ? '…' : 'Approve'}
                            </button>
                            <button
                                type="button"
                                style={btnStyle}
                                disabled={moderating === s.id}
                                onClick={() => handleReject(s)}
                            >
                                {moderating === s.id ? '…' : 'Reject'}
                            </button>
                        </div>
                    ))}
                </div>
            )}

            {!isAdmin && !authLoading && (
                <div style={{ margin: '1rem 0' }}>
                    {showAdminLogin ? (
                        <div
                            style={{
                                padding: '0.75rem 1rem',
                                border: '1px solid rgba(127,127,127,0.4)',
                                borderRadius: 6,
                            }}
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    justifyContent: 'space-between',
                                    alignItems: 'center',
                                    marginBottom: '0.5rem',
                                }}
                            >
                                <span style={{ fontWeight: 600 }}>Admin sign in</span>
                                <button type="button" style={btnStyle} onClick={() => setShowAdminLogin(false)}>
                                    Cancel
                                </button>
                            </div>
                            <AdminLogin onSuccess={() => setShowAdminLogin(false)} />
                        </div>
                    ) : (
                        <button type="button" style={btnStyle} onClick={() => setShowAdminLogin(true)}>
                            Admin sign in
                        </button>
                    )}
                </div>
            )}

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
                                        <div
                                            style={{
                                                opacity: 0.7,
                                                fontSize: '0.85rem',
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: '0.4rem',
                                            }}
                                        >
                                            <span>{s.visibility}</span>
                                            <StatusBadge status={s.status} />
                                            <span>{s.created_at}</span>
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
                            {isAdmin && s.status !== 'approved' && s.visibility === 'public' && (
                                <button
                                    type="button"
                                    style={btnStyle}
                                    disabled={moderating === s.id}
                                    onClick={() => handleApprove(s)}
                                >
                                    {moderating === s.id ? '…' : 'Approve'}
                                </button>
                            )}
                            {isAdmin && s.status === 'pending' && (
                                <button
                                    type="button"
                                    style={btnStyle}
                                    disabled={moderating === s.id}
                                    onClick={() => handleReject(s)}
                                >
                                    {moderating === s.id ? '…' : 'Reject'}
                                </button>
                            )}
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
