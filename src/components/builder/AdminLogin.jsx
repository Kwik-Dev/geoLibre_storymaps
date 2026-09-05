import React, { useState } from 'react';
import { adminLogin } from '../../api/client.js';
import { useAuth } from '../../auth/AuthContext.jsx';

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
const labelStyle = { display: 'block', marginBottom: '0.25rem', opacity: 0.9 };

// Local email/password sign-in for the seeded admin account (ADMIN_EMAIL /
// ADMIN_PASSWORD). On success it stores the returned access token and refreshes
// the shared auth state (whoami), which flips `user.admin` to true and reveals
// the moderation controls. A 401 surfaces as an inline error.
export default function AdminLogin({ onSuccess }) {
    const { refresh } = useAuth();
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState(null);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!email.trim() || !password) {
            setError('Email and password are required.');
            return;
        }
        setSubmitting(true);
        setError(null);
        try {
            await adminLogin(email.trim(), password);
            await refresh();
            setPassword('');
            if (typeof onSuccess === 'function') onSuccess();
        } catch (err) {
            setError((err && err.message) || 'Sign in failed.');
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} noValidate style={{ maxWidth: 360 }}>
            <div>
                <label htmlFor="admin-email" style={labelStyle}>
                    Admin email
                </label>
                <input
                    id="admin-email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="admin@example.com"
                    autoComplete="username"
                    style={inputStyle}
                />
            </div>
            <div>
                <label htmlFor="admin-password" style={labelStyle}>
                    Password
                </label>
                <input
                    id="admin-password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoComplete="current-password"
                    style={inputStyle}
                />
            </div>
            {error && <p style={{ color: '#c0392b', margin: '0.25rem 0' }}>{error}</p>}
            <button
                type="submit"
                disabled={submitting}
                style={{
                    padding: '0.5rem 1.1rem',
                    fontSize: '0.95rem',
                    cursor: 'pointer',
                    borderRadius: 4,
                    border: '1px solid currentColor',
                    background: 'transparent',
                    color: 'inherit',
                }}
            >
                {submitting ? 'Signing in…' : 'Sign in as admin'}
            </button>
        </form>
    );
}
