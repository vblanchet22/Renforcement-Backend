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
    const { colocation_id, ...queryParams } = params;
    const response = await api.get<{ payments: Payment[] }>(
      `/colocations/${colocation_id}/payments`,
      { params: queryParams }
    );
    return response.data.payments || [];
  },

  async get(colocationId: string, id: string): Promise<Payment> {
    const response = await api.get<Payment>(`/colocations/${colocationId}/payments/${id}`);
    return response.data;
  },

  async create(data: CreatePaymentRequest): Promise<Payment> {
    const { colocation_id, ...body } = data;
    const response = await api.post<Payment>(`/colocations/${colocation_id}/payments`, body);
    return response.data;
  },

  async confirm(colocationId: string, id: string): Promise<Payment> {
    const response = await api.post<Payment>(
      `/colocations/${colocationId}/payments/${id}/confirm`
    );
    return response.data;
  },

  async reject(colocationId: string, id: string): Promise<Payment> {
    const response = await api.post<Payment>(
      `/colocations/${colocationId}/payments/${id}/reject`
    );
    return response.data;
  },

  async cancel(colocationId: string, id: string): Promise<void> {
    await api.delete(`/colocations/${colocationId}/payments/${id}`);
  },
};
