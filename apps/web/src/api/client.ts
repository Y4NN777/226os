import createClient from 'openapi-fetch';
import type { paths } from './schema';

/** Typed client for the contract in api/openapi.yaml (servers: /api/v1, same origin; absolute URL so Request resolves outside a browser too). */
export function createApiClient(fetchImpl?: typeof fetch) {
  return createClient<paths>({
    baseUrl: new URL('/api/v1', globalThis.location.origin).href,
    credentials: 'include',
    ...(fetchImpl ? { fetch: fetchImpl } : {}),
  });
}

export const apiClient = createApiClient();
