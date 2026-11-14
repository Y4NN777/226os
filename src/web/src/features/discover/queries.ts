import { useInfiniteQuery, useMutation, useQuery } from '@tanstack/react-query';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { listQuery } from './region';
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

/** The contract's `country` is form style without explode: `country=BF,CI`, never repeated. */
const commaArrays = { array: { style: 'form', explode: false } } as const;

/** Cursor pages of the catalogue; the region and country filters run on the server. */
export function useResourceList(search: DiscoverSearch) {
  return useInfiniteQuery({
    queryKey: resourceKeys.list(search),
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query =
        pageParam === null ? listQuery(search) : { ...listQuery(search), cursor: pageParam };
      const page = await api.call(() =>
        api.client.GET('/resources', { params: { query }, querySerializer: commaArrays }),
      );
      return { items: page.items, nextCursor: page.next_cursor };
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

export type ErrorReportCreate = components['schemas']['ErrorReportCreate'];

/** Members only; the CSRF header comes from the shared client. Nothing cached changes. */
export function useCreateErrorReport(resourceId: string) {
  return useMutation({
    mutationFn: (body: ErrorReportCreate) =>
      api.call(() =>
        api.client.POST('/resources/{resourceId}/error-reports', {
          params: { path: { resourceId } },
          body,
        }),
      ),
  });
}
