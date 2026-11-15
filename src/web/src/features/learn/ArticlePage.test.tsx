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

  it('heads the lesson with type, level and language pills, the country and the date', async () => {
    setLanguages(['fr-FR']);
    mockApi(
      route(
        article({
          author: { ...article().author, country: 'SN' },
          published_at: '2026-11-12T09:00:00Z',
        }),
      ),
    );
    await renderApp(`/apprendre/articles/${articleId}`);
    const heading = await screen.findByRole('heading', { level: 1 });
    const head = heading.closest('header') as HTMLElement;
    expect(within(head).getByText('Article')).toBeInTheDocument();
    expect(within(head).getByText('Intermédiaire')).toBeInTheDocument();
    expect(within(head).getByText('Français')).toBeInTheDocument();
    expect(within(head).getByText('Sénégal')).toBeInTheDocument();
    expect(within(head).getByText('12 nov. 2026')).toHaveAttribute(
      'datetime',
      '2026-11-12T09:00:00Z',
    );
    expect(within(head).getByText('Calculer le HMAC et comparer en temps constant.')).toBeVisible();
  });

  it('shows code blocks with a copy button, the keywords and an outline', async () => {
    setLanguages(['fr-FR']);
    mockApi(
      route(
        article({
          body_md:
            '## Lire le corps brut\n\nLisez les octets.\n\n```go\nbody, err := io.ReadAll(r.Body)\n```\n\n## Comparer\n\nEn temps constant.',
        }),
      ),
    );
    await renderApp(`/apprendre/articles/${articleId}`);
    await screen.findByRole('heading', { level: 1, name: 'Vérifier un webhook signé' });
    // Highlighting splits the code into tokens: match the code element by its whole text.
    expect(
      screen.getAllByText(
        (_, element) =>
          element?.tagName === 'CODE' &&
          element.textContent.includes('body, err := io.ReadAll(r.Body)'),
      ).length,
    ).toBeGreaterThan(0);
    expect(screen.getByRole('button', { name: 'Copier le code' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Lire le corps brut' })).toBeVisible();
    const tags = screen.getByRole('list', { name: 'Mots-clés' });
    expect(within(tags).getByText('webhook')).toBeInTheDocument();
    const outline = screen.getByRole('navigation', { name: 'Sur cette page' });
    expect(within(outline).getByRole('link', { name: 'Comparer' })).toBeInTheDocument();
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
