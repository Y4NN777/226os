import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { resource, rowOf, source } from './testFixtures';

const id = 'c0000000-0000-4000-8000-000000000001';
const detail = resource({
  id,
  name: 'Wave Pay',
  countries: ['SN', 'CI'],
  sdk_languages: [
    { language: 'python', name: 'wave-python', url: 'https://sdk.example/python' },
    { language: 'go', url: 'https://sdk.example/go' },
  ],
  doc_languages: ['fr', 'en'],
  docs_url: 'https://docs.provider.example/',
  limitations: ['Plafond de 1 000 000 XOF par jour', 'Pas de remboursement partiel'],
  related: [
    { type: 'resource', id: 'c0000000-0000-4000-8000-000000000002', title: 'Orange Money API' },
    { type: 'question', id: 'c0000000-0000-4000-8000-000000000003', title: 'Webhook non reçu' },
  ],
});

// A claim without a title (shown by its host) that the data wrongly marks as verified.
const untitledClaim = source({
  id: '22222222-2222-4222-8222-000000000002',
  url: 'https://provider.example/blog/annonce',
  kind: 'claim',
  status: 'verified',
  verified_at: null,
});
delete untitledClaim.title;

const sources = {
  items: [
    source(),
    untitledClaim,
    source({
      id: '22222222-2222-4222-8222-000000000003',
      url: 'https://prices.example/tarifs',
      title: 'Grille tarifaire',
      kind: 'pricing',
      status: 'stale',
    }),
  ],
};

function routes(overrides: Record<string, () => Response> = {}) {
  return mockApi({
    [`GET /resources/${id}`]: () => jsonResponse(detail),
    [`GET /resources/${id}/sources`]: () => jsonResponse(sources),
    ...overrides,
  });
}

