import { QueryClient } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { api } from '../api/http';
import { jsonResponse } from '../test/render';
import { applySession, sessionQuery, sessionQueryKey, type SessionState } from './session';

function account(id: string, username: string) {
  return {
    id,
    email: `${username}@example.test`,
    email_verified: true,
    username,
    display_name: username,
    roles: ['member' as const],
    created_at: '2026-11-14T00:00:00Z',
  };
}

const anonymous: SessionState = { authenticated: false, user: null, csrf_token: 'token-anon' };
const alice: SessionState = {
  authenticated: true,
  user: account('00000000-0000-4000-8000-00000000000a', 'alice'),
  csrf_token: 'token-alice',
};
const bob: SessionState = {
  authenticated: true,
  user: account('00000000-0000-4000-8000-00000000000b', 'bob'),
  csrf_token: 'token-bob',
};

function stubSessionResponses(...bodies: SessionState[]) {
  const releases: Array<() => void> = [];
  const fetchMock = vi.fn<typeof fetch>(
    () =>
      new Promise<Response>((resolve) => {
        const body = bodies[releases.length];
        releases.push(() => {
          resolve(jsonResponse(body));
        });
      }),
  );
  vi.stubGlobal('fetch', fetchMock);
  return (index = 0) => {
    const release = releases[index];
    if (release === undefined) throw new Error(`no request ${String(index)} yet`);
    release();
  };
}

afterEach(() => {
  vi.unstubAllGlobals();
  api.csrf.set(null);
});

describe('session query', () => {
  it('ignores a session read that started before a sign-in', async () => {
    const queryClient = new QueryClient();
    const releaseStale = stubSessionResponses(anonymous);
    const stale = queryClient.query(sessionQuery);
    await vi.waitFor(() => {
      expect(globalThis.fetch).toHaveBeenCalledTimes(1);
    });

    applySession(queryClient, alice);
    releaseStale();
    await stale;

    expect(queryClient.getQueryData(sessionQueryKey)).toEqual(alice);
    await expect(api.csrf.get()).resolves.toBe('token-alice');
  });

  it('forgets private data when a refetch returns another member', async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(sessionQueryKey, alice);
    queryClient.setQueryData(['me'], { private: 'alice' });
    const release = stubSessionResponses(bob);

    const refetch = queryClient.query({ ...sessionQuery, staleTime: 0 });
    await vi.waitFor(() => {
      expect(globalThis.fetch).toHaveBeenCalledTimes(1);
    });
    release();
    await refetch;

    expect(queryClient.getQueryData(['me'])).toBeUndefined();
    expect(queryClient.getQueryData(sessionQueryKey)).toEqual(bob);
  });

  it('keeps cached data when a refetch returns the same member', async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(sessionQueryKey, alice);
    queryClient.setQueryData(['me'], { private: 'alice' });
    const release = stubSessionResponses(alice);

    const refetch = queryClient.query({ ...sessionQuery, staleTime: 0 });
    await vi.waitFor(() => {
      expect(globalThis.fetch).toHaveBeenCalledTimes(1);
    });
    release();
    await refetch;

    expect(queryClient.getQueryData(['me'])).toEqual({ private: 'alice' });
  });
});
