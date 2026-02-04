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

const palette = [
  'bg-blue-100 text-blue-700',
  'bg-emerald-100 text-emerald-700',
  'bg-violet-100 text-violet-700',
  'bg-amber-100 text-amber-700',
  'bg-rose-100 text-rose-700',
  'bg-cyan-100 text-cyan-700',
  'bg-orange-100 text-orange-700',
  'bg-indigo-100 text-indigo-700',
];

function getInitials(name: string): string {
  return name.split(' ').map((p) => p[0]).join('').toUpperCase().slice(0, 2);
}

function colorFromName(name: string): string {
  const hash = name.split('').reduce((a, c) => a + c.charCodeAt(0), 0);
  return palette[hash % palette.length];
}

export function Avatar({ src, name, size = 'md', className = '' }: AvatarProps) {
  if (src) {
    return <img src={src} alt={name} className={`${sizeStyles[size]} rounded-full object-cover ${className}`} />;
  }
  return (
    <div className={`${sizeStyles[size]} ${colorFromName(name)} rounded-full flex items-center justify-center font-semibold select-none ${className}`} title={name}>
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
  const visible = users.slice(0, max);
  const remaining = users.length - max;
  return (
    <div className="flex -space-x-2">
      {visible.map((u, i) => (
        <div key={i} className="ring-2 ring-white rounded-full">
          <Avatar src={u.avatar_url} name={u.name} size={size} />
        </div>
      ))}
      {remaining > 0 && (
        <div className={`${sizeStyles[size]} rounded-full ring-2 ring-white bg-slate-200 text-slate-600 font-medium flex items-center justify-center`}>
          +{remaining}
        </div>
      )}
    </div>
  );
}
