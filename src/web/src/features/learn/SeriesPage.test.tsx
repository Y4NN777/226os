import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { articleId, secondArticleId, series, seriesId } from './testFixtures';

const route = (body: unknown) => ({ [`GET /learn/series/${seriesId}`]: () => jsonResponse(body) });

describe('SeriesPage', () => {
  it('lists the lessons in order, each linking to its article', async () => {
    setLanguages(['fr-FR']);
    mockApi(route(series()));
    await renderApp(`/apprendre/formations/${seriesId}`);

    expect(
      await screen.findByRole('heading', {
        level: 1,
        name: 'Premiers pas avec une API de paiement',
      }),
    ).toBeVisible();
    const lessons = within(screen.getByRole('list', { name: 'Leçons' })).getAllByRole('listitem');
    expect(lessons).toHaveLength(2);
    const first = within(lessons[0] as HTMLElement).getByRole('link', {
      name: 'Créer le paiement',
    });
    expect(first).toHaveAttribute('href', `/apprendre/articles/${articleId}`);
    expect(within(lessons[0] as HTMLElement).getByText('10 min')).toBeVisible();
    expect(
      within(lessons[1] as HTMLElement).getByRole('link', { name: 'Recevoir le webhook' }),
    ).toHaveAttribute('href', `/apprendre/articles/${secondArticleId}`);
    expect(screen.getByText('Par Fatou Diallo')).toBeVisible();
    expect(screen.getByRole('link', { name: '@fatou' })).toHaveAttribute('href', '/profils/fatou');
  });

  it('heads the course with its facts and starts at the first lesson', async () => {
    setLanguages(['fr-FR']);
    mockApi(route(series()));
    await renderApp(`/apprendre/formations/${seriesId}`);
    const heading = await screen.findByRole('heading', { level: 1 });
    const hero = heading.closest('header') as HTMLElement;
    expect(within(hero).getByText('Formation')).toBeInTheDocument();
    expect(within(hero).getByText('2 leçons')).toBeInTheDocument();
    expect(within(hero).getByText('Débutant')).toBeInTheDocument();
    expect(within(hero).getByText('Français')).toBeInTheDocument();
    expect(within(hero).getByRole('link', { name: 'Commencer la formation' })).toHaveAttribute(
      'href',
      `/apprendre/articles/${articleId}`,
    );
  });

  it('adds up the course duration when every lesson has one', async () => {
    setLanguages(['fr-FR']);
    const base = series();
    mockApi(
      route(
        series({
          lessons: base.lessons.map((lesson) => ({ ...lesson, duration_minutes: 45 })),
        }),
      ),
    );
    await renderApp(`/apprendre/formations/${seriesId}`);
    const heading = await screen.findByRole('heading', { level: 1 });
    const hero = heading.closest('header') as HTMLElement;
    expect(within(hero).getByText('1 h 30')).toBeInTheDocument();
  });

  it('offers to ask about the first lesson, which is the lowest position, not the first listed', async () => {
    setLanguages(['fr-FR']);
    mockApi(route(series()));
    await renderApp(`/apprendre/formations/${seriesId}`);
    expect(
      await screen.findByRole('link', { name: 'Poser une question sur la première leçon' }),
    ).toHaveAttribute('href', `/solve?link=article:${articleId}`);
  });

  it('offers no question action when the series has no lesson', async () => {
    setLanguages(['en-US']);
    mockApi(route(series({ lessons: [] })));
    await renderApp(`/apprendre/formations/${seriesId}`);
    await screen.findByRole('heading', { level: 1 });
    expect(
      screen.queryByRole('link', { name: 'Ask a question about the first lesson' }),
    ).not.toBeInTheDocument();
  });

  it('renders no section for the always-empty tags and related', async () => {
    setLanguages(['fr-FR']);
    mockApi(route(series({ tags: [], related: [] })));
    await renderApp(`/apprendre/formations/${seriesId}`);
    await screen.findByRole('heading', { level: 1 });
    expect(screen.queryByRole('heading', { name: 'Contenus liés' })).not.toBeInTheDocument();
    expect(screen.queryByRole('list', { name: 'Mots-clés' })).not.toBeInTheDocument();
  });

  it('says the series does not exist on a 404', async () => {
    setLanguages(['fr-FR']);
    mockApi({ [`GET /learn/series/${seriesId}`]: () => problemResponse(404, 'NOT_FOUND') });
    await renderApp(`/apprendre/formations/${seriesId}`);
    expect(await screen.findByText('Formation introuvable')).toBeVisible();
    expect(screen.getByRole('link', { name: 'Retour à Apprendre' })).toHaveAttribute(
      'href',
      '/apprendre',
    );
  });

  it('shows a readable error with a retry on another failure', async () => {
    setLanguages(['en-US']);
    mockApi({ [`GET /learn/series/${seriesId}`]: () => problemResponse(500, 'INTERNAL_ERROR') });
    await renderApp(`/apprendre/formations/${seriesId}`);
    expect(await screen.findByText('The content could not be loaded')).toBeVisible();
  });
});
