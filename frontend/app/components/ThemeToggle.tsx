'use client';

import { useTheme } from '../context/ThemeContext';
import { Sun, Moon } from 'lucide-react';
import { motion } from 'framer-motion';

export function ThemeToggle() {
    const { theme, toggleTheme } = useTheme();

    return (
        <button
            onClick={toggleTheme}
            className="relative w-14 h-7 rounded-full p-1 transition-colors duration-300 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
            style={{
                backgroundColor: theme === 'dark' ? '#1e293b' : '#e2e8f0'
            }}
            aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
        >
            <motion.div
                className="w-5 h-5 rounded-full flex items-center justify-center"
                style={{
                    backgroundColor: theme === 'dark' ? '#a960ee' : '#ffcb57'
                }}
                animate={{
                    x: theme === 'dark' ? 0 : 28
                }}
                transition={{ type: 'spring', stiffness: 500, damping: 30 }}
            >
                {theme === 'dark' ? (
                    <Moon size={12} className="text-white" />
                ) : (
                    <Sun size={12} className="text-amber-800" />
                )}
            </motion.div>
        </button>
    );
}
