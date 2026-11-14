import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

type Resource = components['schemas']['Resource'];

function resource(n: number, overrides: Partial<Resource> = {}): Resource {
  return {
    id: `7a1c2a34-0000-4000-8000-00000000000${String(n)}`,
    name: `Passerelle ${String(n)}`,
    slug: `passerelle-${String(n)}`,
    type: 'api',
    provenance: 'live',
    category: 'payments',
    provider: { name: `Fournisseur ${String(n)}`, slug: `fournisseur-${String(n)}` },
    summary: `Résumé de la passerelle ${String(n)}.`,
    countries: ['BF'],
    sdk_languages: [],
    doc_languages: ['fr'],
    verification_status: 'verified',
    last_verified_at: '2026-11-01T10:00:00Z',
    source_count: 1,
    related: [],
    updated_at: '2026-11-01T10:00:00Z',
    ...overrides,
  };
}

const page = (items: Resource[]) => jsonResponse({ items, next_cursor: null });

async function open() {
  const view = await renderApp('/');
  const main = await screen.findByRole('main');
  return { ...view, main: within(main) };
}

describe('LandingPage', () => {
  it('shows the promise, the positioning and the two calls to action', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([resource(1)]) });
    const { main } = await open();
    expect(
      await main.findByRole('heading', {
        level: 1,
        name: 'Apprenez, échangez et construisez entre développeurs africains',
      }),
    ).toBeInTheDocument();
    expect(main.getByText('Afrique de l’Ouest d’abord, puis toute l’Afrique')).toBeInTheDocument();
    const signups = main.getAllByRole('link', { name: 'Créer un compte' });
    expect(signups.length).toBeGreaterThanOrEqual(2);
    for (const link of signups) expect(link).toHaveAttribute('href', '/inscription');
    expect(main.getByRole('link', { name: 'Parcourir le catalogue' })).toHaveAttribute(
      'href',
      '/discover',
    );
  });

  it('lists the pillars with Apprendre and Connect first, and only Découvrir links', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([resource(1)]) });
    const { main } = await open();
    const headings = (await main.findAllByRole('heading', { level: 3 })).map((h) => h.textContent);
    expect(headings.slice(0, 5)).toEqual([
      'Apprendre',
      'Connect',
      'Résoudre',
      'Découvrir',
      'Construire',
    ]);
    const link = main.getByRole('link', { name: 'Ouvrir le catalogue' });
    expect(link).toHaveAttribute('href', '/discover');
    // Every built pillar links to its page; only Connect is still announced.
    for (const [name, href] of [
      ['Lire les articles et les formations', '/apprendre'],
      ['Voir les questions', '/solve'],
      ['Voir les projets', '/build'],
    ] as const) {
      expect(main.getByRole('link', { name })).toHaveAttribute('href', href);
    }
    expect(main.getAllByText('Bientôt')).toHaveLength(1);
    expect(main.queryByRole('link', { name: /Connect/ })).not.toBeInTheDocument();
  });

  it('searches the catalogue with a trimmed, capped query and ignores an empty one', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([resource(1)]) });
    const user = userEvent.setup({ delay: null });
    const { main, router } = await open();
    const field = await main.findByRole('searchbox', { name: 'Rechercher dans le catalogue' });
    await user.type(field, '   {Enter}');
    expect(router.state.location.pathname).toBe('/');
    await user.clear(field);
    await user.type(field, '  paiement mobile ');
    await user.click(main.getByRole('button', { name: 'Rechercher' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/recherche');
    });
    expect(router.state.location.search).toEqual({ q: 'paiement mobile' });
  });

  it('caps the query at 100 characters', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([resource(1)]) });
    const user = userEvent.setup({ delay: null });
    const { main, router } = await open();
    const field = await main.findByRole('searchbox', { name: 'Rechercher dans le catalogue' });
    await user.click(field);
    await user.paste('a'.repeat(150));
    await user.keyboard('{Enter}');
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/recherche');
    });
    expect((router.state.location.search as { q: string }).q).toHaveLength(100);
  });

  it('renders the latest resources as links to their detail page', async () => {
    setLanguages(['fr-FR']);
    const items = [resource(1), resource(2), resource(3)];
    mockApi({ 'GET /resources': () => page(items) });
    const { main } = await open();
    const link = await main.findByRole('link', { name: 'Passerelle 2' });
    expect(link).toHaveAttribute('href', `/discover/${items[1]?.id ?? ''}`);
    expect(main.getByText('Fournisseur 2')).toBeInTheDocument();
    expect(main.getByText('Résumé de la passerelle 3.')).toBeInTheDocument();
  });

  it('shows a loading state while the catalogue loads', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => new Promise<Response>(() => undefined) });
    const { main } = await open();
    const catalogue = within(await main.findByRole('region', { name: 'Dans le catalogue' }));
    expect(await catalogue.findByRole('status')).toBeInTheDocument();
  });

  it('shows an error with a retry that reloads the catalogue', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    mockApi({
      'GET /resources': () => {
        calls += 1;
        return calls === 1 ? problemResponse(503, 'SERVICE_UNAVAILABLE') : page([resource(1)]);
      },
    });
    const user = userEvent.setup({ delay: null });
    const { main } = await open();
    await main.findByRole('alert');
    await user.click(main.getByRole('button', { name: 'Réessayer' }));
    expect(await main.findByRole('link', { name: 'Passerelle 1' })).toBeInTheDocument();
  });

  it('shows an empty state when the catalogue has no resource', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([]) });
    const { main } = await open();
    expect(
      await main.findByText('Le catalogue est vide pour le moment. Revenez bientôt.'),
    ).toBeInTheDocument();
  });

  it('reads in English when the browser asks for it', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /resources': () => page([resource(1)]) });
    const { main } = await open();
    expect(
      await main.findByRole('heading', {
        level: 1,
        name: 'Learn, connect and build with African developers',
      }),
    ).toBeInTheDocument();
    expect(main.getByText('West Africa first, then the whole continent')).toBeInTheDocument();
    expect(main.getAllByRole('link', { name: 'Create an account' })[0]).toHaveAttribute(
      'href',
      '/inscription',
    );
    expect(main.getAllByText('Soon')).toHaveLength(1);
  });
});
