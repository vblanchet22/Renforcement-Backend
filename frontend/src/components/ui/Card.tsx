import type { ReactNode } from 'react';
import { motion } from 'framer-motion';

interface CardProps {
  children: ReactNode;
  className?: string;
  padding?: 'none' | 'sm' | 'md' | 'lg';
  hover?: boolean;
  onClick?: () => void;
}

export function Card({
  children,
  className = '',
  padding = 'md',
  hover = false,
  onClick,
}: CardProps) {
  const paddingStyles = {
    none: '',
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8',
  };

  const Component = onClick || hover ? motion.div : 'div';
  const motionProps = onClick || hover
    ? {
        whileHover: { y: -2, boxShadow: 'var(--shadow-lg)' },
        transition: { duration: 0.2 },
      }
    : {};

  return (
    <Component
      className={`
        bg-[var(--color-surface)]
        rounded-[var(--radius-md)]
        shadow-[var(--shadow-md)]
        border border-[var(--color-border-light)]
        ${paddingStyles[padding]}
        ${onClick ? 'cursor-pointer' : ''}
        ${className}
      `}
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
    <div className={`flex items-start justify-between mb-4 ${className}`}>
      <div>
        <h3 className="text-lg font-medium text-[var(--color-text)]">{title}</h3>
        {subtitle && (
          <p className="text-sm text-[var(--color-text-secondary)] mt-0.5">{subtitle}</p>
        )}
      </div>
      {action && <div>{action}</div>}
    </div>
  );
}

interface StatCardProps {
  label: string;
  value: string | number;
  change?: {
    value: number;
    isPositive: boolean;
  };
  icon?: ReactNode;
  color?: 'primary' | 'success' | 'warning' | 'danger' | 'accent';
}

export function StatCard({ label, value, change, icon, color = 'primary' }: StatCardProps) {
  const colorStyles = {
    primary: 'bg-[var(--color-primary-light)] text-[var(--color-primary)]',
    success: 'bg-[var(--color-success-light)] text-[var(--color-success)]',
    warning: 'bg-[var(--color-warning-light)] text-[var(--color-warning)]',
    danger: 'bg-[var(--color-danger-light)] text-[var(--color-danger)]',
    accent: 'bg-amber-50 text-[var(--color-accent-hover)]',
  };

  return (
    <Card className="relative overflow-hidden">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm text-[var(--color-text-secondary)] mb-1">{label}</p>
          <p className="text-2xl font-semibold text-[var(--color-text)] tracking-tight">{value}</p>
          {change && (
            <p
              className={`text-sm mt-1 ${
                change.isPositive ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]'
              }`}
            >
              {change.isPositive ? '+' : ''}
              {change.value}% vs mois dernier
            </p>
          )}
        </div>
        {icon && (
          <div className={`p-3 rounded-[var(--radius-sm)] ${colorStyles[color]}`}>{icon}</div>
        )}
      </div>
    </Card>
  );
}
