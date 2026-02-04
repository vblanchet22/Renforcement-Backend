import type { ReactNode, MouseEvent } from 'react';
import { motion } from 'framer-motion';

interface ButtonProps {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  children: ReactNode;
  className?: string;
  disabled?: boolean;
  type?: 'button' | 'submit' | 'reset';
  onClick?: (e: MouseEvent<HTMLButtonElement>) => void;
}

const variants = {
  primary: 'bg-primary text-white shadow-sm hover:bg-primary-dark active:shadow-none',
  secondary: 'bg-white text-slate-700 border border-slate-200 hover:bg-slate-50 hover:border-slate-300',
  ghost: 'bg-transparent text-slate-600 hover:bg-slate-100 hover:text-slate-800',
  danger: 'bg-red-500 text-white shadow-sm hover:bg-red-600 active:shadow-none',
};

const sizes = {
  sm: 'h-8 px-3 text-xs gap-1.5 rounded-lg',
  md: 'h-10 px-4 text-sm gap-2 rounded-xl',
  lg: 'h-12 px-5 text-sm gap-2 rounded-xl',
};

export function Button({
  variant = 'primary',
  size = 'md',
  isLoading = false,
  leftIcon,
  rightIcon,
  children,
  className = '',
  disabled,
  type = 'button',
  onClick,
}: ButtonProps) {
  const isDisabled = disabled || isLoading;

  return (
    <motion.button
      whileHover={isDisabled ? undefined : { scale: 1.01 }}
      whileTap={isDisabled ? undefined : { scale: 0.98 }}
      className={`
        inline-flex items-center justify-center font-medium
        transition-colors duration-150 cursor-pointer
        disabled:opacity-50 disabled:cursor-not-allowed
        ${variants[variant]} ${sizes[size]} ${className}
      `}
      disabled={isDisabled}
      type={type}
      onClick={onClick}
    >
      {isLoading ? (
        <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      ) : (
        leftIcon
      )}
      {children}
      {!isLoading && rightIcon}
    </motion.button>
  );
}
