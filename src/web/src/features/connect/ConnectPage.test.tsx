import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { validateConnectSearch } from './routeSearch';

type Summary = components['schemas']['ProfileSummary'];

const awa: Summary = {
  username: 'awa',
  display_name: 'Awa Traoré',
  country: 'BF',
  languages: ['fr', 'en'],
  skills: ['go', 'postgresql'],
  github: 'awa-dev',
};
// Hidden fields are absent from the payload, never null.
const kofi: Summary = { username: 'kofi', display_name: 'Kofi Mensah' };

function page(items: Summary[], next: string | null = null) {
  return jsonResponse({ items, next_cursor: next });
}

function params(request: Request | undefined) {
  return new URL(request?.url ?? 'http://x').searchParams;
}

const base = { 'GET /profiles': () => page([awa, kofi]) };

describe('ConnectPage list', () => {
  it('lists people with a profile link, username and only the fields present (FR)', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/connect');
    expect(await screen.findByRole('heading', { level: 1, name: 'Connect' })).toBeVisible();
    const link = await screen.findByRole('link', { name: 'Awa Traoré' });
    expect(link).toHaveAttribute('href', '/profils/awa');
    expect(screen.getByText('@awa')).toBeVisible();
    const card = link.closest('li') as HTMLElement;
    expect(within(card).getByText('Burkina Faso')).toBeVisible();
    expect(within(card).getByText('français, anglais')).toBeVisible();
    expect(within(card).getByText('go')).toBeVisible();
    expect(within(card).getByText('postgresql')).toBeVisible();
    expect(within(card).getByRole('link', { name: 'GitHub de Awa Traoré' })).toHaveAttribute(
      'href',
      'https://github.com/awa-dev',
    );
  });

  it('renders nothing for fields that are absent', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/connect');
    const card = (await screen.findByRole('link', { name: 'Kofi Mensah' })).closest(
      'li',
    ) as HTMLElement;
    expect(within(card).getByText('@kofi')).toBeVisible();
    expect(within(card).queryByRole('link', { name: /GitHub/ })).not.toBeInTheDocument();
    expect(within(card).queryByRole('list')).not.toBeInTheDocument();
    expect(card.textContent).toBe('Kofi Mensah@kofi');
  });

  it('speaks English when the browser does', async () => {
    setLanguages(['en-US']);
    mockApi(base);
    await renderApp('/connect?country=BF');
    const card = (await screen.findByRole('link', { name: 'Awa Traoré' })).closest(
      'li',
    ) as HTMLElement;
    expect(within(card).getByText('French, English')).toBeVisible();
    expect(screen.getByRole('button', { name: 'Reset filters' })).toBeVisible();
  });

  it('shows a quiet « Bientôt » line for mentoring, spaces and opportunities, without links', async () => {
    setLanguages(['fr-FR']);
    mockApi(base);
    await renderApp('/connect');
    await screen.findByRole('link', { name: 'Awa Traoré' });
    expect(screen.getByText('Bientôt : mentorat, espaces et opportunités.')).toBeVisible();
  });
});

