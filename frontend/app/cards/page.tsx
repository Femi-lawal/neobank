'use client';

import { useState, useEffect, useCallback } from 'react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { motion } from 'framer-motion';
import { Plus, Snowflake, Lock, MoreVertical, Eye, EyeOff, Settings, AlertTriangle } from 'lucide-react';
import { useTheme } from '../context/ThemeContext';

interface CardData {
    id: string;
    card_number: string;
    expiration_date: string;
    status: string;
    isFrozen?: boolean;
}

export default function CardsPage() {
    const [cards, setCards] = useState<CardData[]>([]);
    const [accountID, setAccountID] = useState('');
    const [accounts, setAccounts] = useState([]);
    const [frozenCards, setFrozenCards] = useState<Set<string>>(new Set());
    const [hiddenNumbers, setHiddenNumbers] = useState<Set<string>>(new Set());
    const [showCardMenu, setShowCardMenu] = useState<string | null>(null);
    const { theme } = useTheme();
    const isDark = theme === 'dark';

    const fetchCards = useCallback((accId: string) => {
        const token = localStorage.getItem('token');
        fetch(`/api/card/cards?account_id=${accId}`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
            .then(res => res.json())
            .then(data => setCards(data || []));
    }, []);

    useEffect(() => {
        const token = localStorage.getItem('token');
        fetch('/api/ledger/accounts', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
            .then(res => res.json())
            .then(data => {
                setAccounts(data || []);
                if (data && data.length > 0) {
                    setAccountID(data[0].id);
                    fetchCards(data[0].id);
                }
            });
    }, [fetchCards]);

    const handleIssueCard = async () => {
        if (!accountID) return;
        const token = localStorage.getItem('token');
        await fetch('/api/card/cards', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ account_id: accountID })
        });
        fetchCards(accountID);
    };

    const toggleFreeze = (cardId: string) => {
        setFrozenCards(prev => {
            const newSet = new Set(prev);
            if (newSet.has(cardId)) {
                newSet.delete(cardId);
            } else {
                newSet.add(cardId);
            }
            return newSet;
        });
    };

    const toggleHideNumber = (cardId: string) => {
        setHiddenNumbers(prev => {
            const newSet = new Set(prev);
            if (newSet.has(cardId)) {
                newSet.delete(cardId);
            } else {
                newSet.add(cardId);
            }
            return newSet;
        });
    };

    const formatCardNumber = (number: string, cardId: string) => {
        if (hiddenNumbers.has(cardId)) {
            return '•••• •••• •••• ' + number.slice(-4);
        }
        return number.match(/.{1,4}/g)?.join(' ') || number;
    };

    const cardGradients = [
        'from-purple-600 via-pink-500 to-red-500',
        'from-blue-600 via-cyan-500 to-teal-400',
        'from-emerald-600 via-green-500 to-lime-400',
        'from-orange-600 via-amber-500 to-yellow-400',
        'from-indigo-600 via-violet-500 to-purple-400',
    ];

    return (
        <div className="space-y-8">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <div>
                    <h1 className={`text-2xl sm:text-3xl font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>Your Cards</h1>
                    <p className={`mt-1 ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Manage physical and virtual cards.</p>
                </div>

                <div className="flex items-center gap-4">
                    <select
                        value={accountID}
                        onChange={(e) => {
                            setAccountID(e.target.value);
                            fetchCards(e.target.value);
                        }}
                        className={`rounded-xl px-4 py-2.5 focus:outline-none focus:ring-1 focus:ring-brand-500 ${isDark
                            ? 'bg-surface-800 border-surface-700 text-white'
                            : 'bg-white border border-slate-200 text-slate-900'
                            }`}
                    >
                        {accounts.map((a: { id: string; name: string }) => <option key={a.id} value={a.id}>{a.name}</option>)}
                    </select>
                    <Button onClick={handleIssueCard}>
                        <Plus size={18} className="mr-2" /> Issue Virtual Card
                    </Button>
                </div>
            </div>

            {/* Card Stats */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                <Card className="text-center">
                    <div className={`text-3xl font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>{cards.length}</div>
                    <div className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Total Cards</div>
                </Card>
                <Card className="text-center">
                    <div className="text-3xl font-bold text-emerald-500">{cards.length - frozenCards.size}</div>
                    <div className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Active</div>
                </Card>
                <Card className="text-center">
                    <div className="text-3xl font-bold text-blue-500">{frozenCards.size}</div>
                    <div className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Frozen</div>
                </Card>
                <Card className="text-center">
                    <div className="text-3xl font-bold text-purple-500">$10,000</div>
                    <div className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Monthly Limit</div>
                </Card>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-8">
                {cards.map((c, i) => {
                    const isFrozen = frozenCards.has(c.id);
                    const gradientClass = cardGradients[i % cardGradients.length];

                    return (
                        <motion.div
                            key={c.id}
                            initial={{ scale: 0.9, opacity: 0 }}
                            animate={{ scale: 1, opacity: 1 }}
                            transition={{ delay: i * 0.1 }}
                            className="space-y-4"
                        >
                            {/* Credit Card */}
                            <div className={`relative w-full aspect-[1.586] rounded-2xl p-6 text-white shadow-2xl overflow-hidden group select-none ${isFrozen ? 'grayscale opacity-80' : ''
                                }`}>
                                {/* Card Background Gradient */}
                                <div className={`absolute inset-0 bg-gradient-to-br ${gradientClass} z-0`}></div>

                                {/* Glassmorphism Effect */}
                                <div className="absolute inset-0 bg-white/10 backdrop-blur-[1px] z-0"></div>

                                {/* Decorative Circles */}
                                <div className="absolute -top-20 -right-20 w-64 h-64 bg-white/10 rounded-full z-0"></div>
                                <div className="absolute -bottom-10 -left-10 w-48 h-48 bg-white/5 rounded-full z-0"></div>

                                {/* Frozen Overlay */}
                                {isFrozen && (
                                    <div className="absolute inset-0 bg-blue-900/40 z-10 flex items-center justify-center">
                                        <div className="text-center">
                                            <Snowflake size={40} className="mx-auto mb-2 text-blue-200" />
                                            <span className="text-blue-100 font-medium">Card Frozen</span>
                                        </div>
                                    </div>
                                )}

                                <div className="relative z-20 h-full flex flex-col justify-between">
                                    <div className="flex justify-between items-start">
                                        <div className="font-bold text-lg tracking-tight text-white/90">NeoBank</div>
                                        <div className="flex items-center gap-2">
                                            {isFrozen && (
                                                <span className="px-2 py-0.5 bg-blue-500/30 rounded-full text-[10px] font-medium text-blue-100">
                                                    FROZEN
                                                </span>
                                            )}
                                            <span className="font-bold italic text-white/70">VISA</span>
                                        </div>
                                    </div>

                                    <div className="space-y-6">
                                        {/* Chip */}
                                        <div className="w-12 h-9 rounded-md bg-gradient-to-br from-yellow-200 to-yellow-500 shadow-lg flex items-center justify-center">
                                            <div className="w-8 h-5 border-2 border-yellow-700/30 rounded grid grid-cols-3 gap-px p-0.5">
                                                {[...Array(6)].map((_, i) => (
                                                    <div key={i} className="bg-yellow-600/30 rounded-sm"></div>
                                                ))}
                                            </div>
                                        </div>

                                        <div className="font-mono text-xl tracking-[0.15em] text-shadow-sm">
                                            {formatCardNumber(c.card_number, c.id)}
                                        </div>
                                    </div>

                                    <div className="flex justify-between items-end text-xs uppercase tracking-wider">
                                        <div>
                                            <div className="text-[10px] text-white/60 mb-1">Card Holder</div>
                                            <div className="font-semibold text-white">James Doe</div>
                                        </div>
                                        <div>
                                            <div className="text-[10px] text-white/60 mb-1">Expires</div>
                                            <div className="font-semibold text-white">{c.expiration_date}</div>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            {/* Card Actions */}
                            <div className="flex gap-2">
                                <Button
                                    variant={isFrozen ? "primary" : "secondary"}
                                    size="sm"
                                    className="flex-1"
                                    onClick={() => toggleFreeze(c.id)}
                                >
                                    {isFrozen ? (
                                        <>
                                            <Lock size={14} className="mr-2" /> Unfreeze
                                        </>
                                    ) : (
                                        <>
                                            <Snowflake size={14} className="mr-2" /> Freeze
                                        </>
                                    )}
                                </Button>
                                <Button
                                    variant="secondary"
                                    size="sm"
                                    className="flex-1"
                                    onClick={() => toggleHideNumber(c.id)}
                                >
                                    {hiddenNumbers.has(c.id) ? (
                                        <>
                                            <Eye size={14} className="mr-2" /> Show
                                        </>
                                    ) : (
                                        <>
                                            <EyeOff size={14} className="mr-2" /> Hide
                                        </>
                                    )}
                                </Button>
                                <div className="relative">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => setShowCardMenu(showCardMenu === c.id ? null : c.id)}
                                    >
                                        <MoreVertical size={18} />
                                    </Button>

                                    {showCardMenu === c.id && (
                                        <motion.div
                                            initial={{ opacity: 0, y: -10 }}
                                            animate={{ opacity: 1, y: 0 }}
                                            className={`absolute right-0 top-full mt-2 w-48 rounded-xl shadow-xl border overflow-hidden z-50 ${isDark
                                                ? 'bg-surface-800 border-surface-700'
                                                : 'bg-white border-slate-200'
                                                }`}
                                        >
                                            <button className={`w-full px-4 py-3 text-left text-sm flex items-center gap-3 transition-colors ${isDark
                                                ? 'text-surface-300 hover:bg-surface-700'
                                                : 'text-slate-600 hover:bg-slate-50'
                                                }`}>
                                                <Settings size={16} /> Card Settings
                                            </button>
                                            <button className={`w-full px-4 py-3 text-left text-sm flex items-center gap-3 transition-colors ${isDark
                                                ? 'text-surface-300 hover:bg-surface-700'
                                                : 'text-slate-600 hover:bg-slate-50'
                                                }`}>
                                                <Lock size={16} /> Set Spending Limit
                                            </button>
                                            <button className={`w-full px-4 py-3 text-left text-sm flex items-center gap-3 transition-colors ${isDark
                                                ? 'text-red-400 hover:bg-surface-700'
                                                : 'text-red-500 hover:bg-red-50'
                                                }`}>
                                                <AlertTriangle size={16} /> Report Lost
                                            </button>
                                        </motion.div>
                                    )}
                                </div>
                            </div>
                        </motion.div>
                    );
                })}
            </div>

            {cards.length === 0 && (
                <Card className="text-center py-16">
                    <div className={`w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-4 ${isDark ? 'bg-surface-800' : 'bg-slate-100'
                        }`}>
                        <Plus size={32} className={isDark ? 'text-surface-400' : 'text-slate-400'} />
                    </div>
                    <h3 className={`text-xl font-semibold mb-2 ${isDark ? 'text-white' : 'text-slate-900'}`}>No Cards Yet</h3>
                    <p className={`mb-6 ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Issue your first virtual card to get started.</p>
                    <Button onClick={handleIssueCard}>
                        <Plus size={18} className="mr-2" /> Issue Virtual Card
                    </Button>
                </Card>
            )}
        </div>
    );
}
