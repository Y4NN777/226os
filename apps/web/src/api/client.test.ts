import { describe, expect, it, vi } from 'vitest';
import { jsonResponse } from '../test/render';
import { createApiClient } from './client';

describe('createApiClient', () => {
  it('calls contract paths under /api/v1 and sends the session cookie', async () => {
    const fetchMock = vi.fn<typeof fetch>(() => Promise.resolve(jsonResponse({})));
    const client = createApiClient(fetchMock);
    await client.GET('/auth/session');
    const request = fetchMock.mock.calls[0]?.[0];
    expect(request).toBeInstanceOf(Request);
    const sent = request as Request;
    expect(new URL(sent.url).pathname).toBe('/api/v1/auth/session');
    expect(sent.method).toBe('GET');
    expect(sent.credentials).toBe('include');
  });

  it('uses the global fetch by default', async () => {
    const fetchMock = vi.fn<typeof fetch>(() => Promise.resolve(jsonResponse({})));
    vi.stubGlobal('fetch', fetchMock);
    await createApiClient().GET('/auth/session');
    expect(fetchMock).toHaveBeenCalledOnce();
  });
});
