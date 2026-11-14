import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
  type QueryClient,
} from '@tanstack/react-query';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import type { SolveSearch } from './routeSearch';

export type Question = components['schemas']['Question'];
export type Answer = components['schemas']['Answer'];
export type QuestionCreate = components['schemas']['QuestionCreate'];
export type QuestionUpdate = components['schemas']['QuestionUpdate'];
export type AnswerCreate = components['schemas']['AnswerCreate'];
export type VoteState = components['schemas']['VoteState'];
export type VoteTarget = components['schemas']['VoteTargetType'];

/** One key factory; every variable a query function reads is in its key. */
export const solveKeys = {
  list: (search: SolveSearch) => ['questions', 'list', search] as const,
  detail: (id: string) => ['questions', 'detail', id] as const,
  answers: (id: string) => ['questions', 'answers', id] as const,
  interactions: ['questions', 'interactions'] as const,
};

function retryOnce(failures: number, error: unknown): boolean {
  return failures < 1 && error instanceof ApiError && error.code === 'NETWORK';
}

export function useQuestionList(search: SolveSearch) {
  return useInfiniteQuery({
    queryKey: solveKeys.list(search),
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = {
        ...(search.q ? { q: search.q } : {}),
        ...(search.tag ? { tag: search.tag } : {}),
        ...(pageParam === null ? {} : { cursor: pageParam }),
      };
      const page = await api.call(() => api.client.GET('/questions', { params: { query } }));
      return { items: page.items, nextCursor: page.next_cursor };
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}

/** The question and the ETag it came with, which an edit must send back in If-Match. */
export interface QuestionWithTag {
  question: Question;
  etag: string | null;
}

export function useQuestion(id: string) {
  return useQuery({
    queryKey: solveKeys.detail(id),
    retry: retryOnce,
    queryFn: async (): Promise<QuestionWithTag> => {
      let etag: string | null = null;
      const question = await api.call(async () => {
        const result = await api.client.GET('/questions/{questionId}', {
          params: { path: { questionId: id } },
        });
        etag = result.response.headers.get('ETag');
        return result;
      });
      return { question, etag };
    },
  });
}

export function useAnswers(id: string) {
  return useInfiniteQuery({
    queryKey: solveKeys.answers(id),
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = pageParam === null ? {} : { cursor: pageParam };
      const page = await api.call(() =>
        api.client.GET('/questions/{questionId}/answers', {
          params: { path: { questionId: id }, query },
        }),
      );
      return { items: page.items, nextCursor: page.next_cursor };
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}

export function useCreateQuestion() {
  return useMutation({
    mutationFn: ({ body, key }: { body: QuestionCreate; key: string }) =>
      api.call(() =>
        api.client.POST('/questions', { params: { header: { 'Idempotency-Key': key } }, body }),
      ),
  });
}

export function useCreateAnswer(questionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ body, key }: { body: AnswerCreate; key: string }) =>
      api.call(() =>
        api.client.POST('/questions/{questionId}/answers', {
          params: { path: { questionId }, header: { 'Idempotency-Key': key } },
          body,
        }),
      ),
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: solveKeys.answers(questionId) }),
        queryClient.invalidateQueries({ queryKey: solveKeys.detail(questionId) }),
      ]),
  });
}

/** PATCH with If-Match; the answer carries the new ETag, which replaces the cached one. */
export function useUpdateQuestion(questionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ body, etag }: { body: QuestionUpdate; etag: string }) => {
      let next: string | null = null;
      const question = await api.call(async () => {
        const result = await api.client.PATCH('/questions/{questionId}', {
          params: { path: { questionId }, header: { 'If-Match': etag } },
          body,
        });
        next = result.response.headers.get('ETag');
        return result;
      });
      return { question, etag: next } satisfies QuestionWithTag;
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(solveKeys.detail(questionId), updated);
    },
  });
}

function refresh(queryClient: QueryClient, questionId: string) {
  return Promise.all([
    queryClient.invalidateQueries({ queryKey: solveKeys.detail(questionId) }),
    queryClient.invalidateQueries({ queryKey: solveKeys.answers(questionId) }),
  ]);
}

export function useCastVote() {
  return useMutation({
    mutationFn: (vote: { target_type: VoteTarget; target_id: string; value: 1 | -1 }) =>
      api.call(() => api.client.POST('/votes', { body: vote })),
  });
}

export function useRetractVote() {
  return useMutation({
    mutationFn: (target: { target_type: VoteTarget; target_id: string }) =>
      api.call(() => api.client.DELETE('/votes', { params: { query: target } })),
  });
}

export function useAccept(questionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (answerId: string | null) =>
      answerId === null
        ? api.call(() =>
            api.client.POST('/questions/{questionId}/unaccept', {
              params: { path: { questionId } },
            }),
          )
        : api.call(() =>
            api.client.POST('/questions/{questionId}/accept', {
              params: { path: { questionId } },
              body: { answer_id: answerId },
            }),
          ),
    onSuccess: () => refresh(queryClient, questionId),
  });
}

/**
 * Votes and acceptance answer 501 until they are live. Nothing is sent to find out: the controls
 * show for members, and the first 501 hides them for the rest of the session (no error shown).
 */
export function useInteractions() {
  const queryClient = useQueryClient();
  const { data } = useQuery({
    queryKey: solveKeys.interactions,
    queryFn: () => true,
    initialData: true,
    staleTime: Infinity,
  });
  return {
    available: data,
    /** True (and the controls go away) when the failure is a 501; false for any other error. */
    absorbNotLive: (error: unknown) => {
      const notLive = error instanceof ApiError && error.status === 501;
      if (notLive) queryClient.setQueryData(solveKeys.interactions, false);
      return notLive;
    },
  };
}
