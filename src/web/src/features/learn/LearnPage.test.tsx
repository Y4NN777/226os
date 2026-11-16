import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { selectSegment } from '../../test/segmented';
import { question } from '../solve/testFixtures';
import {
  articleId,
  page,
  secondArticleId,
  series,
  seriesId,
  seriesSummary,
  summary,
} from './testFixtures';

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
    // As in the prototype, the featured training heads the page and is listed with the rest.
    const formationLinks = await screen.findAllByRole('link', {
      name: 'Premiers pas avec une API de paiement',
    });
    expect(formationLinks).toHaveLength(2);
    for (const formation of formationLinks) {
      expect(formation).toHaveAttribute('href', `/apprendre/formations/${seriesId}`);
    }
    // Dense rows as in the prototype: type badge, author, level, language, date.
    const browse = screen.getByRole('region', { name: 'Tous les contenus' });
    const row = within(browse)
      .getByRole('link', { name: 'Vérifier un webhook signé' })
      .closest('li') as HTMLElement;
    expect(within(row).getByText('Article')).toBeInTheDocument();
    expect(within(row).getByText('Fatou Diallo')).toBeInTheDocument();
    expect(within(row).getByText('Intermédiaire')).toBeInTheDocument();
    expect(within(row).getByText('Français')).toBeInTheDocument();
    expect(within(row).getByText('12 nov. 2026')).toHaveAttribute(
      'datetime',
      '2026-11-12T09:00:00Z',
    );
    expect(
      within(row).getByText('Calculer le HMAC et comparer en temps constant.'),
    ).toBeInTheDocument();
    // The featured training is its own region, named by its title (prototype `.ln-feature`).
    const formations = screen.getByRole('region', {
      name: 'Premiers pas avec une API de paiement',
    });
    expect(within(formations).getByText('2 leçons')).toBeInTheDocument();
    expect(within(formations).getByText('Formation de Fatou Diallo')).toBeInTheDocument();
    expect(within(formations).getByRole('link', { name: 'Voir le programme' })).toHaveAttribute(
      'href',
      `/apprendre/formations/${seriesId}`,
    );
  });

  it('shows the lessons of the featured series and starts at the first one', async () => {
    setLanguages(['fr-FR']);
    mockApi({ ...base, [`GET /learn/series/${seriesId}`]: () => jsonResponse(series()) });
    await renderApp('/apprendre');
    const steps = await screen.findByRole('group', { name: 'Au programme' });
    const lessons = within(steps).getAllByRole('link');
    expect(lessons.map((link) => link.textContent)).toEqual([
      '1Créer le paiement10 min',
      '2Recevoir le webhook',
    ]);
    expect(lessons[0]).toHaveAttribute('href', `/apprendre/articles/${articleId}`);
    expect(screen.getByRole('link', { name: 'Commencer' })).toHaveAttribute(
      'href',
      `/apprendre/articles/${articleId}`,
    );
  });

  it('lists the other series as rows with their lesson count', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      ...base,
      'GET /learn/series': () =>
        page([
          seriesSummary(),
          seriesSummary({
            id: 'e0000000-0000-4000-8000-000000000002',
            title: 'Déployer sur un VPS',
            lesson_count: 1,
          }),
        ]),
    });
    await renderApp('/apprendre');
    const link = await screen.findByRole('link', { name: 'Déployer sur un VPS' });
    const row = link.closest('li') as HTMLElement;
    expect(within(row).getByText('Formation')).toBeInTheDocument();
    expect(within(row).getByText('1 leçon')).toBeInTheDocument();
  });

  it('is a live link in the navigation', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/apprendre');
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
    await user.click(screen.getByRole('button', { name: 'Afficher plus' }));

    expect(await screen.findByRole('link', { name: 'Mobile Money et faible débit' })).toBeVisible();
    expect(callsTo(fetchMock, 'GET /learn/articles')).toHaveLength(2);
    expect(screen.queryByRole('button', { name: 'Afficher plus' })).not.toBeInTheDocument();
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
    await screen.findAllByRole('link', { name: 'Premiers pas avec une API de paiement' });
    await user.click(screen.getByRole('button', { name: 'Afficher plus' }));
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

  it('clears the topic filter with the reset button', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/apprendre?topic=mobile');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    // One reset for every filter, as in the prototype (owner decision, 15 Nov 2026).
    await user.click(screen.getByRole('button', { name: 'Effacer' }));
    await waitFor(() => {
      expect(router.state.location.search).not.toHaveProperty('topic');
    });
  });

  it('says so when nothing is published yet, without a reset', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /learn/articles': () => page([]), 'GET /learn/series': () => page([]) });
    await renderApp('/apprendre');
    // Articles and series share one list now, so the empty state speaks of both.
    expect(await screen.findByText('Aucun contenu')).toBeVisible();
    expect(screen.queryByRole('link', { name: 'Voir le programme' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Effacer/ })).not.toBeInTheDocument();
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
    // The header action and the invitation carry the same label.
    const writeLinks = await screen.findAllByRole('link', { name: 'Write an article' });
    expect(writeLinks).toHaveLength(2);
    for (const link of writeLinks) expect(link).toHaveAttribute('href', '/ecrire');
    expect(screen.getByRole('heading', { level: 1, name: 'Learn' })).toBeInTheDocument();
    const invite = screen.getByRole('region', { name: 'Share what you know' });
    expect(within(invite).getByRole('link', { name: 'Write an article' })).toHaveAttribute(
      'href',
      '/ecrire',
    );
  });

  it('does not offer writing to a visitor', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    expect(screen.queryByRole('link', { name: 'Écrire un article' })).not.toBeInTheDocument();
    expect(
      screen.queryByRole('region', { name: 'Partagez ce que vous savez' }),
    ).not.toBeInTheDocument();
  });
});

