import { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Plus,
  Search,
  Filter,
  Receipt,
  Calendar,
  ChevronDown,
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
      setNewExpense({ title: '', amount: '', category_id: '', split_type: 'equal', expense_date: new Date().toISOString().split('T')[0], description: '' });
    } catch (error) {
      console.error('Error creating expense:', error);
    }
  };

  const getSplitTypeLabel = (type: SplitType) => {
    switch (type) {
      case 'equal': return 'Égal';
      case 'percentage': return 'Pourcentage';
      case 'custom': return 'Personnalisé';
    }
  };

  const totalAmount = filteredExpenses.reduce((sum, e) => sum + e.amount, 0);

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-semibold text-slate-800">Dépenses</h1>
          <p className="text-slate-500 text-lg mt-1">Gérez les dépenses de votre colocation</p>
        </div>
        <Button size="lg" leftIcon={<Plus className="w-5 h-5" />} onClick={() => setShowNewExpenseModal(true)}>
          Nouvelle dépense
        </Button>
      </div>

      {/* Filters */}
      <Card className="p-5">
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <Input placeholder="Rechercher une dépense..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} leftIcon={<Search className="w-5 h-5" />} />
          </div>
          <div className="relative group">
            <Button variant="secondary" leftIcon={<Filter className="w-4 h-4" />}>
              Catégorie <ChevronDown className="w-4 h-4 ml-2" />
            </Button>
            <div className="absolute top-full right-0 mt-2 w-56 bg-white border border-slate-200 rounded-xl shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-50 py-2">
              <button className={`w-full text-left px-4 py-3 text-sm hover:bg-slate-50 ${!selectedCategory ? 'text-primary font-medium' : 'text-slate-700'}`} onClick={() => setSelectedCategory(null)}>
                Toutes les catégories
              </button>
              {categories.map((cat) => (
                <button key={cat.id} className={`w-full text-left px-4 py-3 text-sm hover:bg-slate-50 ${selectedCategory === cat.id ? 'text-primary font-medium' : 'text-slate-700'}`} onClick={() => setSelectedCategory(cat.id)}>
                  <span className="flex items-center gap-3">
                    <span className="w-3 h-3 rounded-full" style={{ backgroundColor: cat.color }} />
                    {cat.name}
                  </span>
                </button>
              ))}
            </div>
          </div>
        </div>
      </Card>

      {/* Stats bar */}
      <div className="flex items-center gap-10 px-6 py-5 rounded-2xl bg-primary/5 border border-primary/10">
        <div>
          <p className="text-sm text-primary font-medium mb-1">Total</p>
          <p className="text-3xl font-semibold text-primary">{totalAmount.toFixed(2)} €</p>
        </div>
        <div className="w-px h-12 bg-primary/20" />
        <div>
          <p className="text-sm text-primary font-medium mb-1">Nombre de dépenses</p>
          <p className="text-3xl font-semibold text-primary">{filteredExpenses.length}</p>
        </div>
      </div>

      {/* Expenses List */}
      <Card className="p-0 overflow-hidden">
        <div className="divide-y divide-slate-100">
          <AnimatePresence>
            {filteredExpenses.length === 0 ? (
              <div className="p-16 text-center">
                <div className="w-20 h-20 rounded-2xl bg-slate-50 flex items-center justify-center mx-auto mb-6">
                  <Receipt className="w-10 h-10 text-slate-300" />
                </div>
                <p className="text-slate-400 text-lg">Aucune dépense trouvée</p>
              </div>
            ) : (
              filteredExpenses.map((expense, index) => (
                <motion.div
                  key={expense.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  transition={{ delay: index * 0.03 }}
                  className="flex items-center gap-5 px-6 py-5 hover:bg-slate-50/50 transition-colors"
                >
                  <div className="w-14 h-14 rounded-xl flex items-center justify-center shrink-0" style={{ backgroundColor: expense.category?.color ? `${expense.category.color}12` : '#f1f5f9' }}>
                    <span className="text-xl">{expense.category?.icon || '💰'}</span>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3">
                      <p className="font-medium text-slate-800 truncate text-base">{expense.title}</p>
                      <Badge variant="default" size="sm">{getSplitTypeLabel(expense.split_type)}</Badge>
                    </div>
                    <div className="flex items-center gap-4 mt-1.5 text-sm text-slate-400">
                      <span className="flex items-center gap-2">
                        <Avatar name={expense.payer ? `${expense.payer.prenom} ${expense.payer.nom}` : 'Utilisateur'} size="sm" className="w-5 h-5 text-[9px]" />
                        {expense.payer?.prenom || 'Inconnu'}
                      </span>
                      <span className="flex items-center gap-1.5">
                        <Calendar className="w-4 h-4" />
                        {new Date(expense.expense_date).toLocaleDateString('fr-FR')}
                      </span>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-lg font-semibold text-slate-800">{expense.amount.toFixed(2)} €</p>
                    {expense.splits && expense.splits.length > 0 && (
                      <p className="text-sm text-slate-400 mt-0.5">{(expense.amount / expense.splits.length).toFixed(2)} €/pers</p>
                    )}
                  </div>
                  <div className="flex items-center gap-1">
                    <button className="p-3 rounded-xl text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors">
                      <Edit2 className="w-5 h-5" />
                    </button>
                    <button className="p-3 rounded-xl text-slate-400 hover:text-red-500 hover:bg-red-50 transition-colors">
                      <Trash2 className="w-5 h-5" />
                    </button>
                  </div>
                </motion.div>
              ))
            )}
          </AnimatePresence>
        </div>
      </Card>

      {/* New Expense Modal */}
      <Modal isOpen={showNewExpenseModal} onClose={() => setShowNewExpenseModal(false)} title="Nouvelle dépense" size="lg">
        <form onSubmit={handleCreateExpense} className="space-y-5">
          <Input label="Titre" placeholder="Ex: Courses Carrefour" value={newExpense.title} onChange={(e) => setNewExpense({ ...newExpense, title: e.target.value })} required />
          <div className="grid grid-cols-2 gap-5">
            <Input label="Montant (€)" type="number" step="0.01" placeholder="0.00" value={newExpense.amount} onChange={(e) => setNewExpense({ ...newExpense, amount: e.target.value })} required />
            <div className="space-y-2">
              <label className="block text-sm font-medium text-slate-700">Catégorie</label>
              <select className="w-full h-12 px-4 rounded-xl border border-slate-200 bg-white text-sm text-slate-800 outline-none focus:border-primary focus:ring-2 focus:ring-primary/20" value={newExpense.category_id} onChange={(e) => setNewExpense({ ...newExpense, category_id: e.target.value })} required>
                <option value="">Sélectionner...</option>
                {categories.map((cat) => (<option key={cat.id} value={cat.id}>{cat.icon} {cat.name}</option>))}
              </select>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-5">
            <Input label="Date" type="date" value={newExpense.expense_date} onChange={(e) => setNewExpense({ ...newExpense, expense_date: e.target.value })} required />
            <div className="space-y-2">
              <label className="block text-sm font-medium text-slate-700">Mode de partage</label>
              <select className="w-full h-12 px-4 rounded-xl border border-slate-200 bg-white text-sm text-slate-800 outline-none focus:border-primary focus:ring-2 focus:ring-primary/20" value={newExpense.split_type} onChange={(e) => setNewExpense({ ...newExpense, split_type: e.target.value as SplitType })}>
                <option value="equal">Égal entre tous</option>
                <option value="percentage">Par pourcentage</option>
                <option value="custom">Montants personnalisés</option>
              </select>
            </div>
          </div>
          <div className="space-y-2">
            <label className="block text-sm font-medium text-slate-700">Description (optionnel)</label>
            <textarea className="w-full px-4 py-3 rounded-xl border border-slate-200 bg-white text-sm text-slate-800 outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 resize-none" rows={3} placeholder="Ajouter une description..." value={newExpense.description} onChange={(e) => setNewExpense({ ...newExpense, description: e.target.value })} />
          </div>
          <div className="flex justify-end gap-4 pt-4">
            <Button type="button" variant="secondary" onClick={() => setShowNewExpenseModal(false)}>Annuler</Button>
            <Button type="submit">Créer la dépense</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
