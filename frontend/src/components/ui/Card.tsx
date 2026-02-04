import type { ReactNode } from 'react';
import { motion } from 'framer-motion';

interface CardProps {
  children: ReactNode;
  className?: string;
  padding?: 'none' | 'sm' | 'md' | 'lg';
  hover?: boolean;
  onClick?: () => void;
}

const paddingMap = { none: '', sm: 'p-5', md: 'p-6', lg: 'p-8' };

export function Card({ children, className = '', padding = 'md', hover = false, onClick }: CardProps) {
  const Component = onClick || hover ? motion.div : 'div';
  const motionProps = onClick || hover ? { whileHover: { y: -2 }, transition: { duration: 0.15 } } : {};

  return (
    <Component
      className={`bg-white rounded-2xl border border-slate-100 shadow-sm ${paddingMap[padding]} ${onClick ? 'cursor-pointer' : ''} ${className}`}
      onClick={onClick}
      {...motionProps}
    >
      {children}
    </Component>
  );
}

interface CardHeaderProps {
  title: string;
  subtitle?: string;
  action?: ReactNode;
  className?: string;
}

export function CardHeader({ title, subtitle, action, className = '' }: CardHeaderProps) {
  return (
    <div className={`flex items-start justify-between ${className}`}>
      <div>
        <h3 className="text-base font-semibold text-slate-800">{title}</h3>
        {subtitle && <p className="text-sm text-slate-400 mt-1">{subtitle}</p>}
      </div>
      {action && <div>{action}</div>}
    </div>
  );
}

interface StatCardProps {
  label: string;
  value: string | number;
  change?: { value: number; isPositive: boolean };
  icon?: ReactNode;
  color?: 'primary' | 'success' | 'warning' | 'danger' | 'accent';
}

const iconBg = {
  primary: 'bg-primary/10 text-primary',
  success: 'bg-emerald-50 text-emerald-600',
  warning: 'bg-amber-50 text-amber-600',
  danger: 'bg-red-50 text-red-500',
  accent: 'bg-accent/20 text-accent-dark',
};

export function StatCard({ label, value, change, icon, color = 'primary' }: StatCardProps) {
  return (
    <Card padding="lg">
      <div className="flex items-start justify-between">
        {icon && (
          <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${iconBg[color]}`}>
            {icon}
          </div>
        )}
        {change && (
          <span className={`inline-flex items-center text-xs font-medium px-2.5 py-1 rounded-full ${change.isPositive ? 'text-emerald-700 bg-emerald-50' : 'text-red-600 bg-red-50'}`}>
            {change.isPositive ? '+' : ''}{change.value}%
          </span>
        )}
      </div>
      <div className="mt-5">
        <p className="text-3xl font-semibold text-slate-800 tracking-tight">{value}</p>
        <p className="text-sm text-slate-400 mt-1">{label}</p>
      </div>
    </Card>
  );
}
