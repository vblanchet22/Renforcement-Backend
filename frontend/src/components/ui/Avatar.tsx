interface AvatarProps {
  src?: string | null;
  name: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
}

const sizeStyles = {
  sm: 'w-8 h-8 text-xs',
  md: 'w-10 h-10 text-sm',
  lg: 'w-12 h-12 text-base',
  xl: 'w-16 h-16 text-lg',
};

const colors = [
  'bg-blue-100 text-blue-700',
  'bg-emerald-100 text-emerald-700',
  'bg-amber-100 text-amber-700',
  'bg-rose-100 text-rose-700',
  'bg-purple-100 text-purple-700',
  'bg-cyan-100 text-cyan-700',
  'bg-orange-100 text-orange-700',
  'bg-teal-100 text-teal-700',
];

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((part) => part[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

function getColorFromName(name: string): string {
  const hash = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
  return colors[hash % colors.length];
}

export function Avatar({ src, name, size = 'md', className = '' }: AvatarProps) {
  if (src) {
    return (
      <img
        src={src}
        alt={name}
        className={`
          ${sizeStyles[size]}
          rounded-full object-cover
          ring-2 ring-[var(--color-surface)] ring-offset-1
          ${className}
        `}
      />
    );
  }

  return (
    <div
      className={`
        ${sizeStyles[size]}
        ${getColorFromName(name)}
        rounded-full
        flex items-center justify-center
        font-medium
        ring-2 ring-[var(--color-surface)] ring-offset-1
        ${className}
      `}
    >
      {getInitials(name)}
    </div>
  );
}

interface AvatarGroupProps {
  users: Array<{ name: string; avatar_url?: string | null }>;
  max?: number;
  size?: 'sm' | 'md' | 'lg';
}

export function AvatarGroup({ users, max = 4, size = 'sm' }: AvatarGroupProps) {
  const visibleUsers = users.slice(0, max);
  const remaining = users.length - max;

  return (
    <div className="flex -space-x-2">
      {visibleUsers.map((user, i) => (
        <Avatar
          key={i}
          src={user.avatar_url}
          name={user.name}
          size={size}
          className="ring-[var(--color-surface)]"
        />
      ))}
      {remaining > 0 && (
        <div
          className={`
            ${sizeStyles[size]}
            rounded-full
            bg-[var(--color-surface-hover)]
            text-[var(--color-text-secondary)]
            flex items-center justify-center
            font-medium
            ring-2 ring-[var(--color-surface)]
          `}
        >
          +{remaining}
        </div>
      )}
    </div>
  );
}
