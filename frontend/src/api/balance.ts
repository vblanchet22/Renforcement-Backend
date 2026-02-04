import api from './client';
import type { UserBalance, SimplifiedDebt } from '../types';

interface BalanceHistoryEntry {
  date: string;
  description: string;
  amount: number;
  type: 'expense_paid' | 'expense_owed' | 'payment_made' | 'payment_received';
  running_balance: number;
}

export const balanceApi = {
  async getBalances(colocationId: string): Promise<UserBalance[]> {
    const response = await api.get<{ balances: UserBalance[] }>(
      `/colocations/${colocationId}/balances`
    );
    return response.data.balances || [];
  },

  async getSimplifiedDebts(colocationId: string): Promise<SimplifiedDebt[]> {
    const response = await api.get<{ debts: SimplifiedDebt[] }>(
      `/colocations/${colocationId}/balances/simplified`
    );
    return response.data.debts || [];
  },

  async getHistory(colocationId: string, userId?: string): Promise<BalanceHistoryEntry[]> {
    const response = await api.get<{ history: BalanceHistoryEntry[] }>(
      `/colocations/${colocationId}/balances/history`,
      { params: userId ? { user_id: userId } : undefined }
    );
    return response.data.history || [];
  },
};
