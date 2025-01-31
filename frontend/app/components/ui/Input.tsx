import { forwardRef } from 'react';
import { clsx } from 'clsx';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
    label?: string;
    error?: string;
    icon?: React.ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(({ label, error, icon, className, ...props }, ref) => {
    return (
        <div className="w-full">
            {label && <label className="block text-sm font-medium text-surface-300 mb-1.5 ml-1">{label}</label>}
            <div className="relative group">
                {icon && (
                    <div className="absolute inset-y-0 left-0 flex items-center pl-4 pointer-events-none text-surface-500 group-focus-within:text-brand-400 transition-colors">
                        {icon}
                    </div>
                )}
                <input
                    ref={ref}
                    className={clsx(
                        "block w-full rounded-xl border border-surface-700 bg-surface-900/50 py-3.5 text-white placeholder:text-surface-600 focus:border-brand-500 focus:ring-1 focus:ring-brand-500 transition-all duration-200 outline-none [&:-webkit-autofill]:shadow-[0_0_0_1000px_#0f172a_inset] [&:-webkit-autofill]:-webkit-text-fill-color-white",
                        icon ? "pl-11" : "pl-4",
                        error ? "border-red-500/50 focus:border-red-500 focus:ring-red-500" : "hover:border-surface-600",
                        className
                    )}
                    {...props}
                />
            </div>
            {error && <p className="mt-1 text-xs text-red-400 ml-1">{error}</p>}
        </div>
    );
});

Input.displayName = "Input";
