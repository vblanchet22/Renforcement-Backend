import axios from 'axios';
import type { AxiosError, InternalAxiosRequestConfig } from 'axios';
import type { AuthResponse, ApiError } from '../types';

const API_BASE_URL = '/api';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Token management
const sanitizeToken = (value: string | null) =>
  value && value !== 'undefined' && value !== 'null' ? value : null;

let accessToken: string | null = sanitizeToken(localStorage.getItem('access_token'));
let refreshToken: string | null = sanitizeToken(localStorage.getItem('refresh_token'));

if (!accessToken && localStorage.getItem('access_token')) {
  localStorage.removeItem('access_token');
}
if (!refreshToken && localStorage.getItem('refresh_token')) {
  localStorage.removeItem('refresh_token');
}

export const setTokens = (access?: string | null, refresh?: string | null) => {
  accessToken = sanitizeToken(access ?? null);
  refreshToken = sanitizeToken(refresh ?? null);

  if (accessToken) {
    localStorage.setItem('access_token', accessToken);
  } else {
    localStorage.removeItem('access_token');
  }

  if (refreshToken) {
    localStorage.setItem('refresh_token', refreshToken);
  } else {
    localStorage.removeItem('refresh_token');
  }
};

export const clearTokens = () => {
  accessToken = null;
  refreshToken = null;
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
};

export const getAccessToken = () => accessToken;
export const getRefreshToken = () => refreshToken;

// Request interceptor to add auth header
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (accessToken && config.headers) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor to handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // If 401 and we have a refresh token, try to refresh
    if (error.response?.status === 401 && refreshToken && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const response = await axios.post<AuthResponse>(`${API_BASE_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });

        const { access_token, refresh_token } = response.data;
        setTokens(access_token, refresh_token);

        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
        }

        return api(originalRequest);
      } catch (refreshError) {
        clearTokens();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default api;
