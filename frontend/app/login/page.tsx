'use client';

import { useState } from "react";
import { useRouter } from "next/navigation";
import axios from "axios";
import { Mail, Lock, ArrowRight } from "lucide-react";
import Link from 'next/link';
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";
import { motion } from "framer-motion";

export default function LoginPage() {
    const router = useRouter();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError("");

        try {
            const res = await axios.post("/api/identity/auth/login", {
                email,
                password,
            });
            localStorage.setItem("token", res.data.token);
            router.push("/dashboard");
        } catch (err: any) {
            setError(err.response?.data?.error || "Login failed");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen grid lg:grid-cols-2">
            {/* Left: Form */}
            <div className="flex items-center justify-center p-8 xs:p-12 bg-surface-950">
                <div className="w-full max-w-sm space-y-8">
                    <motion.div
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.5 }}
                    >
                        <div className="h-12 w-12 rounded-xl bg-gradient-to-br from-gradient-purple to-gradient-red flex items-center justify-center text-white font-bold text-2xl shadow-lg mb-6">
                            N
                        </div>
                        <h2 className="text-3xl font-bold tracking-tight text-white mb-2">Welcome back</h2>
                        <p className="text-surface-400">Enter your credentials to access your account.</p>
                    </motion.div>

                    <form onSubmit={handleLogin} className="space-y-6">
                        <Input
                            icon={<Mail size={18} />}
                            type="email"
                            placeholder="name@example.com"
                            value={email}
                            onChange={e => setEmail(e.target.value)}
                            required
                            autoFocus
                        />
                        <Input
                            icon={<Lock size={18} />}
                            type="password"
                            placeholder="Password"
                            value={password}
                            onChange={e => setPassword(e.target.value)}
                            required
                        />

                        {error && (
                            <div className="text-red-400 text-sm bg-red-500/10 p-3 rounded-lg border border-red-500/20">
                                {error}
                            </div>
                        )}

                        <Button type="submit" className="w-full" isLoading={loading} size="lg">
                            Sign in to Account
                        </Button>
                    </form>

                    <div className="text-center text-sm text-surface-500">
                        Don&apos;t have an account?{" "}
                        <Link href="/register" className="text-gradient-cyan font-medium hover:text-white transition-colors">
                            Create an account
                        </Link>
                    </div>

                    <div className="pt-8 border-t border-surface-800 text-center">
                        <p className="text-xs text-surface-600 uppercase tracking-widest mb-4">Demo Credentials</p>
                        <div className="bg-surface-900/50 p-4 rounded-xl border border-surface-800 text-xs text-surface-400 font-mono">
                            user@example.com <br /> password
                        </div>
                    </div>
                </div>
            </div>

            {/* Right: Feature/Art with Stripe Gradient */}
            <div className="hidden lg:flex flex-col justify-center p-12 relative overflow-hidden">
                {/* Stripe Gradient Background */}
                <div className="absolute inset-0 bg-stripe-gradient"></div>

                <div className="relative z-10 max-w-lg mx-auto text-center">
                    <motion.div
                        initial={{ opacity: 0, scale: 0.95 }}
                        animate={{ opacity: 1, scale: 1 }}
                        transition={{ delay: 0.2, duration: 0.8 }}
                    >
                        {/* Mock Dashboard Card */}
                        <div className="bg-white/95 backdrop-blur rounded-2xl shadow-2xl p-6 mb-8 text-left">
                            <div className="flex items-center gap-2 mb-4">
                                <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-pink-500 rounded-lg flex items-center justify-center text-white font-bold text-sm">N</div>
                                <span className="font-bold text-slate-800">NeoBank Dashboard</span>
                            </div>
                            <div className="text-3xl font-bold text-slate-900 mb-1">$24,580.00</div>
                            <div className="text-sm text-emerald-500 font-medium flex items-center gap-1">
                                <ArrowRight size={14} className="rotate-[-45deg]" /> +12.5% this month
                            </div>
                            <div className="mt-4 h-12 bg-gradient-to-t from-purple-100 to-transparent rounded">
                                <svg className="w-full h-full" viewBox="0 0 200 40" preserveAspectRatio="none">
                                    <path d="M0,30 Q40,10 80,20 T160,15 T200,20" fill="none" stroke="#a960ee" strokeWidth="2" />
                                </svg>
                            </div>
                        </div>
                    </motion.div>
                    <h3 className="text-2xl font-bold text-white mb-4">Banking for the Modern Era</h3>
                    <p className="text-white/80">Experience lightning fast transfers, real-time insights, and premium cards. All in one place.</p>
                </div>
            </div>
        </div>
    );
}
