import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import type { LearnSearch } from './routeSearch';

export type ArticleSummary = components['schemas']['ArticleSummary'];
export type Article = components['schemas']['Article'];
export type SeriesSummary = components['schemas']['SeriesSummary'];
export type TrainingSeries = components['schemas']['TrainingSeries'];

/** One key factory; every variable a query function reads is in its key. */
export const learnKeys = {
  articles: (search: LearnSearch) => ['learn', 'articles', search] as const,
  article: (id: string) => ['learn', 'article', id] as const,
  seriesList: (search: LearnSearch) => ['learn', 'series', search] as const,
  series: (id: string) => ['learn', 'series-detail', id] as const,
};

function retryOnce(failures: number, error: unknown): boolean {
  return failures < 1 && error instanceof ApiError && error.code === 'NETWORK';
}

export function useArticleList(search: LearnSearch) {
  return useInfiniteQuery({
    queryKey: learnKeys.articles(search),
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = {
        ...(search.topic ? { topic: search.topic } : {}),
        ...(pageParam === null ? {} : { cursor: pageParam }),
      };
      const page = await api.call(() => api.client.GET('/learn/articles', { params: { query } }));
      return { items: page.items, nextCursor: page.next_cursor };
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}

export function useSeriesList(search: LearnSearch) {
  return useInfiniteQuery({
    queryKey: learnKeys.seriesList(search),
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = {
        ...(search.topic ? { topic: search.topic } : {}),
        ...(pageParam === null ? {} : { cursor: pageParam }),
      };
      const page = await api.call(() => api.client.GET('/learn/series', { params: { query } }));
      return { items: page.items, nextCursor: page.next_cursor };
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
