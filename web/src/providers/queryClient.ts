import { QueryClient } from '@tanstack/react-query';
import { ApiError } from '../utils/apiClient';

const shouldRetryRequest = (failureCount: number, error: unknown) => {
  if (error instanceof ApiError) {
    if (error.status === 401 || error.status === 403 || error.status === 404) {
      return false;
    }
    if (error.status >= 500) {
      return failureCount < 3;
    }
    return false;
  }
  
  return false;
};

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      refetchOnReconnect: false,
      retry: shouldRetryRequest,
      staleTime: 30 * 1000,
      gcTime: 5 * 60 * 1000,
      networkMode: 'online',
    },
    mutations: {
      retry: shouldRetryRequest,
      networkMode: 'online',
    },
  },
});
