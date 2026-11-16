import { describe, expect, it, vi } from 'vitest';
import { jsonResponse } from '../test/render';
import { ApiError, createApi } from './http';

function problem(status: number, code: string, extra: Record<string, unknown> = {}, headers = {}) {
  return new Response(
    JSON.stringify({
      type: `urn:226os:problem:${code}`,
      title: code,
      status,
      code,
      request_id: 'req-123',
      ...extra,
    }),
    { status, headers: { 'Content-Type': 'application/problem+json', ...headers } },
  );
}

type FetchMock = ReturnType<typeof vi.fn<typeof fetch>>;

function sent(fetchMock: FetchMock): Request[] {
  return fetchMock.mock.calls.map(([input]) => input as Request);
}

function pathOf(input: unknown): string {
  return new URL((input as Request).url).pathname;
}

async function failure(promise: Promise<unknown>): Promise<ApiError> {
  const error = await promise.then(
    () => undefined,
    (e: unknown) => e,
  );
  if (!(error instanceof ApiError)) throw new Error('expected an ApiError');
  return error;
}

describe('createApi', () => {
  it('returns the data of a successful call', async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(jsonResponse({ authenticated: false, user: null, csrf_token: 't1' })),
    );
    const api = createApi(fetchMock);
    const session = await api.call(() => api.client.GET('/auth/session'));
    expect(session.csrf_token).toBe('t1');
  });

  it('never sends a CSRF header on safe methods', async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(jsonResponse({ items: [], next_cursor: null })),
    );
    const api = createApi(fetchMock);
    await api.call(() => api.client.GET('/resources'));
    expect(fetchMock).toHaveBeenCalledOnce();
    expect(sent(fetchMock)[0]?.headers.get('X-CSRF-Token')).toBeNull();
  });

  it('fetches a CSRF token once, then sends it in X-CSRF-Token on every unsafe call', async () => {
    const fetchMock = vi.fn<typeof fetch>((input) =>
      Promise.resolve(
        pathOf(input) === '/api/v1/auth/csrf'
          ? jsonResponse({ csrf_token: 'pre-1' })
          : jsonResponse({ status: 'accepted' }, { status: 202 }),
      ),
    );
    const api = createApi(fetchMock);
    const body = { email: 'a@example.org' };
    await api.call(() => api.client.POST('/auth/password-reset/request', { body }));
    await api.call(() => api.client.POST('/auth/password-reset/request', { body }));
    expect(sent(fetchMock).map((r) => pathOf(r))).toEqual([
      '/api/v1/auth/csrf',
      '/api/v1/auth/password-reset/request',
      '/api/v1/auth/password-reset/request',
    ]);
    const posts = sent(fetchMock).filter((r) => r.method === 'POST');
    expect(posts.map((r) => r.headers.get('X-CSRF-Token'))).toEqual(['pre-1', 'pre-1']);
  });

  it('uses a token set from a session response without fetching another', async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(new Response(null, { status: 204 })),
    );
    const api = createApi(fetchMock);
    api.csrf.set('session-token');
    await api.call(() => api.client.POST('/auth/sign-out'));
    expect(fetchMock).toHaveBeenCalledOnce();
    expect(sent(fetchMock)[0]?.headers.get('X-CSRF-Token')).toBe('session-token');
  });

  it('refreshes the token and retries once after 403 CSRF_FAILED', async () => {
    let posts = 0;
    const fetchMock = vi.fn<typeof fetch>((input) => {
      if (pathOf(input) === '/api/v1/auth/csrf') {
        return Promise.resolve(jsonResponse({ csrf_token: 'fresh' }));
      }
      posts += 1;
      return Promise.resolve(
        posts === 1 ? problem(403, 'CSRF_FAILED') : new Response(null, { status: 204 }),
      );
    });
    const api = createApi(fetchMock);
    api.csrf.set('stale');
    await api.call(() =>
      api.client.POST('/auth/verify-email', { body: { token: 'x'.repeat(16) } }),
    );
    const tokens = sent(fetchMock)
      .filter((r) => r.method === 'POST')
      .map((r) => r.headers.get('X-CSRF-Token'));
    expect(tokens).toEqual(['stale', 'fresh']);
  });

  it('gives up after one CSRF retry', async () => {
    const fetchMock = vi.fn<typeof fetch>((input) =>
      Promise.resolve(
        pathOf(input) === '/api/v1/auth/csrf'
          ? jsonResponse({ csrf_token: 'fresh' })
          : problem(403, 'CSRF_FAILED'),
      ),
    );
    const api = createApi(fetchMock);
    const error = await failure(api.call(() => api.client.POST('/auth/sign-out')));
    expect(error.code).toBe('CSRF_FAILED');
    expect(sent(fetchMock).filter((r) => r.method === 'POST')).toHaveLength(2);
  });

  it('maps a problem+json body to an ApiError with code, request id and violations', async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(
        problem(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/email', code: 'invalid_format' }],
        }),
      ),
    );
    const api = createApi(fetchMock);
    api.csrf.set('t');
    const error = await failure(
      api.call(() => api.client.POST('/auth/sign-in', { body: { email: 'x', password: 'y' } })),
    );
    expect(error.status).toBe(422);
    expect(error.code).toBe('VALIDATION_ERROR');
    expect(error.requestId).toBe('req-123');
    expect(error.violations).toEqual([{ field: '/email', code: 'invalid_format' }]);
    expect(error.retryAfter).toBeUndefined();
  });

  it('reads Retry-After in seconds on 429', async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(problem(429, 'RATE_LIMITED', {}, { 'Retry-After': '42' })),
    );
    const api = createApi(fetchMock);
    const error = await failure(api.call(() => api.client.GET('/resources')));
    expect(error.code).toBe('RATE_LIMITED');
    expect(error.retryAfter).toBe(42);
  });

  it('maps a non-problem error body to UNKNOWN with the HTTP status', async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(new Response('Bad gateway', { status: 502 })),
    );
    const api = createApi(fetchMock);
    const error = await failure(api.call(() => api.client.GET('/resources')));
    expect(error.status).toBe(502);
    expect(error.code).toBe('UNKNOWN');
    expect(error.requestId).toBeUndefined();
  });

  it('maps a network failure to NETWORK', async () => {
    const fetchMock = vi.fn<typeof fetch>(() => Promise.reject(new TypeError('Failed to fetch')));
    const api = createApi(fetchMock);
    const error = await failure(api.call(() => api.client.GET('/resources')));
    expect(error.code).toBe('NETWORK');
    expect(error.status).toBe(0);
  });

  it('never sends the mutation when no CSRF token can be obtained', async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(new Response('nope', { status: 500 })),
    );
    const api = createApi(fetchMock);
    await failure(api.call(() => api.client.POST('/auth/sign-out')));
    expect(sent(fetchMock).every((r) => r.method === 'GET')).toBe(true);
  });

  it('forgets the token on clear, so the next unsafe call fetches a new one', async () => {
    const fetchMock = vi.fn<typeof fetch>((input) =>
      Promise.resolve(
        pathOf(input) === '/api/v1/auth/csrf'
          ? jsonResponse({ csrf_token: 'new' })
          : new Response(null, { status: 204 }),
      ),
    );
    const api = createApi(fetchMock);
    api.csrf.set('old');
    api.csrf.set(null);
    await api.call(() => api.client.POST('/auth/sign-out'));
    const post = sent(fetchMock).find((r) => r.method === 'POST');
    expect(post?.headers.get('X-CSRF-Token')).toBe('new');
  });

  it('sends X-CSRF-Token on PUT, PATCH and DELETE too', async () => {
    const fetchMock = vi.fn<typeof fetch>(() => Promise.resolve(jsonResponse({}, { status: 200 })));
    const api = createApi(fetchMock);
    api.csrf.set('tok');
    await api.call(() =>
      api.client.PUT('/learn/series/{seriesId}/lessons/{lessonId}/progress', {
        params: { path: { seriesId: 'a', lessonId: 'b' } },
        body: { status: 'completed' },
      }),
    );
    await api.call(() => api.client.PATCH('/me/profile', { body: {} }));
    await api.call(() =>
      api.client.DELETE('/votes', {
        params: { query: { target_type: 'question', target_id: 'x' } },
      }),
    );
    expect(sent(fetchMock).map((r) => [r.method, r.headers.get('X-CSRF-Token')])).toEqual([
      ['PUT', 'tok'],
      ['PATCH', 'tok'],
      ['DELETE', 'tok'],
    ]);
  });

  it('keeps the problem code even without a request id (edge or proxy errors)', async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(
        new Response(JSON.stringify({ status: 429, code: 'RATE_LIMITED', title: 'x' }), {
          status: 429,
          headers: { 'Content-Type': 'application/problem+json', 'Retry-After': '9' },
        }),
      ),
    );
    const api = createApi(fetchMock);
    const error = await failure(api.call(() => api.client.GET('/resources')));
    expect(error.code).toBe('RATE_LIMITED');
    expect(error.requestId).toBeUndefined();
  });

  it('treats a bare 429 as RATE_LIMITED and caps Retry-After at five minutes', async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(
        new Response('slow down', { status: 429, headers: { 'Retry-After': '86400' } }),
      ),
    );
    const api = createApi(fetchMock);
    const error = await failure(api.call(() => api.client.GET('/resources')));
    expect(error.code).toBe('RATE_LIMITED');
    expect(error.retryAfter).toBe(300);
  });

  it('never lets a token fetched earlier overwrite one set while the fetch was in flight', async () => {
    let release: (r: Response) => void = () => undefined;
    const fetchMock = vi.fn<typeof fetch>((input) => {
      if (pathOf(input) === '/api/v1/auth/csrf') {
        return new Promise<Response>((resolve) => {
          release = resolve;
        });
      }
      return Promise.resolve(new Response(null, { status: 204 }));
    });
    const api = createApi(fetchMock);
    const first = api.csrf.get();
    const second = api.csrf.get();
    api.csrf.set('from-session');
    release(jsonResponse({ csrf_token: 'stale-fetch' }));
    await Promise.all([first, second]);
    expect(sent(fetchMock).filter((r) => pathOf(r) === '/api/v1/auth/csrf')).toHaveLength(1);
    expect(await api.csrf.get()).toBe('from-session');
  });
});
