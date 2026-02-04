import api from './client';
import type { Colocation, ColocationMember, ColocationWithMembers } from '../types';

interface CreateColocationRequest {
  name: string;
  description?: string;
  address?: string;
}

interface UpdateColocationRequest {
  name?: string;
  description?: string;
  address?: string;
}

export const colocationApi = {
  async list(): Promise<Colocation[]> {
    const response = await api.get<{ colocations: Colocation[] }>('/colocations');
    return response.data.colocations || [];
  },

  async get(id: string): Promise<ColocationWithMembers> {
    const response = await api.get<ColocationWithMembers>(`/colocations/${id}`);
    return response.data;
  },

  async create(data: CreateColocationRequest): Promise<Colocation> {
    const response = await api.post<Colocation>('/colocations', data);
    return response.data;
  },

  async update(id: string, data: UpdateColocationRequest): Promise<Colocation> {
    const response = await api.put<Colocation>(`/colocations/${id}`, data);
    return response.data;
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/colocations/${id}`);
  },

  async join(inviteCode: string): Promise<ColocationMember> {
    const response = await api.post<ColocationMember>('/colocations/join', {
      invite_code: inviteCode,
    });
    return response.data;
  },

  async leave(id: string): Promise<void> {
    await api.post(`/colocations/${id}/leave`);
  },

  async removeMember(colocationId: string, userId: string): Promise<void> {
    await api.delete(`/colocations/${colocationId}/members/${userId}`);
  },

  async regenerateInviteCode(id: string): Promise<{ invite_code: string }> {
    const response = await api.post<{ invite_code: string }>(
      `/colocations/${id}/regenerate-code`
    );
    return response.data;
  },
};
