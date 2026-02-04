import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import {
  TrendingUp,
  Receipt,
  Users,
  Wallet,
  ArrowRight,
  Plus,
  Calendar,
} from 'lucide-react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import { useColocation } from '../context/ColocationContext';
import { useAuth } from '../context/AuthContext';
import { expenseApi, balanceApi, categoryApi, colocationApi } from '../api';
import { Card, CardHeader, StatCard, Avatar, Badge, Button, Modal, Input } from '../components/ui';
import type { Expense, UserBalance, CategoryStat, SimplifiedDebt } from '../types';

export function Dashboard() {
  const { user } = useAuth();
  const { currentColocation, refreshColocations } = useColocation();
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [balances, setBalances] = useState<UserBalance[]>([]);
  const [categoryStats, setCategoryStats] = useState<CategoryStat[]>([]);
  const [debts, setDebts] = useState<SimplifiedDebt[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [createError, setCreateError] = useState('');
  const [createForm, setCreateForm] = useState({ name: '', description: '', address: '' });

  useEffect(() => {
    const fetchData = async () => {
      if (!currentColocation) return;
      setIsLoading(true);
      try {
        const [expensesRes, balancesRes, statsRes, debtsRes] = await Promise.all([
          expenseApi.list({ colocation_id: currentColocation.id, per_page: 5 }),
          balanceApi.getBalances(currentColocation.id),
          categoryApi.getStats({ colocation_id: currentColocation.id }),
          balanceApi.getSimplifiedDebts(currentColocation.id),
        ]);
        setExpenses(expensesRes.expenses);
        setBalances(balancesRes);
        setCategoryStats(statsRes);
        setDebts(debtsRes);
      } catch (error) {
        console.error('Error fetching dashboard data:', error);
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [currentColocation]);

  const userBalance = balances.find((b) => b.user_id === user?.id);
  const totalExpenses = expenses.reduce((sum, e) => sum + e.amount, 0);

  const monthlyData = [
    { month: 'Jan', amount: 450 },
    { month: 'Fév', amount: 520 },
    { month: 'Mar', amount: 380 },
    { month: 'Avr', amount: 610 },
    { month: 'Mai', amount: 490 },
    { month: 'Juin', amount: 430 },
  ];

  const COLORS = ['#5682F2', '#F1C086', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6'];

  const handleCreateColocation = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!createForm.name.trim()) {
      setCreateError('Le nom est obligatoire');
      return;
    }
    setIsCreating(true);
    setCreateError('');
    try {
      await colocationApi.create({
        name: createForm.name.trim(),
        description: createForm.description.trim() || undefined,
        address: createForm.address.trim() || undefined,
      });
      setIsCreateModalOpen(false);
      setCreateForm({ name: '', description: '', address: '' });
      await refreshColocations();
    } catch (error) {
      console.error(error);
      setCreateError("Impossible de créer la colocation. Réessayez.");
    } finally {
      setIsCreating(false);
    }
  };

  // No colocation state
  if (!currentColocation) {
    return (
      <>
        <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
          <div className="w-24 h-24 rounded-3xl bg-primary/10 flex items-center justify-center mb-8">
            <Users className="w-12 h-12 text-primary" />
          </div>
          <h2 className="text-3xl font-semibold text-slate-800 mb-3">Aucune colocation</h2>
          <p className="text-slate-500 mb-8 max-w-md text-lg">
            Créez ou rejoignez une colocation pour commencer à gérer vos dépenses partagées.
          </p>
          <div className="flex gap-4 flex-wrap justify-center">
            <Button size="lg" leftIcon={<Plus className="w-5 h-5" />} onClick={() => setIsCreateModalOpen(true)}>
              Créer une colocation
            </Button>
            <Button size="lg" variant="secondary">Rejoindre avec un code</Button>
          </div>
        </div>

        <Modal
          isOpen={isCreateModalOpen}
          onClose={() => { setIsCreateModalOpen(false); setCreateError(''); }}
          title="Nouvelle colocation"
          size="md"
        >
          <form className="space-y-5" onSubmit={handleCreateColocation}>
            <Input label="Nom de la colocation" value={createForm.name} onChange={(e) => setCreateForm((prev) => ({ ...prev, name: e.target.value }))} required />
            <Input label="Description" value={createForm.description} onChange={(e) => setCreateForm((prev) => ({ ...prev, description: e.target.value }))} placeholder="Ex : Appartement rue de la Paix" />
            <Input label="Adresse" value={createForm.address} onChange={(e) => setCreateForm((prev) => ({ ...prev, address: e.target.value }))} placeholder="Adresse (optionnel)" />
            {createError && <p className="text-sm text-red-500">{createError}</p>}
            <div className="flex justify-end gap-3 pt-4">
              <Button type="button" variant="secondary" onClick={() => setIsCreateModalOpen(false)}>Annuler</Button>
              <Button type="submit" isLoading={isCreating}>Créer</Button>
            </div>
          </form>
        </Modal>
      </>
    );
  }

  return (
    <div className="space-y-10">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <motion.h1 initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="text-3xl font-semibold text-slate-800 mb-2">
            Bonjour, {user?.prenom} !
          </motion.h1>
          <p className="text-slate-500 text-lg">
            Voici un aperçu de votre colocation {currentColocation.name}
          </p>
        </div>
        <Button size="lg" leftIcon={<Plus className="w-5 h-5" />}>Nouvelle dépense</Button>
      </div>

      {/* Stats Grid */}
      <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
        <StatCard
          label="Votre solde"
          value={`${(userBalance?.net_balance || 0) >= 0 ? '+' : ''}${(userBalance?.net_balance || 0).toFixed(2)} €`}
          icon={<Wallet className="w-6 h-6" />}
          color={(userBalance?.net_balance || 0) >= 0 ? 'success' : 'danger'}
        />
        <StatCard
          label="Total des dépenses"
          value={`${totalExpenses.toFixed(2)} €`}
          icon={<Receipt className="w-6 h-6" />}
          color="primary"
          change={{ value: 12, isPositive: false }}
        />
        <StatCard
          label="Colocataires"
          value={currentColocation.members?.length || 0}
          icon={<Users className="w-6 h-6" />}
          color="accent"
        />
        <StatCard
          label="Dépenses ce mois"
          value={expenses.length}
          icon={<Calendar className="w-6 h-6" />}
          color="warning"
        />
      </motion.div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="xl:col-span-2">
          <Card className="p-6">
            <CardHeader title="Évolution des dépenses" subtitle="6 derniers mois" />
            <div className="h-72 mt-6">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={monthlyData}>
                  <defs>
                    <linearGradient id="colorAmount" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#5682F2" stopOpacity={0.15} />
                      <stop offset="95%" stopColor="#5682F2" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
                  <XAxis dataKey="month" stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                  <YAxis stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} tickFormatter={(v: number) => `${v}€`} />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#fff', border: '1px solid #e2e8f0', borderRadius: '12px', boxShadow: '0 4px 12px rgba(0,0,0,0.08)' }}
                    formatter={(value) => [`${value} €`, 'Dépenses']}
                  />
                  <Area type="monotone" dataKey="amount" stroke="#5682F2" strokeWidth={2} fillOpacity={1} fill="url(#colorAmount)" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </motion.div>

        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.3 }}>
          <Card className="p-6 h-full">
            <CardHeader title="Par catégorie" />
            <div className="h-52 mt-4">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={categoryStats.length > 0 ? categoryStats : [{ name: 'Aucune', value: 1 }]}
                    cx="50%" cy="50%"
                    innerRadius={55} outerRadius={80}
                    dataKey="total_amount" nameKey="category.name"
                    paddingAngle={2}
                  >
                    {categoryStats.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{ backgroundColor: '#fff', border: '1px solid #e2e8f0', borderRadius: '12px' }}
                    formatter={(value) => [`${Number(value).toFixed(2)} €`]}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="flex flex-wrap gap-2 mt-6">
              {categoryStats.slice(0, 4).map((stat, index) => (
                <Badge key={stat.category?.id || index} variant="default">
                  <span className="w-2.5 h-2.5 rounded-full mr-2" style={{ backgroundColor: COLORS[index % COLORS.length] }} />
                  {stat.category?.name}
                </Badge>
              ))}
            </div>
          </Card>
        </motion.div>
      </div>

      {/* Bottom Row */}
      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        {/* Recent Expenses */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.4 }}>
          <Card className="p-6">
            <CardHeader title="Dernières dépenses" action={<Button variant="ghost" size="sm" rightIcon={<ArrowRight className="w-4 h-4" />}>Voir tout</Button>} />
            <div className="space-y-3 mt-6">
              {expenses.length === 0 ? (
                <p className="text-slate-400 text-center py-12">Aucune dépense pour le moment</p>
              ) : (
                expenses.map((expense) => (
                  <div key={expense.id} className="flex items-center gap-4 p-4 rounded-xl hover:bg-slate-50 transition-colors">
                    <div
                      className="w-12 h-12 rounded-xl flex items-center justify-center shrink-0"
                      style={{
                        backgroundColor: expense.category?.color ? `${expense.category.color}15` : '#EEF2FF',
                        color: expense.category?.color || '#5682F2',
                      }}
                    >
                      <Receipt className="w-5 h-5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-slate-800 truncate">{expense.title}</p>
                      <p className="text-sm text-slate-400">Payé par {expense.payer?.prenom || 'Inconnu'}</p>
                    </div>
                    <span className="text-base font-semibold text-slate-800">{expense.amount.toFixed(2)} €</span>
                  </div>
                ))
              )}
            </div>
          </Card>
        </motion.div>

        {/* Debts */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.5 }}>
          <Card className="p-6">
            <CardHeader title="Remboursements suggérés" subtitle="Dettes simplifiées" action={<Button variant="ghost" size="sm" rightIcon={<ArrowRight className="w-4 h-4" />}>Voir tout</Button>} />
            <div className="space-y-3 mt-6">
              {debts.length === 0 ? (
                <div className="text-center py-12">
                  <div className="w-14 h-14 rounded-full bg-emerald-50 flex items-center justify-center mx-auto mb-4">
                    <TrendingUp className="w-7 h-7 text-emerald-500" />
                  </div>
                  <p className="text-slate-400 text-lg">Tous les comptes sont équilibrés !</p>
                </div>
              ) : (
                debts.map((debt, index) => (
                  <div key={index} className="flex items-center gap-4 p-4 rounded-xl bg-slate-50">
                    <Avatar name={debt.from_user ? `${debt.from_user.prenom} ${debt.from_user.nom}` : 'Utilisateur'} size="md" />
                    <div className="flex-1 flex items-center gap-3">
                      <span className="text-sm font-medium text-slate-700">{debt.from_user?.prenom || 'Utilisateur'}</span>
                      <ArrowRight className="w-4 h-4 text-slate-300" />
                      <span className="text-sm font-medium text-slate-700">{debt.to_user?.prenom || 'Utilisateur'}</span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="text-base font-semibold text-red-500">{debt.amount.toFixed(2)} €</span>
                      <Button size="sm" variant="secondary">Payer</Button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </Card>
        </motion.div>
      </div>
    </div>
  );
}
