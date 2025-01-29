'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Plus, ArrowUpRight, Wallet, CreditCard, Send, Receipt, ArrowDownToLine, Eye, EyeOff, TrendingUp, PiggyBank, Building } from 'lucide-react';
import { AreaChart, Area, ResponsiveContainer } from 'recharts';
import { Account } from '../types';
import { useTheme } from '../context/ThemeContext';
import { motion } from 'framer-motion';

const chartData = [
    { value: 1000 }, { value: 2500 }, { value: 1800 },
    { value: 4000 }, { value: 3500 }, { value: 6000 },
    { value: 8000 },
];

// Quick action items
const quickActions = [
    { id: 'transfer', icon: Send, label: 'Transfer', path: '/transfers', gradient: 'from-purple-500 to-pink-500', bg: 'bg-purple-500/10' },
    { id: 'bills', icon: Receipt, label: 'Pay Bills', path: '/transfers', gradient: 'from-blue-500 to-cyan-500', bg: 'bg-blue-500/10' },
    { id: 'deposit', icon: ArrowDownToLine, label: 'Deposit', path: '/dashboard', gradient: 'from-emerald-500 to-teal-500', bg: 'bg-emerald-500/10' },
    { id: 'invest', icon: TrendingUp, label: 'Invest', path: '/products', gradient: 'from-orange-500 to-amber-500', bg: 'bg-orange-500/10' },
];

// Account type to color mapping
const accountTypeColors: Record<string, { gradient: string; icon: any; bg: string }> = {
    'ASSET': { gradient: 'from-emerald-500 to-teal-500', icon: Wallet, bg: 'bg-emerald-500/10' },
    'SAVINGS': { gradient: 'from-blue-500 to-indigo-500', icon: PiggyBank, bg: 'bg-blue-500/10' },
    'CHECKING': { gradient: 'from-purple-500 to-pink-500', icon: Building, bg: 'bg-purple-500/10' },
    'INVESTMENT': { gradient: 'from-orange-500 to-amber-500', icon: TrendingUp, bg: 'bg-orange-500/10' },
};

function getGreeting(): string {
    const hour = new Date().getHours();
    if (hour < 12) return 'Good morning';
    if (hour < 18) return 'Good afternoon';
    return 'Good evening';
}

