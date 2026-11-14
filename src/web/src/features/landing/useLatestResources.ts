import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/http';

const LIMIT = 3;

/** The three latest catalogue entries shown on the landing page. */
export function useLatestResources() {
  return useQuery({
    // Failures surface at once with an explicit retry button, not after silent back-offs.
    retry: false,
    queryKey: ['resources', 'list', { limit: LIMIT }],
    queryFn: () =>
      api.call(() => api.client.GET('/resources', { params: { query: { limit: LIMIT } } })),
  });
}
