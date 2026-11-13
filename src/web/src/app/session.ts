import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { api } from '../api/http';
import type { components } from '../api/schema';

export type SessionState = components['schemas']['SessionState'];

export const sessionQueryKey = ['session'] as const;

/**
 * GET /auth/session never answers 401: anonymous callers get a pre-session CSRF token.
 * Every response refreshes the in-memory token used by unsafe calls (ADR-0003).
 */
export const sessionQuery = queryOptions({
  queryKey: sessionQueryKey,
  queryFn: async (): Promise<SessionState> => {
    const state = await api.call(() => api.client.GET('/auth/session'));
    api.csrf.set(state.csrf_token);
    return state;
  },
  staleTime: 60_000,
  retry: false,
});

export function useSession() {
  return useQuery(sessionQuery);
}

/** Stores a session returned by sign-in: the token rotates with the session. */
export function useApplySession() {
  const queryClient = useQueryClient();
  return useCallback(
    (state: SessionState) => {
      api.csrf.set(state.csrf_token);
      queryClient.setQueryData(sessionQueryKey, state);
    },
    [queryClient],
  );
}

/** POST /auth/sign-out, then forget the token (revoked with the session) and re-read the session. */
export function useSignOut() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.call(() => api.client.POST('/auth/sign-out')),
    onSuccess: async () => {
      api.csrf.set(null);
      await queryClient.invalidateQueries({ queryKey: sessionQueryKey });
    },
  });
}
