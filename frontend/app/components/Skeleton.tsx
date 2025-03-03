'use client';

import { motion } from 'framer-motion';

/**
 * Loading Skeleton Components
 * Animated placeholders for content loading states
 */

// Base skeleton with shimmer animation
function SkeletonBase({ className = '' }: { className?: string }) {
    return (
        <div
            className={`relative overflow-hidden bg-slate-200 dark:bg-slate-700 rounded ${className}`}
        >
            <motion.div
                className="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent"
                animate={{ x: ['-100%', '100%'] }}
                transition={{ duration: 1.5, repeat: Infinity, ease: 'linear' }}
            />
        </div>
    );
}

// Text skeleton
export function SkeletonText({ lines = 1, className = '' }: { lines?: number; className?: string }) {
    return (
        <div className={`space-y-2 ${className}`}>
            {Array.from({ length: lines }).map((_, i) => (
                <SkeletonBase
                    key={i}
                    className={`h-4 ${i === lines - 1 && lines > 1 ? 'w-3/4' : 'w-full'}`}
                />
            ))}
        </div>
    );
}

// Card skeleton
export function SkeletonCard({ className = '' }: { className?: string }) {
    return (
        <div className={`p-6 bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 ${className}`}>
            <div className="flex items-center gap-4 mb-4">
                <SkeletonBase className="w-12 h-12 rounded-full" />
                <div className="flex-1">
                    <SkeletonBase className="h-4 w-32 mb-2" />
                    <SkeletonBase className="h-3 w-24" />
                </div>
            </div>
            <SkeletonText lines={3} />
        </div>
    );
}

// Avatar skeleton
export function SkeletonAvatar({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) {
    const sizes = {
        sm: 'w-8 h-8',
        md: 'w-12 h-12',
        lg: 'w-16 h-16',
    };
    return <SkeletonBase className={`${sizes[size]} rounded-full`} />;
}

// Button skeleton
export function SkeletonButton({ className = '' }: { className?: string }) {
    return <SkeletonBase className={`h-10 w-24 rounded-lg ${className}`} />;
}

// Table row skeleton
export function SkeletonTableRow({ columns = 4 }: { columns?: number }) {
    return (
        <div className="flex items-center gap-4 py-4 border-b border-slate-100 dark:border-slate-700">
            {Array.from({ length: columns }).map((_, i) => (
                <SkeletonBase
                    key={i}
                    className={`h-4 ${i === 0 ? 'w-32' : i === columns - 1 ? 'w-20' : 'flex-1'}`}
                />
            ))}
        </div>
    );
}

// Dashboard stat skeleton
export function SkeletonStat({ className = '' }: { className?: string }) {
    return (
        <div className={`p-6 bg-white dark:bg-slate-800 rounded-xl ${className}`}>
            <SkeletonBase className="h-4 w-24 mb-3" />
            <SkeletonBase className="h-8 w-32 mb-2" />
            <SkeletonBase className="h-3 w-20" />
        </div>
    );
}

// Account card skeleton
export function SkeletonAccountCard() {
    return (
        <div className="p-6 bg-gradient-to-br from-slate-100 to-slate-200 dark:from-slate-800 dark:to-slate-700 rounded-2xl">
            <div className="flex justify-between items-start mb-6">
                <SkeletonBase className="h-5 w-32" />
                <SkeletonBase className="h-6 w-16 rounded-full" />
            </div>
            <SkeletonBase className="h-8 w-40 mb-6" />
            <div className="flex gap-4">
                <SkeletonButton />
                <SkeletonButton />
            </div>
        </div>
    );
}

// Transaction skeleton
export function SkeletonTransaction() {
    return (
        <div className="flex items-center gap-4 py-4">
            <SkeletonBase className="w-10 h-10 rounded-full" />
            <div className="flex-1">
                <SkeletonBase className="h-4 w-40 mb-2" />
                <SkeletonBase className="h-3 w-24" />
            </div>
            <SkeletonBase className="h-5 w-20" />
        </div>
    );
}

export default SkeletonBase;
