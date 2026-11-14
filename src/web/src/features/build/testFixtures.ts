import type { components } from '../../api/schema';
import { memberAccount, memberSession, mockApi } from '../../test/api';
import { jsonResponse } from '../../test/render';

export type Project = components['schemas']['Project'];
export type Need = components['schemas']['ContributionNeed'];
export type Application = components['schemas']['Application'];

export const projectId = 'e0000000-0000-4000-8000-000000000001';
export const needId = 'e1000000-0000-4000-8000-000000000001';
export const applicationId = 'e2000000-0000-4000-8000-000000000001';
export const ownerId = 'e3000000-0000-4000-8000-0000000000aa';

/** Contract-shaped project owned by someone other than the test member; synthetic data only. */
export function project(overrides: Partial<Project> = {}): Project {
  return {
    id: projectId,
    name: 'kora-sdk',
    github_url: 'https://github.com/kora-dev/kora-sdk',
    description: 'Un SDK pour les paiements mobiles.',
    lang: 'fr',
    technologies: ['go', 'mobile-money'],
    visibility: 'public',
    owner: { id: ownerId, username: 'kofi', display_name: 'Kofi Mensah' },
    open_needs_count: 1,
    related: [],
    created_at: '2026-11-10T09:00:00Z',
    updated_at: '2026-11-10T09:00:00Z',
    ...overrides,
  };
}

export function need(overrides: Partial<Need> = {}): Need {
  return {
    id: needId,
    project_id: projectId,
    title: 'Tests d’intégration',
    description: 'Écrire des tests contre le bac à sable.',
    lang: 'fr',
    skills: ['go', 'tests'],
    status: 'open',
    created_at: '2026-11-10T10:00:00Z',
    ...overrides,
  };
}

export function application(overrides: Partial<Application> = {}): Application {
  return {
    id: applicationId,
    need_id: needId,
    project_id: projectId,
    applicant: { id: memberAccount.id, username: 'awa', display_name: 'Awa Traoré' },
    message: 'Je connais bien ce domaine.',
    status: 'pending',
    created_at: '2026-11-11T09:00:00Z',
    decided_at: null,
    ...overrides,
  };
}

export function page(items: unknown[], next: string | null = null) {
  return jsonResponse({ items, next_cursor: next });
}

export function member(routes = {}) {
  return mockApi({ 'GET /auth/session': () => jsonResponse(memberSession), ...routes });
}

/** A signed-in member who owns the project returned by `project()`. */
export function owner(routes = {}) {
  return mockApi({
    'GET /auth/session': () =>
      jsonResponse({
        ...memberSession,
        user: { ...memberAccount, id: ownerId, username: 'kofi', display_name: 'Kofi Mensah' },
      }),
    ...routes,
  });
}

export { memberAccount };
