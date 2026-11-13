import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

type Resource = components['schemas']['Resource'];

function resource(n: string, name: string): Resource {
  return {
    id: `00000000-0000-4000-8000-00000000000${n}`,
    name,
    slug: name.toLowerCase().replace(/\s+/g, '-'),
    type: 'api',
    category: 'payments',
    provider: { name: `Fournisseur ${n}`, slug: `fournisseur-${n}` },
    summary: `Résumé de ${name}`,
    countries: ['BF'],
    sdk_languages: [],
    doc_languages: ['fr'],
    verification_status: 'verified',
    last_verified_at: '2026-11-01T09:00:00Z',
    source_count: 1,
    related: [],
    updated_at: '2026-11-01T09:00:00Z',
  };
}

const page = {
  items: [resource('1', 'Mobile Money Alpha'), resource('2', 'SMS Beta')],
  next_cursor: null,
};

function signedIn(resources: () => Response | Promise<Response>) {
  return mockApi({
    'GET /auth/session': () => jsonResponse(memberSession),
    'GET /resources': resources,
  });
}

function section(name: string) {
  return screen.getByRole('region', { name });
}

describe('HomePage', () => {
  it('greets the member and lists the sections in pillar order (FR)', async () => {
    setLanguages(['fr-FR']);
    signedIn(() => jsonResponse(page));
    await renderApp('/');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Bonjour, Awa Traoré' }),
    ).toBeInTheDocument();
    const names = screen
      .getAllByRole('heading', { level: 2 })
      .map((heading) => heading.textContent);
    expect(names).toEqual([
      'Apprendre',
      'Connect',
      'Questions à résoudre',
      'API du catalogue',
      'Projets',
    ]);
  });

  it('greets in English too', async () => {
    setLanguages(['en-US']);
    signedIn(() => jsonResponse(page));
    await renderApp('/');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Hello, Awa Traoré' }),
    ).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Questions to solve' })).toBeVisible();
    expect(screen.getByRole('searchbox', { name: 'Search 226os' })).toBeVisible();
  });

  it('marks each fixture section with one demo badge, and none on the live catalogue', async () => {
    setLanguages(['fr-FR']);
    signedIn(() => jsonResponse(page));
    await renderApp('/');
    await screen.findByRole('link', { name: 'Mobile Money Alpha' });
    for (const name of ['Apprendre', 'Questions à résoudre', 'Projets']) {
      expect(within(section(name)).getAllByText('Démo')).toHaveLength(1);
      expect(within(section(name)).queryAllByRole('link')).toHaveLength(0);
    }
    expect(within(section('API du catalogue')).queryByText('Démo')).not.toBeInTheDocument();
    expect(within(section('Connect')).queryByText('Démo')).not.toBeInTheDocument();
    expect(
      within(section('Connect')).getByText('Bientôt : l’annuaire des développeurs.'),
    ).toBeVisible();
  });

  it('links the live catalogue entries and the catalogue index', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = signedIn(() => jsonResponse(page));
    await renderApp('/');
    const link = await screen.findByRole('link', { name: 'Mobile Money Alpha' });
    expect(link).toHaveAttribute('href', '/discover/00000000-0000-4000-8000-000000000001');
    expect(screen.getByRole('link', { name: 'SMS Beta' })).toBeVisible();
    expect(screen.getByRole('link', { name: 'Parcourir le catalogue' })).toHaveAttribute(
      'href',
      '/discover',
    );
    const [request] = callsTo(fetchMock, 'GET /resources');
    expect(new URL(request?.url ?? '').searchParams.get('limit')).toBe('4');
  });

  it('shows a loading state, then an error with a retry that recovers', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    signedIn(() => {
      calls += 1;
      return calls === 1 ? problemResponse(500, 'INTERNAL_ERROR') : jsonResponse(page);
    });
    await renderApp('/');
    const alert = await within(
      await screen.findByRole('region', { name: 'API du catalogue' }),
    ).findByRole('alert');
    expect(alert).toHaveTextContent('Impossible de charger le contenu');
    await userEvent.click(within(alert).getByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('link', { name: 'SMS Beta' })).toBeVisible();
  });

  it('says so when the catalogue is empty', async () => {
    setLanguages(['fr-FR']);
    signedIn(() => jsonResponse({ items: [], next_cursor: null }));
    await renderApp('/');
    expect(await screen.findByText('Le catalogue est vide pour l’instant.')).toBeVisible();
  });

  it('shows a labelled loading indicator while the catalogue loads', async () => {
    setLanguages(['fr-FR']);
    signedIn(() => new Promise<Response>(() => undefined));
    await renderApp('/');
    const region = await screen.findByRole('region', { name: 'API du catalogue' });
    expect(within(region).getAllByRole('status')[0]).toHaveTextContent('Chargement…');
  });

  it('searches with a trimmed, capped query and ignores an empty one', async () => {
    setLanguages(['fr-FR']);
    signedIn(() => jsonResponse(page));
    const { router } = await renderApp('/');
    const field = await screen.findByRole('searchbox', { name: 'Rechercher dans 226os' });
    await userEvent.type(field, '   {Enter}');
    expect(router.state.location.pathname).toBe('/');
    await userEvent.type(field, '  go api  {Enter}');
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/recherche');
    });
    expect(router.state.location.search).toEqual({ q: 'go api' });
  });

  it('caps the query at 100 characters', async () => {
    setLanguages(['fr-FR']);
    signedIn(() => jsonResponse(page));
    const { router } = await renderApp('/');
    const field = await screen.findByRole('searchbox', { name: 'Rechercher dans 226os' });
    await userEvent.click(field);
    await userEvent.paste('a'.repeat(120));
    await userEvent.click(screen.getByRole('button', { name: 'Rechercher' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/recherche');
    });
    expect(router.state.location.search).toEqual({ q: 'a'.repeat(100) });
  });
});
