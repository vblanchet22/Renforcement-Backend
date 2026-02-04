import api from './client';
import type { Expense, RecurringExpense, SplitType, Recurrence, MonthlyForecast } from '../types';

interface SplitInput {
  user_id: string;
  amount?: number;
  percentage?: number;
}

interface CreateExpenseRequest {
  colocation_id: string;
  category_id: string;
  title: string;
  description?: string;
  amount: number;
  split_type: SplitType;
  expense_date: string;
  splits?: SplitInput[];
}

interface UpdateExpenseRequest {
  category_id?: string;
  title?: string;
  description?: string;
  amount?: number;
  split_type?: SplitType;
  expense_date?: string;
  splits?: SplitInput[];
}

interface ListExpensesParams {
  colocation_id: string;
  category_id?: string;
  paid_by?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  per_page?: number;
}

interface CreateRecurringExpenseRequest {
  colocation_id: string;
  category_id: string;
  title: string;
  description?: string;
  amount: number;
  split_type: SplitType;
  recurrence: Recurrence;
  start_date: string;
  end_date?: string;
  splits?: SplitInput[];
}

export const expenseApi = {
  async list(params: ListExpensesParams): Promise<{ expenses: Expense[]; total: number }> {
    const response = await api.get<{ expenses: Expense[]; total: number }>('/v1/expenses', {
      params,
    });
    return { expenses: response.data.expenses || [], total: response.data.total || 0 };
  },

  async get(id: string): Promise<Expense> {
    const response = await api.get<Expense>(`/v1/expenses/${id}`);
    return response.data;
  },

  async create(data: CreateExpenseRequest): Promise<Expense> {
    const response = await api.post<Expense>('/v1/expenses', data);
    return response.data;
  },

  async update(id: string, data: UpdateExpenseRequest): Promise<Expense> {
    const response = await api.put<Expense>(`/v1/expenses/${id}`, data);
    return response.data;
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/v1/expenses/${id}`);
  },

  // Recurring expenses
  async listRecurring(colocationId: string): Promise<RecurringExpense[]> {
    const response = await api.get<{ recurring_expenses: RecurringExpense[] }>(
      '/v1/recurring-expenses',
      { params: { colocation_id: colocationId } }
    );
    return response.data.recurring_expenses || [];
  },

  async createRecurring(data: CreateRecurringExpenseRequest): Promise<RecurringExpense> {
    const response = await api.post<RecurringExpense>('/v1/recurring-expenses', data);
    return response.data;
  },

  async updateRecurring(
    id: string,
    data: Partial<CreateRecurringExpenseRequest>
  ): Promise<RecurringExpense> {
    const response = await api.put<RecurringExpense>(`/v1/recurring-expenses/${id}`, data);
    return response.data;
  },

  async deleteRecurring(id: string): Promise<void> {
    await api.delete(`/v1/recurring-expenses/${id}`);
  },

  // Forecast
  async getForecast(colocationId: string, months: number = 3): Promise<MonthlyForecast[]> {
    const response = await api.get<{ forecasts: MonthlyForecast[] }>('/v1/expenses/forecast', {
      params: { colocation_id: colocationId, months },
    });
    return response.data.forecasts || [];
  },
};
