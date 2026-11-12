import { describe, expect, it, vi } from 'vitest';
import { jsonResponse } from '../test/render';
import { fetchHealth } from './health';

describe('fetchHealth', () => {
  it('returns the status reported by GET /healthz and sends credentials', async () => {
    const fetchMock = vi.fn<typeof fetch>(() => Promise.resolve(jsonResponse({ status: 'ok' })));
    await expect(fetchHealth(fetchMock)).resolves.toEqual({ status: 'ok' });
    expect(fetchMock).toHaveBeenCalledWith(
      '/healthz',
      expect.objectContaining({ credentials: 'include' }),
    );
  });

  it('uses the global fetch by default', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(jsonResponse({ status: 'ok' }))),
    );
    await expect(fetchHealth()).resolves.toEqual({ status: 'ok' });
  });

  it.each([
    ['a non-2xx status', () => jsonResponse({ status: 'ok' }, { status: 503 })],
    ['an unexpected status value', () => jsonResponse({ status: 'degraded' })],
    ['a non-object body', () => jsonResponse('ok')],
    ['a null body', () => jsonResponse(null)],
    ['an unparsable body', () => new Response('<html>', { status: 200 })],
  ])('rejects %s', async (_name, make) => {
    await expect(fetchHealth(() => Promise.resolve(make()))).rejects.toThrow();
  });

  it('propagates network failures', async () => {
    await expect(fetchHealth(() => Promise.reject(new TypeError('offline')))).rejects.toThrow(
      'offline',
    );
  });
});
