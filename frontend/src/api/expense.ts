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
    const { colocation_id, ...queryParams } = params;
    const response = await api.get<{ expenses: Expense[]; total: number }>(
      `/colocations/${colocation_id}/expenses`,
      { params: queryParams }
    );
    return { expenses: response.data.expenses || [], total: response.data.total || 0 };
  },

  async get(colocationId: string, id: string): Promise<Expense> {
    const response = await api.get<Expense>(`/colocations/${colocationId}/expenses/${id}`);
    return response.data;
  },

  async create(data: CreateExpenseRequest): Promise<Expense> {
    const { colocation_id, ...body } = data;
    const response = await api.post<Expense>(`/colocations/${colocation_id}/expenses`, body);
    return response.data;
  },

  async update(colocationId: string, id: string, data: UpdateExpenseRequest): Promise<Expense> {
    const response = await api.put<Expense>(`/colocations/${colocationId}/expenses/${id}`, data);
    return response.data;
  },

  async delete(colocationId: string, id: string): Promise<void> {
    await api.delete(`/colocations/${colocationId}/expenses/${id}`);
  },

  // Recurring expenses
  async listRecurring(colocationId: string): Promise<RecurringExpense[]> {
    const response = await api.get<{ recurring_expenses: RecurringExpense[] }>(
      `/colocations/${colocationId}/recurring-expenses`
    );
    return response.data.recurring_expenses || [];
  },

  async createRecurring(data: CreateRecurringExpenseRequest): Promise<RecurringExpense> {
    const { colocation_id, ...body } = data;
    const response = await api.post<RecurringExpense>(
      `/colocations/${colocation_id}/recurring-expenses`,
      body
    );
    return response.data;
  },

  async updateRecurring(
    colocationId: string,
    id: string,
    data: Partial<CreateRecurringExpenseRequest>
  ): Promise<RecurringExpense> {
    const response = await api.put<RecurringExpense>(
      `/colocations/${colocationId}/recurring-expenses/${id}`,
      data
    );
    return response.data;
  },

  async deleteRecurring(colocationId: string, id: string): Promise<void> {
    await api.delete(`/colocations/${colocationId}/recurring-expenses/${id}`);
  },

  // Forecast
  async getForecast(colocationId: string, months: number = 3): Promise<MonthlyForecast[]> {
    const response = await api.get<{ forecasts: MonthlyForecast[] }>(
      `/colocations/${colocationId}/expenses/forecast`,
      { params: { months } }
    );
    return response.data.forecasts || [];
  },
};
