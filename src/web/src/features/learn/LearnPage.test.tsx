import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { articleId, page, secondArticleId, seriesId, seriesSummary, summary } from './testFixtures';

const second = summary({
  id: secondArticleId,
  title: 'Mobile Money et faible débit',
  type: 'tutorial',
  tags: [],
});

const member = { 'GET /auth/session': () => jsonResponse(memberSession) };

function params(request: Request | undefined) {
  return new URL(request?.url ?? 'http://x').searchParams;
}

const base = {
  'GET /learn/articles': () => page([summary()]),
  'GET /learn/series': () => page([seriesSummary()]),
};

describe('LearnPage list', () => {
  it('lists articles and training series as links, in French', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/apprendre');

    expect(await screen.findByRole('heading', { level: 1, name: 'Apprendre' })).toBeInTheDocument();
    const article = await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    expect(article).toHaveAttribute('href', `/apprendre/articles/${articleId}`);
    const formation = await screen.findByRole('link', {
      name: 'Premiers pas avec une API de paiement',
    });
    expect(formation).toHaveAttribute('href', `/apprendre/formations/${seriesId}`);
    const articles = screen.getByRole('region', { name: 'Articles' });
    expect(within(articles).getByText('Par Fatou Diallo')).toBeInTheDocument();
    expect(within(articles).getByText('webhook')).toBeInTheDocument();
    const formations = screen.getByRole('region', { name: 'Formations' });
    expect(within(formations).getByText('Leçons : 2')).toBeInTheDocument();
  });

  it('is a live link in the navigation', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/');
    const nav = await screen.findByRole('navigation', { name: 'Navigation principale' });
    expect(within(nav).getByRole('link', { name: 'Apprendre' })).toHaveAttribute(
      'href',
      '/apprendre',
    );
  });

  it('loads the next page of articles with the opaque cursor', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      ...base,
      'GET /learn/articles': (request) =>
        params(request).get('cursor') === 'cur-2' ? page([second]) : page([summary()], 'cur-2'),
    });
    const user = userEvent.setup({ delay: null });
    await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    await user.click(screen.getByRole('button', { name: 'Afficher plus d’articles' }));

    expect(await screen.findByRole('link', { name: 'Mobile Money et faible débit' })).toBeVisible();
    expect(callsTo(fetchMock, 'GET /learn/articles')).toHaveLength(2);
    expect(
      screen.queryByRole('button', { name: 'Afficher plus d’articles' }),
    ).not.toBeInTheDocument();
  });

  it('loads the next page of series with the opaque cursor', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      ...base,
      'GET /learn/series': (request) =>
        params(request).get('cursor') === 'cur-s'
          ? page([
              seriesSummary({
                id: 'e0000000-0000-4000-8000-000000000002',
                title: 'Déployer sur un VPS',
              }),
            ])
          : page([seriesSummary()], 'cur-s'),
    });
    const user = userEvent.setup({ delay: null });
    await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Premiers pas avec une API de paiement' });
    await user.click(screen.getByRole('button', { name: 'Afficher plus de formations' }));
    expect(await screen.findByRole('link', { name: 'Déployer sur un VPS' })).toBeVisible();
    expect(callsTo(fetchMock, 'GET /learn/series')).toHaveLength(2);
  });

  it('sends the chosen topic code to both lists and keeps it in the route search', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });

    await user.click(screen.getByRole('combobox', { name: 'Sujet' }));
    await user.click(await screen.findByRole('option', { name: 'Paiements et Mobile Money' }));

    await waitFor(() => {
      expect(params(callsTo(fetchMock, 'GET /learn/articles').at(-1)).get('topic')).toBe(
        'payments',
      );
      expect(params(callsTo(fetchMock, 'GET /learn/series').at(-1)).get('topic')).toBe('payments');
    });
    expect(router.state.location.search).toMatchObject({ topic: 'payments' });
  });

  it('reads the topic from the address', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    await renderApp('/apprendre?topic=mobile');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    expect(params(callsTo(fetchMock, 'GET /learn/articles')[0]).get('topic')).toBe('mobile');
  });

  it('drops an unknown topic from the address', async () => {
    setLanguages(['fr-FR']);
    const other = mockApi(base);
    await renderApp('/apprendre?topic=nope');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    expect(params(callsTo(other, 'GET /learn/articles')[0]).has('topic')).toBe(false);
  });

  it('clears the topic filter with its reset button', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/apprendre?topic=mobile');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    await user.click(screen.getByRole('button', { name: 'Tous les sujets' }));
    await waitFor(() => {
      expect(router.state.location.search).not.toHaveProperty('topic');
    });
  });

  it('says so when no article matches', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /learn/articles': () => page([]), 'GET /learn/series': () => page([]) });
    await renderApp('/apprendre');
    expect(await screen.findByText('Aucun article pour le moment')).toBeVisible();
    expect(screen.queryByRole('region', { name: 'Formations' })).not.toBeInTheDocument();
  });

  it('shows a readable error with a retry when the articles fail to load', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    mockApi({
      ...base,
      'GET /learn/articles': () => {
        calls += 1;
        return calls === 1 ? problemResponse(500, 'INTERNAL_ERROR') : page([summary()]);
      },
    });
    const user = userEvent.setup({ delay: null });
    await renderApp('/apprendre');
    expect(await screen.findByText('Impossible de charger le contenu')).toBeVisible();
    await user.click(screen.getByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('link', { name: 'Vérifier un webhook signé' })).toBeVisible();
  });

  it('shows an error when the series fail to load, articles stay visible', async () => {
    setLanguages(['fr-FR']);
    mockApi({ ...base, 'GET /learn/series': () => problemResponse(500, 'INTERNAL_ERROR') });
    await renderApp('/apprendre');
    expect(await screen.findByRole('link', { name: 'Vérifier un webhook signé' })).toBeVisible();
    expect(await screen.findByText('Impossible de charger le contenu')).toBeVisible();
  });

  it('offers writing to a member only, and in English when asked', async () => {
    setLanguages(['en-US']);
    mockApi({ ...base, ...member });
    await renderApp('/apprendre');
    expect(await screen.findByRole('link', { name: 'Write an article' })).toHaveAttribute(
      'href',
      '/ecrire',
    );
    expect(screen.getByRole('heading', { level: 1, name: 'Learn' })).toBeInTheDocument();
  });

  it('does not offer writing to a visitor', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    expect(screen.queryByRole('link', { name: 'Écrire un article' })).not.toBeInTheDocument();
  });
});
