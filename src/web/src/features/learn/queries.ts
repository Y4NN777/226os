import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { api, ApiError } from '../../api/http';
import type { components, operations } from '../../api/schema';
import type { LearnSearch } from './routeSearch';

export type ArticleSummary = components['schemas']['ArticleSummary'];
export type Article = components['schemas']['Article'];
export type SeriesSummary = components['schemas']['SeriesSummary'];
export type TrainingSeries = components['schemas']['TrainingSeries'];

type ArticleQuery = NonNullable<operations['listLearnArticles']['parameters']['query']>;
type SeriesQuery = NonNullable<operations['listTrainingSeries']['parameters']['query']>;

/** The URL filters listLearnArticles serves; the series chip has no article counterpart. */
export type ArticleFilters = Omit<ArticleQuery, 'cursor' | 'limit'>;
/** The URL filters listTrainingSeries serves: no text search, no type. */
export type SeriesFilters = Omit<SeriesQuery, 'cursor' | 'limit'>;

/** Only the set filters, so an unfiltered key is `{}` and two equal filter sets share a cache entry. */
export function seriesFilters(search: LearnSearch): SeriesFilters {
  return {
    ...(search.topic ? { topic: search.topic } : {}),
    ...(search.level ? { level: search.level } : {}),
    ...(search.lang ? { lang: search.lang } : {}),
  };
}

export function articleFilters(search: LearnSearch): ArticleFilters {
  return {
    ...(search.q ? { q: search.q } : {}),
    ...(search.type && search.type !== 'series' ? { type: search.type } : {}),
    ...seriesFilters(search),
  };
}

/** One key factory; every variable a query function reads is in its key. */
export const learnKeys = {
  articles: (filters: ArticleFilters) => ['learn', 'articles', filters] as const,
  article: (id: string) => ['learn', 'article', id] as const,
  seriesList: (filters: SeriesFilters) => ['learn', 'series', filters] as const,
  series: (id: string) => ['learn', 'series-detail', id] as const,
};

function retryOnce(failures: number, error: unknown): boolean {
  return failures < 1 && error instanceof ApiError && error.code === 'NETWORK';
}

/** A filter change is a new key, so the list starts again at the first page (no stale cursor). */
export function useArticleList(filters: ArticleFilters, enabled = true) {
  return useInfiniteQuery({
    queryKey: learnKeys.articles(filters),
    enabled,
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = { ...filters, ...(pageParam === null ? {} : { cursor: pageParam }) };
      const page = await api.call(() => api.client.GET('/learn/articles', { params: { query } }));
      return { items: page.items, nextCursor: page.next_cursor, total: page.total };
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}

export function useSeriesList(filters: SeriesFilters, enabled = true) {
  return useInfiniteQuery({
    queryKey: learnKeys.seriesList(filters),
    enabled,
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = { ...filters, ...(pageParam === null ? {} : { cursor: pageParam }) };
      const page = await api.call(() => api.client.GET('/learn/series', { params: { query } }));
      return { items: page.items, nextCursor: page.next_cursor, total: page.total };
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}

export function useArticle(id: string) {
  return useQuery({
    queryKey: learnKeys.article(id),
    retry: retryOnce,
    queryFn: () =>
      api.call(() =>
        api.client.GET('/learn/articles/{articleId}', { params: { path: { articleId: id } } }),
      ),
  });
}

export function useSeries(id: string) {
  return useQuery({
    queryKey: learnKeys.series(id),
    retry: retryOnce,
    queryFn: () =>
      api.call(() =>
        api.client.GET('/learn/series/{seriesId}', { params: { path: { seriesId: id } } }),
      ),
  });
}

/** Read-only: the questions tagged with a topic code (E4, no new table). */
export function useTopicQuestions(topic: string) {
  return useQuery({
    queryKey: ['questions', 'list', 'topic', topic] as const,
    retry: retryOnce,
    queryFn: () =>
      api.call(() => api.client.GET('/questions', { params: { query: { tag: topic, limit: 5 } } })),
  });
}
