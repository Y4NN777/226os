import { queryOptions, useQuery } from '@tanstack/react-query';
import { fetchHealth } from '../../api/health';

export const healthKeys = { all: ['health'] as const };

export const healthQuery = queryOptions({
  queryKey: healthKeys.all,
  // The probe reports the truth: no automatic retry hides an outage.
  retry: false,
  queryFn: ({ signal }) => fetchHealth(globalThis.fetch, signal),
});

export function useHealth() {
  return useQuery(healthQuery);
}
