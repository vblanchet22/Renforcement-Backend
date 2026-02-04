import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import {
  TrendingUp,
  TrendingDown,
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
import { expenseApi, balanceApi, categoryApi } from '../api';
import { Card, CardHeader, StatCard, Avatar, Badge, Button } from '../components/ui';
import type { Expense, UserBalance, CategoryStat, SimplifiedDebt } from '../types';

export function Dashboard() {
  const { user } = useAuth();
  const { currentColocation } = useColocation();
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [balances, setBalances] = useState<UserBalance[]>([]);
  const [categoryStats, setCategoryStats] = useState<CategoryStat[]>([]);
  const [debts, setDebts] = useState<SimplifiedDebt[]>([]);
  const [isLoading, setIsLoading] = useState(true);

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

  // Sample chart data
  const monthlyData = [
    { month: 'Jan', amount: 450 },
    { month: 'Fév', amount: 520 },
    { month: 'Mar', amount: 380 },
    { month: 'Avr', amount: 610 },
    { month: 'Mai', amount: 490 },
    { month: 'Juin', amount: 430 },
  ];

  const COLORS = ['#5682F2', '#F1C086', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6'];

  if (!currentColocation) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
        <div className="w-20 h-20 rounded-full bg-[var(--color-primary-light)] flex items-center justify-center mb-6">
          <Users className="w-10 h-10 text-[var(--color-primary)]" />
        </div>
        <h2 className="text-display text-2xl text-[var(--color-text)] mb-2">
          Aucune colocation
        </h2>
        <p className="text-[var(--color-text-secondary)] mb-6 max-w-md">
          Créez ou rejoignez une colocation pour commencer à gérer vos dépenses partagées.
        </p>
        <div className="flex gap-3">
          <Button leftIcon={<Plus className="w-4 h-4" />}>Créer une colocation</Button>
          <Button variant="secondary">Rejoindre avec un code</Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <motion.h1
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-display text-3xl text-[var(--color-text)] mb-1"
          >
            Bonjour, {user?.prenom} !
          </motion.h1>
          <p className="text-[var(--color-text-secondary)]">
            Voici un aperçu de votre colocation {currentColocation.name}
          </p>
        </div>
        <Button leftIcon={<Plus className="w-4 h-4" />}>Nouvelle dépense</Button>
      </div>

      {/* Stats Grid */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
      >
        <StatCard
          label="Votre solde"
          value={`${(userBalance?.net_balance || 0) >= 0 ? '+' : ''}${(userBalance?.net_balance || 0).toFixed(2)} €`}
          icon={<Wallet className="w-5 h-5" />}
          color={(userBalance?.net_balance || 0) >= 0 ? 'success' : 'danger'}
        />
        <StatCard
          label="Total des dépenses"
          value={`${totalExpenses.toFixed(2)} €`}
          icon={<Receipt className="w-5 h-5" />}
          color="primary"
          change={{ value: 12, isPositive: false }}
        />
        <StatCard
          label="Colocataires"
          value={currentColocation.members?.length || 0}
          icon={<Users className="w-5 h-5" />}
          color="accent"
        />
        <StatCard
          label="Dépenses ce mois"
          value={expenses.length}
          icon={<Calendar className="w-5 h-5" />}
          color="warning"
        />
      </motion.div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Area Chart */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="lg:col-span-2"
        >
          <Card>
            <CardHeader title="Évolution des dépenses" subtitle="6 derniers mois" />
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={monthlyData}>
                  <defs>
                    <linearGradient id="colorAmount" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="var(--color-primary)" stopOpacity={0.2} />
                      <stop offset="95%" stopColor="var(--color-primary)" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border-light)" />
                  <XAxis
                    dataKey="month"
                    stroke="var(--color-text-muted)"
                    fontSize={12}
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    stroke="var(--color-text-muted)"
                    fontSize={12}
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={(v) => `${v}€`}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'var(--color-surface)',
                      border: '1px solid var(--color-border)',
                      borderRadius: 'var(--radius-sm)',
                    }}
                    formatter={(value) => [`${value} €`, 'Dépenses']}
                  />
                  <Area
                    type="monotone"
                    dataKey="amount"
                    stroke="var(--color-primary)"
                    strokeWidth={2}
                    fillOpacity={1}
                    fill="url(#colorAmount)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </motion.div>

        {/* Pie Chart */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
        >
          <Card>
            <CardHeader title="Par catégorie" />
            <div className="h-48">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={categoryStats.length > 0 ? categoryStats : [{ name: 'Aucune', value: 1 }]}
                    cx="50%"
                    cy="50%"
                    innerRadius={50}
                    outerRadius={70}
                    dataKey="total_amount"
                    nameKey="category.name"
                    paddingAngle={2}
                  >
                    {categoryStats.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'var(--color-surface)',
                      border: '1px solid var(--color-border)',
                      borderRadius: 'var(--radius-sm)',
                    }}
                    formatter={(value) => [`${Number(value).toFixed(2)} €`]}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="flex flex-wrap gap-2 mt-4">
              {categoryStats.slice(0, 4).map((stat, index) => (
                <Badge key={stat.category?.id || index} variant="default">
                  <span
                    className="w-2 h-2 rounded-full mr-1"
                    style={{ backgroundColor: COLORS[index % COLORS.length] }}
                  />
                  {stat.category?.name}
                </Badge>
              ))}
            </div>
          </Card>
        </motion.div>
      </div>

      {/* Bottom Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Expenses */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
        >
          <Card>
            <CardHeader
              title="Dernières dépenses"
              action={
                <Button variant="ghost" size="sm" rightIcon={<ArrowRight className="w-4 h-4" />}>
                  Voir tout
                </Button>
              }
            />
            <div className="space-y-4">
              {expenses.length === 0 ? (
                <p className="text-[var(--color-text-muted)] text-center py-8">
                  Aucune dépense pour le moment
                </p>
              ) : (
                expenses.map((expense) => (
                  <div
                    key={expense.id}
                    className="flex items-center gap-4 p-3 rounded-[var(--radius-sm)] hover:bg-[var(--color-surface-hover)] transition-colors"
                  >
                    <div
                      className="w-10 h-10 rounded-[var(--radius-sm)] flex items-center justify-center"
                      style={{
                        backgroundColor: expense.category?.color
                          ? `${expense.category.color}20`
                          : 'var(--color-primary-light)',
                        color: expense.category?.color || 'var(--color-primary)',
                      }}
                    >
                      <Receipt className="w-5 h-5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-[var(--color-text)] truncate">
                        {expense.title}
                      </p>
                      <p className="text-xs text-[var(--color-text-muted)]">
                        Payé par {expense.payer?.prenom || 'Inconnu'}
                      </p>
                    </div>
                    <span className="text-sm font-semibold text-[var(--color-text)]">
                      {expense.amount.toFixed(2)} €
                    </span>
                  </div>
                ))
              )}
            </div>
          </Card>
        </motion.div>

        {/* Debts to settle */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
        >
          <Card>
            <CardHeader
              title="Remboursements suggérés"
              subtitle="Dettes simplifiées"
              action={
                <Button variant="ghost" size="sm" rightIcon={<ArrowRight className="w-4 h-4" />}>
                  Voir tout
                </Button>
              }
            />
            <div className="space-y-4">
              {debts.length === 0 ? (
                <div className="text-center py-8">
                  <div className="w-12 h-12 rounded-full bg-[var(--color-success-light)] flex items-center justify-center mx-auto mb-3">
                    <TrendingUp className="w-6 h-6 text-[var(--color-success)]" />
                  </div>
                  <p className="text-[var(--color-text-muted)]">Tous les comptes sont équilibrés !</p>
                </div>
              ) : (
                debts.map((debt, index) => (
                  <div
                    key={index}
                    className="flex items-center gap-4 p-3 rounded-[var(--radius-sm)] bg-[var(--color-bg)]"
                  >
                    <Avatar
                      name={
                        debt.from_user
                          ? `${debt.from_user.prenom} ${debt.from_user.nom}`
                          : 'Utilisateur'
                      }
                      size="sm"
                    />
                    <div className="flex-1 flex items-center gap-2">
                      <span className="text-sm text-[var(--color-text)]">
                        {debt.from_user?.prenom || 'Utilisateur'}
                      </span>
                      <ArrowRight className="w-4 h-4 text-[var(--color-text-muted)]" />
                      <span className="text-sm text-[var(--color-text)]">
                        {debt.to_user?.prenom || 'Utilisateur'}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-semibold text-[var(--color-danger)]">
                        {debt.amount.toFixed(2)} €
                      </span>
                      <Button size="sm" variant="secondary">
                        Payer
                      </Button>
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