describe('LearnPage filters', () => {
  const articleCalls = (fetchMock: ReturnType<typeof mockApi>) =>
    callsTo(fetchMock, 'GET /learn/articles').map(params);
  const seriesCalls = (fetchMock: ReturnType<typeof mockApi>) =>
    callsTo(fetchMock, 'GET /learn/series').map(params);

  it('groups the type chips, the search and the three selects in one filter area', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    const browse = screen.getByRole('region', { name: 'Tous les contenus' });
    const chips = within(browse).getByRole('radiogroup', { name: 'Filtrer par type' });
    expect(
      within(chips)
        .getAllByRole('radio')
        .map((radio) => radio.textContent),
    ).toEqual(['Tout', 'Articles', 'Tutoriels', 'Formations']);
    expect(within(chips).getByRole('radio', { name: 'Tout' })).toBeChecked();
    // No counts on the chips: the API returns no totals.
    expect(within(chips).queryByText(/\d/)).not.toBeInTheDocument();
    const field = within(browse).getByRole('searchbox', { name: 'Rechercher' });
    expect(field).toHaveAttribute('placeholder', 'Un mot, un sujet');
    expect(within(browse).getByRole('combobox', { name: 'Sujet' })).toBeInTheDocument();
    expect(within(browse).getByRole('combobox', { name: 'Niveau' })).toBeInTheDocument();
    expect(within(browse).getByRole('combobox', { name: 'Langue' })).toBeInTheDocument();
    // One series and one article, no further page: a total, announced politely.
    expect(within(browse).getByText('2 contenus')).toHaveAttribute('role', 'status');
    expect(screen.queryByRole('button', { name: 'Effacer' })).not.toBeInTheDocument();
  });

  it('sends the tutorial type to the articles and leaves the series out', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      ...base,
      'GET /learn/series': () =>
        page([
          seriesSummary(),
          seriesSummary({ id: 'e0000000-0000-4000-8000-000000000002', title: 'Déployer' }),
        ]),
    });
    const { router } = await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Déployer' });
    await selectSegment(screen.getByRole('radio', { name: 'Tutoriels' }));

    await waitFor(() => {
      expect(articleCalls(fetchMock).at(-1)?.get('type')).toBe('tutorial');
    });
    expect(router.state.location.search).toEqual({ type: 'tutorial' });
    await waitFor(() => {
      expect(screen.queryByRole('link', { name: 'Déployer' })).not.toBeInTheDocument();
    });
    // The featured training stays; the series list is not asked for with a type it cannot serve.
    expect(seriesCalls(fetchMock)).toHaveLength(1);
    expect(screen.getByRole('link', { name: 'Voir le programme' })).toBeVisible();
  });

  it('lists only the series under Formations, without asking for articles or offering a search', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    await renderApp('/apprendre?type=series&q=hmac');
    const browse = await screen.findByRole('region', { name: 'Tous les contenus' });
    expect(
      await within(browse).findByRole('link', { name: 'Premiers pas avec une API de paiement' }),
    ).toBeVisible();
    expect(callsTo(fetchMock, 'GET /learn/articles')).toHaveLength(0);
    // The series endpoint has no text search, so the field is not offered and q is dropped.
    expect(within(browse).queryByRole('searchbox')).not.toBeInTheDocument();
    expect(within(browse).getByRole('radio', { name: 'Formations' })).toBeChecked();
  });

  it('searches the articles on submit and says the series are not searched', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    await user.type(screen.getByRole('searchbox', { name: 'Rechercher' }), '  webhook {Enter}');

    await waitFor(() => {
      expect(articleCalls(fetchMock).at(-1)?.get('q')).toBe('webhook');
    });
    expect(router.state.location.search).toEqual({ q: 'webhook' });
    expect(seriesCalls(fetchMock).every((call) => !call.has('q'))).toBe(true);
    expect(seriesCalls(fetchMock)).toHaveLength(1);
    expect(await screen.findByText('Articles et tutoriels uniquement.')).toBeVisible();
  });

  it('sends the level and the language to both lists, keeping the other filters', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/apprendre?topic=mobile');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });

    await user.click(screen.getByRole('combobox', { name: 'Niveau' }));
    await user.click(await screen.findByRole('option', { name: 'Débutant' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({ topic: 'mobile', level: 'beginner' });
    });
    await user.click(screen.getByRole('combobox', { name: 'Langue' }));
    await user.click(await screen.findByRole('option', { name: 'English' }));

    await waitFor(() => {
      expect(router.state.location.search).toEqual({
        topic: 'mobile',
        level: 'beginner',
        lang: 'en',
      });
    });
    await waitFor(() => {
      const last = articleCalls(fetchMock).at(-1);
      expect(last?.get('level')).toBe('beginner');
      expect(last?.get('lang')).toBe('en');
      expect(last?.get('topic')).toBe('mobile');
    });
    const series = seriesCalls(fetchMock).at(-1);
    expect(series?.get('level')).toBe('beginner');
    expect(series?.get('lang')).toBe('en');
  });

  it('reads every filter from the address and shows it in the controls', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    await renderApp('/apprendre?type=article&q=hmac&topic=mobile&level=advanced&lang=en');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    const call = articleCalls(fetchMock)[0];
    expect(Object.fromEntries(call?.entries() ?? [])).toEqual({
      type: 'article',
      q: 'hmac',
      topic: 'mobile',
      level: 'advanced',
      lang: 'en',
    });
    expect(screen.getByRole('radio', { name: 'Articles' })).toBeChecked();
    expect(screen.getByRole('searchbox', { name: 'Rechercher' })).toHaveValue('hmac');
    expect(screen.getByRole('combobox', { name: 'Niveau' })).toHaveTextContent('Avancé');
    expect(screen.getByRole('combobox', { name: 'Langue' })).toHaveTextContent('English');
  });

  it('drops unknown filter values from the address', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    const { router } = await renderApp('/apprendre?type=video&level=expert&lang=de&q=%20%20');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    expect([...(articleCalls(fetchMock)[0]?.keys() ?? [])]).toEqual([]);
    expect([...(seriesCalls(fetchMock)[0]?.keys() ?? [])]).toEqual([]);
    expect(screen.getByRole('radio', { name: 'Tout' })).toBeChecked();
    expect(screen.queryByRole('button', { name: 'Effacer' })).not.toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/apprendre');
  });

  it('resets every filter at once', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/apprendre?type=tutorial&level=beginner&lang=fr&q=go');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    await user.click(screen.getByRole('button', { name: 'Effacer' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({});
    });
    expect(await screen.findByRole('radio', { name: 'Tout' })).toBeChecked();
    expect(screen.getByRole('searchbox', { name: 'Rechercher' })).toHaveValue('');
  });

  it('starts again from the first page when a filter changes', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      ...base,
      'GET /learn/articles': (request) =>
        params(request).get('cursor') === 'cur-2'
          ? page([second], null, 5)
          : page([summary()], 'cur-2', 5),
    });
    const user = userEvent.setup({ delay: null });
    await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    // The total of both lists (5 articles + 1 training), not the two rows loaded so far.
    expect(screen.getByText('6 contenus')).toHaveAttribute('role', 'status');
    await user.click(screen.getByRole('button', { name: 'Afficher plus' }));
    await screen.findByRole('link', { name: 'Mobile Money et faible débit' });
    expect(articleCalls(fetchMock).at(-1)?.get('cursor')).toBe('cur-2');

    await selectSegment(screen.getByRole('radio', { name: 'Articles' }));
    await waitFor(() => {
      expect(articleCalls(fetchMock).at(-1)?.get('type')).toBe('article');
    });
    expect(articleCalls(fetchMock).at(-1)?.has('cursor')).toBe(false);
  });

  it('says nothing matches the search, with a reset in the empty state', async () => {
    setLanguages(['fr-FR']);
    mockApi({ ...base, 'GET /learn/articles': () => page([]) });
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/apprendre?q=kotlin');
    expect(await screen.findByText('Aucun contenu ne correspond')).toBeVisible();
    expect(
      screen.getByText(
        // Testing Library folds the no-break spaces of « » into plain spaces.
        'Essayez d’autres mots ou retirez un filtre.',
      ),
    ).toBeVisible();
    await user.click(screen.getByRole('button', { name: 'Effacer les filtres' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({});
    });
  });

  it('names the type in the empty state of a type filter', async () => {
    setLanguages(['fr-FR']);
    mockApi({ ...base, 'GET /learn/series': () => page([]) });
    await renderApp('/apprendre?type=series&level=advanced');
    expect(await screen.findByText('Aucune formation ne correspond')).toBeVisible();
    expect(screen.getByText('Retirez un filtre.')).toBeVisible();
  });

  it('names a tutorial filter in its empty state', async () => {
    setLanguages(['fr-FR']);
    mockApi({ ...base, 'GET /learn/articles': () => page([]) });
    await renderApp('/apprendre?type=tutorial');
    expect(await screen.findByText('Aucun tutoriel ne correspond')).toBeVisible();
  });

  it('lists the featured series again when a filter matches it', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/apprendre?level=beginner');
    const browse = await screen.findByRole('region', { name: 'Tous les contenus' });
    expect(
      await within(browse).findByRole('link', { name: 'Premiers pas avec une API de paiement' }),
    ).toBeVisible();
    expect(within(browse).getByText('2 contenus')).toHaveAttribute('role', 'status');
  });

  it('speaks English', async () => {
    setLanguages(['en-US']);
    mockApi({
      ...base,
      'GET /learn/articles': () => page([]),
      'GET /learn/series': () => page([]),
    });
    await renderApp('/apprendre?type=article&q=go');
    const chips = await screen.findByRole('radiogroup', { name: 'Filter by type' });
    expect(
      within(chips)
        .getAllByRole('radio')
        .map((radio) => radio.textContent),
    ).toEqual(['All', 'Articles', 'Tutorials', 'Courses']);
    expect(screen.getByRole('searchbox', { name: 'Search' })).toHaveAttribute(
      'placeholder',
      'A word, a topic',
    );
    expect(screen.getByRole('combobox', { name: 'Topic' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Level' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Language' })).toBeInTheDocument();
    expect(await screen.findByText('No article matches')).toBeVisible();
    expect(screen.getByText('Try other words or remove a filter.')).toBeVisible();
    expect(screen.getByRole('button', { name: 'Clear filters' })).toBeVisible();
  });
});

