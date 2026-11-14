import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

type SearchResults = components['schemas']['SearchResults'];

const RESOURCE_ID = '7a1c2a34-0000-4000-8000-0000000000a1';

const results: SearchResults = {
  query: 'momo',
  groups: {
    projects: {
      items: [
        {
          type: 'projects',
          id: '7a1c2a34-0000-4000-8000-0000000000d1',
          title: 'Passerelle de paiement',
          snippet: 'Un projet ouvert.',
          lang: 'fr',
        },
      ],
      has_more: false,
    },
    learn: {
      items: [
        {
          type: 'learn',
          id: '7a1c2a34-0000-4000-8000-0000000000c1',
          title: 'Premiers pas avec les paiements mobiles',
          snippet: 'Un guide.',
          lang: null,
        },
      ],
      has_more: false,
    },
    questions: {
      items: [
        {
          type: 'questions',
          id: '7a1c2a34-0000-4000-8000-0000000000b1',
          title: 'Webhook reçu deux fois',
          snippet: 'Le <b>webhook</b> arrive en double.',
          lang: 'en',
        },
      ],
      has_more: true,
    },
    resources: {
      items: [
        {
          type: 'resources',
          id: RESOURCE_ID,
          title: 'API Mobile Money',
          snippet: 'Paiements par téléphone.',
          lang: 'fr',
        },
      ],
      has_more: false,
    },
  },
};

const empty: SearchResults = {
  query: 'zzzz',
  groups: {
    resources: { items: [], has_more: false },
    questions: { items: [], has_more: false },
  },
};

function searchRoute(body: SearchResults) {
  return { 'GET /search': () => jsonResponse(body) };
}

