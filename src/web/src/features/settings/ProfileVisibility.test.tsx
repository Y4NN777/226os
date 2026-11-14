import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, memberAccount, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { findVisibleText } from '../../test/visibleText';

type MyProfile = components['schemas']['MyProfile'];

const profile: MyProfile = {
  username: 'awa',
  display_name: 'Awa Traoré',
  intent: null,
  skills: [],
  country: null,
  languages: [],
  github: null,
  bio: null,
  open_to_collab: false,
  field_visibility: {},
  profile_visibility: 'public',
  checklist_missing: [],
};

function member(routes = {}) {
  return mockApi({ 'GET /auth/session': () => jsonResponse(memberSession), ...routes });
}

describe('Profile visibility setting (parametres)', () => {
  it('invites an anonymous visitor to sign in, coming back here after', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/parametres');
    const link = await screen.findByRole('link', {
      name: 'Connectez-vous pour choisir qui voit votre profil',
    });
    expect(link).toHaveAttribute('href', '/connexion?next=%2Fparametres');
  });

  it('is shown but disabled, with a clear label, while profiles are not live', async () => {
    setLanguages(['fr-FR']);
    member({ 'GET /me': () => new Response(null, { status: 501 }) });
    await renderApp('/parametres');
    expect(
      await screen.findByText(/ce réglage s’activera dès que les profils seront en ligne/),
    ).toBeInTheDocument();
    const group = screen.getByRole('radiogroup', { name: /Qui peut voir votre profil/ });
    for (const radio of within(group).getAllByRole('radio')) {
      expect(radio).toBeDisabled();
    }
    expect(screen.queryByRole('button', { name: 'Enregistrer la visibilité' })).toBeNull();
  });

  it('says that bylines still show the name and username', async () => {
    setLanguages(['en-US']);
    member({ 'GET /me': () => jsonResponse({ account: memberAccount, profile }) });
    await renderApp('/parametres');
    expect(
      await screen.findByText(/your name and username still appear on what you publish/i),
    ).toBeInTheDocument();
  });

  it('saves the chosen visibility with the CSRF token and confirms it', async () => {
    setLanguages(['fr-FR']);
    let saved: unknown;
    const fetchMock = member({
      'GET /me': () => jsonResponse({ account: memberAccount, profile }),
      'PATCH /me/profile': async (request: Request) => {
        saved = await request.json();
        return jsonResponse({ ...profile, profile_visibility: 'members' });
      },
    });
    await renderApp('/parametres');
    expect(await screen.findByRole('radio', { name: /Public/ })).toBeChecked();
    await userEvent.setup({ delay: null }).click(screen.getByRole('radio', { name: /Membres/ }));
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Enregistrer la visibilité' }));
    expect(await findVisibleText('Visibilité enregistrée.')).toBeInTheDocument();
    expect(saved).toEqual({ profile_visibility: 'members' });
    const [patch] = callsTo(fetchMock, 'PATCH /me/profile');
    expect(patch?.headers.get('X-CSRF-Token')).toBe('csrf-member');
  });

  it('keeps the choice and explains when saving fails', async () => {
    setLanguages(['fr-FR']);
    member({
      'GET /me': () => jsonResponse({ account: memberAccount, profile }),
      'PATCH /me/profile': () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 20 }),
    });
    await renderApp('/parametres');
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('radio', { name: /Privé/ }));
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Enregistrer la visibilité' }));
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/Réessayez dans 20/);
    });
    expect(screen.getByRole('radio', { name: /Privé/ })).toBeChecked();
  });
});

describe('Settings route', () => {
  it('redirects the old /settings address to /parametres', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    const { router } = await renderApp('/settings');
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/parametres');
    });
  });
});
