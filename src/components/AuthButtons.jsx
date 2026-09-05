import React from 'react';
import { useAuth } from '../auth/AuthContext.jsx';
import { basePath } from '../basePath.js';

// The start of the GitHub OAuth flow. Same-origin `/api` is proxied to the Go
// backend in dev (vite.config.js) and served by the same binary in prod. The
// base path (if any) is prepended so the flow stays under the app's subpath.
const OAUTH_START_PATH = basePath + '/api/auth/github';

// Sign in / Sign out controls for the header. Shows the current user's name
// (github_login or admin_email) plus a Sign out button when authenticated, or a
// "Sign in with GitHub" link when not. Renders nothing while the whoami check
// is still loading.
export default function AuthButtons() {
    const { user, loading, logout } = useAuth();

    const btnStyle = {
        padding: '0.35rem 0.8rem',
        fontSize: '0.9rem',
        cursor: 'pointer',
        borderRadius: 4,
        border: '1px solid currentColor',
        background: 'transparent',
        color: 'inherit',
        textDecoration: 'none',
        fontFamily: 'inherit',
    };

    if (loading) return null;

    if (user) {
        const name = user.github_login || user.admin_email || `user #${user.id}`;
        return (
            <div style={{ display: 'inline-flex', gap: '0.5rem', alignItems: 'center' }}>
                <a
                    href="#/"
                    title="Go to home"
                    style={{ opacity: 0.85, color: 'inherit', textDecoration: 'none', cursor: 'pointer' }}
                    onMouseEnter={(e) => (e.currentTarget.style.textDecoration = 'underline')}
                    onMouseLeave={(e) => (e.currentTarget.style.textDecoration = 'none')}
                >
                    {name}
                </a>
                <button type="button" style={btnStyle} onClick={logout}>
                    Sign out
                </button>
            </div>
        );
    }

    return (
        <a href={OAUTH_START_PATH} style={btnStyle}>
            Sign in with GitHub
        </a>
    );
}
