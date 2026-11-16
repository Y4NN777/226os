import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/http';

/** Small previews: the home page shows four entries per section and links to the full lists. */
const LIMIT = 4;
const query = { limit: LIMIT } as const;

/** One key per section; a failing section never touches the others. */
export const homeKeys = {
  articles: ['learn', 'articles', 'home', query] as const,
  questions: ['questions', 'list', 'home', query] as const,
  resources: ['resources', 'list', query] as const,
  projects: ['projects', 'list', 'home', query] as const,
};

// The error state offers an explicit retry, so no automatic one here.
export function useHomeArticles() {
  return useQuery({
    queryKey: homeKeys.articles,
    retry: false,
    queryFn: () => api.call(() => api.client.GET('/learn/articles', { params: { query } })),
  });
}

export function useHomeQuestions() {
  return useQuery({
    queryKey: homeKeys.questions,
    retry: false,
    queryFn: () => api.call(() => api.client.GET('/questions', { params: { query } })),
  });
}

export function useHomeResources() {
  return useQuery({
    queryKey: homeKeys.resources,
    retry: false,
    queryFn: () => api.call(() => api.client.GET('/resources', { params: { query } })),
  });
}

export function useHomeProjects() {
  return useQuery({
    queryKey: homeKeys.projects,
    retry: false,
    queryFn: () => api.call(() => api.client.GET('/projects', { params: { query } })),
  });
}