describe('ConnectPage filters', () => {
  it('sends the filters from the address and keeps them there', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    await renderApp('/connect?country=BF&language=fr&skill=go');
    await screen.findByRole('link', { name: 'Awa Traoré' });
    const query = params(callsTo(fetchMock, 'GET /profiles')[0]);
    expect(query.get('country')).toBe('BF');
    expect(query.get('language')).toBe('fr');
    expect(query.get('skill')).toBe('go');
  });

  it('picks a country and puts it in the route search', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    const user = userEvent.setup();
    const { router } = await renderApp('/connect');
    await screen.findByRole('link', { name: 'Awa Traoré' });
    await user.click(screen.getByRole('combobox', { name: 'Pays' }));
    await user.click(await screen.findByRole('option', { name: 'Sénégal' }));
    await waitFor(() => {
      expect(params(callsTo(fetchMock, 'GET /profiles').at(-1)).get('country')).toBe('SN');
    });
    expect(router.state.location.search).toMatchObject({ country: 'SN' });
  });

  it('applies a language and a skill typed in the form', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    const user = userEvent.setup();
    const { router } = await renderApp('/connect');
    await screen.findByRole('link', { name: 'Awa Traoré' });
    await user.type(screen.getByRole('textbox', { name: 'Langue' }), 'FR');
    await user.type(screen.getByRole('textbox', { name: 'Compétence' }), ' Go{Enter}');
    await waitFor(() => {
      const query = params(callsTo(fetchMock, 'GET /profiles').at(-1));
      expect(query.get('language')).toBe('fr');
      expect(query.get('skill')).toBe('go');
    });
    expect(router.state.location.search).toMatchObject({ language: 'fr', skill: 'go' });
  });

  it('ignores a typed value that is not a valid tag', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    const user = userEvent.setup();
    const { router } = await renderApp('/connect');
    await screen.findByRole('link', { name: 'Awa Traoré' });
    await user.type(screen.getByRole('textbox', { name: 'Langue' }), 'français{Enter}');
    expect(router.state.location.search).not.toHaveProperty('language');
    expect(callsTo(fetchMock, 'GET /profiles')).toHaveLength(1);
  });

  it('drops invalid filters from the address', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    await renderApp('/connect?country=burkina&language=FRANCAIS&skill=%20!!');
    await screen.findByRole('link', { name: 'Awa Traoré' });
    const query = params(callsTo(fetchMock, 'GET /profiles')[0]);
    expect(query.has('country')).toBe(false);
    expect(query.has('language')).toBe(false);
    expect(query.has('skill')).toBe(false);
  });

  it('validates the route search', () => {
    expect(validateConnectSearch({ country: 'BF', language: 'fr', skill: 'c++' })).toEqual({
      country: 'BF',
      language: 'fr',
      skill: 'c++',
    });
    expect(validateConnectSearch({ country: 'bf', language: 1, skill: 'x'.repeat(33) })).toEqual(
      {},
    );
  });

  it('resets every filter', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(base);
    const user = userEvent.setup();
    const { router } = await renderApp('/connect?country=BF&skill=go');
    await screen.findByRole('link', { name: 'Awa Traoré' });
    await user.click(screen.getByRole('button', { name: 'Réinitialiser les filtres' }));
    await waitFor(() => {
      expect(params(callsTo(fetchMock, 'GET /profiles').at(-1)).has('country')).toBe(false);
    });
    expect(router.state.location.search).toEqual({});
    expect(screen.getByRole('textbox', { name: 'Compétence' })).toHaveValue('');
  });
});

describe('ConnectPage states', () => {
  it('loads more with next_cursor', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /profiles': (request) =>
        params(request).get('cursor') === 'c2' ? page([kofi]) : page([awa], 'c2'),
    });
    const user = userEvent.setup();
    await renderApp('/connect');
    await screen.findByRole('link', { name: 'Awa Traoré' });
    await user.click(screen.getByRole('button', { name: 'Afficher plus' }));
    expect(await screen.findByRole('link', { name: 'Kofi Mensah' })).toBeVisible();
    expect(screen.getByRole('link', { name: 'Awa Traoré' })).toBeVisible();
    expect(screen.queryByRole('button', { name: 'Afficher plus' })).not.toBeInTheDocument();
    expect(params(callsTo(fetchMock, 'GET /profiles')[1]).get('cursor')).toBe('c2');
  });

  it('says the directory is empty, with no reset when no filter is active', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /profiles': () => page([]) });
    await renderApp('/connect');
    expect(await screen.findByText('Aucun profil pour l’instant')).toBeVisible();
  });

  it('offers a reset when filters match nobody', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /profiles': () => page([]) });
    const user = userEvent.setup();
    const { router } = await renderApp('/connect?skill=cobol');
    expect(await screen.findByText('Personne ne correspond à ces filtres')).toBeVisible();
    const empty = screen
      .getByText('Personne ne correspond à ces filtres')
      .closest('section') as HTMLElement;
    await user.click(within(empty).getByRole('button', { name: 'Réinitialiser les filtres' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({});
    });
  });

  it('shows an error with retry', async () => {
    setLanguages(['fr-FR']);
    let failing = true;
    mockApi({
      'GET /profiles': () => (failing ? problemResponse(500, 'INTERNAL_ERROR') : page([awa])),
    });
    const user = userEvent.setup();
    await renderApp('/connect');
    expect(await screen.findByText('Impossible de charger le contenu')).toBeVisible();
    failing = false;
    await user.click(screen.getByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('link', { name: 'Awa Traoré' })).toBeVisible();
  });

  it('tells how long to wait on a 429', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      'GET /profiles': () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 12 }),
    });
    await renderApp('/connect');
    expect(await screen.findByText('Trop de requêtes, réessayez dans 12 s')).toBeVisible();
    expect(screen.getByRole('button', { name: 'Réessayer' })).toBeVisible();
  });

  it('falls back to a generic wait on a 429 without Retry-After', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /profiles': () => problemResponse(429, 'RATE_LIMITED') });
    await renderApp('/connect');
    expect(await screen.findByText('Too many requests, try again shortly')).toBeVisible();
  });
});
