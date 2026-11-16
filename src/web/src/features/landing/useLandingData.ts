import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/http';

/** How many recent items to read per list: enough to skip the ones the page must not show. */
const RECENT = 20;

interface Page<T> {
  items: T[];
  next_cursor?: string | null | undefined;
  total?: number | null | undefined;
}

/** What the landing keeps of a list: no demo item, and a total that does not count them. */
export interface LivePage<T> {
  items: T[];
  total: number | undefined;
}

/**
 * Demo fixtures are never shown on the landing (CLAUDE.md: demo accounts do not appear outside
 * local-dev). A list that fits in one page is counted exactly from what is kept; a longer one
 * keeps the server total.
 */
function livePage<T>(page: Page<T>, isDemo: (item: T) => boolean): LivePage<T> {
  const items = page.items.filter((item) => !isDemo(item));
  return { items, total: page.next_cursor ? (page.total ?? undefined) : items.length };
}

/** A total is shown from this many: a small number would read as a weak community. */
export const MIN_TOTAL = 10;
const shownTotal = (total: number | undefined) =>
  total !== undefined && total >= MIN_TOTAL ? total : undefined;

const byDemoAccount = (username: string) => username.startsWith('demo-');

// Every request is a public read. Failures are silent here on purpose: a block without data is
// simply not rendered (a landing page never shows a skeleton that never resolves or a zero).
export function useLandingQuestions(enabled = true) {
  return useQuery({
    enabled,
    retry: false,
    queryKey: ['questions', 'list', 'landing', { limit: RECENT }],
    queryFn: () =>
      api.call(() =>
        api.client.GET('/questions', { params: { query: { limit: RECENT, sort: 'new' } } }),
      ),
    select: (page) => livePage(page, (question) => byDemoAccount(question.author.username)),
  });
}

export function useLandingArticles(enabled = true) {
  return useQuery({
    enabled,
    retry: false,
    queryKey: ['learn', 'articles', 'landing', { limit: RECENT }],
    queryFn: () =>
      api.call(() => api.client.GET('/learn/articles', { params: { query: { limit: RECENT } } })),
    select: (page) => livePage(page, (article) => byDemoAccount(article.author.username)),
  });
}

export function useLandingProjects(enabled = true) {
  return useQuery({
    enabled,
    retry: false,
    queryKey: ['projects', 'list', 'landing', { limit: RECENT }],
    queryFn: () =>
      api.call(() =>
        api.client.GET('/projects', { params: { query: { limit: RECENT, sort: 'created' } } }),
      ),
    select: (page) => livePage(page, (project) => byDemoAccount(project.owner.username)),
  });
}

export function useLandingResources(enabled = true) {
  return useQuery({
    enabled,
    retry: false,
    queryKey: ['resources', 'list', 'landing', { limit: RECENT }],
    queryFn: () =>
      api.call(() => api.client.GET('/resources', { params: { query: { limit: RECENT } } })),
    select: (page) => livePage(page, (resource) => resource.provenance === 'demo'),
  });
}

/**
 * What the activity section and the header's « Activité » link both depend on: one source, so
 * the link never points at a section that is not rendered. `enabled` is false until the session
 * is known: a member on `/` must not trigger the landing's requests.
 */
export function useLandingActivity(enabled = true) {
  const questions = useLandingQuestions(enabled).data;
  const articles = useLandingArticles(enabled).data;
  const projects = useLandingProjects(enabled).data;
  const resources = useLandingResources(enabled).data;
  const totals = {
    questions: shownTotal(questions?.total),
    articles: shownTotal(articles?.total),
    resources: shownTotal(resources?.total),
    projects: shownTotal(projects?.total),
  };
  const latest = {
    question: questions?.items[0],
    article: articles?.items[0],
    project: projects?.items[0],
  };
  const hasTotals = Object.values(totals).some(Boolean);
  const hasLatest = Object.values(latest).some(Boolean);
  return { totals, latest, hasTotals, hasLatest, hasActivity: hasTotals || hasLatest };
}

/** How many GitHub developers the landing shows. */
export const TALENTS = 6;
const AVATAR_HOST = 'https://avatars.githubusercontent.com/';

/**
 * Public committers.top entries for Burkina Faso (ADR-0016), read like the Connect directory.
 * An avatar from any other host is dropped (privacy review, DESIGN.md), never loaded.
 */
export function useLandingTalents(enabled = true) {
  return useQuery({
    enabled,
    retry: false,
    queryKey: ['profiles', 'list', 'landing', { source: 'github', country: 'BF', limit: TALENTS }],
    queryFn: () =>
      api.call(() =>
        api.client.GET('/profiles', {
          params: { query: { source: 'github', country: 'BF', limit: TALENTS } },
        }),
      ),
    select: (page) =>
      (page.external ?? []).slice(0, TALENTS).map((developer) => ({
        ...developer,
        avatar_url: developer.avatar_url.startsWith(AVATAR_HOST) ? developer.avatar_url : null,
      })),
  });
}
