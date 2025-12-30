'use client';

import { usePathname } from 'next/navigation';
import { Sidebar } from './components/Sidebar';
import { MobileNav } from './components/MobileNav';
import { useTheme } from './context/ThemeContext';

export default function LayoutWrapper({ children }: { children: React.ReactNode }) {
    const pathname = usePathname();
    const isAuthPage = pathname === '/login' || pathname === '/register' || pathname === '/';
    
    const { theme } = useTheme();
    const isDark = theme === 'dark';

    if (isAuthPage) {
        return <main className="min-h-screen">{children}</main>;
    }

    return (
        <div className={`flex min-h-screen transition-colors duration-300 ${isDark
            ? 'bg-surface-950 text-white'
            : 'bg-slate-50 text-slate-900'
            } selection:bg-purple-500/30`}>
            {/* Sidebar - Hidden on mobile */}
            <div className="hidden md:block">
                <Sidebar />
            </div>

            {/* Main Content - Full width on mobile, offset on desktop */}
            <main className="flex-1 md:ml-64 p-4 md:p-8 pb-24 md:pb-8 overflow-y-auto h-screen relative">
                {/* Ambient Background Glow */}
                <div className={`fixed top-0 left-0 md:left-64 right-0 h-96 rounded-full blur-3xl pointer-events-none -z-10 translate-y-[-50%] ${isDark ? 'bg-purple-900/10' : 'bg-purple-200/30'
                    }`}></div>

                <div className="max-w-7xl mx-auto animate-in fade-in zoom-in duration-500">
                    {children}
                </div>
            </main>

            {/* Mobile Bottom Navigation */}
            <MobileNav />
        </div>
    );
}

