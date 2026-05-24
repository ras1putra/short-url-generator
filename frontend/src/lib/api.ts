import axios from 'axios';
import { API_AUTH_REFRESH, ROUTE_LOGIN } from './constants';

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || '',
  withCredentials: true,
});

const RETRYABLE_MESSAGES = ['Missing token', 'Invalid or expired token'];
let refreshPromise: Promise<void> | null = null;

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    const status = error.response?.status;
    const message = error.response?.data?.message as string | undefined;

    if (
      status === 401 &&
      !originalRequest._retry &&
      message &&
      RETRYABLE_MESSAGES.includes(message)
    ) {
      originalRequest._retry = true;
      try {
        if (!refreshPromise) {
          refreshPromise = api
            .post(API_AUTH_REFRESH)
            .then(() => {})
            .finally(() => {
              refreshPromise = null;
            });
        }

        await refreshPromise;
        return api(originalRequest);
      } catch {
        if (typeof window !== 'undefined') {
          window.location.href = ROUTE_LOGIN;
        }
      }
    }

    return Promise.reject(error);
  }
);
