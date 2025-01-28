'use client';

import Link from 'next/link';
import { motion } from 'framer-motion';
import { ArrowRight } from 'lucide-react';

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-white text-slate-900 font-sans overflow-x-hidden">

      {/* HERO with Stripe-style Gradient - Skewed */}
      <div className="relative">
        {/* The Gradient Background - Skewed */}
        <div
          className="absolute inset-0 bg-stripe-gradient origin-top-left"
          style={{ transform: 'skewY(-6deg)', transformOrigin: '0 0' }}
        />

        {/* Content Container - Not Skewed */}
        <div className="relative z-10">
          {/* Navbar */}
          <nav className="flex items-center justify-between px-6 lg:px-12 py-5 max-w-[1080px] mx-auto">
            <div className="flex items-center gap-12">
              <Link href="/" className="flex items-center gap-2">
                <div className="h-8 w-8 rounded-lg bg-white/20 backdrop-blur flex items-center justify-center text-white font-bold text-lg">
                  N
                </div>
                <span className="text-white text-xl font-bold tracking-tight">NeoBank</span>
              </Link>
              <div className="hidden md:flex items-center gap-7 text-[15px] font-medium text-white/90">
                <a href="#" className="hover:text-white transition-opacity">Products</a>
                <a href="#" className="hover:text-white transition-opacity">Solutions</a>
                <a href="#" className="hover:text-white transition-opacity">Developers</a>
                <a href="#" className="hover:text-white transition-opacity">Pricing</a>
              </div>
            </div>
            <div className="flex items-center gap-6">
              <Link href="/login" className="text-[15px] font-medium text-white/90 hover:text-white flex items-center gap-1">
                Sign in <ArrowRight size={14} />
              </Link>
              <Link href="/register">
                <button className="bg-white text-slate-900 px-5 py-2 rounded-full text-[15px] font-semibold hover:bg-slate-100 transition-colors flex items-center gap-1">
                  Get Started <ArrowRight size={14} />
                </button>
              </Link>
            </div>
          </nav>

          {/* Hero Content */}
          <div className="max-w-[1080px] mx-auto px-6 pt-20 pb-56 grid grid-cols-1 lg:grid-cols-2 gap-16 items-start">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8 }}
              className="pt-8"
            >
              <h1 className="text-5xl lg:text-[72px] leading-[1.05] font-bold text-white tracking-[-0.03em] mb-8">
                Banking infrastructure for the digital age
              </h1>
              <p className="text-lg lg:text-xl text-white/90 mb-10 leading-relaxed max-w-lg font-normal">
                Join millions using NeoBank to manage accounts, send payments, and grow their wealth with modern financial tools.
              </p>
              <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
                {/* Email Input Style */}
                <div className="flex-1 w-full sm:max-w-sm relative">
                  <input
                    type="email"
                    placeholder="Email address"
                    className="w-full px-5 py-3.5 pr-36 rounded-full bg-white/10 backdrop-blur border border-white/30 text-white placeholder:text-white/60 focus:outline-none focus:ring-2 focus:ring-white/50 text-[15px]"
                  />
                  <Link href="/register" className="absolute right-1.5 top-1.5 bottom-1.5">
                    <button className="h-full px-5 bg-slate-900 text-white rounded-full text-[15px] font-semibold hover:bg-slate-800 transition-colors flex items-center gap-1">
                      Start now <ArrowRight size={14} />
                    </button>
                  </Link>
                </div>
              </div>
            </motion.div>

            {/* Floating UI Graphic - Phone + Dashboard */}
            <motion.div
              initial={{ opacity: 0, x: 50 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 1, delay: 0.3 }}
              className="relative hidden lg:block"
            >
              {/* Phone Mockup */}
              <div className="absolute left-0 top-0 w-[250px] z-20">
                <div className="bg-white rounded-[2.5rem] p-2 shadow-2xl">
                  <div className="bg-white rounded-[2rem] overflow-hidden">
                    {/* Phone Notch */}
                    <div className="h-6 bg-white flex justify-center">
                      <div className="w-20 h-5 bg-black rounded-b-2xl"></div>
                    </div>
                    {/* Phone Content */}
                    <div className="p-4 space-y-3">
                      {/* Product Card */}
                      <div className="bg-gradient-to-br from-purple-400 via-pink-400 to-orange-300 rounded-xl p-4 flex flex-col items-center">
                        <div className="w-10 h-10 bg-yellow-400 rounded-lg mb-2 flex items-center justify-center text-lg">💳</div>
                        <span className="text-xs font-semibold text-white">Premium Card</span>
                        <span className="text-[10px] text-white/80">$9.99/month</span>
                      </div>
                      {/* Apple Pay Button */}
                      <button className="w-full bg-black text-white py-2.5 rounded-lg text-sm font-medium">
                        Pay
                      </button>
                      <div className="text-center text-xs text-slate-400">Or pay by card</div>
                      {/* Form Fields */}
                      <div className="space-y-2">
                        <input className="w-full px-3 py-2 border border-slate-200 rounded-md text-sm placeholder:text-slate-400" placeholder="Email" />
                        <div className="grid grid-cols-2 gap-2">
                          <input className="px-3 py-2 border border-slate-200 rounded-md text-sm placeholder:text-slate-400" placeholder="Card" />
                          <input className="px-3 py-2 border border-slate-200 rounded-md text-sm placeholder:text-slate-400" placeholder="CVC" />
                        </div>
                      </div>
                      <button className="w-full bg-gradient-to-r from-purple-500 to-pink-500 text-white py-2.5 rounded-lg text-sm font-semibold">
                        Pay $9.99
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              {/* Dashboard Mockup */}
              <div className="ml-36 mt-6 w-[550px] bg-white/95 backdrop-blur rounded-xl p-5 shadow-2xl border border-white/50">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                    <div className="w-6 h-6 bg-gradient-to-br from-purple-500 to-pink-500 rounded flex items-center justify-center text-white text-xs font-bold">N</div>
                    <span className="text-xs font-bold text-slate-700 uppercase tracking-wider">NeoBank</span>
                  </div>
                  <div className="text-xs text-slate-400">Today</div>
                </div>
                <div className="grid grid-cols-2 gap-6">
                  {/* Left: Volume Chart */}
                  <div>
                    <div className="text-xs text-slate-500 mb-1">Total Balance</div>
                    <div className="text-2xl font-bold text-slate-900">$24,580.00</div>
                    <div className="text-xs text-emerald-500 font-medium">+12.5% this month</div>
                    {/* Mock Chart */}
                    <div className="mt-3 h-16 bg-gradient-to-t from-purple-100 to-transparent rounded relative">
                      <svg className="w-full h-full" viewBox="0 0 200 60" preserveAspectRatio="none">
                        <defs>
                          <linearGradient id="chartGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                            <stop offset="0%" stopColor="#a960ee" />
                            <stop offset="100%" stopColor="#ff333d" />
                          </linearGradient>
                        </defs>
                        <path d="M0,50 Q30,30 60,35 T120,20 T200,25" fill="none" stroke="url(#chartGradient)" strokeWidth="2.5" />
                      </svg>
                    </div>
                  </div>
                  {/* Right: Stats */}
                  <div className="space-y-3">
                    <div className="p-3 bg-slate-50 rounded-lg">
                      <div className="text-xs text-slate-500">Savings</div>
                      <div className="text-lg font-bold text-slate-900">$8,240.00</div>
                    </div>
                    <div className="p-3 bg-slate-50 rounded-lg">
                      <div className="text-xs text-slate-500">Investments</div>
                      <div className="text-lg font-bold text-slate-900">$16,340.00</div>
                    </div>
                  </div>
                </div>
              </div>
            </motion.div>
          </div>
        </div>
      </div>

      {/* White Section - Features */}
      <div className="relative z-20 bg-white py-20 px-6">
        <div className="max-w-[1080px] mx-auto">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
            {[
              { title: "Instant Transfers", desc: "Send money globally in seconds." },
              { title: "Bank-grade Security", desc: "Your funds are always protected." },
              { title: "Smart Analytics", desc: "Track spending with AI insights." },
              { title: "Multi-Currency", desc: "Hold 50+ currencies in one account." }
            ].map((f, i) => (
              <div key={i} className="border-l-2 border-purple-500 pl-4">
                <h4 className="font-bold text-[15px] text-slate-900 mb-2">{f.title}</h4>
                <p className="text-slate-600 text-sm leading-relaxed">{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Testimonials / Trust Section */}
      <div className="bg-slate-50 py-16 px-6">
        <div className="max-w-[1080px] mx-auto text-center">
          <p className="text-slate-500 text-sm mb-8">Trusted by innovative companies worldwide</p>
          <div className="flex items-center justify-center gap-12 flex-wrap">
            {["Shopify", "Notion", "Figma", "Vercel", "Linear", "Raycast"].map((logo, i) => (
              <span key={i} className="text-xl font-bold text-slate-300 hover:text-slate-400 transition-colors">{logo}</span>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
