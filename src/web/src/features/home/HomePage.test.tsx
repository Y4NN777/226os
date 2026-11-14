import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { findVisibleText } from '../../test/visibleText';

type Resource = components['schemas']['Resource'];

function resource(n: string, name: string): Resource {
  return {
    id: `00000000-0000-4000-8000-00000000000${n}`,
    name,
    slug: name.toLowerCase().replace(/\s+/g, '-'),
    type: 'api',
    provenance: 'live',
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

const author = {
  id: 'a0000000-0000-4000-8000-000000000001',
  username: 'kofi',
  display_name: 'Kofi',
};

const articles = {
  items: [
    {
      id: 'a1000000-0000-4000-8000-000000000001',
      type: 'tutorial',
      title: 'Recevoir un webhook sans doublon',
      summary: 'Vérifier la signature, puis répondre vite.',
      lang: 'fr',
      level: 'intermediate',
      tags: ['webhooks'],
      author,
      published_at: '2026-11-01T09:00:00Z',
    },
  ],
  next_cursor: null,
};

const questions = {
  items: [
    {
      id: 'b1000000-0000-4000-8000-000000000001',
      title: 'Webhook non reçu après un paiement',
      body_md: 'Rien ne arrive.',
      lang: 'fr',
      tags: ['webhook', 'paiement'],
      snippets: [],
      state: 'open',
      author,
      answer_count: 0,
      vote_score: 0,
      accepted_answer_id: null,
      related: [],
      created_at: '2026-11-10T09:00:00Z',
      updated_at: '2026-11-10T09:00:00Z',
    },
  ],
  next_cursor: null,
};

const projects = {
  items: [
    {
      id: 'c1000000-0000-4000-8000-000000000001',
      name: 'momo-sdk',
      github_url: 'https://github.com/example/momo-sdk',
      description: 'Un SDK léger pour Mobile Money.',
      lang: 'fr',
      technologies: ['go', 'sdk'],
      visibility: 'public',
      owner: author,
      open_needs_count: 1,
      related: [],
      created_at: '2026-11-01T09:00:00Z',
      updated_at: '2026-11-01T09:00:00Z',
    },
  ],
  next_cursor: null,
};

type Handler = () => Response | Promise<Response>;

function signedIn(
  resources: Handler,
  others: { articles?: Handler; questions?: Handler; projects?: Handler } = {},
) {
  return mockApi({
    'GET /auth/session': () => jsonResponse(memberSession),
    'GET /resources': resources,
    'GET /learn/articles': others.articles ?? (() => jsonResponse(articles)),
    'GET /questions': others.questions ?? (() => jsonResponse(questions)),
    'GET /projects': others.projects ?? (() => jsonResponse(projects)),
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
      'Dernières questions',
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
    expect(screen.getByRole('heading', { level: 2, name: 'Latest questions' })).toBeVisible();
    expect(screen.getByRole('searchbox', { name: 'Search 226os' })).toBeVisible();
  });

  it('shows live articles, questions and projects with links, and no demo content', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = signedIn(() => jsonResponse(page));
    await renderApp('/');
    await screen.findByRole('region', { name: 'Apprendre' });
    expect(
      await within(section('Apprendre')).findByRole('link', {
        name: 'Recevoir un webhook sans doublon',
      }),
    ).toHaveAttribute('href', '/apprendre/articles/a1000000-0000-4000-8000-000000000001');
    expect(
      await within(section('Dernières questions')).findByRole('link', {
        name: 'Webhook non reçu après un paiement',
      }),
    ).toHaveAttribute('href', '/questions/b1000000-0000-4000-8000-000000000001');
    expect(
      await within(section('Projets')).findByRole('link', { name: 'momo-sdk' }),
    ).toHaveAttribute('href', '/build/c1000000-0000-4000-8000-000000000001');
    expect(screen.queryByText('Démo')).not.toBeInTheDocument();
    expect(
      within(section('Connect')).getByText('Bientôt : l’annuaire des développeurs.'),
    ).toBeVisible();
    for (const key of ['GET /learn/articles', 'GET /questions', 'GET /projects']) {
      const [request] = callsTo(fetchMock, key);
      expect(new URL(request?.url ?? '').searchParams.get('limit')).toBe('4');
    }
  });

  it.each([
    ['Apprendre', 'articles', 'Aucun article pour l’instant.'],
    ['Dernières questions', 'questions', 'Aucune question pour l’instant.'],
    ['Projets', 'projects', 'Aucun projet pour l’instant.'],
  ] as const)('says so when %s is empty', async (name, key, text) => {
    setLanguages(['fr-FR']);
    signedIn(() => jsonResponse(page), {
      [key]: () => jsonResponse({ items: [], next_cursor: null }),
    });
    await renderApp('/');
    const region = await screen.findByRole('region', { name });
    expect(await within(region).findByText(text)).toBeVisible();
  });

  it('keeps the page alive when one section fails, with a retry for it only', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    signedIn(() => jsonResponse(page), {
      questions: () => {
        calls += 1;
        return calls === 1 ? problemResponse(500, 'INTERNAL_ERROR') : jsonResponse(questions);
      },
    });
    await renderApp('/');
    const region = await screen.findByRole('region', { name: 'Dernières questions' });
    const alert = await within(region).findByRole('alert');
    expect(alert).toHaveTextContent('Impossible de charger le contenu');
    expect(await within(section('Projets')).findByRole('link', { name: 'momo-sdk' })).toBeVisible();
    expect(screen.getAllByRole('alert')).toHaveLength(1);
    await userEvent.click(within(alert).getByRole('button', { name: 'Réessayer' }));
    expect(
      await within(region).findByRole('link', { name: 'Webhook non reçu après un paiement' }),
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
    await userEvent
      .setup({ delay: null })
      .click(within(alert).getByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('link', { name: 'SMS Beta' })).toBeVisible();
  });

  it('says so when the catalogue is empty', async () => {
    setLanguages(['fr-FR']);
    signedIn(() => jsonResponse({ items: [], next_cursor: null }));
    await renderApp('/');
    expect(await findVisibleText('Le catalogue est vide pour l’instant.')).toBeVisible();
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
    await userEvent.setup({ delay: null }).type(field, '   {Enter}');
    expect(router.state.location.pathname).toBe('/');
    await userEvent.setup({ delay: null }).type(field, '  go api  {Enter}');
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
    await userEvent.setup({ delay: null }).click(field);
    await userEvent.setup({ delay: null }).paste('a'.repeat(120));
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Rechercher' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/recherche');
    });
    expect(router.state.location.search).toEqual({ q: 'a'.repeat(100) });
  });
});
