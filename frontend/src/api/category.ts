import api from './client';
import type { Category, CategoryStat } from '../types';

interface CreateCategoryRequest {
  colocation_id: string;
  name: string;
  icon: string;
  color: string;
}

interface UpdateCategoryRequest {
  name?: string;
  icon?: string;
  color?: string;
}

interface CategoryStatsParams {
  colocation_id: string;
  start_date?: string;
  end_date?: string;
}

export const categoryApi = {
  async list(colocationId: string): Promise<Category[]> {
    const response = await api.get<{ categories: Category[] }>('/v1/categories', {
      params: { colocation_id: colocationId },
    });
    return response.data.categories || [];
  },

  async create(data: CreateCategoryRequest): Promise<Category> {
    const response = await api.post<Category>('/v1/categories', data);
    return response.data;
  },

  async update(id: string, data: UpdateCategoryRequest): Promise<Category> {
    const response = await api.put<Category>(`/v1/categories/${id}`, data);
    return response.data;
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/v1/categories/${id}`);
  },

  async getStats(params: CategoryStatsParams): Promise<CategoryStat[]> {
    const response = await api.get<{ stats: CategoryStat[] }>('/v1/categories/stats', { params });
    return response.data.stats || [];
  },
};
