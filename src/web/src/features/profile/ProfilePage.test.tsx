import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, memberAccount, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

type PublicProfile = components['schemas']['PublicProfile'];
type MyProfile = components['schemas']['MyProfile'];
type ProfileField = components['schemas']['ProfileField'];

function publicProfile(overrides: Partial<PublicProfile> = {}): PublicProfile {
  return {
    username: 'awa',
    display_name: 'Awa Traoré',
    bio: 'Je construis des outils de paiement.',
    country: 'BF',
    languages: ['fr', 'mos'],
    skills: ['go', 'react'],
    github: 'awa-dev',
    open_to_collab: true,
    profile_visibility: 'public',
    joined_at: '2026-11-01T10:00:00Z',
    ...overrides,
  };
}

function myProfile(missing: ProfileField[]): MyProfile {
  return {
    username: 'awa',
    display_name: 'Awa Traoré',
    intent: missing.includes('intent') ? null : 'learn',
    skills: missing.includes('skills') ? [] : ['go'],
    country: missing.includes('country') ? null : 'BF',
    languages: missing.includes('languages') ? [] : ['fr'],
    github: missing.includes('github') ? null : 'awa-dev',
    bio: missing.includes('bio') ? null : 'Bonjour',
    open_to_collab: false,
    field_visibility: {},
    profile_visibility: 'public',
    checklist_missing: missing,
  };
}

function member(routes = {}) {
  return mockApi({ 'GET /auth/session': () => jsonResponse(memberSession), ...routes });
}

async function expectAlert(text: string) {
  await waitFor(() => {
    const alerts = screen.getAllByRole('alert').map((alert) => alert.textContent);
    expect(alerts.join(' ')).toContain(text);
  });
}

describe('ProfilePage view', () => {
  it('shows the public profile in French, with links to GitHub', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /profiles/awa': () => jsonResponse(publicProfile()) });
    await renderApp('/profils/awa');
    expect(await screen.findByRole('heading', { level: 1, name: 'Awa Traoré' })).toBeVisible();
    expect(screen.getByText('@awa')).toBeVisible();
    expect(screen.getByText('Je construis des outils de paiement.')).toBeVisible();
    expect(screen.getByText('Burkina Faso')).toBeVisible();
    expect(screen.getByText('react')).toBeVisible();
    expect(screen.getByText('Ouvert aux collaborations')).toBeVisible();
    expect(screen.getByRole('link', { name: /awa-dev/ })).toHaveAttribute(
      'href',
      'https://github.com/awa-dev',
    );
  });

  it('renders no GitHub link for a malformed github value', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      'GET /profiles/awa': () => jsonResponse(publicProfile({ github: '../evil?x=1' })),
    });
    await renderApp('/profils/awa');
    expect(await screen.findByRole('heading', { level: 1, name: 'Awa Traoré' })).toBeVisible();
    expect(screen.queryByRole('link', { name: /evil/ })).not.toBeInTheDocument();
    expect(screen.queryByText(/evil/)).not.toBeInTheDocument();
  });

  it('omits fields the server hid, without placeholders', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      'GET /profiles/awa': () =>
        jsonResponse({
          username: 'awa',
          display_name: 'Awa Traoré',
          profile_visibility: 'public',
          joined_at: '2026-11-01T10:00:00Z',
        }),
    });
    await renderApp('/profils/awa');
    await screen.findByRole('heading', { level: 1, name: 'Awa Traoré' });
    expect(screen.queryByText('Pays')).not.toBeInTheDocument();
    expect(screen.queryByText('Compétences')).not.toBeInTheDocument();
    expect(screen.getByText('Awa Traoré n’a pas encore complété son profil.')).toBeVisible();
  });

  it('says nothing about an empty profile when something is filled in', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /profiles/awa': () => jsonResponse(publicProfile()) });
    await renderApp('/profils/awa');
    await screen.findByRole('heading', { level: 1, name: 'Awa Traoré' });
    expect(screen.queryByText(/pas encore complété son profil/)).not.toBeInTheDocument();
  });

  it('shows the member their checklist, not the empty-profile line, on their own bare profile', async () => {
    setLanguages(['fr-FR']);
    member({
      'GET /profiles/awa': () =>
        jsonResponse({
          username: 'awa',
          display_name: 'Awa Traoré',
          profile_visibility: 'public',
          joined_at: '2026-11-01T10:00:00Z',
        }),
      'GET /me': () =>
        jsonResponse({
          account: memberAccount,
          profile: myProfile(['bio', 'github', 'country', 'languages', 'skills', 'intent']),
        }),
    });
    await renderApp('/profils/awa');
    const heading = await screen.findByRole('heading', { level: 1, name: 'Awa Traoré' });
    const card = await screen.findByRole('region', { name: 'Complétez votre profil' });
    // Who the page is about comes first, the checklist after.
    expect(heading.compareDocumentPosition(card) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(screen.queryByText(/pas encore complété son profil/)).not.toBeInTheDocument();
  });

  it('shows the same not-found state for a hidden profile, an unknown one and a malformed name', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /profiles/hidden': () => problemResponse(404, 'NOT_FOUND'),
      'GET /profiles/unknown': () => problemResponse(404, 'NOT_FOUND'),
    });
    const texts: (string | null)[] = [];
    for (const name of ['hidden', 'unknown', 'A!']) {
      const { unmount } = await renderApp(`/profils/${encodeURIComponent(name)}`);
      const heading = await screen.findByRole('heading', { level: 1, name: 'Profil introuvable' });
      texts.push(heading.parentElement?.textContent ?? null);
      unmount();
    }
    expect(texts[0]).toBe(texts[1]);
    expect(texts[1]).toBe(texts[2]);
    expect(callsTo(fetchMock, 'GET /profiles/A!')).toHaveLength(0);
  });

  it('reads a 422 on the username as the same not-found state', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /profiles/awa': () => problemResponse(422, 'VALIDATION_ERROR') });
    await renderApp('/profils/awa');
    expect(await screen.findByText('Profil introuvable')).toBeVisible();
  });

  it('shows a retryable error on a server failure', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /profiles/awa': () => problemResponse(500, 'INTERNAL_ERROR') });
    await renderApp('/profils/awa');
    await expectAlert('Une erreur est survenue');
  });

  it('reads in English too', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /profiles/awa': () => problemResponse(404, 'NOT_FOUND') });
    await renderApp('/profils/awa');
    expect(await screen.findByText('Profile not found')).toBeVisible();
  });

  it('adds Mon profil to the account menu, pointing at the own profile', async () => {
    setLanguages(['fr-FR']);
    member({ 'GET /profiles/awa': () => jsonResponse(publicProfile()) });
    const user = userEvent.setup();
    await renderApp('/profils/awa');
    await user.click(await screen.findByRole('button', { name: /Awa Traoré/ }));
    expect(await screen.findByRole('menuitem', { name: 'Mon profil' })).toHaveAttribute(
      'href',
      '/profils/awa',
    );
  });
});

