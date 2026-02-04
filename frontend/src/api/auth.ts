import api, { setTokens, clearTokens } from './client';
import type { AuthResponse, LoginRequest, RegisterRequest } from '../types';

export const authApi = {
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/auth/login', data);
    setTokens(response.data.access_token, response.data.refresh_token);
    return response.data;
  },

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/auth/register', data);
    setTokens(response.data.access_token, response.data.refresh_token);
    return response.data;
  },

  async logout(refreshToken: string): Promise<void> {
    try {
      await api.post('/auth/logout', { refresh_token: refreshToken });
    } finally {
      clearTokens();
    }
  },

  async refreshToken(token: string): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/auth/refresh', {
      refresh_token: token,
    });
    setTokens(response.data.access_token, response.data.refresh_token);
    return response.data;
  },
};
