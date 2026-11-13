import type { components } from '../../api/schema';
import { callsTo, type mockApi } from '../../test/api';

export type Resource = components['schemas']['Resource'];
export type ResourceSource = components['schemas']['ResourceSource'];

/** Contract-shaped resource; synthetic data only. */
export function resource(overrides: Partial<Resource> = {}): Resource {
  return {
    id: '11111111-1111-4111-8111-111111111111',
    name: 'Passerelle Paiement',
    slug: 'passerelle-paiement',
    type: 'api',
    category: 'payments',
    provenance: 'live',
    provider: { name: 'Exemple SA', slug: 'exemple-sa', url: 'https://provider.example/' },
    summary: 'Encaissement mobile money.',
    countries: ['SN', 'CI'],
    sdk_languages: [],
    doc_languages: ['fr'],
    verification_status: 'verified',
    last_verified_at: '2026-11-10T09:00:00Z',
    source_count: 2,
    related: [],
    updated_at: '2026-11-10T09:00:00Z',
    ...overrides,
  };
}

export function source(overrides: Partial<ResourceSource> = {}): ResourceSource {
  return {
    id: '22222222-2222-4222-8222-222222222222',
    url: 'https://docs.provider.example/api',
    title: 'Référence de l’API',
    kind: 'official_docs',
    status: 'verified',
    verified_at: '2026-11-09T09:00:00Z',
    ...overrides,
  };
}

/** The list item holding an element; fails the test loudly instead of returning null. */
export function rowOf(element: HTMLElement): HTMLElement {
  const row = element.closest('li');
  if (row === null) throw new Error('element is not inside a list item');
  return row;
}

/** Query string of the first or last GET /resources the fake received. */
export function listParams(fetchMock: ReturnType<typeof mockApi>, which: 'first' | 'last') {
  const calls = callsTo(fetchMock, 'GET /resources');
  const call = which === 'first' ? calls[0] : calls.at(-1);
  if (call === undefined) throw new Error('GET /resources was never called');
  return new URL(call.url).searchParams;
}