describe('SearchPage', () => {
  it('without a query shows a helpful empty state and sends no request', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi();
    await renderApp('/recherche');
    expect(await screen.findByRole('heading', { level: 1, name: 'Recherche' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /Que cherchez-vous/ })).toBeInTheDocument();
    expect(screen.getByRole('searchbox', { name: 'Votre recherche' })).toHaveValue('');
    expect(callsTo(fetchMock, 'GET /search')).toHaveLength(0);
  });

  it('does not search for a single character', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi();
    await renderApp('/recherche?q=a');
    expect(await screen.findByText('Saisissez au moins deux caractères.')).toBeInTheDocument();
    expect(callsTo(fetchMock, 'GET /search')).toHaveLength(0);
  });

  it('groups results in a fixed order and links only resources', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /search': (request) => {
        expect(new URL(request.url).searchParams.get('q')).toBe('momo');
        return jsonResponse(results);
      },
    });
    await renderApp('/recherche?q=momo');
    await screen.findByRole('heading', { level: 2, name: 'Catalogue' });
    const titles = screen
      .getAllByRole('heading', { level: 2 })
      .map((heading) => heading.textContent)
      .filter((text) => ['Catalogue', 'Questions', 'Apprendre', 'Projets'].includes(text));
    expect(titles).toEqual(['Catalogue', 'Questions', 'Apprendre', 'Projets']);

    expect(screen.getByRole('link', { name: 'API Mobile Money' })).toHaveAttribute(
      'href',
      `/discover/${RESOURCE_ID}`,
    );
    for (const name of [
      'Webhook reçu deux fois',
      'Premiers pas avec les paiements mobiles',
      'Passerelle de paiement',
    ]) {
      expect(screen.getByText(name)).toBeInTheDocument();
      expect(screen.queryByRole('link', { name })).not.toBeInTheDocument();
    }
    expect(callsTo(fetchMock, 'GET /search')).toHaveLength(1);
  });

  it('renders snippets as plain text, never as HTML', async () => {
    setLanguages(['fr-FR']);
    mockApi(searchRoute(results));
    await renderApp('/recherche?q=momo');
    expect(await screen.findByText('Le <b>webhook</b> arrive en double.')).toBeInTheDocument();
    expect(document.querySelector('b')).toBeNull();
  });

  it('marks a hit whose language differs from the interface language', async () => {
    setLanguages(['fr-FR']);
    mockApi(searchRoute(results));
    await renderApp('/recherche?q=momo');
    const english = (await screen.findByText('Webhook reçu deux fois')).closest('li');
    expect(english).toHaveAttribute('lang', 'en');
    const french = screen.getByRole('link', { name: 'API Mobile Money' }).closest('li');
    expect(french).not.toHaveAttribute('lang');
    const unknown = screen.getByText('Premiers pas avec les paiements mobiles').closest('li');
    expect(unknown).not.toHaveAttribute('lang');
  });

  it('notes, without a link, that a group has more results', async () => {
    setLanguages(['fr-FR']);
    mockApi(searchRoute(results));
    await renderApp('/recherche?q=momo');
    const note = await screen.findByText(
      'D’autres résultats existent. Affinez votre recherche pour les voir.',
    );
    expect(note.closest('a')).toBeNull();
    expect(screen.getAllByText(/D’autres résultats existent/)).toHaveLength(1);
  });

  it('suggests other words when every group is empty', async () => {
    setLanguages(['fr-FR']);
    mockApi(searchRoute(empty));
    await renderApp('/recherche?q=zzzz');
    expect((await screen.findAllByText(/Aucun résultat pour .*zzzz/)).length).toBeGreaterThan(0);
    // Astryx may also announce the message in its aria-live status region, so match the visible copy only.
    const visible = screen
      .getAllByText(/Essayez d’autres mots/)
      .filter((el) => !el.closest('[data-astryx-live-region]'));
    expect(visible).toHaveLength(1);
    expect(screen.queryByRole('heading', { level: 2, name: 'Catalogue' })).not.toBeInTheDocument();
  });

  it('shows the wait on a 429 and recovers on retry', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    mockApi({
      'GET /search': () => {
        calls += 1;
        return calls === 1
          ? problemResponse(429, 'RATE_LIMITED', { retryAfter: 12 })
          : jsonResponse(results);
      },
    });
    await renderApp('/recherche?q=momo');
    expect(await screen.findByText(/Réessayez dans 12 s\./)).toBeInTheDocument();
    expect(screen.getByText(/Référence.*req-test-1/)).toBeInTheDocument();
    await userEvent.setup({ delay: null }).click(screen.getByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('link', { name: 'API Mobile Money' })).toBeInTheDocument();
  });

  it('offers a retry after a network failure', async () => {
    setLanguages(['en-US']);
    let calls = 0;
    mockApi({
      'GET /search': () => {
        calls += 1;
        return calls === 1 ? Promise.reject(new TypeError('offline')) : jsonResponse(results);
      },
    });
    await renderApp('/recherche?q=momo');
    expect(await screen.findByText('Search failed')).toBeInTheDocument();
    await userEvent.setup({ delay: null }).click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByRole('link', { name: 'API Mobile Money' })).toBeInTheDocument();
    expect(calls).toBe(2);
  });

  it('puts the submitted query in the URL and searches for it', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /search': (request) =>
        jsonResponse({ ...results, query: new URL(request.url).searchParams.get('q') ?? '' }),
    });
    const { router } = await renderApp('/recherche');
    const box = await screen.findByRole('searchbox', { name: 'Votre recherche' });
    await userEvent.setup({ delay: null }).type(box, '  orange money  ');
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Rechercher' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({ q: 'orange money' });
    });
    await screen.findByRole('link', { name: 'API Mobile Money' });
    expect(
      callsTo(fetchMock, 'GET /search').map((r) => new URL(r.url).searchParams.get('q')),
    ).toEqual(['orange money']);

    act(() => {
      router.history.back();
    });
    await waitFor(() => {
      expect(router.state.location.search).toEqual({});
    });
    expect(await screen.findByRole('heading', { name: /Que cherchez-vous/ })).toBeInTheDocument();
    expect(screen.getByRole('searchbox', { name: 'Votre recherche' })).toHaveValue('');
  });

  it('speaks English', async () => {
    setLanguages(['en-US']);
    mockApi(searchRoute(results));
    await renderApp('/recherche?q=momo');
    expect(await screen.findByRole('heading', { level: 2, name: 'Learn' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Projects' })).toBeInTheDocument();
    expect(screen.getByText(/There are more results/)).toBeInTheDocument();
    expect(await screen.findByText(/Sign in to ask the assistant/)).toBeInTheDocument();
  });
});
