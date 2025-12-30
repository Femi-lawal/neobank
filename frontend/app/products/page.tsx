'use client';

import { useState, useEffect } from 'react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { ChevronRight, PiggyBank, Banknote, TrendingUp, CreditCard, Check, Sparkles, LucideIcon } from 'lucide-react';
import { useTheme } from '../context/ThemeContext';
import { motion } from 'framer-motion';

interface Product {
    Code: string;
    Name: string;
    Type: string;
    InterestRate: string;
}

const productCategories = [
    { id: 'all', label: 'All Products', icon: Sparkles },
    { id: 'savings', label: 'Savings', icon: PiggyBank },
    { id: 'checking', label: 'Checking', icon: Banknote },
    { id: 'investment', label: 'Investment', icon: TrendingUp },
    { id: 'credit', label: 'Credit', icon: CreditCard },
];

const productTypeIcons: Record<string, LucideIcon> = {
    'SAVINGS': PiggyBank,
    'CHECKING': Banknote,
    'INVESTMENT': TrendingUp,
    'CREDIT': CreditCard,
};

const productTypeGradients: Record<string, string> = {
    'SAVINGS': 'from-emerald-500 to-teal-500',
    'CHECKING': 'from-blue-500 to-indigo-500',
    'INVESTMENT': 'from-purple-500 to-pink-500',
    'CREDIT': 'from-orange-500 to-amber-500',
};

