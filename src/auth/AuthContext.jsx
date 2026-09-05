import React, { createContext, useCallback, useContext, useEffect, useState } from 'react';
import { apiFetch } from '../api/client.js';

// AuthContext — tracks the current user via GET /api/auth/whoami (the httpOnly
// refresh cookie authenticates the request). Provides `logout` (POST
// /api/auth/logout) and `refresh`. Login is a plain navigation to the GitHub
// OAuth start path, so it's just a link — no state needed here.
const AuthContext = createContext(null);

export function AuthProvider({ children }) {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);

    const refresh = useCallback(async () => {
        try {
            const resp = await apiFetch('/auth/whoami');
            const data = await resp.json();
            setUser(data);
        } catch (_) {
            // 401 (or server down) → not signed in.
            setUser(null);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        refresh();
    }, [refresh]);

    const logout = useCallback(async () => {
        try {
            await apiFetch('/auth/logout', { method: 'POST' });
        } catch (_) {
            // Even if the request fails, drop the local user state.
        }
        setUser(null);
        // Reload so auth-dependent UI (e.g. the story list, which shows the
        // owner's private stories only when signed in) re-initializes cleanly.
        if (typeof window !== 'undefined') window.location.reload();
    }, []);

    return (
        <AuthContext.Provider value={{ user, loading, refresh, logout }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    return useContext(AuthContext);
}
