'use client';

import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { type ReactNode, type ButtonHTMLAttributes, forwardRef } from 'react';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger';
    size?: 'sm' | 'md' | 'lg';
    isLoading?: boolean;
    children?: ReactNode;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
    { className, variant = 'primary', size = 'md', isLoading, children, disabled, ...props },
    ref
) {
    const variants = {
        primary: 'bg-gradient-to-r from-[#a960ee] to-[#ff333d] hover:from-[#9050dd] hover:to-[#e6222c] text-white shadow-lg shadow-purple-500/20 border-0',
        secondary: 'bg-surface-800 hover:bg-surface-700 text-white border border-surface-700',
        outline: 'bg-transparent border border-surface-600 text-surface-200 hover:border-purple-500 hover:text-white',
        ghost: 'bg-transparent hover:bg-surface-800 text-surface-400 hover:text-white',
        danger: 'bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/50'
    };

    const sizes = {
        sm: 'px-3 py-1.5 text-xs',
        md: 'px-6 py-3 text-sm',
        lg: 'px-8 py-4 text-base font-semibold'
    };

    return (
        <button
            ref={ref}
            className={cn(
                'relative inline-flex items-center justify-center rounded-xl font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-purple-500/50 disabled:opacity-50 disabled:cursor-not-allowed overflow-hidden',
                'hover:scale-[1.02] active:scale-[0.98]',
                variants[variant],
                sizes[size],
                className
            )}
            disabled={isLoading || disabled}
            {...props}
        >
            {isLoading && (
                <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
            )}
            {children}
        </button>
    );
});
