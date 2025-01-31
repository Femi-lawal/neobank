'use client';

import { usePathname, useRouter } from 'next/navigation';
import { Home, CreditCard, ArrowLeftRight, Package, Menu } from 'lucide-react';
import { motion } from 'framer-motion';
import { useState } from 'react';

const navItems = [
    { icon: Home, label: 'Home', path: '/dashboard' },
    { icon: CreditCard, label: 'Cards', path: '/cards' },
    { icon: ArrowLeftRight, label: 'Transfer', path: '/transfers' },
    { icon: Package, label: 'Products', path: '/products' },
];

export function MobileNav() {
    const pathname = usePathname();
    const router = useRouter();
    const [showMenu, setShowMenu] = useState(false);

    return (
        <>
            {/* Bottom Navigation Bar */}
            <nav className="md:hidden fixed bottom-0 left-0 right-0 z-50 bg-surface-900/95 backdrop-blur-xl border-t border-surface-800 safe-area-bottom">
                <div className="flex justify-around items-center h-16 px-2">
                    {navItems.map((item) => {
                        const isActive = pathname === item.path;
                        const Icon = item.icon;

                        return (
                            <button
                                key={item.path}
                                onClick={() => router.push(item.path)}
                                className="flex flex-col items-center justify-center flex-1 h-full relative group"
                            >
                                {isActive && (
                                    <motion.div
                                        layoutId="mobile-nav-active"
                                        className="absolute -top-0.5 w-12 h-1 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full"
                                        transition={{ type: 'spring', stiffness: 500, damping: 30 }}
                                    />
                                )}
                                <div className={`p-2 rounded-xl transition-all ${isActive
                                        ? 'text-white bg-surface-800'
                                        : 'text-surface-500 group-hover:text-surface-300'
                                    }`}>
                                    <Icon size={22} />
                                </div>
                                <span className={`text-[10px] mt-0.5 font-medium ${isActive ? 'text-white' : 'text-surface-500'
                                    }`}>
                                    {item.label}
                                </span>
                            </button>
                        );
                    })}

                    {/* More Menu Button */}
                    <button
                        onClick={() => setShowMenu(!showMenu)}
                        className="flex flex-col items-center justify-center flex-1 h-full"
                    >
                        <div className={`p-2 rounded-xl transition-all ${showMenu
                                ? 'text-white bg-surface-800'
                                : 'text-surface-500 hover:text-surface-300'
                            }`}>
                            <Menu size={22} />
                        </div>
                        <span className={`text-[10px] mt-0.5 font-medium ${showMenu ? 'text-white' : 'text-surface-500'
                            }`}>
                            More
                        </span>
                    </button>
                </div>
            </nav>

            {/* More Menu Overlay */}
            {showMenu && (
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: 20 }}
                    className="md:hidden fixed bottom-20 left-4 right-4 z-50 bg-surface-800 rounded-2xl p-4 shadow-2xl border border-surface-700"
                >
                    <div className="grid grid-cols-3 gap-4">
                        <button
                            onClick={() => { router.push('/settings'); setShowMenu(false); }}
                            className="flex flex-col items-center p-4 rounded-xl hover:bg-surface-700 transition-colors"
                        >
                            <div className="w-10 h-10 bg-purple-500/20 rounded-full flex items-center justify-center text-purple-400 mb-2">
                                ⚙️
                            </div>
                            <span className="text-xs text-surface-300">Settings</span>
                        </button>
                        <button
                            onClick={() => { router.push('/help'); setShowMenu(false); }}
                            className="flex flex-col items-center p-4 rounded-xl hover:bg-surface-700 transition-colors"
                        >
                            <div className="w-10 h-10 bg-blue-500/20 rounded-full flex items-center justify-center text-blue-400 mb-2">
                                ❓
                            </div>
                            <span className="text-xs text-surface-300">Help</span>
                        </button>
                        <button
                            onClick={() => { router.push('/login'); setShowMenu(false); }}
                            className="flex flex-col items-center p-4 rounded-xl hover:bg-surface-700 transition-colors"
                        >
                            <div className="w-10 h-10 bg-red-500/20 rounded-full flex items-center justify-center text-red-400 mb-2">
                                🚪
                            </div>
                            <span className="text-xs text-surface-300">Logout</span>
                        </button>
                    </div>
                </motion.div>
            )}

            {/* Backdrop */}
            {showMenu && (
                <div
                    className="md:hidden fixed inset-0 bg-black/50 z-40"
                    onClick={() => setShowMenu(false)}
                />
            )}
        </>
    );
}