export default function ProductsPage() {
    const [products, setProducts] = useState<Product[]>([]);
    const [activeCategory, setActiveCategory] = useState('all');
    const [compareList, setCompareList] = useState<Set<string>>(new Set());
    const { theme } = useTheme();
    const isDark = theme === 'dark';

    useEffect(() => {
        const token = localStorage.getItem('token');
        fetch('/api/product/products', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
            .then(res => res.json())
            .then(data => setProducts(data || []))
            .catch((err) => console.error(err));
    }, []);

    const filteredProducts = activeCategory === 'all'
        ? products
        : products.filter(p => p.Type.toLowerCase() === activeCategory);

    const toggleCompare = (code: string) => {
        setCompareList(prev => {
            const newSet = new Set(prev);
            if (newSet.has(code)) {
                newSet.delete(code);
            } else if (newSet.size < 3) {
                newSet.add(code);
            }
            return newSet;
        });
    };

    return (
        <div className="space-y-8">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4">
                <div>
                    <h1 className={`text-2xl sm:text-3xl font-bold ${isDark ? 'text-white' : 'text-slate-900'}`}>
                        Financial Products
                    </h1>
                    <p className={isDark ? 'text-surface-400' : 'text-slate-500'}>
                        Explore our premium banking solutions tailored for you.
                    </p>
                </div>
                {compareList.size > 0 && (
                    <Button className="gap-2">
                        Compare ({compareList.size})
                    </Button>
                )}
            </div>

            {/* Category Filter Tabs */}
            <div className="flex flex-wrap gap-2">
                {productCategories.map((cat) => {
                    const Icon = cat.icon;
                    const isActive = activeCategory === cat.id;
                    return (
                        <button
                            key={cat.id}
                            onClick={() => setActiveCategory(cat.id)}
                            className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${isActive
                                ? 'bg-brand-500 text-white shadow-lg shadow-brand-500/25'
                                : isDark
                                    ? 'bg-surface-800 text-surface-300 hover:bg-surface-700'
                                    : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'
                                }`}
                        >
                            <Icon size={16} />
                            {cat.label}
                        </button>
                    );
                })}
            </div>

            {/* Products Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {filteredProducts.map((p, i) => {
                    const productType = p.Type.toUpperCase();
                    const Icon = productTypeIcons[productType] || PiggyBank;
                    const gradient = productTypeGradients[productType] || 'from-purple-500 to-pink-500';
                    const isComparing = compareList.has(p.Code);

                    return (
                        <motion.div
                            key={p.Code}
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: i * 0.1 }}
                        >
                            <Card
                                delay={i * 0.1}
                                className={`relative group overflow-hidden transition-all ${isComparing
                                    ? 'ring-2 ring-brand-500 ring-offset-2 ring-offset-surface-950'
                                    : 'border-brand-500/20 hover:border-brand-500/40'
                                    }`}
                            >
                                {/* Background Glow */}
                                <div className={`absolute top-0 right-0 w-32 h-32 bg-gradient-to-br ${gradient} opacity-10 rounded-full blur-2xl -mr-10 -mt-10 transition-all group-hover:opacity-20`}></div>

                                <div className="relative z-10">
                                    {/* Header with Icon and Type */}
                                    <div className="flex items-center justify-between mb-6">
                                        <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${gradient} flex items-center justify-center text-white shadow-lg`}>
                                            <Icon size={24} />
                                        </div>
                                        <span className={`px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider ${isDark ? 'bg-surface-800 text-brand-400' : 'bg-slate-100 text-brand-600'
                                            }`}>
                                            {p.Type}
                                        </span>
                                    </div>

                                    {/* Product Name */}
                                    <h2 className={`text-xl font-bold mb-2 ${isDark ? 'text-white' : 'text-slate-900'}`}>
                                        {p.Name}
                                    </h2>

                                    {/* Interest Rate */}
                                    <div className="flex items-baseline gap-1 mb-6">
                                        <span className={`text-4xl font-bold bg-gradient-to-r ${gradient} bg-clip-text text-transparent`}>
                                            {(parseFloat(p.InterestRate) * 100).toFixed(2)}%
                                        </span>
                                        <span className={isDark ? 'text-surface-500' : 'text-slate-500'}>APY</span>
                                    </div>

                                    {/* Features */}
                                    <ul className={`space-y-3 mb-6 text-sm ${isDark ? 'text-surface-400' : 'text-slate-600'}`}>
                                        <li className="flex items-center gap-2">
                                            <div className={`w-5 h-5 rounded-full bg-gradient-to-br ${gradient} flex items-center justify-center`}>
                                                <Check size={12} className="text-white" />
                                            </div>
                                            Compounded Daily
                                        </li>
                                        <li className="flex items-center gap-2">
                                            <div className={`w-5 h-5 rounded-full bg-gradient-to-br ${gradient} flex items-center justify-center`}>
                                                <Check size={12} className="text-white" />
                                            </div>
                                            No Monthly Fees
                                        </li>
                                        <li className="flex items-center gap-2">
                                            <div className={`w-5 h-5 rounded-full bg-gradient-to-br ${gradient} flex items-center justify-center`}>
                                                <Check size={12} className="text-white" />
                                            </div>
                                            FDIC Insured
                                        </li>
                                    </ul>

                                    {/* Actions */}
                                    <div className="flex gap-2">
                                        <Button className="flex-1 group-hover:bg-brand-500 transition-colors">
                                            Apply Now <ChevronRight size={16} className="ml-2" />
                                        </Button>
                                        <Button
                                            variant={isComparing ? "primary" : "outline"}
                                            size="sm"
                                            onClick={() => toggleCompare(p.Code)}
                                            className="px-3"
                                            title={isComparing ? "Remove from compare" : "Add to compare"}
                                        >
                                            {isComparing ? <Check size={16} /> : '+'}
                                        </Button>
                                    </div>
                                </div>
                            </Card>
                        </motion.div>
                    );
                })}
            </div>

            {filteredProducts.length === 0 && (
                <Card className="text-center py-16">
                    <div className={`w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4 ${isDark ? 'bg-surface-800' : 'bg-slate-100'
                        }`}>
                        <Sparkles size={28} className={isDark ? 'text-surface-400' : 'text-slate-400'} />
                    </div>
                    <h3 className={`text-xl font-semibold mb-2 ${isDark ? 'text-white' : 'text-slate-900'}`}>
                        No Products Found
                    </h3>
                    <p className={isDark ? 'text-surface-400' : 'text-slate-500'}>
                        Try selecting a different category.
                    </p>
                </Card>
            )}
        </div>
    );
}

