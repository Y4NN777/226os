import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { useIdempotencyKey } from '../solve/idempotency';

export type Project = components['schemas']['Project'];
export type ProjectCreate = components['schemas']['ProjectCreate'];
export type ProjectUpdate = components['schemas']['ProjectUpdate'];
export type Need = components['schemas']['ContributionNeed'];
export type NeedCreate = components['schemas']['NeedCreate'];
export type Application = components['schemas']['Application'];
export type ApplicationStatus = components['schemas']['ApplicationStatus'];

/** One key factory; every variable a query function reads is in its key. */
export const buildKeys = {
  list: ['projects', 'list'] as const,
  detail: (id: string) => ['projects', 'detail', id] as const,
  needs: (id: string) => ['projects', 'needs', id] as const,
  applications: (id: string) => ['projects', 'applications', id] as const,
};

function retryOnce(failures: number, error: unknown): boolean {
  return failures < 1 && error instanceof ApiError && error.code === 'NETWORK';
}

export function useProjectList() {
  return useInfiniteQuery({
    queryKey: buildKeys.list,
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = pageParam === null ? {} : { cursor: pageParam };
      const page = await api.call(() => api.client.GET('/projects', { params: { query } }));
      return { items: page.items, nextCursor: page.next_cursor };
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}

export function useProject(id: string) {
  return useQuery({
    queryKey: buildKeys.detail(id),
    retry: retryOnce,
    queryFn: () =>
      api.call(() =>
        api.client.GET('/projects/{projectId}', { params: { path: { projectId: id } } }),
      ),
  });
}

export function useNeeds(id: string) {
  return useInfiniteQuery({
    queryKey: buildKeys.needs(id),
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = pageParam === null ? {} : { cursor: pageParam };
      const page = await api.call(() =>
        api.client.GET('/projects/{projectId}/contribution-needs', {
          params: { path: { projectId: id }, query },
        }),
      );
      return { items: page.items, nextCursor: page.next_cursor };
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}

/** Only mounted for the project owner: a third party never asks, so nothing can leak. */
export function useApplications(id: string) {
  return useInfiniteQuery({
    queryKey: buildKeys.applications(id),
    retry: retryOnce,
    initialPageParam: null as string | null,
    queryFn: async ({ pageParam }) => {
      const query = pageParam === null ? {} : { cursor: pageParam };
      const page = await api.call(() =>
        api.client.GET('/projects/{projectId}/applications', {
          params: { path: { projectId: id }, query },
        }),
      );
      return { items: page.items, nextCursor: page.next_cursor };
    },
    getNextPageParam: (last) => last.nextCursor,
  });
}

const PROJECT_PATH =
  /^\/api\/v1\/projects\/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$/i;

/** The project id in a Location header, only when it points at a project of this very API. */
function projectIdOf(location: string | null): string | undefined {
  if (location === null) return undefined;
  try {
    const url = new URL(location, globalThis.location.origin);
    if (url.origin !== globalThis.location.origin) return undefined;
    return PROJECT_PATH.exec(url.pathname)?.[1];
  } catch {
    return undefined;
  }
}

/**
 * POST /projects with one Idempotency-Key per submission, kept across retries (a 503 from GitHub
 * is retried with the same key). A replay too large for the store answers 201 with
 * `Idempotency-Body-Omitted: true` and no body: the id then comes from `Location`.
 */
export function useCreateProject() {
  const queryClient = useQueryClient();
  const { keyFor, settle } = useIdempotencyKey();
  return useMutation({
    mutationFn: async (body: ProjectCreate): Promise<{ id: string }> => {
      const key = keyFor(body);
      let response: Response | undefined;
      const project = await api.call(async () => {
        const result = await api.client.POST('/projects', {
          body,
          params: { header: { 'Idempotency-Key': key } },
        });
        response = result.response;
        return result;
      });
      settle();
      if (response?.headers.get('Idempotency-Body-Omitted') !== 'true') return { id: project.id };
      const id = projectIdOf(response.headers.get('Location'));
      if (id === undefined) throw new ApiError({ status: response.status, code: 'UNKNOWN' });
      return { id };
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: buildKeys.list }),
  });
}

export function useUpdateProject(projectId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: ProjectUpdate) =>
      api.call(() =>
        api.client.PATCH('/projects/{projectId}', { params: { path: { projectId } }, body }),
      ),
    onSuccess: (project) => {
      queryClient.setQueryData(buildKeys.detail(projectId), project);
      return queryClient.invalidateQueries({ queryKey: buildKeys.list });
    },
  });
}

export function useCreateNeed(projectId: string) {
  const queryClient = useQueryClient();
  const { keyFor, settle } = useIdempotencyKey();
  return useMutation({
    mutationFn: (body: NeedCreate) => {
      const key = keyFor(body);
      return api.call(() =>
        api.client.POST('/projects/{projectId}/contribution-needs', {
          params: { path: { projectId }, header: { 'Idempotency-Key': key } },
          body,
        }),
      );
    },
    onSuccess: () => {
      settle();
      return Promise.all([
        queryClient.invalidateQueries({ queryKey: buildKeys.needs(projectId) }),
        queryClient.invalidateQueries({ queryKey: buildKeys.detail(projectId) }),
      ]);
    },
  });
}

export function useCreateApplication(needId: string) {
  const { keyFor, settle } = useIdempotencyKey();
  return useMutation({
    mutationFn: async (message: string) => {
      const body = { message };
      const key = keyFor(body);
      const created = await api.call(() =>
        api.client.POST('/contribution-needs/{needId}/applications', {
          params: { path: { needId }, header: { 'Idempotency-Key': key } },
          body,
        }),
      );
      settle();
      return created;
    },
  });
}

export function useDecideApplication(projectId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'accepted' | 'declined' | 'withdrawn' }) =>
      api.call(() =>
        api.client.PATCH('/applications/{applicationId}', {
          params: { path: { applicationId: id } },
          body: { status },
        }),
      ),
    onSettled: () => queryClient.invalidateQueries({ queryKey: buildKeys.applications(projectId) }),
  });
}
