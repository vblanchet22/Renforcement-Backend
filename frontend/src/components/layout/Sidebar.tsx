import { NavLink, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  LayoutDashboard,
  Receipt,
  Wallet,
  Users,
  Vote,
  PiggyBank,
  Bell,
  Settings,
  LogOut,
  ChevronDown,
  Home,
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useColocation } from '../../context/ColocationContext';
import { Avatar } from '../ui';

const navItems = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Tableau de bord' },
  { to: '/expenses', icon: Receipt, label: 'Dépenses' },
  { to: '/balances', icon: Wallet, label: 'Soldes' },
  { to: '/payments', icon: Users, label: 'Remboursements' },
  { to: '/decisions', icon: Vote, label: 'Décisions' },
  { to: '/funds', icon: PiggyBank, label: 'Cagnottes' },
  { to: '/notifications', icon: Bell, label: 'Notifications' },
];

export function Sidebar() {
  const { user, logout } = useAuth();
  const { currentColocation, colocations, selectColocation } = useColocation();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <aside className="fixed left-0 top-0 h-screen w-64 bg-[var(--color-surface)] border-r border-[var(--color-border-light)] flex flex-col z-40">
      {/* Logo */}
      <div className="p-6 border-b border-[var(--color-border-light)]">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-[var(--radius-sm)] bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-accent)] flex items-center justify-center">
            <Home className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-display text-xl text-[var(--color-text)]">ColocApp</h1>
            <p className="text-xs text-[var(--color-text-muted)]">Gestion de colocation</p>
          </div>
        </div>
      </div>

      {/* Colocation Selector */}
      {currentColocation && (
        <div className="p-4 border-b border-[var(--color-border-light)]">
          <div className="relative group">
            <button className="w-full flex items-center justify-between p-3 rounded-[var(--radius-sm)] bg-[var(--color-bg)] hover:bg-[var(--color-surface-hover)] transition-colors">
              <div className="flex items-center gap-2 min-w-0">
                <div className="w-8 h-8 rounded-full bg-[var(--color-primary-light)] flex items-center justify-center text-[var(--color-primary)] text-sm font-medium">
                  {currentColocation.name.charAt(0).toUpperCase()}
                </div>
                <span className="text-sm font-medium text-[var(--color-text)] truncate">
                  {currentColocation.name}
                </span>
              </div>
              <ChevronDown className="w-4 h-4 text-[var(--color-text-muted)] group-hover:text-[var(--color-text)]" />
            </button>

            {/* Dropdown */}
            <div className="absolute top-full left-0 right-0 mt-1 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-sm)] shadow-[var(--shadow-lg)] opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-50">
              {colocations.map((coloc) => (
                <button
                  key={coloc.id}
                  onClick={() => selectColocation(coloc.id)}
                  className={`
                    w-full flex items-center gap-2 p-3
                    hover:bg-[var(--color-surface-hover)]
                    transition-colors text-left
                    ${coloc.id === currentColocation.id ? 'bg-[var(--color-primary-light)]' : ''}
                  `}
                >
                  <div className="w-6 h-6 rounded-full bg-[var(--color-primary-light)] flex items-center justify-center text-[var(--color-primary)] text-xs font-medium">
                    {coloc.name.charAt(0).toUpperCase()}
                  </div>
                  <span className="text-sm text-[var(--color-text)] truncate">{coloc.name}</span>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
        {navItems.map((item, index) => (
          <motion.div
            key={item.to}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.05 }}
          >
            <NavLink
              to={item.to}
              className={({ isActive }) => `
                flex items-center gap-3 px-3 py-2.5 rounded-[var(--radius-sm)]
                text-sm font-medium transition-all duration-200
                ${
                  isActive
                    ? 'bg-[var(--color-primary)] text-white shadow-[var(--shadow-sm)]'
                    : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-hover)] hover:text-[var(--color-text)]'
                }
              `}
            >
              <item.icon className="w-5 h-5" />
              {item.label}
            </NavLink>
          </motion.div>
        ))}
      </nav>

      {/* User section */}
      <div className="p-4 border-t border-[var(--color-border-light)]">
        <div className="flex items-center gap-3 p-3 rounded-[var(--radius-sm)] bg-[var(--color-bg)]">
          <Avatar
            name={user ? `${user.prenom} ${user.nom}` : 'User'}
            src={user?.avatar_url}
            size="md"
          />
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-[var(--color-text)] truncate">
              {user?.prenom} {user?.nom}
            </p>
            <p className="text-xs text-[var(--color-text-muted)] truncate">{user?.email}</p>
          </div>
        </div>

        <div className="flex gap-2 mt-3">
          <NavLink
            to="/settings"
            className="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-[var(--radius-sm)] text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-hover)] transition-colors"
          >
            <Settings className="w-4 h-4" />
            Paramètres
          </NavLink>
          <button
            onClick={handleLogout}
            className="flex items-center justify-center gap-2 px-3 py-2 rounded-[var(--radius-sm)] text-sm text-[var(--color-danger)] hover:bg-[var(--color-danger-light)] transition-colors"
          >
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </div>
    </aside>
  );
}
