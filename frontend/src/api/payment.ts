import api from './client';
import type { Payment, PaymentStatus } from '../types';

interface CreatePaymentRequest {
  colocation_id: string;
  to_user_id: string;
  amount: number;
  note?: string;
}

interface ListPaymentsParams {
  colocation_id: string;
  status?: PaymentStatus;
  as_sender?: boolean;
  as_recipient?: boolean;
}

export const paymentApi = {
  async list(params: ListPaymentsParams): Promise<Payment[]> {
    const response = await api.get<{ payments: Payment[] }>('/v1/payments', { params });
    return response.data.payments || [];
  },

  async get(id: string): Promise<Payment> {
    const response = await api.get<Payment>(`/v1/payments/${id}`);
    return response.data;
  },

  async create(data: CreatePaymentRequest): Promise<Payment> {
    const response = await api.post<Payment>('/v1/payments', data);
    return response.data;
  },

  async confirm(id: string): Promise<Payment> {
    const response = await api.post<Payment>(`/v1/payments/${id}/confirm`);
    return response.data;
  },

  async reject(id: string): Promise<Payment> {
    const response = await api.post<Payment>(`/v1/payments/${id}/reject`);
    return response.data;
  },

  async cancel(id: string): Promise<void> {
    await api.delete(`/v1/payments/${id}`);
  },
};
