import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { selectSegment } from '../../test/segmented';
import { listParams, resource, rowOf } from './testFixtures';
import { findVisibleText, getVisibleText } from '../../test/visibleText';

const senegal = resource({ id: 'a0000000-0000-4000-8000-000000000001', name: 'Wave Pay' });
const kenya = resource({
  id: 'a0000000-0000-4000-8000-000000000002',
  name: 'Safari Hub',
  countries: ['KE'],
  verification_status: 'claimed',
  last_verified_at: null,
});
const pan = resource({
  id: 'a0000000-0000-4000-8000-000000000003',
  name: 'Sans Pays',
  countries: [],
  verification_status: 'stale',
});
const ghana = resource({
  id: 'a0000000-0000-4000-8000-000000000004',
  name: 'Accra Gate',
  countries: ['GH'],
  verification_status: 'disputed',
});

function page(items: ReturnType<typeof resource>[], next: string | null = null) {
  return jsonResponse({ items, next_cursor: next });
}

function searchOf(request: Request) {
  return new URL(request.url).searchParams;
}

describe('DiscoverPage', () => {
  it('asks the server for West Africa in one request and shows what it returns', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /resources': () => page([senegal, ghana]) });
    await renderApp('/discover');

    expect(await screen.findByRole('link', { name: 'Wave Pay' })).toHaveAttribute(
      'href',
      '/discover/a0000000-0000-4000-8000-000000000001',
    );
    expect(screen.getByRole('link', { name: 'Accra Gate' })).toBeInTheDocument();
    expect(
      screen.getByText('Afrique de l’Ouest d’abord, puis toute l’Afrique'),
    ).toBeInTheDocument();
    const params = listParams(fetchMock, 'first');
    expect(params.get('limit')).toBe('50');
    // One comma-separated parameter (form, explode false), as the contract requires.
    expect(params.getAll('country')).toEqual(['BJ,BF,CV,CI,GM,GH,GN,GW,LR,ML,MR,NE,NG,SN,SL,TG']);
    expect(callsTo(fetchMock, 'GET /resources')).toHaveLength(1);
  });

  it('marks demo fixtures with a badge and live entries without one', async () => {
    setLanguages(['fr-FR']);
    const demo = resource({
      id: 'a0000000-0000-4000-8000-000000000009',
      name: 'Exemple Démo',
      provenance: 'demo',
    });
    mockApi({ 'GET /resources': () => page([senegal, demo]) });
    await renderApp('/discover');
    const demoRow = rowOf(await screen.findByRole('link', { name: 'Exemple Démo' }));
    expect(within(demoRow).getByText('Démo')).toBeInTheDocument();
    const liveRow = rowOf(screen.getByRole('link', { name: 'Wave Pay' }));
    expect(within(liveRow).queryByText('Démo')).not.toBeInTheDocument();
  });

  it('shows provider, countries, verification text, date and source count without invented metrics', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([senegal]) });
    await renderApp('/discover');
    const row = rowOf(await screen.findByRole('link', { name: 'Wave Pay' }));
    // Prototype sub-line: the provider name alone, then category and type.
    expect(within(row).getByText('Exemple SA')).toBeInTheDocument();
    expect(within(row).getByText('Vérifié')).toBeInTheDocument();
    expect(within(row).getByText(/Vérifié le 10 nov/)).toBeInTheDocument();
    expect(within(row).getByText('2 sources')).toBeInTheDocument();
    expect(within(row).getByText(/Sénégal/)).toBeInTheDocument();
    expect(within(row).getByText('Encaissement mobile money.')).toBeInTheDocument();
  });

  it('labels every verification status with text', async () => {
    setLanguages(['fr-FR']);
    const withStatus = (
      n: number,
      verification_status: ReturnType<typeof resource>['verification_status'],
    ) =>
      resource({
        id: `b0000000-0000-4000-8000-00000000000${String(n)}`,
        name: `Fiche ${String(n)}`,
        verification_status,
      });
    mockApi({
      'GET /resources': () =>
        page([
          withStatus(1, 'verified'),
          withStatus(2, 'claimed'),
          withStatus(3, 'unverified'),
          withStatus(4, 'stale'),
          withStatus(5, 'disputed'),
        ]),
    });
    await renderApp('/discover');
    for (const [n, text] of [
      [1, 'Vérifié'],
      [2, 'Déclaré par le fournisseur'],
      [3, 'Non vérifié'],
      [4, 'À revérifier'],
      [5, 'Contesté'],
    ] as const) {
      const row = rowOf(await screen.findByRole('link', { name: `Fiche ${String(n)}` }));
      expect(within(row).getByText(text)).toBeInTheDocument();
    }
    const claimed = rowOf(screen.getByRole('link', { name: 'Fiche 2' }));
    expect(within(claimed).queryByText(/Vérifié le/)).not.toBeInTheDocument();
  });

  it('loads the next page with its cursor and the same region, then hides the button', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /resources': (request) =>
        searchOf(request).get('cursor') === null ? page([senegal], 'c1') : page([ghana], null),
    });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });

    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Afficher plus' }));
    expect(await screen.findByRole('link', { name: 'Accra Gate' })).toBeInTheDocument();
    const calls = callsTo(fetchMock, 'GET /resources');
    expect(calls).toHaveLength(2);
    expect(searchOf(calls[1] as Request).get('cursor')).toBe('c1');
    expect(searchOf(calls[1] as Request).get('country')).toContain('BF');
    expect(screen.queryByRole('button', { name: 'Afficher plus' })).not.toBeInTheDocument();
  });

  it('shows everything for "Toute l’Afrique" and keeps the filter in the URL', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /resources': () => page([senegal, kenya, pan]) });
    const { router } = await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });

    await selectSegment(screen.getByRole('radio', { name: 'Toute l’Afrique' }));
    expect(await screen.findByRole('link', { name: 'Safari Hub' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Sans Pays' })).toBeInTheDocument();
    expect(router.state.location.search).toMatchObject({ region: 'all' });
    expect(listParams(fetchMock, 'last').has('country')).toBe(false);
  });

  it('sends a single picked country to the server and does not filter client-side', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /resources': (request) =>
        // The server owns the country filter: whatever it returns is shown as is.
        searchOf(request).get('country') === 'GH' ? page([kenya]) : page([senegal]),
    });
    const { router } = await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });

    await userEvent.setup({ delay: null }).click(screen.getByRole('button', { name: /^Pays/ }));
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('option', { name: 'Ghana' }));
    expect(await screen.findByRole('link', { name: 'Safari Hub' })).toBeInTheDocument();
    expect(router.state.location.search).toMatchObject({ country: 'GH' });
    expect(listParams(fetchMock, 'last').get('country')).toBe('GH');
  });

  it('offers West African countries by their localized names', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /resources': () => page([senegal]) });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    await userEvent.setup({ delay: null }).click(screen.getByRole('button', { name: /^Country/ }));
    expect(await screen.findByRole('option', { name: 'Senegal' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Côte d’Ivoire' })).toBeInTheDocument();
  });

  it('submits the text search on Enter and the type filter, then resets everything', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /resources': () => page([senegal]) });
    const { router } = await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(
      screen.queryByRole('button', { name: 'Réinitialiser les filtres' }),
    ).not.toBeInTheDocument();

    await userEvent
      .setup({ delay: null })
      .type(screen.getByRole('searchbox', { name: /Rechercher une API/ }), 'momo{Enter}');
    await waitFor(() => {
      expect(listParams(fetchMock, 'last').get('q')).toBe('momo');
    });
    await selectSegment(screen.getByRole('radio', { name: 'SDK' }));
    await waitFor(() => {
      expect(listParams(fetchMock, 'last').get('type')).toBe('sdk');
    });
    expect(router.state.location.search).toMatchObject({ q: 'momo', type: 'sdk' });

    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Réinitialiser les filtres' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({});
    });
    expect(screen.getByRole('searchbox', { name: /Rechercher une API/ })).toHaveValue('');
  });

  it('submits the search with the button too', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /resources': () => page([senegal]) });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    await userEvent
      .setup({ delay: null })
      .type(screen.getByRole('searchbox', { name: /Rechercher une API/ }), 'ussd');
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Lancer la recherche' }));
    await waitFor(() => {
      expect(listParams(fetchMock, 'last').get('q')).toBe('ussd');
    });
  });

  it('shows and removes an active category filter coming from the URL', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /resources': () => page([senegal]) });
    const { router } = await renderApp('/discover?category=payments');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(listParams(fetchMock, 'first').get('category')).toBe('payments');
    expect(screen.getByText('Catégorie : payments')).toBeInTheDocument();
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Retirer le filtre de catégorie' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({});
    });
  });

  it('offers the whole continent when the West Africa view is empty', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      // The server applies the region: nothing for West Africa, Kenya for the whole continent.
      'GET /resources': (request) => (searchOf(request).has('country') ? page([]) : page([kenya])),
    });
    const { router } = await renderApp('/discover');
    expect(await findVisibleText('Aucune fiche ne correspond')).toBeInTheDocument();
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Voir toute l’Afrique' }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ region: 'all' });
    });
    expect(await screen.findByRole('link', { name: 'Safari Hub' })).toBeInTheDocument();
  });

  it('suggests removing filters when nothing matches elsewhere', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([]) });
    await renderApp('/discover?region=all&q=zzz');
    expect(await findVisibleText('Aucune fiche ne correspond')).toBeInTheDocument();
    // Astryx may also announce the message in its aria-live status region, so match the visible copy only.
    const visible = screen
      .getAllByText(/Retirez un filtre/)
      .filter((el) => !el.closest('[data-astryx-live-region]'));
    expect(visible).toHaveLength(1);
    expect(screen.queryByRole('button', { name: 'Voir toute l’Afrique' })).not.toBeInTheDocument();
  });

  it('offers to clear the filters from the empty state', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([]) });
    const { router } = await renderApp('/discover?region=all&q=zzz');
    expect(await findVisibleText('Aucune fiche ne correspond')).toBeInTheDocument();
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Réinitialiser les filtres' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({});
    });
  });

  it('shows the error with the request reference and retries', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    mockApi({
      'GET /resources': () =>
        calls++ === 0 ? problemResponse(500, 'INTERNAL_ERROR') : page([senegal]),
    });
    await renderApp('/discover');
    expect(await findVisibleText('Référence : req-test-1')).toBeInTheDocument();
    await userEvent.setup({ delay: null }).click(screen.getByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('link', { name: 'Wave Pay' })).toBeInTheDocument();
  });

  it('retries once on its own after a dropped connection', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    mockApi({
      'GET /resources': () => {
        if (calls++ === 0) throw new TypeError('Failed to fetch');
        return page([senegal]);
      },
    });
    await renderApp('/discover');
    expect(
      await screen.findByRole('link', { name: 'Wave Pay' }, { timeout: 4000 }),
    ).toBeInTheDocument();
    expect(calls).toBe(2);
  });

  it('tells how long to wait on a rate limit', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 12 }) });
    await renderApp('/discover');
    expect(await findVisibleText(/Réessayez dans 12 s/)).toBeInTheDocument();
  });

  it('keeps the list and reports an error when loading more fails', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      'GET /resources': (request) =>
        searchOf(request).get('cursor') === null
          ? page([senegal], 'c1')
          : problemResponse(500, 'INTERNAL_ERROR'),
    });
    await renderApp('/discover');
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('button', { name: 'Afficher plus' }));
    expect((await screen.findAllByRole('alert')).length).toBeGreaterThan(0);
    expect(screen.getByRole('link', { name: 'Wave Pay' })).toBeInTheDocument();
  });

  it('lays each listing out as a compact row: title heading, provider line, codes and status', async () => {
    setLanguages(['fr-FR']);
    const momo = resource({
      id: 'a0000000-0000-4000-8000-000000000005',
      name: 'Kalimba Pay',
      category: 'mobile-money',
      countries: ['SN', 'BF'],
    });
    mockApi({ 'GET /resources': () => page([momo]) });
    await renderApp('/discover');
    const link = await screen.findByRole('link', { name: 'Kalimba Pay' });
    expect(screen.getByRole('heading', { level: 3, name: 'Kalimba Pay' })).toContainElement(link);
    const row = rowOf(link);
    expect(within(row).getByText('Paiements Mobile Money')).toBeInTheDocument();
    expect(within(row).getByText('API')).toBeInTheDocument();
    // Country codes stay text (no flags), each with its full name for assistive technology.
    expect(within(row).getByText('SN')).toBeInTheDocument();
    expect(within(row).getByText('Burkina Faso')).toBeInTheDocument();
    expect(within(row).getByText('2 sources')).toBeInTheDocument();
  });

  it('keeps the heading outline: h1, a hidden « Résultats » h2, then one h3 per row', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([senegal]) });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    const results = screen.getByRole('region', { name: 'Résultats' });
    expect(
      within(results).getByRole('heading', { level: 2, name: 'Résultats' }),
    ).toBeInTheDocument();
    expect(
      within(results).getByRole('heading', { level: 3, name: 'Wave Pay' }),
    ).toBeInTheDocument();
  });

  it('says « Choisir un pays » in the West Africa view and « Tous les pays » for the continent', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([senegal]) });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(screen.getByRole('button', { name: /^Pays/ })).toHaveTextContent('Choisir un pays');
    await selectSegment(screen.getByRole('radio', { name: 'Toute l’Afrique' }));
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^Pays/ })).toHaveTextContent('Tous les pays');
    });
  });

  it('shows the raw category code when the label is unknown', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([senegal]) });
    await renderApp('/discover');
    const row = rowOf(await screen.findByRole('link', { name: 'Wave Pay' }));
    expect(within(row).getByText('payments')).toBeInTheDocument();
  });

  it('counts the listings shown, without claiming a total while more pages exist', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      'GET /resources': (request) =>
        searchOf(request).get('cursor') === null ? page([senegal, ghana], 'c1') : page([pan]),
    });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(getVisibleText('Les 2 premières fiches')).toBeInTheDocument();
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Afficher plus' }));
    await screen.findByRole('link', { name: 'Sans Pays' });
    expect(getVisibleText('3 fiches')).toBeInTheDocument();
  });

  it('says « 1 fiche » in the singular', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([senegal]) });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(getVisibleText('1 fiche')).toBeInTheDocument();
  });

  it('filters by category with a labelled select and sends the code to the server', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /resources': () => page([senegal]) });
    const { router } = await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: /^Catégorie/ }));
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('option', { name: 'Modèles et API d’IA' }));
    await waitFor(() => {
      expect(listParams(fetchMock, 'last').get('category')).toBe('ai-models');
    });
    expect(router.state.location.search).toMatchObject({ category: 'ai-models' });
  });

  it('renders in English', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /resources': () => page([senegal]) });
    await renderApp('/discover');
    expect(
      await screen.findByText('West Africa first, then the whole continent'),
    ).toBeInTheDocument();
    const row = rowOf(await screen.findByRole('link', { name: 'Wave Pay' }));
    expect(within(row).getByText('Exemple SA')).toBeInTheDocument();
    expect(within(row).getByText('Senegal')).toBeInTheDocument();
    expect(within(row).getByText('2 sources')).toBeInTheDocument();
    expect(within(row).getByText(/Verified on Nov 10, 2026/)).toBeInTheDocument();
  });
});