export default function Dashboard() {
    const [accounts, setAccounts] = useState<Account[]>([]);
    const [showBalance, setShowBalance] = useState(true);
    const router = useRouter();
    const { theme } = useTheme();
    const isDark = theme === 'dark';

    useEffect(() => {
        fetch('/api/ledger/accounts')
            .then((res) => {
                if (res.status === 401) router.push('/login');
                return res.json();
            })
            .then((data) => {
                if (Array.isArray(data)) setAccounts(data);
            })
            .catch((err) => console.error(err));
    }, []);

    const createMockAccount = async () => {
        await fetch('/api/ledger/accounts', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                account_number: "ACC-" + Math.floor(Math.random() * 10000),
                name: "New Checking",
                currency: "USD",
                type: "ASSET"
            })
        });
        const res = await fetch('/api/ledger/accounts');
        const data = await res.json();
        setAccounts(data);
    };

    const totalBalance = accounts.reduce((sum, acc) => sum + parseFloat(acc.CachedBalance), 0);

    const formatBalance = (amount: number) => {
        if (!showBalance) return '••••••';
        return `$${amount.toLocaleString(undefined, { minimumFractionDigits: 2 })}`;
    };

    return (
        <div className="space-y-8">
            {/* Header with Personalized Greeting */}
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4">
                <div>
                    <h1 className={`text-2xl sm:text-3xl font-bold mb-2 ${isDark ? 'text-white' : 'text-slate-900'}`}>
                        {getGreeting()}, Alex 👋
                    </h1>
                    <p className={isDark ? 'text-surface-400' : 'text-slate-500'}>Here's what's happening with your finances.</p>
                </div>
                <div className="flex gap-3">
                    <Button variant="secondary" size="md" onClick={() => router.push('/transfers')}>
                        Send Money
                    </Button>
                    <Button onClick={createMockAccount} size="md" className="hidden sm:flex">
                        <Plus size={18} className="mr-2" /> New Account
                    </Button>
                </div>
            </div>

            {/* Hero Cards */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <Card className="lg:col-span-2 relative overflow-hidden group min-h-[200px] lg:min-h-[300px] flex flex-col justify-between border-gradient-purple/20">
                    {/* Chart Background */}
                    <div className="absolute inset-0 pt-20 px-0 opacity-40 group-hover:opacity-50 transition-opacity duration-500">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={chartData}>
                                <defs>
                                    <linearGradient id="colorBalance" x1="0" y1="0" x2="1" y2="0">
                                        <stop offset="0%" stopColor="#a960ee" stopOpacity={isDark ? 0.3 : 0.2} />
                                        <stop offset="50%" stopColor="#ff333d" stopOpacity={isDark ? 0.3 : 0.2} />
                                        <stop offset="100%" stopColor="#90e0ff" stopOpacity={isDark ? 0.3 : 0.2} />
                                    </linearGradient>
                                    <linearGradient id="strokeGradient" x1="0" y1="0" x2="1" y2="0">
                                        <stop offset="0%" stopColor="#a960ee" />
                                        <stop offset="100%" stopColor="#ff333d" />
                                    </linearGradient>
                                </defs>
                                <Area type="monotone" dataKey="value" stroke="url(#strokeGradient)" strokeWidth={3} fillOpacity={1} fill="url(#colorBalance)" />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>

                    <div className="relative z-10 p-2">
                        <div className="flex items-center gap-2 mb-1">
                            <p className={`font-medium ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Total Balance</p>
                            <button
                                onClick={() => setShowBalance(!showBalance)}
                                className={`p-1 rounded-lg hover:bg-surface-800/50 transition-colors ${isDark ? 'text-surface-400' : 'text-slate-500'}`}
                            >
                                {showBalance ? <Eye size={16} /> : <EyeOff size={16} />}
                            </button>
                        </div>
                        <div className={`text-4xl sm:text-5xl font-bold tracking-tight ${isDark ? 'text-white' : 'text-slate-900'}`}>
                            {formatBalance(totalBalance)}
                        </div>
                    </div>

                    <div className="relative z-10 p-2 mt-auto flex gap-4">
                        <div className="flex items-center gap-2 text-emerald-500 bg-emerald-500/10 px-3 py-1 rounded-lg text-sm border border-emerald-500/20">
                            <ArrowUpRight size={16} /> +$1,240.50 (12%)
                        </div>
                    </div>
                </Card>

                <Card className="flex flex-col justify-center items-center text-center space-y-4">
                    <div className="h-16 w-16 bg-gradient-to-br from-gradient-purple to-gradient-red rounded-full flex items-center justify-center text-white mb-2 group-hover:scale-110 transition-transform duration-300">
                        <CreditCard size={32} />
                    </div>
                    <div>
                        <h3 className={`font-semibold text-lg ${isDark ? 'text-white' : 'text-slate-900'}`}>My Cards</h3>
                        <p className={`text-sm ${isDark ? 'text-surface-500' : 'text-slate-500'}`}>Manage your 3 active cards</p>
                    </div>
                    <Button variant="outline" size="sm" onClick={() => router.push('/cards')}>View All</Button>
                </Card>
            </div>

            {/* Quick Actions */}
            <div>
                <h2 className={`text-xl font-bold mb-4 ${isDark ? 'text-white' : 'text-slate-900'}`}>Quick Actions</h2>
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                    {quickActions.map((action, i) => {
                        const Icon = action.icon;
                        return (
                            <motion.button
                                key={action.id}
                                initial={{ opacity: 0, y: 20 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: i * 0.1 }}
                                onClick={() => router.push(action.path)}
                                className={`p-4 sm:p-6 rounded-2xl border transition-all hover:scale-105 hover:shadow-lg group ${isDark
                                    ? 'bg-surface-900/50 border-surface-800 hover:border-surface-700'
                                    : 'bg-white border-slate-200 hover:border-slate-300'
                                    }`}
                            >
                                <div className={`w-12 h-12 rounded-xl ${action.bg} flex items-center justify-center mb-3 group-hover:scale-110 transition-transform`}>
                                    <div className={`bg-gradient-to-br ${action.gradient} bg-clip-text`}>
                                        <Icon size={24} className="text-current" style={{ color: action.gradient.includes('purple') ? '#a855f7' : action.gradient.includes('blue') ? '#3b82f6' : action.gradient.includes('emerald') ? '#10b981' : '#f97316' }} />
                                    </div>
                                </div>
                                <span className={`font-medium ${isDark ? 'text-white' : 'text-slate-900'}`}>{action.label}</span>
                            </motion.button>
                        );
                    })}
                </div>
            </div>

            {/* Account List */}
            <div>
                <div className="flex justify-between items-center mb-4">
                    <h2 className={`text-xl font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>Your Accounts</h2>
                    <Button variant="ghost" size="sm" onClick={createMockAccount} className="sm:hidden">
                        <Plus size={16} className="mr-1" /> Add
                    </Button>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {accounts.map((acc, i) => {
                        const accountType = (acc.Type || 'ASSET').toUpperCase();
                        const typeConfig = accountTypeColors[accountType] || accountTypeColors['ASSET'];
                        const Icon = typeConfig.icon;

                        return (
                            <Card key={acc.ID} delay={i * 0.1} className="hover:border-gradient-purple/50 transition-colors cursor-pointer group" data-testid="account-card">
                                <div className="flex justify-between items-start mb-6">
                                    <div className={`h-10 w-10 rounded-full flex items-center justify-center transition-all ${typeConfig.bg} group-hover:scale-110`}>
                                        <Icon size={20} className={`bg-gradient-to-br ${typeConfig.gradient}`} style={{ color: typeConfig.gradient.includes('emerald') ? '#10b981' : typeConfig.gradient.includes('blue') ? '#3b82f6' : '#a855f7' }} />
                                    </div>
                                    <span className={`text-xs font-mono px-2 py-1 rounded-full ${isDark ? 'bg-surface-800 text-surface-400' : 'bg-slate-100 text-slate-500'}`}>
                                        {acc.CurrencyCode}
                                    </span>
                                </div>
                                <div>
                                    <h3 className={`font-medium mb-1 ${isDark ? 'text-surface-300' : 'text-slate-600'}`}>{acc.Name}</h3>
                                    <div className={`text-2xl font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>
                                        {formatBalance(parseFloat(acc.CachedBalance))}
                                    </div>
                                    <p className={`text-xs mt-2 font-mono truncate ${isDark ? 'text-surface-600' : 'text-slate-400'}`}>{acc.ID}</p>
                                </div>
                            </Card>
                        );
                    })}
                </div>
            </div>
        </div>
    );
}

