import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "var(--background)",
        foreground: "var(--foreground)",
        // Custom Stripe-like Palette
        brand: {
             50: '#f0f9ff',
             100: '#e0f2fe',
             200: '#bae6fd',
             300: '#7dd3fc',
             400: '#38bdf8',
             500: '#0ea5e9',
             600: '#0284c7', // Primary Brand Color
             700: '#0369a1',
             800: '#075985',
             900: '#0c4a6e',
        },
        accent: {
             500: '#6366f1', // Indigo
             600: '#4f46e5',
        },
        surface: {
            50: '#f8fafc',
            100: '#f1f5f9',
            200: '#e2e8f0',
            300: '#cbd5e1',
            800: '#1e293b',
            900: '#0f172a',
            950: '#020617', // Main Dark BG
        }
      },
      fontFamily: {
        sans: ['var(--font-inter)', 'sans-serif'],
      },
      backgroundImage: {
         'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
         'hero-glow': 'conic-gradient(from 180deg at 50% 50%, #2a8af6 0deg, #a853ba 180deg, #e92a67 360deg)',
      }
    },
  },
  plugins: [],
};
export default config;
