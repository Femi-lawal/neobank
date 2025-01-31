'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Home, CreditCard, ArrowRightLeft, ShoppingBag, LogOut } from 'lucide-react';
import { motion } from 'framer-motion';
import { ThemeToggle } from './ThemeToggle';
import { useTheme } from '../context/ThemeContext';

const menuItems = [
    { name: 'Dashboard', href: '/dashboard', icon: Home },
    { name: 'Transfers', href: '/transfers', icon: ArrowRightLeft },
    { name: 'Cards', href: '/cards', icon: CreditCard },
    { name: 'Products', href: '/products', icon: ShoppingBag },
];

export function Sidebar() {
    const pathname = usePathname();
    const { theme } = useTheme();
    const isDark = theme === 'dark';

    return (
        <aside className={`fixed left-0 top-0 h-full w-64 p-6 flex flex-col z-50 border-r transition-colors duration-300 ${isDark ? 'bg-surface-950 border-surface-800' : 'bg-white border-slate-200'
            }`}>
            <div className="mb-10 flex items-center gap-3 px-2">
                {/* Logo with Stripe gradient */}
                <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-gradient-purple to-gradient-red flex items-center justify-center text-white font-bold text-xl shadow-lg">
                    N
                </div>
                <span className={`text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r ${isDark ? 'from-white to-surface-400' : 'from-slate-900 to-slate-600'
                    }`}>
                    NeoBank
                </span>
            </div>

            <nav className="flex-1 space-y-2">
                {menuItems.map((item) => {
                    const isActive = pathname === item.href;
                    return (
                        <Link key={item.href} href={item.href}>
                            <div className={`relative px-4 py-3 rounded-xl flex items-center gap-3 transition-all duration-200 group ${isActive
                                    ? (isDark ? 'text-white' : 'text-slate-900')
                                    : (isDark ? 'text-surface-400 hover:text-white hover:bg-surface-900' : 'text-slate-500 hover:text-slate-900 hover:bg-slate-100')
                                }`}>
                                {isActive && (
                                    <motion.div
                                        layoutId="activeTab"
                                        className={`absolute inset-0 rounded-xl ${isDark
                                                ? 'bg-gradient-purple/10 border border-gradient-purple/20'
                                                : 'bg-purple-50 border border-purple-200'
                                            }`}
                                        transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                                    />
                                )}
                                <item.icon size={20} className={
                                    isActive
                                        ? 'text-gradient-purple'
                                        : (isDark ? 'text-surface-500 group-hover:text-gradient-cyan' : 'text-slate-400 group-hover:text-purple-500') + ' transition-colors'
                                } />
                                <span className="relative z-10 font-medium">{item.name}</span>
                            </div>
                        </Link>
                    )
                })}
            </nav>

            {/* Theme Toggle */}
            <div className={`py-4 border-t ${isDark ? 'border-surface-800' : 'border-slate-200'}`}>
                <div className="flex items-center justify-between px-2">
                    <span className={`text-sm font-medium ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>
                        {isDark ? 'Dark Mode' : 'Light Mode'}
                    </span>
                    <ThemeToggle />
                </div>
            </div>

            {/* Sign Out */}
            <div className={`pt-4 border-t ${isDark ? 'border-surface-800' : 'border-slate-200'}`}>
                <button
                    onClick={() => {
                        localStorage.removeItem('token');
                        window.location.href = '/login';
                    }}
                    className={`w-full px-4 py-3 rounded-xl flex items-center gap-3 transition-all ${isDark
                            ? 'text-surface-400 hover:text-red-400 hover:bg-red-500/5'
                            : 'text-slate-500 hover:text-red-500 hover:bg-red-50'
                        }`}
                >
                    <LogOut size={20} />
                    <span className="font-medium">Sign Out</span>
                </button>
            </div>
        </aside>
    );
}
