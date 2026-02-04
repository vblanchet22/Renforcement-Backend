import api, { setTokens, clearTokens } from './client';
import type { AuthResponse, LoginRequest, RegisterRequest, UserInfo } from '../types';

type RawAuthResponse = {
  access_token?: string | null;
  refresh_token?: string | null;
  expires_in?: number | null;
  user?: UserInfo;
  accessToken?: string | null;
  refreshToken?: string | null;
  expiresIn?: number | null;
};

const normalizeAuthResponse = (data: RawAuthResponse): AuthResponse => {
  const accessToken = data.access_token ?? data.accessToken ?? null;
  const refreshToken = data.refresh_token ?? data.refreshToken ?? null;
  const expiresIn = data.expires_in ?? data.expiresIn ?? 0;
  const user = data.user;

  if (!user) {
    throw new Error("Réponse d'authentification invalide: utilisateur manquant");
  }

  setTokens(accessToken, refreshToken);

  return {
    access_token: accessToken ?? '',
    refresh_token: refreshToken ?? '',
    expires_in: expiresIn ?? 0,
    user,
  };
};

export const authApi = {
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await api.post<RawAuthResponse>('/auth/login', data);
    return normalizeAuthResponse(response.data);
  },

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await api.post<RawAuthResponse>('/auth/register', data);
    return normalizeAuthResponse(response.data);
  },

  async logout(refreshToken: string): Promise<void> {
    try {
      await api.post('/auth/logout', { refresh_token: refreshToken });
    } finally {
      clearTokens();
    }
  },

  async refreshToken(token: string): Promise<AuthResponse> {
    const response = await api.post<RawAuthResponse>('/auth/refresh', {
      refresh_token: token,
    });
    return normalizeAuthResponse(response.data);
  },
};
