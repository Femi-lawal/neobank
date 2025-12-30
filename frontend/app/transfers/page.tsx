'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { ArrowLeft, ArrowRight, CheckCircle2, AlertCircle, X, Wallet, User, DollarSign, FileText, Sparkles } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { Account } from '../types';
import { useTheme } from '../context/ThemeContext';

interface RecipientAccount {
    id: string;
    name: string;
    type: 'own' | 'recent' | 'external';
    accountId: string;
}

// Compact step progress indicator
const StepIndicator = ({ currentStep, isDark }: { currentStep: number; isDark: boolean }) => {
    const steps = [
        { num: 1, label: 'Details' },
        { num: 2, label: 'Review' },
        { num: 3, label: 'Complete' },
    ];

    return (
        <div className="flex items-center justify-center mb-4">
            {steps.map((step, i) => (
                <div key={step.num} className="flex items-center">
                    <div className="flex flex-col items-center">
                        <motion.div
                            initial={false}
                            animate={{
                                scale: currentStep === step.num ? 1.1 : 1,
                                backgroundColor: currentStep >= step.num ? 'rgb(139 92 246)' : (isDark ? 'rgb(30 30 35)' : 'rgb(226 232 240)'),
                            }}
                            className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold transition-colors ${currentStep >= step.num ? 'text-white' : (isDark ? 'text-surface-500' : 'text-slate-400')
                                }`}
                        >
                            {currentStep > step.num ? <CheckCircle2 size={16} /> : step.num}
                        </motion.div>
                        <span className={`text-[10px] mt-1 ${currentStep >= step.num
                            ? (isDark ? 'text-white' : 'text-slate-900')
                            : (isDark ? 'text-surface-500' : 'text-slate-400')
                            }`}>
                            {step.label}
                        </span>
                    </div>
                    {i < steps.length - 1 && (
                        <div className={`relative w-12 sm:w-16 h-0.5 mx-1 mb-4 overflow-hidden ${isDark ? 'bg-surface-800' : 'bg-slate-200'}`}>
                            <motion.div
                                initial={false}
                                animate={{ width: currentStep > step.num ? '100%' : '0%' }}
                                className="absolute inset-y-0 left-0 bg-brand-500"
                                transition={{ duration: 0.3 }}
                            />
                        </div>
                    )}
                </div>
            ))}
        </div>
    );
};

// Confetti animation - particles are pre-generated at module level
const CONFETTI_COLORS = ['#8B5CF6', '#EC4899', '#10B981', '#F59E0B', '#3B82F6'];

// Generate particles once at module load time (pure)
const CONFETTI_PARTICLES = Array.from({ length: 30 }, (_, i) => ({
    id: i,
    x: (i * 3.33) % 100, // Deterministic positioning
    color: CONFETTI_COLORS[i % CONFETTI_COLORS.length],
    delay: (i * 0.01) % 0.3,
    size: 3 + (i % 6),
}));

const Confetti = () => {
    return (
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
            {CONFETTI_PARTICLES.map((p) => (
                <motion.div
                    key={p.id}
                    initial={{ y: -10, x: `${p.x}%`, opacity: 1 }}
                    animate={{ y: '100%', opacity: 0 }}
                    transition={{ duration: 1.5, delay: p.delay, ease: 'easeOut' }}
                    style={{
                        position: 'absolute',
                        width: p.size,
                        height: p.size,
                        backgroundColor: p.color,
                        borderRadius: '50%',
                    }}
                />
            ))}
        </div>
    );
};

export default function TransfersPage() {
    const [accounts, setAccounts] = useState<Account[]>([]);
    const [fromAccount, setFromAccount] = useState('');
    const [toAccount, setToAccount] = useState('');
    const [recipientName, setRecipientName] = useState('');
    const [amount, setAmount] = useState('');
    const [description, setDescription] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [currentStep, setCurrentStep] = useState(1);
    const [showConfirmModal, setShowConfirmModal] = useState(false);
    const [error, setError] = useState('');
    const router = useRouter();
    const { theme } = useTheme();
    const isDark = theme === 'dark';

    const presetAmounts = [50, 100, 250, 500];

    const recentRecipients: RecipientAccount[] = [
        { id: 'recent-1', name: 'John Doe', type: 'recent', accountId: 'b0000002-0002-0002-0002-000000000001' },
    ];

    useEffect(() => {
        const token = localStorage.getItem('token');
        fetch('/api/ledger/accounts', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
            .then((res) => res.json())
            .then((data) => {
                if (Array.isArray(data)) {
                    setAccounts(data);
                    if (data.length > 0) setFromAccount(data[0].id);
                }
            })
            .catch((err) => console.error(err));
    }, []);

    const getAccountName = (id: string) => {
        const acc = accounts.find(a => a.id === id);
        return acc?.name || 'Unknown Account';
    };

    const getFromAccountBalance = () => {
        const acc = accounts.find(a => a.id === fromAccount);
        return acc ? parseFloat(acc.balance || '0') : 0;
    };

    const validateTransfer = (): boolean => {
        if (!toAccount.trim()) {
            setError('Please enter a recipient account');
            return false;
        }
        if (!amount || parseFloat(amount) <= 0) {
            setError('Please enter a valid amount');
            return false;
        }
        if (parseFloat(amount) > getFromAccountBalance()) {
            setError('Insufficient funds');
            return false;
        }
        if (fromAccount === toAccount) {
            setError('Cannot transfer to the same account');
            return false;
        }
        setError('');
        return true;
    };

    const handleReviewTransfer = (e: React.FormEvent) => {
        e.preventDefault();
        if (validateTransfer()) {
            setCurrentStep(2);
            setShowConfirmModal(true);
        }
    };

    const handleConfirmTransfer = async () => {
        setIsLoading(true);
        setShowConfirmModal(false);

        try {
            const token = localStorage.getItem('token');
            const res = await fetch('/api/payment/transfer', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    from_account_id: fromAccount,
                    to_account_id: toAccount,
                    amount: amount,
                    currency: 'USD',
                    description: description,
                }),
            });

            if (!res.ok) throw new Error('Transfer failed');

            setCurrentStep(3);
            setTimeout(() => router.push('/dashboard'), 3000);
        } catch (error) {
            console.error(error);
            setError('Transfer failed. Please try again.');
            setCurrentStep(1);
            setIsLoading(false);
        }
    };

    const selectOwnAccount = (accId: string) => {
        setToAccount(accId);
        setRecipientName(getAccountName(accId));
    };

    // Success state
    if (currentStep === 3) {
        return (
            <div className="max-w-lg mx-auto">
                <StepIndicator currentStep={3} isDark={isDark} />
                <div className="relative flex flex-col items-center justify-center text-center py-12">
                    <Confetti />
                    <motion.div
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                        transition={{ type: 'spring', stiffness: 200 }}
                        className="w-16 h-16 bg-gradient-to-br from-emerald-400 to-emerald-600 rounded-full flex items-center justify-center text-white mb-4 shadow-lg"
                    >
                        <CheckCircle2 size={32} />
                    </motion.div>
                    <h2 className={`text-2xl font-bold mb-2 ${isDark ? 'text-white' : 'text-slate-900'}`}>
                        Transfer Successful! 🎉
                    </h2>
                    <p className={`text-sm mb-4 ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>
                        Your funds have been sent.
                    </p>
                    <div className={`rounded-xl p-4 text-left w-full max-w-xs border ${isDark ? 'bg-surface-800/80 border-surface-700' : 'bg-white border-slate-200 shadow-sm'
                        }`}>
                        <div className={`flex justify-between items-center mb-3 pb-3 border-b ${isDark ? 'border-surface-700' : 'border-slate-200'}`}>
                            <span className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Amount</span>
                            <span className="text-xl font-bold text-emerald-500">${parseFloat(amount).toFixed(2)}</span>
                        </div>
                        <div className="space-y-2 text-sm">
                            <div className="flex justify-between">
                                <span className={isDark ? 'text-surface-500' : 'text-slate-400'}>To</span>
                                <span className={isDark ? 'text-surface-200' : 'text-slate-700'}>{recipientName || 'Account'}</span>
                            </div>
                            <div className="flex justify-between">
                                <span className={isDark ? 'text-surface-500' : 'text-slate-400'}>Fee</span>
                                <span className="text-emerald-500">Free</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-lg mx-auto">
            <Button variant="ghost" onClick={() => router.back()} className={`mb-2 pl-0 hover:bg-transparent text-sm ${isDark ? 'hover:text-white' : 'hover:text-slate-900'}`}>
                <ArrowLeft size={14} className="mr-1" /> Back
            </Button>

            <StepIndicator currentStep={currentStep} isDark={isDark} />

            <Card className="p-5">
                <div className="mb-4">
                    <h1 className={`text-xl font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>Send Money</h1>
                    <p className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Transfer funds securely.</p>
                </div>

                {error && (
                    <motion.div
                        initial={{ opacity: 0, y: -10 }}
                        animate={{ opacity: 1, y: 0 }}
                        className="flex items-center gap-2 p-2 mb-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-500 text-sm"
                    >
                        <AlertCircle size={16} />
                        <span>{error}</span>
                    </motion.div>
                )}

                <form onSubmit={handleReviewTransfer} className="space-y-4">
                    {/* From Account */}
                    <div>
                        <label className={`block text-xs font-medium mb-1 ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>From</label>
                        <select
                            value={fromAccount}
                            onChange={(e) => setFromAccount(e.target.value)}
                            className={`block w-full rounded-lg border py-2.5 px-3 text-sm appearance-none focus:border-brand-500 outline-none ${isDark
                                ? 'border-surface-700 bg-surface-900/50 text-white'
                                : 'border-slate-200 bg-white text-slate-900'
                                }`}
                        >
                            {accounts.map(acc => (
                                <option key={acc.id} value={acc.id}>
                                    {acc.name} • ${parseFloat(acc.balance || '0').toFixed(2)}
                                </option>
                            ))}
                        </select>
                    </div>

                    {/* Arrow */}
                    <div className="flex items-center justify-center">
                        <motion.div
                            animate={{ y: [0, 3, 0] }}
                            transition={{ repeat: Infinity, duration: 1.2 }}
                            className="h-8 w-8 rounded-full bg-gradient-to-br from-brand-500 to-accent-500 flex items-center justify-center text-white shadow-md"
                        >
                            <ArrowRight size={14} className="rotate-90" />
                        </motion.div>
                    </div>

                    {/* To Account */}
                    <div className="space-y-2">
                        <label className={`block text-xs font-medium ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>To</label>

                        {/* Quick Select Buttons */}
                        <div className="flex flex-wrap gap-1.5">
                            {accounts.filter(a => a.id !== fromAccount).slice(0, 3).map(acc => (
                                <button
                                    key={acc.id}
                                    type="button"
                                    onClick={() => selectOwnAccount(acc.id)}
                                    className={`px-2.5 py-1.5 rounded-md text-xs transition-all ${toAccount === acc.id
                                        ? 'bg-brand-500 text-white'
                                        : (isDark
                                            ? 'bg-surface-800 text-surface-300 hover:bg-surface-700'
                                            : 'bg-slate-100 text-slate-600 hover:bg-slate-200')
                                        }`}
                                >
                                    {acc.name}
                                </button>
                            ))}
                            {recentRecipients.slice(0, 2).map(r => (
                                <button
                                    key={r.id}
                                    type="button"
                                    onClick={() => selectOwnAccount(r.accountId)}
                                    className={`px-2.5 py-1.5 rounded-md text-xs font-medium border transition-all ${toAccount === r.accountId
                                        ? 'bg-brand-500 text-white border-brand-500'
                                        : (isDark
                                            ? 'bg-surface-800 text-surface-300 border-surface-700 hover:bg-surface-700 hover:border-surface-600'
                                            : 'bg-white text-slate-700 border-slate-200 hover:bg-slate-50 hover:border-slate-300')
                                        }`}
                                >
                                    {r.name}
                                </button>
                            ))}
                        </div>

                        <input
                            type="text"
                            placeholder="Enter destination UUID"
                            value={toAccount}
                            onChange={(e) => { setToAccount(e.target.value); setRecipientName(''); }}
                            className={`block w-full rounded-lg border py-2.5 px-3 text-sm focus:border-brand-500 outline-none ${isDark
                                ? 'border-surface-700 bg-surface-900/50 text-white placeholder:text-surface-600'
                                : 'border-slate-200 bg-white text-slate-900 placeholder:text-slate-400'
                                }`}
                            required
                        />
                    </div>

                    {/* Amount */}
                    <div className="space-y-2">
                        <div className="flex items-center justify-between">
                            <label className={`block text-xs font-medium ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Amount</label>
                            <span className={`text-xs ${isDark ? 'text-surface-500' : 'text-slate-400'}`}>
                                Available: <span className="text-emerald-500">${getFromAccountBalance().toFixed(2)}</span>
                            </span>
                        </div>

                        <div className="relative">
                            <span className={`absolute left-3 top-1/2 -translate-y-1/2 text-xl ${isDark ? 'text-surface-500' : 'text-slate-400'}`}>$</span>
                            <input
                                type="number"
                                placeholder="0.00"
                                value={amount}
                                onChange={(e) => setAmount(e.target.value)}
                                className={`block w-full rounded-lg border py-3 pl-8 pr-3 text-2xl font-bold text-center focus:border-brand-500 outline-none ${isDark
                                    ? 'border-surface-700 bg-surface-900/50 text-white placeholder:text-surface-600'
                                    : 'border-slate-200 bg-white text-slate-900 placeholder:text-slate-300'
                                    }`}
                                required
                            />
                        </div>

                        {/* Preset Buttons */}
                        <div className="flex gap-1.5">
                            {presetAmounts.map(p => (
                                <button
                                    key={p}
                                    type="button"
                                    onClick={() => setAmount(p.toString())}
                                    className={`flex-1 py-1.5 rounded-md text-xs font-medium transition-all ${amount === p.toString()
                                        ? 'bg-brand-500 text-white'
                                        : (isDark
                                            ? 'bg-surface-800 text-surface-300 hover:bg-surface-700'
                                            : 'bg-slate-100 text-slate-600 hover:bg-slate-200')
                                        }`}
                                >
                                    ${p}
                                </button>
                            ))}
                            <button
                                type="button"
                                onClick={() => setAmount(getFromAccountBalance().toFixed(2))}
                                className={`flex-1 py-1.5 rounded-md text-xs font-medium ${isDark
                                    ? 'bg-surface-800 text-brand-400 hover:bg-surface-700'
                                    : 'bg-slate-100 text-brand-500 hover:bg-slate-200'
                                    }`}
                            >
                                Max
                            </button>
                        </div>
                    </div>

                    {/* Reference */}
                    <input
                        type="text"
                        placeholder="What's this for? (optional)"
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        className={`block w-full rounded-lg border py-2.5 px-3 text-sm focus:border-brand-500 outline-none ${isDark
                            ? 'border-surface-700 bg-surface-900/50 text-white placeholder:text-surface-500'
                            : 'border-slate-200 bg-white text-slate-900 placeholder:text-slate-400'
                            }`}
                    />

                    <Button type="submit" className="w-full" size="lg" isLoading={isLoading}>
                        <Sparkles size={16} className="mr-2" />
                        Review Transfer
                    </Button>
                </form>
            </Card>

            {/* Confirmation Modal */}
            <AnimatePresence>
                {showConfirmModal && (
                    <>
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            className="fixed inset-0 bg-black/70 backdrop-blur-sm z-50"
                            onClick={() => { setShowConfirmModal(false); setCurrentStep(1); }}
                        />
                        <motion.div
                            initial={{ opacity: 0, scale: 0.95 }}
                            animate={{ opacity: 1, scale: 1 }}
                            exit={{ opacity: 0, scale: 0.95 }}
                            className="fixed inset-4 sm:inset-auto sm:top-1/2 sm:left-1/2 sm:-translate-x-1/2 sm:-translate-y-1/2 sm:w-full sm:max-w-sm z-50 flex items-center justify-center"
                        >
                            <div className={`rounded-2xl p-5 border shadow-2xl w-full max-h-[90vh] overflow-y-auto ${isDark
                                ? 'bg-surface-900 border-surface-700'
                                : 'bg-white border-slate-200'
                                }`}>
                                <button
                                    onClick={() => { setShowConfirmModal(false); setCurrentStep(1); }}
                                    className={`absolute top-3 right-3 ${isDark ? 'text-surface-400 hover:text-white' : 'text-slate-400 hover:text-slate-900'}`}
                                >
                                    <X size={18} />
                                </button>

                                <h2 className={`text-lg font-bold mb-1 ${isDark ? 'text-white' : 'text-slate-900'}`}>Confirm Transfer</h2>
                                <p className={`text-xs mb-4 ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Review before confirming.</p>

                                <div className="space-y-2 mb-4">
                                    <div className="flex items-center justify-between p-3 rounded-lg bg-gradient-to-r from-brand-500/10 to-accent-500/10 border border-brand-500/20">
                                        <div className="flex items-center gap-2">
                                            <DollarSign size={16} className="text-brand-500" />
                                            <span className={`text-sm ${isDark ? 'text-surface-300' : 'text-slate-600'}`}>Amount</span>
                                        </div>
                                        <span className={`text-xl font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>${parseFloat(amount || '0').toFixed(2)}</span>
                                    </div>

                                    <div className={`flex items-center justify-between p-2.5 rounded-lg ${isDark ? 'bg-surface-800/50' : 'bg-slate-50'}`}>
                                        <div className="flex items-center gap-2">
                                            <Wallet size={14} className={isDark ? 'text-surface-500' : 'text-slate-400'} />
                                            <span className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>From</span>
                                        </div>
                                        <span className={`text-sm ${isDark ? 'text-surface-200' : 'text-slate-700'}`}>{getAccountName(fromAccount)}</span>
                                    </div>

                                    <div className={`flex items-center justify-between p-2.5 rounded-lg ${isDark ? 'bg-surface-800/50' : 'bg-slate-50'}`}>
                                        <div className="flex items-center gap-2">
                                            <User size={14} className={isDark ? 'text-surface-500' : 'text-slate-400'} />
                                            <span className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>To</span>
                                        </div>
                                        <span className={`text-sm truncate ml-2 ${isDark ? 'text-surface-200' : 'text-slate-700'}`}>
                                            {recipientName || toAccount.slice(0, 12) + '...'}
                                        </span>
                                    </div>

                                    {description && (
                                        <div className={`flex items-center justify-between p-2.5 rounded-lg ${isDark ? 'bg-surface-800/50' : 'bg-slate-50'}`}>
                                            <div className="flex items-center gap-2">
                                                <FileText size={14} className={isDark ? 'text-surface-500' : 'text-slate-400'} />
                                                <span className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Note</span>
                                            </div>
                                            <span className={`text-sm ${isDark ? 'text-surface-200' : 'text-slate-700'}`}>{description}</span>
                                        </div>
                                    )}

                                    <div className={`flex items-center justify-between p-2.5 rounded-lg ${isDark ? 'bg-surface-800/50' : 'bg-slate-50'}`}>
                                        <span className={`text-sm ${isDark ? 'text-surface-400' : 'text-slate-500'}`}>Fee</span>
                                        <span className="text-sm font-medium text-emerald-500">Free</span>
                                    </div>
                                </div>

                                <div className="flex gap-2">
                                    <Button
                                        variant="secondary"
                                        className="flex-1"
                                        size="sm"
                                        onClick={() => { setShowConfirmModal(false); setCurrentStep(1); }}
                                    >
                                        Cancel
                                    </Button>
                                    <Button
                                        className="flex-1"
                                        size="sm"
                                        onClick={handleConfirmTransfer}
                                        isLoading={isLoading}
                                    >
                                        Confirm & Send
                                    </Button>
                                </div>
                            </div>
                        </motion.div>
                    </>
                )}
            </AnimatePresence>
        </div>
    );
}
