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
