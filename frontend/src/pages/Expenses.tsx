import { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Plus,
  Search,
  Filter,
  Receipt,
  Calendar,
  ChevronDown,
  MoreHorizontal,
  Trash2,
  Edit2,
} from 'lucide-react';
import { useColocation } from '../context/ColocationContext';
import { expenseApi, categoryApi } from '../api';
import { Card, CardHeader, Button, Input, Badge, Avatar, Modal } from '../components/ui';
import type { Expense, Category, SplitType } from '../types';

export function Expenses() {
  const { currentColocation } = useColocation();
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [showNewExpenseModal, setShowNewExpenseModal] = useState(false);

  // New expense form
  const [newExpense, setNewExpense] = useState({
    title: '',
    amount: '',
    category_id: '',
    split_type: 'equal' as SplitType,
    expense_date: new Date().toISOString().split('T')[0],
    description: '',
  });

  useEffect(() => {
    const fetchData = async () => {
      if (!currentColocation) return;

      setIsLoading(true);
      try {
        const [expensesRes, categoriesRes] = await Promise.all([
          expenseApi.list({ colocation_id: currentColocation.id }),
          categoryApi.list(currentColocation.id),
        ]);

        setExpenses(expensesRes.expenses);
        setCategories(categoriesRes);
      } catch (error) {
        console.error('Error fetching expenses:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [currentColocation]);

  const filteredExpenses = expenses.filter((expense) => {
    const matchesSearch = expense.title.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = !selectedCategory || expense.category_id === selectedCategory;
    return matchesSearch && matchesCategory;
  });

  const handleCreateExpense = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentColocation) return;

    try {
      const expense = await expenseApi.create({
        colocation_id: currentColocation.id,
        title: newExpense.title,
        amount: parseFloat(newExpense.amount),
        category_id: newExpense.category_id,
        split_type: newExpense.split_type,
        expense_date: newExpense.expense_date,
        description: newExpense.description || undefined,
      });

      setExpenses([expense, ...expenses]);
      setShowNewExpenseModal(false);
      setNewExpense({
        title: '',
        amount: '',
        category_id: '',
        split_type: 'equal',
        expense_date: new Date().toISOString().split('T')[0],
        description: '',
      });
    } catch (error) {
      console.error('Error creating expense:', error);
    }
  };

  const getSplitTypeLabel = (type: SplitType) => {
    switch (type) {
      case 'equal':
        return 'Égal';
      case 'percentage':
        return 'Pourcentage';
      case 'custom':
        return 'Personnalisé';
    }
  };

  const totalAmount = filteredExpenses.reduce((sum, e) => sum + e.amount, 0);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-display text-3xl text-[var(--color-text)]">Dépenses</h1>
          <p className="text-[var(--color-text-secondary)]">
            Gérez les dépenses de votre colocation
          </p>
        </div>
        <Button leftIcon={<Plus className="w-4 h-4" />} onClick={() => setShowNewExpenseModal(true)}>
          Nouvelle dépense
        </Button>
      </div>

      {/* Filters */}
      <Card padding="sm">
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <Input
              placeholder="Rechercher une dépense..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              leftIcon={<Search className="w-5 h-5" />}
            />
          </div>

          <div className="relative group">
            <Button variant="secondary" leftIcon={<Filter className="w-4 h-4" />}>
              Catégorie
              <ChevronDown className="w-4 h-4 ml-1" />
            </Button>
            <div className="absolute top-full right-0 mt-2 w-48 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-sm)] shadow-[var(--shadow-lg)] opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-50">
              <button
                className={`w-full text-left px-4 py-2 text-sm hover:bg-[var(--color-surface-hover)] ${
                  !selectedCategory ? 'text-[var(--color-primary)]' : 'text-[var(--color-text)]'
                }`}
                onClick={() => setSelectedCategory(null)}
              >
                Toutes les catégories
              </button>
              {categories.map((cat) => (
                <button
                  key={cat.id}
                  className={`w-full text-left px-4 py-2 text-sm hover:bg-[var(--color-surface-hover)] ${
                    selectedCategory === cat.id
                      ? 'text-[var(--color-primary)]'
                      : 'text-[var(--color-text)]'
                  }`}
                  onClick={() => setSelectedCategory(cat.id)}
                >
                  <span className="flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: cat.color }}
                    />
                    {cat.name}
                  </span>
                </button>
              ))}
            </div>
          </div>
        </div>
      </Card>

      {/* Stats bar */}
      <div className="flex items-center justify-between p-4 rounded-[var(--radius-md)] bg-[var(--color-primary-light)]">
        <div className="flex items-center gap-6">
          <div>
            <p className="text-sm text-[var(--color-primary)]">Total</p>
            <p className="text-2xl font-semibold text-[var(--color-primary)]">
              {totalAmount.toFixed(2)} €
            </p>
          </div>
          <div className="w-px h-10 bg-[var(--color-primary)]/20" />
          <div>
            <p className="text-sm text-[var(--color-primary)]">Nombre de dépenses</p>
            <p className="text-2xl font-semibold text-[var(--color-primary)]">
              {filteredExpenses.length}
            </p>
          </div>
        </div>
      </div>

      {/* Expenses List */}
      <Card padding="none">
        <div className="divide-y divide-[var(--color-border-light)]">
          <AnimatePresence>
            {filteredExpenses.length === 0 ? (
              <div className="p-12 text-center">
                <div className="w-16 h-16 rounded-full bg-[var(--color-surface-hover)] flex items-center justify-center mx-auto mb-4">
                  <Receipt className="w-8 h-8 text-[var(--color-text-muted)]" />
                </div>
                <p className="text-[var(--color-text-secondary)]">Aucune dépense trouvée</p>
              </div>
            ) : (
              filteredExpenses.map((expense, index) => (
                <motion.div
                  key={expense.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  transition={{ delay: index * 0.05 }}
                  className="flex items-center gap-4 p-4 hover:bg-[var(--color-surface-hover)] transition-colors"
                >
                  {/* Category icon */}
                  <div
                    className="w-12 h-12 rounded-[var(--radius-sm)] flex items-center justify-center flex-shrink-0"
                    style={{
                      backgroundColor: expense.category?.color
                        ? `${expense.category.color}15`
                        : 'var(--color-surface-hover)',
                    }}
                  >
                    <span className="text-xl">{expense.category?.icon || '💰'}</span>
                  </div>

                  {/* Info */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-[var(--color-text)] truncate">
                        {expense.title}
                      </p>
                      <Badge variant="default" size="sm">
                        {getSplitTypeLabel(expense.split_type)}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-3 mt-1 text-sm text-[var(--color-text-muted)]">
                      <span className="flex items-center gap-1">
                        <Avatar
                          name={
                            expense.payer
                              ? `${expense.payer.prenom} ${expense.payer.nom}`
                              : 'Utilisateur'
                          }
                          size="sm"
                          className="w-5 h-5"
                        />
                        {expense.payer?.prenom || 'Inconnu'}
                      </span>
                      <span className="flex items-center gap-1">
                        <Calendar className="w-4 h-4" />
                        {new Date(expense.expense_date).toLocaleDateString('fr-FR')}
                      </span>
                    </div>
                  </div>

                  {/* Amount */}
                  <div className="text-right">
                    <p className="text-lg font-semibold text-[var(--color-text)]">
                      {expense.amount.toFixed(2)} €
                    </p>
                    {expense.splits && expense.splits.length > 0 && (
                      <p className="text-sm text-[var(--color-text-muted)]">
                        {(expense.amount / expense.splits.length).toFixed(2)} €/pers
                      </p>
                    )}
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-1">
                    <button className="p-2 rounded-[var(--radius-sm)] text-[var(--color-text-muted)] hover:text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] transition-colors">
                      <Edit2 className="w-4 h-4" />
                    </button>
                    <button className="p-2 rounded-[var(--radius-sm)] text-[var(--color-text-muted)] hover:text-[var(--color-danger)] hover:bg-[var(--color-danger-light)] transition-colors">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </motion.div>
              ))
            )}
          </AnimatePresence>
        </div>
      </Card>

      {/* New Expense Modal */}
      <Modal
        isOpen={showNewExpenseModal}
        onClose={() => setShowNewExpenseModal(false)}
        title="Nouvelle dépense"
        size="lg"
      >
        <form onSubmit={handleCreateExpense} className="space-y-4">
          <Input
            label="Titre"
            placeholder="Ex: Courses Carrefour"
            value={newExpense.title}
            onChange={(e) => setNewExpense({ ...newExpense, title: e.target.value })}
            required
          />

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Montant (€)"
              type="number"
              step="0.01"
              placeholder="0.00"
              value={newExpense.amount}
              onChange={(e) => setNewExpense({ ...newExpense, amount: e.target.value })}
              required
            />

            <div>
              <label className="block text-sm font-medium text-[var(--color-text)] mb-1.5">
                Catégorie
              </label>
              <select
                className="w-full px-4 py-2.5 rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-surface)] text-[var(--color-text)] focus:outline-none focus:ring-2 focus:ring-[var(--color-primary-light)] focus:border-[var(--color-primary)]"
                value={newExpense.category_id}
                onChange={(e) => setNewExpense({ ...newExpense, category_id: e.target.value })}
                required
              >
                <option value="">Sélectionner...</option>
                {categories.map((cat) => (
                  <option key={cat.id} value={cat.id}>
                    {cat.icon} {cat.name}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Date"
              type="date"
              value={newExpense.expense_date}
              onChange={(e) => setNewExpense({ ...newExpense, expense_date: e.target.value })}
              required
            />

            <div>
              <label className="block text-sm font-medium text-[var(--color-text)] mb-1.5">
                Mode de partage
              </label>
              <select
                className="w-full px-4 py-2.5 rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-surface)] text-[var(--color-text)] focus:outline-none focus:ring-2 focus:ring-[var(--color-primary-light)] focus:border-[var(--color-primary)]"
                value={newExpense.split_type}
                onChange={(e) =>
                  setNewExpense({ ...newExpense, split_type: e.target.value as SplitType })
                }
              >
                <option value="equal">Égal entre tous</option>
                <option value="percentage">Par pourcentage</option>
                <option value="custom">Montants personnalisés</option>
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text)] mb-1.5">
              Description (optionnel)
            </label>
            <textarea
              className="w-full px-4 py-2.5 rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-surface)] text-[var(--color-text)] focus:outline-none focus:ring-2 focus:ring-[var(--color-primary-light)] focus:border-[var(--color-primary)] resize-none"
              rows={3}
              placeholder="Ajouter une description..."
              value={newExpense.description}
              onChange={(e) => setNewExpense({ ...newExpense, description: e.target.value })}
            />
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <Button type="button" variant="secondary" onClick={() => setShowNewExpenseModal(false)}>
              Annuler
            </Button>
            <Button type="submit">Créer la dépense</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
