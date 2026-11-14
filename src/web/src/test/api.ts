import { vi } from 'vitest';
import type { components } from '../api/schema';
import { jsonResponse } from './render';

type Handler = (request: Request) => Response | Promise<Response>;

/** Routes keyed "METHOD /path" with the path relative to /api/v1, e.g. "GET /auth/session". */
export type Routes = Record<string, Handler>;

export type SessionState = components['schemas']['SessionState'];
export type Account = components['schemas']['Account'];

export const anonymousSession: SessionState = {
  authenticated: false,
  user: null,
  csrf_token: 'csrf-anon',
};

export const memberAccount: Account = {
  id: '6f1c2a34-0000-4000-8000-000000000001',
  email: 'awa@example.org',
  email_verified: true,
  username: 'awa',
  display_name: 'Awa Traoré',
  roles: ['member'],
  created_at: '2026-11-01T10:00:00Z',
};

export const memberSession: SessionState = {
  authenticated: true,
  user: memberAccount,
  csrf_token: 'csrf-member',
};

/** RFC 9457 problem response, as the API sends it. */
export function problemResponse(
  status: number,
  code: components['schemas']['ErrorCode'],
  extra: { violations?: components['schemas']['Violation'][]; retryAfter?: number } = {},
) {
  const headers: Record<string, string> = { 'Content-Type': 'application/problem+json' };
  if (extra.retryAfter !== undefined) headers['Retry-After'] = String(extra.retryAfter);
  return new Response(
    JSON.stringify({
      type: `urn:226os:problem:${code}`,
      title: code,
      status,
      code,
      request_id: 'req-test-1',
      ...(extra.violations ? { violations: extra.violations } : {}),
    }),
    { status, headers },
  );
}

function toRequest(input: RequestInfo | URL, init?: RequestInit): Request {
  if (input instanceof Request) return input;
  return new Request(new URL(String(input), globalThis.location.origin), init);
}

function keyOf(request: Request): string {
  const path = new URL(request.url).pathname.replace(/^\/api\/v1/, '');
  return `${request.method} ${path}`;
}

/**
 * Stubs global fetch with a contract-level fake. GET /auth/session (anonymous) and
 * GET /auth/csrf are answered by default; any other unmatched call is a 404 problem,
 * so a missing handler fails loudly instead of hanging.
 */
export function mockApi(routes: Routes = {}) {
  const all: Routes = {
    'GET /auth/session': () => jsonResponse(anonymousSession),
    'GET /auth/csrf': () => jsonResponse({ csrf_token: 'csrf-pre' }),
    'GET /healthz': () => jsonResponse({ status: 'ok' }),
    ...routes,
  };
  const fetchMock = vi.fn<typeof fetch>((input, init) => {
    const request = toRequest(input, init);
    const handler = all[keyOf(request)];
    return Promise.resolve(handler ? handler(request) : problemResponse(404, 'NOT_FOUND'));
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

/** Requests the fake received for "METHOD /path". */
export function callsTo(fetchMock: ReturnType<typeof mockApi>, key: string): Request[] {
  return fetchMock.mock.calls
    .map(([input, init]) => toRequest(input, init))
    .filter((request) => keyOf(request) === key);
}
