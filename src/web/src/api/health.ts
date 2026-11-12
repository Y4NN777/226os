export interface Health {
  status: 'ok';
}

/**
 * GET /healthz is an infrastructure probe served outside /api/v1 and is not described in
 * api/openapi.yaml, so it has no generated type. The body is narrowed, never cast.
 */
export async function fetchHealth(
  fetchImpl: typeof fetch = globalThis.fetch,
  signal?: AbortSignal,
): Promise<Health> {
  const response = await fetchImpl('/healthz', {
    credentials: 'include',
    ...(signal ? { signal } : {}),
  });
  if (!response.ok) throw new Error(`healthz responded ${String(response.status)}`);
  const body: unknown = await response.json();
  if (typeof body === 'object' && body !== null && 'status' in body && body.status === 'ok') {
    return { status: 'ok' };
  }
  throw new Error('healthz returned an unexpected body');
}
