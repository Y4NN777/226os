import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
  type QueryClient,
} from '@tanstack/react-query';
import { useCallback } from 'react';
import { api, ApiError } from '../api/http';
import type { components } from '../api/schema';

export type SessionState = components['schemas']['SessionState'];

export const sessionQueryKey = ['session'] as const;

// Bumped whenever this tab changes the session itself (sign-in, sign-out, reset), so a
// session read that started before cannot bring back the old token or the old user.
let sessionEpoch = 0;

/** Marks a session change made by this tab: reads already in flight are now stale. */
export function bumpSessionEpoch() {
  sessionEpoch += 1;
}

/**
 * GET /auth/session never answers 401: anonymous callers get a pre-session CSRF token.
 * Every response refreshes the in-memory token used by unsafe calls (ADR-0003).
 */
export const sessionQuery = queryOptions({
  queryKey: sessionQueryKey,
  queryFn: async ({ client }): Promise<SessionState> => {
    const started = sessionEpoch;
    const state = await api.call(() => api.client.GET('/auth/session'));
    const current = client.getQueryData<SessionState>(sessionQueryKey);
    if (started !== sessionEpoch && current !== undefined) return current;
    // Another tab may have signed in as someone else: never keep their predecessor's data.
    if (current !== undefined && current.user?.id !== state.user?.id) forgetPrivateData(client);
    api.csrf.set(state.csrf_token);
    return state;
  },
  staleTime: 60_000,
  retry: false,
});

export function useSession() {
  return useQuery(sessionQuery);
}

/**
 * Forgets every cached server answer except the session itself, so the next person on a
 * shared device never sees the previous member's data. Called whenever the user changes.
 */
export function forgetPrivateData(queryClient: QueryClient) {
  queryClient.removeQueries({ predicate: (query) => query.queryKey[0] !== sessionQueryKey[0] });
}

/** Stores a session returned by sign-in: the token rotates with the session. */
export function applySession(queryClient: QueryClient, state: SessionState) {
  bumpSessionEpoch();
  api.csrf.set(state.csrf_token);
  forgetPrivateData(queryClient);
  queryClient.setQueryData(sessionQueryKey, state);
}

export function useApplySession() {
  const queryClient = useQueryClient();
  return useCallback(
    (state: SessionState) => {
      applySession(queryClient, state);
    },
    [queryClient],
  );
}

/** POST /auth/sign-out, then forget the token (revoked with the session) and re-read the session. */
export function useSignOut() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      try {
        await api.call(() => api.client.POST('/auth/sign-out'));
      } catch (error) {
        // 401: the session had already expired, so the member is signed out anyway.
        if (!(error instanceof ApiError && error.status === 401)) throw error;
      }
    },
    onSuccess: async () => {
      bumpSessionEpoch();
      api.csrf.set(null);
      forgetPrivateData(queryClient);
      await queryClient.invalidateQueries({ queryKey: sessionQueryKey });
    },
  });
}
