import { useInfiniteQuery } from '@tanstack/react-query';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import type { ConnectSearch } from './routeSearch';

export type ProfileSummary = components['schemas']['ProfileSummary'];

export const connectKeys = {
  list: (filters: ConnectSearch) => ['profiles', 'list', filters] as const,
};

function retryOnce(failures: number, error: unknown): boolean {
  return failures < 1 && error instanceof ApiError && error.code === 'NETWORK';
}

export function useProfileDirectory(filters: ConnectSearch) {
  return useInfiniteQuery({
    queryKey: connectKeys.list(filters),
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = { ...filters, ...(pageParam === null ? {} : { cursor: pageParam }) };
      const page = await api.call(() => api.client.GET('/profiles', { params: { query } }));
      return { items: page.items, nextCursor: page.next_cursor };
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}
