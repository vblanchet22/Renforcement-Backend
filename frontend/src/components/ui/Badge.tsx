import type { ReactNode } from 'react';

interface BadgeProps {
  children: ReactNode;
  variant?: 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'accent';
  size?: 'sm' | 'md';
  dot?: boolean;
  className?: string;
}

const variantClasses = {
  default: 'bg-slate-100 text-slate-600',
  primary: 'bg-primary-light text-primary',
  success: 'bg-emerald-50 text-emerald-700',
  warning: 'bg-amber-50 text-amber-700',
  danger: 'bg-red-50 text-red-600',
  accent: 'bg-accent-light text-amber-700',
};

const dotClasses = {
  default: 'bg-slate-400',
  primary: 'bg-primary',
  success: 'bg-emerald-500',
  warning: 'bg-amber-500',
  danger: 'bg-red-500',
  accent: 'bg-accent',
};

const sizeClasses = {
  sm: 'px-2 py-0.5 text-[10px]',
  md: 'px-2.5 py-0.5 text-xs',
};

export function Badge({ children, variant = 'default', size = 'sm', dot = false, className = '' }: BadgeProps) {
  return (
    <span className={`inline-flex items-center gap-1 font-medium rounded-full whitespace-nowrap ${variantClasses[variant]} ${sizeClasses[size]} ${className}`}>
      {dot && <span className={`w-1.5 h-1.5 rounded-full ${dotClasses[variant]}`} />}
      {children}
    </span>
  );
}