describe('ProfilePage completion checklist', () => {
  const profileRoutes = { 'GET /profiles/awa': () => jsonResponse(publicProfile()) };

  it('is absent for an anonymous visitor and on someone else’s profile, and asks nothing', async () => {
    setLanguages(['fr-FR']);
    const anonymous = mockApi({
      'GET /profiles/awa': () => jsonResponse(publicProfile()),
    });
    const first = await renderApp('/profils/awa');
    await screen.findByRole('heading', { level: 1, name: 'Awa Traoré' });
    expect(screen.queryByText('Complétez votre profil')).not.toBeInTheDocument();
    expect(callsTo(anonymous, 'GET /me')).toHaveLength(0);
    first.unmount();

    const signedIn = member({
      'GET /profiles/kofi': () =>
        jsonResponse(publicProfile({ username: 'kofi', display_name: 'Kofi' })),
      'GET /me': () => jsonResponse({ account: memberAccount, profile: myProfile(['bio']) }),
    });
    await renderApp('/profils/kofi');
    await screen.findByRole('heading', { level: 1, name: 'Kofi' });
    expect(screen.queryByText('Complétez votre profil')).not.toBeInTheDocument();
    expect(callsTo(signedIn, 'GET /me')).toHaveLength(0);
  });

  it('shows progress and one item per missing field, each with its value, on the own profile', async () => {
    setLanguages(['fr-FR']);
    member({
      ...profileRoutes,
      'GET /me': () =>
        jsonResponse({ account: memberAccount, profile: myProfile(['bio', 'github', 'country']) }),
    });
    await renderApp('/profils/awa');
    const card = await screen.findByRole('region', { name: 'Complétez votre profil' });
    expect(within(card).getByRole('progressbar', { name: 'Profil complété' })).toHaveAttribute(
      'aria-valuenow',
      '50',
    );
    expect(within(card).getAllByRole('listitem')).toHaveLength(3);
    expect(within(card).getByText(/Quelques lignes pour que les autres/)).toBeVisible();
    expect(within(card).queryByRole('textbox', { name: /Compétences/ })).not.toBeInTheDocument();
  });

  it('saves a field in place, updates the progress and removes the item', async () => {
    setLanguages(['fr-FR']);
    let missing: ProfileField[] = ['bio', 'github'];
    const fetchMock = member({
      ...profileRoutes,
      'GET /me': () => jsonResponse({ account: memberAccount, profile: myProfile(missing) }),
      'PATCH /me/profile': () => {
        missing = ['github'];
        return jsonResponse(myProfile(missing));
      },
    });
    const user = userEvent.setup();
    await renderApp('/profils/awa');
    const card = await screen.findByRole('region', { name: 'Complétez votre profil' });
    await user.type(within(card).getByRole('textbox', { name: /Présentation/ }), 'Développeuse.');
    await user.click(within(card).getByRole('button', { name: 'Enregistrer Présentation' }));

    await waitFor(() => {
      expect(within(card).queryByRole('textbox', { name: /Présentation/ })).not.toBeInTheDocument();
    });
    expect(within(card).getByRole('progressbar')).toHaveAttribute('aria-valuenow', '83');
    const [request] = callsTo(fetchMock, 'PATCH /me/profile');
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(request?.headers.get('Content-Type')).toContain('application/json');
    expect(await request?.clone().json()).toEqual({ bio: 'Développeuse.' });
  });

  it('normalises country, languages and skills before sending', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      ...profileRoutes,
      'GET /me': () =>
        jsonResponse({
          account: memberAccount,
          profile: myProfile(['country', 'languages', 'skills']),
        }),
      'PATCH /me/profile': () => jsonResponse(myProfile(['languages', 'skills'])),
    });
    const user = userEvent.setup();
    await renderApp('/profils/awa');
    const card = await screen.findByRole('region', { name: 'Complétez votre profil' });
    await user.type(within(card).getByRole('textbox', { name: /Pays/ }), 'bf');
    await user.click(within(card).getByRole('button', { name: 'Enregistrer Pays' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, 'PATCH /me/profile')).toHaveLength(1);
    });
    expect(await callsTo(fetchMock, 'PATCH /me/profile')[0]?.clone().json()).toEqual({
      country: 'BF',
    });
  });

  it('refuses a malformed value locally and shows server violations on the field', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      ...profileRoutes,
      'GET /me': () =>
        jsonResponse({ account: memberAccount, profile: myProfile(['github', 'languages']) }),
      'PATCH /me/profile': () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/github', code: 'invalid_format' }],
        }),
    });
    const user = userEvent.setup();
    await renderApp('/profils/awa');
    const card = await screen.findByRole('region', { name: 'Complétez votre profil' });

    await user.type(within(card).getByRole('textbox', { name: /Langues parlées/ }), 'francais');
    await user.click(within(card).getByRole('button', { name: 'Enregistrer Langues parlées' }));
    expect(await within(card).findByText(/Indiquez des codes de langue/)).toBeVisible();
    expect(callsTo(fetchMock, 'PATCH /me/profile')).toHaveLength(0);

    await user.type(within(card).getByRole('textbox', { name: /Identifiant GitHub/ }), 'awa-dev');
    await user.click(within(card).getByRole('button', { name: 'Enregistrer Identifiant GitHub' }));
    expect(await within(card).findByText(/Indiquez votre identifiant GitHub/)).toBeVisible();
    expect(within(card).getByRole('textbox', { name: /Identifiant GitHub/ })).toBeInvalid();
  });

  it('removes the whole card at 100 %', async () => {
    setLanguages(['fr-FR']);
    let missing: ProfileField[] = ['intent'];
    member({
      ...profileRoutes,
      'GET /me': () => jsonResponse({ account: memberAccount, profile: myProfile(missing) }),
      'PATCH /me/profile': () => {
        missing = [];
        return jsonResponse(myProfile(missing));
      },
    });
    const user = userEvent.setup();
    await renderApp('/profils/awa');
    const card = await screen.findByRole('region', { name: 'Complétez votre profil' });
    await user.click(within(card).getByRole('radio', { name: 'Apprendre' }));
    await user.click(within(card).getByRole('button', { name: 'Enregistrer Ce qui vous amène' }));
    await waitFor(() => {
      expect(screen.queryByText('Complétez votre profil')).not.toBeInTheDocument();
    });
  });

  it('is not shown when nothing is missing', async () => {
    setLanguages(['fr-FR']);
    member({
      ...profileRoutes,
      'GET /me': () => jsonResponse({ account: memberAccount, profile: myProfile([]) }),
    });
    await renderApp('/profils/awa');
    await screen.findByRole('heading', { level: 1, name: 'Awa Traoré' });
    await waitFor(() => {
      expect(screen.queryByText('Complétez votre profil')).not.toBeInTheDocument();
    });
  });

  it('explains an unverified or refused save on the card', async () => {
    setLanguages(['fr-FR']);
    member({
      ...profileRoutes,
      'GET /me': () => jsonResponse({ account: memberAccount, profile: myProfile(['bio']) }),
      'PATCH /me/profile': () => problemResponse(403, 'EMAIL_NOT_VERIFIED'),
    });
    const user = userEvent.setup();
    await renderApp('/profils/awa');
    const card = await screen.findByRole('region', { name: 'Complétez votre profil' });
    await user.type(within(card).getByRole('textbox', { name: /Présentation/ }), 'Bonjour');
    await user.click(within(card).getByRole('button', { name: 'Enregistrer Présentation' }));
    await expectAlert('Confirmez d’abord votre adresse e-mail.');
  });
});
