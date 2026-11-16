import { queryOptions } from '@tanstack/react-query';
import { api } from '../../api/http';

/** The server truncates `q` to 100 runes (ADR-0004 follow-up C4); do the same before sending. */
const MAX_LENGTH = 100;
/** `q` has minLength 2 in api/openapi.yaml: a shorter one would only earn a 422. */
const MIN_LENGTH = 2;

export function normalizeQuery(raw: string): string {
  return Array.from(raw.trim()).slice(0, MAX_LENGTH).join('');
}

export function isSearchable(q: string): boolean {
  return Array.from(q).length >= MIN_LENGTH;
}

export const searchKeys = { results: (q: string) => ['search', q] as const };

export function searchQuery(q: string) {
  return queryOptions({
    queryKey: searchKeys.results(q),
    enabled: isSearchable(q),
    // A failed search is shown, with a retry button, rather than silently repeated.
    retry: false,
    queryFn: () => api.call(() => api.client.GET('/search', { params: { query: { q } } })),
  });
}