describe('LearnPage topic feed of questions', () => {
  const asked = question({
    id: 'b0000000-0000-4000-8000-0000000000f1',
    title: 'Quel SDK pour Wave ?',
  });

  it('lists the questions tagged with the topic code next to the articles, read-only', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ ...base, 'GET /questions': () => page([asked]) });
    await renderApp('/apprendre?topic=payments');
    const region = await screen.findByRole('region', { name: 'Questions sur ce sujet' });
    const link = await within(region).findByRole('link', { name: 'Quel SDK pour Wave ?' });
    expect(link).toHaveAttribute('href', '/questions/b0000000-0000-4000-8000-0000000000f1');
    expect(within(region).getByRole('link', { name: 'Toutes les questions' })).toHaveAttribute(
      'href',
      '/solve?tag=payments',
    );
    expect(within(region).queryByRole('button')).not.toBeInTheDocument();
    expect(params(callsTo(fetchMock, 'GET /questions')[0]).get('tag')).toBe('payments');
    const row = link.closest('li') as HTMLElement;
    expect(within(row).getByText('Sans réponse')).toBeInTheDocument();
    expect(within(row).getByText('10 nov. 2026')).toBeInTheDocument();
  });

  it('counts the answers of an answered question', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      ...base,
      'GET /questions': () => page([{ ...asked, state: 'answered', answer_count: 2 }]),
    });
    await renderApp('/apprendre?topic=payments');
    const link = await screen.findByRole('link', { name: 'Quel SDK pour Wave ?' });
    const row = link.closest('li') as HTMLElement;
    expect(within(row).getByText('Répondue')).toBeInTheDocument();
    expect(within(row).getByText('2 réponses')).toBeInTheDocument();
  });

  it('says so when no question carries the topic', async () => {
    setLanguages(['fr-FR']);
    mockApi({ ...base, 'GET /questions': () => page([]) });
    await renderApp('/apprendre?topic=payments');
    const region = await screen.findByRole('region', { name: 'Questions sur ce sujet' });
    expect(await within(region).findByText('Aucune question sur ce sujet.')).toBeVisible();
  });

  it('keeps the articles when the questions fail, with a retry for the questions', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    mockApi({
      ...base,
      'GET /questions': () => {
        calls += 1;
        return calls === 1 ? problemResponse(500, 'INTERNAL_ERROR') : page([asked]);
      },
    });
    await renderApp('/apprendre?topic=payments');
    expect(await screen.findByRole('link', { name: 'Vérifier un webhook signé' })).toBeVisible();
    const region = await screen.findByRole('region', { name: 'Questions sur ce sujet' });
    await userEvent.click(await within(region).findByRole('button', { name: 'Réessayer' }));
    expect(await within(region).findByRole('link', { name: 'Quel SDK pour Wave ?' })).toBeVisible();
  });

  it('asks for no questions without a topic', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    expect(callsTo(fetchMock, 'GET /questions')).toHaveLength(0);
    expect(
      screen.queryByRole('region', { name: 'Questions sur ce sujet' }),
    ).not.toBeInTheDocument();
  });
});
