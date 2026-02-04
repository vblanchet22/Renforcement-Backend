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
    const response = await api.get<{ categories: Category[] }>(
      `/colocations/${colocationId}/categories`
    );
    return response.data.categories || [];
  },

  async create(data: CreateCategoryRequest): Promise<Category> {
    const { colocation_id, ...body } = data;
    const response = await api.post<Category>(
      `/colocations/${colocation_id}/categories`,
      body
    );
    return response.data;
  },

  async update(colocationId: string, id: string, data: UpdateCategoryRequest): Promise<Category> {
    const response = await api.put<Category>(
      `/colocations/${colocationId}/categories/${id}`,
      data
    );
    return response.data;
  },

  async delete(colocationId: string, id: string): Promise<void> {
    await api.delete(`/colocations/${colocationId}/categories/${id}`);
  },

  async getStats(params: CategoryStatsParams): Promise<CategoryStat[]> {
    const { colocation_id, ...queryParams } = params;
    const response = await api.get<{ stats: CategoryStat[] }>(
      `/colocations/${colocation_id}/categories/stats`,
      { params: queryParams }
    );
    return response.data.stats || [];
  },
};
