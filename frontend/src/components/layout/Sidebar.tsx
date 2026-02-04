import { NavLink, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { LayoutDashboard, Receipt, Wallet, Settings, LogOut, ChevronDown, Home } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useColocation } from '../../context/ColocationContext';

const navItems = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Synthèse' },
  { to: '/expenses', icon: Receipt, label: 'Dépenses' },
  { to: '/balances', icon: Wallet, label: 'Soldes' },
];

export function Sidebar() {
  const { user, logout } = useAuth();
  const { currentColocation, colocations, selectColocation } = useColocation();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const userInitials = user
    ? `${user.prenom?.charAt(0) || ''}${user.nom?.charAt(0) || ''}`.toUpperCase()
    : 'U';

  return (
    <aside className="w-72 min-h-screen bg-white border-r border-slate-200 flex flex-col">
      {/* Logo */}
      <div className="px-6 pt-8 pb-6">
        <div className="flex items-center gap-3">
          <div className="w-11 h-11 rounded-xl bg-primary flex items-center justify-center shadow-lg shadow-primary/30">
            <Home className="w-5 h-5 text-white" />
          </div>
          <span className="text-xl font-bold text-slate-800">ColocApp</span>
        </div>
      </div>

      {/* Colocation Selector */}
      {currentColocation && (
        <div className="px-5 pb-6">
          <div className="relative group">
            <button className="w-full flex items-center justify-between px-4 py-3.5 rounded-xl bg-slate-50 hover:bg-slate-100 transition-colors border border-slate-200">
              <div className="flex items-center gap-3 min-w-0">
                <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary font-bold">
                  {currentColocation.name.charAt(0).toUpperCase()}
                </div>
                <span className="text-sm font-semibold text-slate-700 truncate">
                  {currentColocation.name}
                </span>
              </div>
              <ChevronDown className="w-4 h-4 text-slate-400 group-hover:text-slate-600 shrink-0 ml-2" />
            </button>

            <div className="absolute top-full left-0 right-0 mt-2 bg-white border border-slate-200 rounded-xl shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-50 py-2">
              {colocations.map((coloc) => (
                <button
                  key={coloc.id}
                  onClick={() => selectColocation(coloc.id)}
                  className={`w-full flex items-center gap-3 px-4 py-3 hover:bg-slate-50 transition-colors text-left ${
                    coloc.id === currentColocation.id ? 'bg-primary/5' : ''
                  }`}
                >
                  <div className="w-8 h-8 rounded-md bg-primary/10 flex items-center justify-center text-primary text-sm font-bold">
                    {coloc.name.charAt(0).toUpperCase()}
                  </div>
                  <span className="text-sm text-slate-600 truncate">{coloc.name}</span>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 px-5 overflow-y-auto">
        <ul className="flex flex-col gap-2">
          {navItems.map((item, index) => (
            <motion.li
              key={item.to}
              initial={{ opacity: 0, x: -12 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.04 }}
            >
              <NavLink
                to={item.to}
                className={({ isActive }) =>
                  `flex items-center gap-4 px-4 py-4 rounded-xl text-sm font-medium transition-all duration-200 ${
                    isActive
                      ? 'bg-primary text-white shadow-lg shadow-primary/30'
                      : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
                  }`
                }
              >
                <item.icon className="w-5 h-5 shrink-0" />
                <span>{item.label}</span>
              </NavLink>
            </motion.li>
          ))}
        </ul>
      </nav>

      {/* User section */}
      <div className="px-5 py-6 mt-auto border-t border-slate-200">
        <div className="flex items-center gap-3 mb-5">
          <div className="w-12 h-12 rounded-full bg-accent/20 flex items-center justify-center text-accent-dark font-bold text-sm">
            {userInitials}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-semibold text-slate-800 truncate">
              {user?.prenom} {user?.nom}
            </p>
            <p className="text-xs text-slate-400 truncate">{user?.email}</p>
          </div>
        </div>

        <div className="flex gap-3">
          <NavLink
            to="/settings"
            className="flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-xl text-sm font-medium text-slate-600 bg-slate-100 hover:bg-slate-200 transition-all"
          >
            <Settings className="w-4 h-4" />
            Paramètres
          </NavLink>
          <button
            onClick={handleLogout}
            className="flex items-center justify-center px-4 py-3 rounded-xl text-slate-500 bg-slate-100 hover:bg-red-100 hover:text-red-500 transition-all"
          >
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </div>
    </aside>
  );
}
