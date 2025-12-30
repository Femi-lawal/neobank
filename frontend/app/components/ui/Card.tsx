'use client';

import { motion } from 'framer-motion';
import { useTheme } from '../../context/ThemeContext';

interface CardProps {
    children: React.ReactNode;
    className?: string;
    delay?: number;
    'data-testid'?: string;
}

export function Card({ children, className = '', delay = 0, ...props }: CardProps) {
    const { theme } = useTheme();
    const isDark = theme === 'dark';

    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: delay, ease: "easeOut" }}
            className={`backdrop-blur-xl rounded-2xl p-6 shadow-xl transition-colors duration-300 ${isDark
                    ? 'bg-surface-900/40 border border-surface-800'
                    : 'bg-white border border-slate-200'
                } ${className}`}
            {...props}
        >
            {children}
        </motion.div>
    );
}
