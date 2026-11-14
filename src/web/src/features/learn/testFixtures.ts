import type { components } from '../../api/schema';
import { jsonResponse } from '../../test/render';

export type Article = components['schemas']['Article'];
export type ArticleSummary = components['schemas']['ArticleSummary'];
export type SeriesSummary = components['schemas']['SeriesSummary'];
export type TrainingSeries = components['schemas']['TrainingSeries'];

export const articleId = 'd0000000-0000-4000-8000-000000000001';
export const secondArticleId = 'd0000000-0000-4000-8000-000000000002';
export const seriesId = 'e0000000-0000-4000-8000-000000000001';

const author = {
  id: 'f0000000-0000-4000-8000-000000000001',
  username: 'fatou',
  display_name: 'Fatou Diallo',
};

/** Contract-shaped article summary; synthetic data only. */
export function summary(overrides: Partial<ArticleSummary> = {}): ArticleSummary {
  return {
    id: articleId,
    type: 'article',
    title: 'Vérifier un webhook signé',
    summary: 'Calculer le HMAC et comparer en temps constant.',
    lang: 'fr',
    level: 'intermediate',
    tags: ['webhook', 'go'],
    author,
    published_at: '2026-11-12T09:00:00Z',
    ...overrides,
  };
}

export function article(overrides: Partial<Article> = {}): Article {
  return {
    ...summary(),
    body_md: 'Calculez le **HMAC** puis comparez.',
    related: [],
    status: 'published',
    updated_at: '2026-11-12T09:00:00Z',
    ...overrides,
  };
}

export function seriesSummary(overrides: Partial<SeriesSummary> = {}): SeriesSummary {
  return {
    id: seriesId,
    title: 'Premiers pas avec une API de paiement',
    summary: 'Quatre leçons courtes pour intégrer un paiement.',
    lang: 'fr',
    level: 'beginner',
    tags: [],
    author,
    lesson_count: 2,
    ...overrides,
  };
}

export function series(overrides: Partial<TrainingSeries> = {}): TrainingSeries {
  return {
    ...seriesSummary(),
    lessons: [
      {
        id: 'a0000000-0000-4000-8000-000000000002',
        position: 2,
        title: 'Recevoir le webhook',
        article_id: secondArticleId,
      },
      {
        id: 'a0000000-0000-4000-8000-000000000001',
        position: 1,
        title: 'Créer le paiement',
        article_id: articleId,
        duration_minutes: 10,
      },
    ],
    related: [],
    published_at: '2026-11-12T09:00:00Z',
    ...overrides,
  };
}

export function page(items: unknown[], next: string | null = null) {
  return jsonResponse({ items, next_cursor: next });
}