describe('ResourcePage', () => {
  it('shows the listing with provider, coverage, SDKs, docs, limitations and verification', async () => {
    setLanguages(['fr-FR']);
    routes();
    await renderApp(`/discover/${id}`);
    expect(await screen.findByRole('heading', { level: 1, name: 'Wave Pay' })).toBeInTheDocument();
    expect(screen.getByText('Encaissement mobile money.')).toBeInTheDocument();
    expect(screen.getByText('Sénégal, Côte d’Ivoire')).toBeInTheDocument();
    expect(screen.getByText('français, anglais')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Exemple SA/ })).toHaveAttribute(
      'href',
      'https://provider.example/',
    );
    expect(screen.getByRole('link', { name: /wave-python/ })).toHaveAttribute(
      'href',
      'https://sdk.example/python',
    );
    expect(screen.getByRole('link', { name: /^go/ })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Lire la documentation/ })).toHaveAttribute(
      'href',
      'https://docs.provider.example/',
    );
    const limits = screen.getByRole('list', { name: 'Limites connues' });
    expect(within(limits).getAllByRole('listitem')).toHaveLength(2);
    expect(screen.getByText(/Vérifié le 10 nov/)).toBeInTheDocument();
  });

  it('opens every external link safely in a new tab', async () => {
    setLanguages(['fr-FR']);
    routes();
    await renderApp(`/discover/${id}`);
    await screen.findByRole('link', { name: /Grille tarifaire/ });
    const external = screen
      .getAllByRole('link')
      .filter((link) => link.getAttribute('href')?.startsWith('https://'));
    expect(external.length).toBeGreaterThanOrEqual(7);
    for (const link of external) {
      expect(link).toHaveAttribute('target', '_blank');
      expect(link.getAttribute('rel')).toMatch(/noopener/);
      expect(link.getAttribute('rel')).toMatch(/noreferrer/);
      expect(link).toHaveTextContent(/nouvel onglet/);
    }
  });

  it('links only related listings and shows other related items as plain text', async () => {
    setLanguages(['fr-FR']);
    routes();
    await renderApp(`/discover/${id}`);
    expect(await screen.findByRole('link', { name: 'Orange Money API' })).toHaveAttribute(
      'href',
      '/discover/c0000000-0000-4000-8000-000000000002',
    );
    expect(screen.getByText(/Webhook non reçu/)).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /Webhook non reçu/ })).not.toBeInTheDocument();
  });

  it('lists sources as evidence, never presenting a claim as verified', async () => {
    setLanguages(['fr-FR']);
    routes();
    await renderApp(`/discover/${id}`);
    const docs = await screen.findByRole('link', { name: /Référence de l’API/ });
    const official = rowOf(docs);
    expect(within(official).getByText('Documentation officielle')).toBeInTheDocument();
    expect(within(official).getByText('Vérifié')).toBeInTheDocument();
    expect(within(official).getByText(/Vérifié le 9 nov/)).toBeInTheDocument();

    const claim = rowOf(screen.getByRole('link', { name: /provider\.example/ }));
    expect(within(claim).getByText('Déclaration du fournisseur')).toBeInTheDocument();
    expect(within(claim).getByText('Déclaré par le fournisseur')).toBeInTheDocument();
    expect(within(claim).queryByText('Vérifié')).not.toBeInTheDocument();

    const pricing = rowOf(screen.getByRole('link', { name: /Grille tarifaire/ }));
    expect(within(pricing).getByText('Tarifs')).toBeInTheDocument();
    expect(within(pricing).getByText('À revérifier')).toBeInTheDocument();
  });

  it('says so when no source is recorded', async () => {
    setLanguages(['fr-FR']);
    routes({ [`GET /resources/${id}/sources`]: () => jsonResponse({ items: [] }) });
    await renderApp(`/discover/${id}`);
    expect(
      await screen.findByText('Aucune source enregistrée pour le moment.'),
    ).toBeInTheDocument();
  });

  it('keeps the listing readable when only the sources fail, with a retry', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    routes({
      [`GET /resources/${id}/sources`]: () =>
        calls++ === 0 ? problemResponse(500, 'INTERNAL_ERROR') : jsonResponse(sources),
    });
    await renderApp(`/discover/${id}`);
    expect(await screen.findByRole('heading', { level: 1, name: 'Wave Pay' })).toBeInTheDocument();
    await userEvent.click(await screen.findByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('link', { name: /Grille tarifaire/ })).toBeInTheDocument();
  });

  it('has a way back to the catalogue and omits empty sections', async () => {
    setLanguages(['fr-FR']);
    const bare = resource({
      id,
      name: 'Nu',
      provider: { name: 'Sans Site', slug: 'sans-site' },
      countries: [],
    });
    routes({ [`GET /resources/${id}`]: () => jsonResponse(bare) });
    await renderApp(`/discover/${id}`);
    await screen.findByRole('heading', { level: 1, name: 'Nu' });
    expect(screen.getByRole('link', { name: 'Découvrir' })).toHaveAttribute('href', '/discover');
    expect(screen.getByText('Sans Site')).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /Sans Site/ })).not.toBeInTheDocument();
    expect(screen.getByText('Non précisé')).toBeInTheDocument();
    expect(screen.queryByText('Limites connues')).not.toBeInTheDocument();
    expect(screen.queryByText('À voir aussi')).not.toBeInTheDocument();
    expect(screen.queryByText('SDK officiels')).not.toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /Lire la documentation/ })).not.toBeInTheDocument();
  });

  it('shows a friendly not-found state with a link back', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      [`GET /resources/${id}`]: () => problemResponse(404, 'NOT_FOUND'),
      [`GET /resources/${id}/sources`]: () => problemResponse(404, 'NOT_FOUND'),
    });
    const { router } = await renderApp(`/discover/${id}`);
    expect(await screen.findByRole('heading', { name: 'Fiche introuvable' })).toBeInTheDocument();
    await userEvent.click(screen.getByRole('link', { name: 'Retour au catalogue' }));
    await screen.findByRole('heading', { level: 1, name: 'Découvrir' });
    expect(router.state.location.pathname).toBe('/discover');
  });

  it('shows other errors with the request reference and retries', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    const fetchMock = routes({
      [`GET /resources/${id}`]: () =>
        calls++ === 0 ? problemResponse(500, 'INTERNAL_ERROR') : jsonResponse(detail),
    });
    await renderApp(`/discover/${id}`);
    expect(await screen.findByText('Référence : req-test-1')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('heading', { level: 1, name: 'Wave Pay' })).toBeInTheDocument();
    expect(callsTo(fetchMock, `GET /resources/${id}`)).toHaveLength(2);
  });

  it('renders in English', async () => {
    setLanguages(['en-US']);
    routes();
    await renderApp(`/discover/${id}`);
    expect(await screen.findByRole('heading', { level: 1, name: 'Wave Pay' })).toBeInTheDocument();
    expect(screen.getByText('Senegal, Côte d’Ivoire')).toBeInTheDocument();
    expect(screen.getByText('French, English')).toBeInTheDocument();
    expect(await screen.findByText('Official documentation')).toBeInTheDocument();
    expect(screen.getByText(/Verified on Nov 10, 2026/)).toBeInTheDocument();
    expect(screen.getByRole('list', { name: 'Known limitations' })).toBeInTheDocument();
  });
});
