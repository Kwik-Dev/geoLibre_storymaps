import React, { useState } from 'react';
import { apiFetch } from '../../api/client.js';
import { navigateToEdit } from '../../hashRoute.js';
import { basePath } from '../../basePath.js';

// The themes the renderer supports (CSS classes .dark / .light in styles.css).
export const THEMES = ['dark', 'light'];

// The start of the GitHub OAuth flow (§7.2). Same-origin `/api` is proxied to
// the Go backend in dev (vite.config.js) and served by the same binary in prod.
// The base path (if any) is prepended so the flow stays under the app's subpath.
const OAUTH_START_PATH = basePath + '/api/auth/github';

// P6.1 — create a story.
//
// Fields: title (required), subtitle, byline, theme, visibility.
// On submit → POST /api/stories (via the shared client). On success, stores the
// new story's id and navigates to #/stories/<id> (the export renderer path,
// which accepts both numeric id and slug). A 401 surfaces a "Sign in with
// GitHub" link (the OAuth flow) instead of a generic error.
export default function StoryForm({ onCreated }) {
    const [title, setTitle] = useState('');
    const [subtitle, setSubtitle] = useState('');
    const [byline, setByline] = useState('');
    const [theme, setTheme] = useState(THEMES[0]);
    const [visibility, setVisibility] = useState('private');
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState(null);
    const [authRequired, setAuthRequired] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        // Client-side validation: title is required.
        if (!title.trim()) {
            setError('Title is required.');
            return;
        }
        setError(null);
        setAuthRequired(false);
        setSubmitting(true);
        try {
            const resp = await apiFetch('/stories', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    title: title.trim(),
                    subtitle: subtitle.trim(),
                    byline: byline.trim(),
                    visibility,
                }),
            });
            const story = await resp.json();
            const id = story && story.id;
            if (!id) {
                throw new Error('The server did not return a story id.');
            }
            if (typeof onCreated === 'function') onCreated(story);
            // After creating, go straight to the chapter editor so the user can
            // add chapters + media (the P6.4 happy path: create → add chapter).
            navigateToEdit(String(id));
        } catch (err) {
            if (err && err.status === 401) {
                setAuthRequired(true);
                setError('Please sign in to create a story.');
            } else {
                setError((err && err.message) || 'Failed to create the story.');
            }
        } finally {
            setSubmitting(false);
        }
    };

    const inputStyle = {
        boxSizing: 'border-box',
        width: '100%',
        padding: '0.5rem 0.6rem',
        marginBottom: '0.75rem',
        fontSize: '1rem',
        borderRadius: 4,
        border: '1px solid currentColor',
        background: 'transparent',
        color: 'inherit',
    };
    const labelStyle = { display: 'block', marginBottom: '0.25rem', opacity: 0.9 };
    const rowStyle = { marginBottom: '0.25rem' };

    return (
        <div
            style={{
                maxWidth: 560,
                margin: '0 auto',
                padding: '2rem 1.5rem',
                color: 'inherit',
            }}
        >
            <h1>Create a story</h1>
            <p style={{ opacity: 0.85 }}>
                Give your story a title and set its visibility. You'll add chapters
                and media next.
            </p>

            {authRequired && (
                <p style={{ fontStyle: 'italic', opacity: 0.9 }}>
                    You need to be signed in to create a story.{' '}
                    <a
                        href={OAUTH_START_PATH}
                        style={{ textDecoration: 'underline' }}
                    >
                        Sign in with GitHub
                    </a>
                </p>
            )}

            <form onSubmit={handleSubmit} noValidate>
                <div style={rowStyle}>
                    <label htmlFor="sf-title" style={labelStyle}>
                        Title *
                    </label>
                    <input
                        id="sf-title"
                        type="text"
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        placeholder="e.g. My Ocean Journey"
                        style={inputStyle}
                        maxLength={120}
                        required
                    />
                </div>

                <div style={rowStyle}>
                    <label htmlFor="sf-subtitle" style={labelStyle}>
                        Subtitle
                    </label>
                    <input
                        id="sf-subtitle"
                        type="text"
                        value={subtitle}
                        onChange={(e) => setSubtitle(e.target.value)}
                        placeholder="Optional subtitle"
                        style={inputStyle}
                        maxLength={200}
                    />
                </div>

                <div style={rowStyle}>
                    <label htmlFor="sf-byline" style={labelStyle}>
                        Byline
                    </label>
                    <input
                        id="sf-byline"
                        type="text"
                        value={byline}
                        onChange={(e) => setByline(e.target.value)}
                        placeholder="Optional byline / author"
                        style={inputStyle}
                        maxLength={120}
                    />
                </div>

                <div style={rowStyle}>
                    <label htmlFor="sf-theme" style={labelStyle}>
                        Theme
                    </label>
                    <select
                        id="sf-theme"
                        value={theme}
                        onChange={(e) => setTheme(e.target.value)}
                        style={inputStyle}
                    >
                        {THEMES.map((t) => (
                            <option key={t} value={t}>
                                {t === 'dark' ? 'Dark' : 'Light'}
                            </option>
                        ))}
                    </select>
                </div>

                <div style={rowStyle}>
                    <span style={labelStyle}>Visibility</span>
                    <label style={{ display: 'block', marginBottom: '0.25rem' }}>
                        <input
                            type="radio"
                            name="sf-visibility"
                            value="private"
                            checked={visibility === 'private'}
                            onChange={() => setVisibility('private')}
                        />{' '}
                        Private — only you (and admins) can view it
                    </label>
                    <label style={{ display: 'block' }}>
                        <input
                            type="radio"
                            name="sf-visibility"
                            value="public"
                            checked={visibility === 'public'}
                            onChange={() => setVisibility('public')}
                        />{' '}
                        Public — anyone can view it
                    </label>
                </div>

                {error && <p style={{ color: '#c0392b', margin: '0.5rem 0' }}>{error}</p>}

                <button
                    type="submit"
                    disabled={submitting}
                    style={{
                        padding: '0.6rem 1.2rem',
                        fontSize: '1rem',
                        cursor: 'pointer',
                        borderRadius: 4,
                        marginTop: '0.5rem',
                        border: '1px solid currentColor',
                        background: 'transparent',
                        color: 'inherit',
                    }}
                >
                    {submitting ? 'Creating…' : 'Create story'}
                </button>
            </form>
        </div>
    );
}
