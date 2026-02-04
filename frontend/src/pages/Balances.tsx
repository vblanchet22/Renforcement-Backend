import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import {
  Wallet,
  TrendingUp,
  TrendingDown,
  ArrowRight,
  RefreshCcw,
  ChevronRight,
} from 'lucide-react';
import { useColocation } from '../context/ColocationContext';
import { useAuth } from '../context/AuthContext';
import { balanceApi } from '../api';
import { Card, CardHeader, Button, Avatar, Badge } from '../components/ui';
import type { UserBalance, SimplifiedDebt } from '../types';

export function Balances() {
  const { user } = useAuth();
  const { currentColocation } = useColocation();
  const [balances, setBalances] = useState<UserBalance[]>([]);
  const [debts, setDebts] = useState<SimplifiedDebt[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      if (!currentColocation) return;

      setIsLoading(true);
      try {
        const [balancesRes, debtsRes] = await Promise.all([
          balanceApi.getBalances(currentColocation.id),
          balanceApi.getSimplifiedDebts(currentColocation.id),
        ]);

        setBalances(balancesRes);
        setDebts(debtsRes);
      } catch (error) {
        console.error('Error fetching balances:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [currentColocation]);

  const positiveBalances = balances.filter((b) => b.net_balance > 0);
  const negativeBalances = balances.filter((b) => b.net_balance < 0);
  const userBalance = balances.find((b) => b.user_id === user?.id);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-display text-3xl text-[var(--color-text)]">Soldes</h1>
          <p className="text-[var(--color-text-secondary)]">
            Vue d'ensemble des soldes de la colocation
          </p>
        </div>
        <Button variant="secondary" leftIcon={<RefreshCcw className="w-4 h-4" />}>
          Actualiser
        </Button>
      </div>

      {/* Your Balance Card */}
      {userBalance && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <Card
            className={`
              relative overflow-hidden
              ${userBalance.net_balance >= 0 ? 'bg-gradient-to-br from-emerald-50 to-teal-50' : 'bg-gradient-to-br from-red-50 to-orange-50'}
            `}
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)] mb-1">Votre solde</p>
                <p
                  className={`text-4xl font-semibold ${
                    userBalance.net_balance >= 0
                      ? 'text-[var(--color-success)]'
                      : 'text-[var(--color-danger)]'
                  }`}
                >
                  {userBalance.net_balance >= 0 ? '+' : ''}
                  {userBalance.net_balance.toFixed(2)} €
                </p>
                <p className="text-sm text-[var(--color-text-muted)] mt-2">
                  {userBalance.net_balance >= 0
                    ? 'On vous doit de l\'argent'
                    : 'Vous devez de l\'argent'}
                </p>
              </div>
              <div
                className={`w-16 h-16 rounded-full flex items-center justify-center ${
                  userBalance.net_balance >= 0
                    ? 'bg-[var(--color-success)]/10'
                    : 'bg-[var(--color-danger)]/10'
                }`}
              >
                {userBalance.net_balance >= 0 ? (
                  <TrendingUp
                    className={`w-8 h-8 ${
                      userBalance.net_balance >= 0
                        ? 'text-[var(--color-success)]'
                        : 'text-[var(--color-danger)]'
                    }`}
                  />
                ) : (
                  <TrendingDown className="w-8 h-8 text-[var(--color-danger)]" />
                )}
              </div>
            </div>

            <div className="grid grid-cols-2 gap-6 mt-6 pt-6 border-t border-black/5">
              <div>
                <p className="text-sm text-[var(--color-text-muted)]">Total payé</p>
                <p className="text-xl font-semibold text-[var(--color-text)]">
                  {userBalance.total_paid.toFixed(2)} €
                </p>
              </div>
              <div>
                <p className="text-sm text-[var(--color-text-muted)]">Total dû</p>
                <p className="text-xl font-semibold text-[var(--color-text)]">
                  {userBalance.total_owed.toFixed(2)} €
                </p>
              </div>
            </div>
          </Card>
        </motion.div>
      )}

      {/* Simplified Debts */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
      >
        <Card>
          <CardHeader
            title="Remboursements optimisés"
            subtitle="Algorithme de simplification des dettes"
          />

          {debts.length === 0 ? (
            <div className="text-center py-12">
              <div className="w-16 h-16 rounded-full bg-[var(--color-success-light)] flex items-center justify-center mx-auto mb-4">
                <TrendingUp className="w-8 h-8 text-[var(--color-success)]" />
              </div>
              <p className="text-lg font-medium text-[var(--color-text)]">Tout est équilibré !</p>
              <p className="text-[var(--color-text-muted)] mt-1">
                Aucun remboursement nécessaire pour le moment
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {debts.map((debt, index) => {
                const isUserDebtor = debt.from_user_id === user?.id;
                const isUserCreditor = debt.to_user_id === user?.id;

                return (
                  <motion.div
                    key={index}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.1 }}
                    className={`
                      flex items-center gap-4 p-4 rounded-[var(--radius-md)]
                      ${isUserDebtor ? 'bg-[var(--color-danger-light)]/50' : ''}
                      ${isUserCreditor ? 'bg-[var(--color-success-light)]/50' : ''}
                      ${!isUserDebtor && !isUserCreditor ? 'bg-[var(--color-bg)]' : ''}
                    `}
                  >
                    <Avatar
                      name={
                        debt.from_user
                          ? `${debt.from_user.prenom} ${debt.from_user.nom}`
                          : 'Utilisateur'
                      }
                      size="md"
                    />

                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-[var(--color-text)]">
                          {debt.from_user?.prenom || 'Utilisateur'}
                          {isUserDebtor && (
                            <Badge variant="danger" size="sm" className="ml-2">
                              Vous
                            </Badge>
                          )}
                        </span>
                        <ArrowRight className="w-4 h-4 text-[var(--color-text-muted)]" />
                        <span className="font-medium text-[var(--color-text)]">
                          {debt.to_user?.prenom || 'Utilisateur'}
                          {isUserCreditor && (
                            <Badge variant="success" size="sm" className="ml-2">
                              Vous
                            </Badge>
                          )}
                        </span>
                      </div>
                      <p className="text-sm text-[var(--color-text-muted)]">
                        {isUserDebtor
                          ? 'Vous devez rembourser'
                          : isUserCreditor
                            ? 'Vous allez recevoir'
                            : 'Transaction entre colocataires'}
                      </p>
                    </div>

                    <div className="text-right">
                      <p className="text-xl font-semibold text-[var(--color-text)]">
                        {debt.amount.toFixed(2)} €
                      </p>
                    </div>

                    {isUserDebtor && (
                      <Button size="sm">
                        Rembourser
                        <ChevronRight className="w-4 h-4" />
                      </Button>
                    )}
                  </motion.div>
                );
              })}
            </div>
          )}
        </Card>
      </motion.div>

      {/* All Balances */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Creditors */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
        >
          <Card>
            <CardHeader
              title="Créanciers"
              subtitle="Membres avec un solde positif"
              action={
                <div className="flex items-center gap-1 text-[var(--color-success)]">
                  <TrendingUp className="w-4 h-4" />
                  <span className="text-sm font-medium">
                    {positiveBalances.reduce((sum, b) => sum + b.net_balance, 0).toFixed(2)} €
                  </span>
                </div>
              }
            />

            {positiveBalances.length === 0 ? (
              <p className="text-[var(--color-text-muted)] text-center py-6">
                Aucun créancier
              </p>
            ) : (
              <div className="space-y-3">
                {positiveBalances.map((balance) => (
                  <div
                    key={balance.user_id}
                    className="flex items-center gap-3 p-3 rounded-[var(--radius-sm)] bg-[var(--color-bg)]"
                  >
                    <Avatar
                      name={
                        balance.user
                          ? `${balance.user.prenom} ${balance.user.nom}`
                          : 'Utilisateur'
                      }
                      size="md"
                    />
                    <div className="flex-1">
                      <p className="font-medium text-[var(--color-text)]">
                        {balance.user?.prenom} {balance.user?.nom}
                        {balance.user_id === user?.id && (
                          <Badge variant="primary" size="sm" className="ml-2">
                            Vous
                          </Badge>
                        )}
                      </p>
                      <p className="text-sm text-[var(--color-text-muted)]">
                        A payé {balance.total_paid.toFixed(2)} €
                      </p>
                    </div>
                    <p className="text-lg font-semibold text-[var(--color-success)]">
                      +{balance.net_balance.toFixed(2)} €
                    </p>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </motion.div>

        {/* Debtors */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
        >
          <Card>
            <CardHeader
              title="Débiteurs"
              subtitle="Membres avec un solde négatif"
              action={
                <div className="flex items-center gap-1 text-[var(--color-danger)]">
                  <TrendingDown className="w-4 h-4" />
                  <span className="text-sm font-medium">
                    {negativeBalances.reduce((sum, b) => sum + b.net_balance, 0).toFixed(2)} €
                  </span>
                </div>
              }
            />

            {negativeBalances.length === 0 ? (
              <p className="text-[var(--color-text-muted)] text-center py-6">
                Aucun débiteur
              </p>
            ) : (
              <div className="space-y-3">
                {negativeBalances.map((balance) => (
                  <div
                    key={balance.user_id}
                    className="flex items-center gap-3 p-3 rounded-[var(--radius-sm)] bg-[var(--color-bg)]"
                  >
                    <Avatar
                      name={
                        balance.user
                          ? `${balance.user.prenom} ${balance.user.nom}`
                          : 'Utilisateur'
                      }
                      size="md"
                    />
                    <div className="flex-1">
                      <p className="font-medium text-[var(--color-text)]">
                        {balance.user?.prenom} {balance.user?.nom}
                        {balance.user_id === user?.id && (
                          <Badge variant="primary" size="sm" className="ml-2">
                            Vous
                          </Badge>
                        )}
                      </p>
                      <p className="text-sm text-[var(--color-text-muted)]">
                        Doit {balance.total_owed.toFixed(2)} €
                      </p>
                    </div>
                    <p className="text-lg font-semibold text-[var(--color-danger)]">
                      {balance.net_balance.toFixed(2)} €
                    </p>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </motion.div>
      </div>
    </div>
  );
}
