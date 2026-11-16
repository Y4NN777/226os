import { useCallback } from 'react';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { useIdempotencyKey } from './idempotency';

type ArticleCreate = components['schemas']['ArticleCreate'];

const ARTICLE_PATH =
  /^\/api\/v1\/learn\/articles\/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$/i;

/** The article id in a Location header, only when it points at an article of this very API. */
function articleIdOf(location: string | null): string | undefined {
  if (location === null) return undefined;
  try {
    const url = new URL(location, globalThis.location.origin);
    if (url.origin !== globalThis.location.origin) return undefined;
    return ARTICLE_PATH.exec(url.pathname)?.[1];
  } catch {
    return undefined;
  }
}

/**
 * POST /learn/articles with one Idempotency-Key per submission, kept across retries. A replay of a
 * body too large for the idempotency store answers 201 with `Idempotency-Body-Omitted: true` and no
 * body: the article is then read through `Location`, never from the empty response.
 */
export function useCreateArticle() {
  const { keyFor, settle } = useIdempotencyKey();
  return useCallback(
    async (draft: ArticleCreate): Promise<{ id: string }> => {
      // Outside the closure: `api.call` runs it again after a CSRF refresh, and the key must not change.
      const key = keyFor(draft);
      let response: Response | undefined;
      const article = await api.call(async () => {
        const result = await api.client.POST('/learn/articles', {
          body: draft,
          params: { header: { 'Idempotency-Key': key } },
        });
        response = result.response;
        return result;
      });

      if (response?.headers.get('Idempotency-Body-Omitted') !== 'true') {
        settle();
        return { id: article.id };
      }
      const id = articleIdOf(response.headers.get('Location'));
      if (id === undefined) throw new ApiError({ status: response.status, code: 'UNKNOWN' });
      const found = await api.call(() =>
        api.client.GET('/learn/articles/{articleId}', { params: { path: { articleId: id } } }),
      );
      settle();
      return { id: found.id };
    },
    [keyFor, settle],
  );
}
