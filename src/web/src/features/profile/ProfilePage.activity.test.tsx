import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { page as projectPage, project } from '../build/testFixtures';
import { page as questionPage, question } from '../solve/testFixtures';

type PublicProfile = components['schemas']['PublicProfile'];

const profile: PublicProfile = {
  username: 'awa',
  display_name: 'Awa Traoré',
  bio: 'Je construis des outils de paiement.',
  country: 'BF',
  languages: ['fr'],
  skills: ['go'],
  github: null,
  open_to_collab: true,
  profile_visibility: 'public',
  joined_at: '2026-11-01T10:00:00Z',
};

const asked = question({ title: 'Webhook non reçu' });
const answered = question({
  id: 'b0000000-0000-4000-8000-000000000002',
  title: 'Erreur CORS depuis Go',
});
const built = project({ name: 'kora-sdk' });

function routes(overrides = {}) {
  return {
    'GET /profiles/awa': () => jsonResponse(profile),
    'GET /questions': (request: Request) => {
      const params = new URL(request.url).searchParams;
      if (params.get('author') === 'awa') return questionPage([asked], null, 7);
      if (params.get('answers_by') === 'awa') return questionPage([answered], null, 3);
      return questionPage([]);
    },
    'GET /projects': () => projectPage([built], null, 2),
    ...overrides,
  };
}

describe('ProfilePage activity tabs', () => {
  it('shows Questions, Réponses and Projets with their totals', async () => {
    setLanguages(['fr-FR']);
    mockApi(routes());
    await renderApp('/profils/awa');
    const tabs = await screen.findByRole('tablist', { name: 'Activité' });
    await waitFor(() => {
      expect(within(tabs).getByRole('tab', { name: /^Questions\s*7$/ })).toBeVisible();
    });
    expect(within(tabs).getByRole('tab', { name: /^Réponses\s*3$/ })).toBeVisible();
    expect(within(tabs).getByRole('tab', { name: /^Projets\s*2$/ })).toBeVisible();
    expect(within(tabs).getByRole('tab', { name: /^Questions/ })).toHaveAttribute(
      'aria-selected',
      'true',
    );
  });

  it('asks for the questions the member wrote, answered and owns, by username', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(routes());
    await renderApp('/profils/awa');
    await screen.findByRole('link', { name: 'Webhook non reçu' });
    const questionCalls = callsTo(fetchMock, 'GET /questions').map(
      (call) => new URL(call.url).searchParams,
    );
    expect(questionCalls.some((params) => params.get('author') === 'awa')).toBe(true);
    expect(questionCalls.some((params) => params.get('answers_by') === 'awa')).toBe(true);
    const projectCalls = callsTo(fetchMock, 'GET /projects');
    expect(new URL(projectCalls[0]?.url ?? 'http://x').searchParams.get('owner')).toBe('awa');
  });

  it('switches panels: questions asked, questions answered, projects', async () => {
    setLanguages(['fr-FR']);
    mockApi(routes());
    const user = userEvent.setup({ delay: null });
    await renderApp('/profils/awa');
    expect(await screen.findByRole('link', { name: 'Webhook non reçu' })).toBeVisible();
    await user.click(screen.getByRole('tab', { name: /^Réponses/ }));
    expect(await screen.findByRole('link', { name: 'Erreur CORS depuis Go' })).toBeVisible();
    expect(screen.queryByRole('link', { name: 'Webhook non reçu' })).not.toBeInTheDocument();
    await user.click(screen.getByRole('tab', { name: /^Projets/ }));
    expect(await screen.findByRole('link', { name: 'kora-sdk' })).toHaveAttribute(
      'href',
      `/build/${built.id}`,
    );
  });

  it('loads the next page of a list with its cursor', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(
      routes({
        'GET /questions': (request: Request) => {
          const params = new URL(request.url).searchParams;
          if (params.get('author') !== 'awa') return questionPage([]);
          return params.get('cursor') === 'c2'
            ? questionPage([answered], null, 2)
            : questionPage([asked], 'c2', 2);
        },
      }),
    );
    const user = userEvent.setup({ delay: null });
    await renderApp('/profils/awa');
    await screen.findByRole('link', { name: 'Webhook non reçu' });
    await user.click(screen.getByRole('button', { name: 'Afficher plus' }));
    expect(await screen.findByRole('link', { name: 'Erreur CORS depuis Go' })).toBeVisible();
    expect(
      callsTo(fetchMock, 'GET /questions').some(
        (call) => new URL(call.url).searchParams.get('cursor') === 'c2',
      ),
    ).toBe(true);
  });

  it('says so when a list is empty', async () => {
    setLanguages(['fr-FR']);
    mockApi(
      routes({
        'GET /questions': () => questionPage([]),
        'GET /projects': () => projectPage([]),
      }),
    );
    const user = userEvent.setup({ delay: null });
    await renderApp('/profils/awa');
    expect(await screen.findByText('Aucune question')).toBeVisible();
    await user.click(screen.getByRole('tab', { name: /^Réponses/ }));
    expect(await screen.findByText('Aucune réponse')).toBeVisible();
    await user.click(screen.getByRole('tab', { name: /^Projets/ }));
    expect(await screen.findByText('Aucun projet')).toBeVisible();
  });

  it('shows an empty list, never a 404, for a hidden profile', async () => {
    setLanguages(['fr-FR']);
    mockApi(
      routes({
        'GET /questions': () => questionPage([]),
        'GET /projects': () => projectPage([]),
      }),
    );
    await renderApp('/profils/awa');
    expect(await screen.findByText('Aucune question')).toBeVisible();
  });

  it('offers a retry when one list fails, and keeps the others', async () => {
    setLanguages(['fr-FR']);
    let failures = 0;
    mockApi(
      routes({
        'GET /projects': () =>
          ++failures === 1 ? problemResponse(500, 'INTERNAL_ERROR') : projectPage([built], null, 1),
      }),
    );
    const user = userEvent.setup({ delay: null });
    await renderApp('/profils/awa');
    expect(await screen.findByRole('link', { name: 'Webhook non reçu' })).toBeVisible();
    await user.click(screen.getByRole('tab', { name: /^Projets/ }));
    await user.click(await screen.findByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('link', { name: 'kora-sdk' })).toBeVisible();
  });

  it('writes the tabs in English', async () => {
    setLanguages(['en-US']);
    mockApi(routes());
    await renderApp('/profils/awa');
    const tabs = await screen.findByRole('tablist', { name: 'Activity' });
    expect(within(tabs).getByRole('tab', { name: /^Answers/ })).toBeVisible();
    expect(within(tabs).getByRole('tab', { name: /^Projects/ })).toBeVisible();
  });

  it('asks for no activity when the profile is not found', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /profiles/awa': () => problemResponse(404, 'NOT_FOUND'),
    });
    await renderApp('/profils/awa');
    expect(await screen.findByText('Profil introuvable')).toBeVisible();
    expect(callsTo(fetchMock, 'GET /questions')).toHaveLength(0);
    expect(callsTo(fetchMock, 'GET /projects')).toHaveLength(0);
  });
});
