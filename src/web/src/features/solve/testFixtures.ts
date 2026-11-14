import type { components } from '../../api/schema';
import { memberAccount, memberSession, mockApi } from '../../test/api';
import { jsonResponse } from '../../test/render';

export type Question = components['schemas']['Question'];
export type Answer = components['schemas']['Answer'];

export const questionId = 'b0000000-0000-4000-8000-000000000001';
export const otherAuthorId = 'b0000000-0000-4000-8000-0000000000aa';

/** Contract-shaped question; synthetic data only. */
export function question(overrides: Partial<Question> = {}): Question {
  return {
    id: questionId,
    title: 'Webhook non reçu après un paiement',
    body_md: 'Mon webhook **ne répond pas** depuis hier.',
    lang: 'fr',
    tags: ['webhook', 'paiement'],
    snippets: [],
    state: 'open',
    author: { id: otherAuthorId, username: 'kofi', display_name: 'Kofi Mensah' },
    answer_count: 0,
    vote_score: 2,
    viewer_vote: 0,
    accepted_answer_id: null,
    related: [],
    created_at: '2026-11-10T09:00:00Z',
    updated_at: '2026-11-10T09:00:00Z',
    ...overrides,
  };
}

export function answer(overrides: Partial<Answer> = {}): Answer {
  return {
    id: 'c0000000-0000-4000-8000-000000000001',
    question_id: questionId,
    body_md: 'Vérifiez la signature du webhook.',
    lang: 'fr',
    snippets: [],
    sources: [],
    author: { id: otherAuthorId, username: 'fatou', display_name: 'Fatou Diallo' },
    vote_score: 0,
    viewer_vote: 0,
    is_accepted: false,
    created_at: '2026-11-10T10:00:00Z',
    updated_at: '2026-11-10T10:00:00Z',
    ...overrides,
  };
}

export function page(items: unknown[], next: string | null = null) {
  return jsonResponse({ items, next_cursor: next });
}

/** A 200 that also carries the entity tag the edit sends back in If-Match. */
export function withEtag(body: unknown, etag: string, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json', ETag: etag },
  });
}

export function member(routes = {}) {
  return mockApi({ 'GET /auth/session': () => jsonResponse(memberSession), ...routes });
}

export const memberId = memberAccount.id;
