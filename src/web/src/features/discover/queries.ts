import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { api, ApiError } from '../../api/http';
import { inWestAfrica, listQuery } from './region';
import type { DiscoverSearch } from './routeSearch';

/** One key factory; every variable a query function reads is in its key. */
export const resourceKeys = {
  list: (search: DiscoverSearch) => ['resources', 'list', search] as const,
  detail: (id: string) => ['resources', 'detail', id] as const,
  sources: (id: string) => ['resources', 'sources', id] as const,
};

/** One automatic retry for a dropped connection; any other failure waits for the retry button. */
function retryOnce(failures: number, error: unknown): boolean {
  return failures < 1 && error instanceof ApiError && error.code === 'NETWORK';
}

/** The default view (no region, no country) is West Africa, filtered client-side. */
export function isWestAfricaView(search: DiscoverSearch): boolean {
  return search.region !== 'all' && search.country === undefined;
}

/**
 * One logical page: in the West Africa view it keeps reading server pages until one has
 * a match (or the cursor ends), so "Afficher plus" never comes back empty-handed.
 */
export function useResourceList(search: DiscoverSearch) {
  const westOnly = isWestAfricaView(search);
  return useInfiniteQuery({
    queryKey: resourceKeys.list(search),
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      let cursor = pageParam;
      for (;;) {
        const query = cursor === null ? listQuery(search) : { ...listQuery(search), cursor };
        const page = await api.call(() => api.client.GET('/resources', { params: { query } }));
        const items = westOnly
          ? page.items.filter((item) => inWestAfrica(item.countries))
          : page.items;
        if (items.length > 0 || page.next_cursor === null) {
          return { items, nextCursor: page.next_cursor };
        }
        cursor = page.next_cursor;
      }
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}

export function useResource(id: string) {
  return useQuery({
    queryKey: resourceKeys.detail(id),
    retry: retryOnce,
    queryFn: () =>
      api.call(() =>
        api.client.GET('/resources/{resourceId}', { params: { path: { resourceId: id } } }),
      ),
  });
}

export function useResourceSources(id: string) {
  return useQuery({
    queryKey: resourceKeys.sources(id),
    retry: retryOnce,
    queryFn: () =>
      api.call(() =>
        api.client.GET('/resources/{resourceId}/sources', { params: { path: { resourceId: id } } }),
      ),
  });
}
