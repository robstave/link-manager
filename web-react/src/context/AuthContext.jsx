import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
    const [user, setUser] = useState(null);
    const [isLoading, setIsLoading] = useState(true);

    // On mount, check if we have a stored token and validate it
    useEffect(() => {
        async function checkAuth() {
            if (!api.isAuthenticated()) {
                setIsLoading(false);
                return;
            }
            try {
                const me = await api.getMe();
                if (me) {
                    setUser(me);
                }
            } catch {
                api.token = null;
                localStorage.removeItem('token');
            }
            setIsLoading(false);
        }
        checkAuth();
    }, []);

    const login = useCallback(async (username, password) => {
        const data = await api.login(username, password);
        if (data) {
            const me = await api.getMe();
            setUser(me);
        }
        return data;
    }, []);

    const logout = useCallback(() => {
        api.token = null;
        localStorage.removeItem('token');
        setUser(null);
    }, []);

    const value = {
        user,
        isLoading,
        isAuthenticated: !!user,
        login,
        logout,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error('useAuth must be used within AuthProvider');
    return ctx;
}
