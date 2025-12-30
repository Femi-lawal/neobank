'use client';

import { createContext, useContext, useEffect, ReactNode, useSyncExternalStore, useCallback } from 'react';

type Theme = 'light' | 'dark';

interface ThemeContextType {
    theme: Theme;
    toggleTheme: () => void;
    setTheme: (theme: Theme) => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

// Subscribe to localStorage changes (for cross-tab sync)
const subscribeToStorage = (callback: () => void) => {
    window.addEventListener('storage', callback);
    return () => window.removeEventListener('storage', callback);
};

// Get theme from localStorage
const getStoredTheme = (): Theme => {
    if (typeof window === 'undefined') return 'dark';
    const saved = localStorage.getItem('neobank-theme') as Theme | null;
    if (saved) return saved;
    return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
};

// Server snapshot always returns dark (default)
const getServerSnapshot = (): Theme => 'dark';

export function ThemeProvider({ children }: { children: ReactNode }) {
    const theme = useSyncExternalStore(subscribeToStorage, getStoredTheme, getServerSnapshot);
    const mounted = useSyncExternalStore(() => () => {}, () => true, () => false);

    // Apply theme to document when it changes
    useEffect(() => {
        if (mounted) {
            document.documentElement.setAttribute('data-theme', theme);
        }
    }, [theme, mounted]);

    const setTheme = useCallback((newTheme: Theme) => {
        localStorage.setItem('neobank-theme', newTheme);
        // Trigger storage event for useSyncExternalStore
        window.dispatchEvent(new StorageEvent('storage', { key: 'neobank-theme' }));
    }, []);

    const toggleTheme = useCallback(() => {
        const newTheme = theme === 'dark' ? 'light' : 'dark';
        setTheme(newTheme);
    }, [theme, setTheme]);

    // Prevent flash of wrong theme
    if (!mounted) {
        return null;
    }

    return (
        <ThemeContext.Provider value={{ theme, toggleTheme, setTheme }}>
            {children}
        </ThemeContext.Provider>
    );
}

// Default theme values for when ThemeProvider is not available
const defaultThemeContext: ThemeContextType = {
    theme: 'dark',
    toggleTheme: () => {},
    setTheme: () => {},
};

export function useTheme() {
    const context = useContext(ThemeContext);
    // Return default values if context is not available (e.g., during SSR or outside provider)
    return context ?? defaultThemeContext;
}
