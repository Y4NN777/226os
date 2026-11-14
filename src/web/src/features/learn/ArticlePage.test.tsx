import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { article, articleId } from './testFixtures';

const ids = {
  question: 'b0000000-0000-4000-8000-0000000000q1',
  resource: 'c0000000-0000-4000-8000-0000000000r1',
  project: 'c0000000-0000-4000-8000-0000000000p1',
  article: 'd0000000-0000-4000-8000-0000000000a9',
  series: 'e0000000-0000-4000-8000-0000000000s1',
};

const linked = article({
  related: [
    { type: 'question', id: ids.question, title: 'Webhook non reçu' },
    { type: 'resource', id: ids.resource, title: 'Wave Pay' },
    { type: 'project', id: ids.project, title: 'Passerelle SMS' },
    { type: 'article', id: ids.article, title: 'Signer une requête' },
    { type: 'series', id: ids.series, title: 'Paiement pas à pas' },
  ],
});

const route = (body: unknown) => ({
  [`GET /learn/articles/${articleId}`]: () => jsonResponse(body),
});

describe('ArticlePage', () => {
  it('shows the article in its original language with the byline', async () => {
    setLanguages(['en-US']);
    mockApi(route(article({ lang: 'fr' })));
    await renderApp(`/apprendre/articles/${articleId}`);

    const heading = await screen.findByRole('heading', {
      level: 1,
      name: 'Vérifier un webhook signé',
    });
    const content = heading.closest('[lang]');
    expect(content).toHaveAttribute('lang', 'fr');
    expect(within(content as HTMLElement).getByText('HMAC')).toBeInTheDocument();
    expect(screen.getByText('By Fatou Diallo')).toBeVisible();
    expect(screen.getByRole('link', { name: '@fatou' })).toHaveAttribute('href', '/profils/fatou');
  });

  it('marks English content as English inside a French interface', async () => {
    setLanguages(['fr-FR']);
    mockApi(route(article({ lang: 'en', title: 'Verifying a signed webhook' })));
    await renderApp(`/apprendre/articles/${articleId}`);
    const heading = await screen.findByRole('heading', {
      level: 1,
      name: 'Verifying a signed webhook',
    });
    expect(heading.closest('[lang]')).toHaveAttribute('lang', 'en');
    expect(screen.getByText('Par Fatou Diallo')).toBeVisible();
    expect(screen.getByRole('link', { name: '@fatou' })).toHaveAttribute('href', '/profils/fatou');
  });

  it('renders resolved links as links to their pages', async () => {
    setLanguages(['fr-FR']);
    mockApi(route(linked));
    await renderApp(`/apprendre/articles/${articleId}`);
    await screen.findByRole('heading', { level: 1, name: 'Vérifier un webhook signé' });

    expect(screen.getByRole('link', { name: 'Webhook non reçu' })).toHaveAttribute(
      'href',
      `/questions/${ids.question}`,
    );
    expect(screen.getByRole('link', { name: 'Wave Pay' })).toHaveAttribute(
      'href',
      `/discover/${ids.resource}`,
    );
    expect(screen.getByRole('link', { name: 'Passerelle SMS' })).toHaveAttribute(
      'href',
      `/build/${ids.project}`,
    );
    expect(screen.getByRole('link', { name: 'Signer une requête' })).toHaveAttribute(
      'href',
      `/apprendre/articles/${ids.article}`,
    );
    expect(screen.getByRole('link', { name: 'Paiement pas à pas' })).toHaveAttribute(
      'href',
      `/apprendre/formations/${ids.series}`,
    );
  });

  it('has no linked-content section when there is no link', async () => {
    setLanguages(['fr-FR']);
    mockApi(route(article()));
    await renderApp(`/apprendre/articles/${articleId}`);
    await screen.findByRole('heading', { level: 1, name: 'Vérifier un webhook signé' });
    expect(screen.queryByRole('heading', { name: 'Contenus liés' })).not.toBeInTheDocument();
  });

  it('opens Résoudre with the article as the link of a new question', async () => {
    setLanguages(['fr-FR']);
    mockApi(route(article()));
    await renderApp(`/apprendre/articles/${articleId}`);
    expect(
      await screen.findByRole('link', { name: 'Poser une question sur cette leçon' }),
    ).toHaveAttribute('href', `/solve?link=article:${articleId}`);
  });

  it('says the article does not exist on a 404', async () => {
    setLanguages(['fr-FR']);
    mockApi({ [`GET /learn/articles/${articleId}`]: () => problemResponse(404, 'NOT_FOUND') });
    await renderApp(`/apprendre/articles/${articleId}`);
    expect(await screen.findByText('Article introuvable')).toBeVisible();
    expect(screen.getByRole('link', { name: 'Retour aux articles' })).toHaveAttribute(
      'href',
      '/apprendre',
    );
  });

  it('shows a readable error with a retry on another failure', async () => {
    setLanguages(['fr-FR']);
    mockApi({ [`GET /learn/articles/${articleId}`]: () => problemResponse(500, 'INTERNAL_ERROR') });
    await renderApp(`/apprendre/articles/${articleId}`);
    expect(await screen.findByText('Impossible de charger le contenu')).toBeVisible();
    expect(screen.getByRole('button', { name: 'Réessayer' })).toBeVisible();
  });
});
