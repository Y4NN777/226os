import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, memberAccount, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

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
    await userEvent.click(screen.getByRole('radio', { name: /Membres/ }));
    await userEvent.click(screen.getByRole('button', { name: 'Enregistrer la visibilité' }));
    expect(await screen.findByText('Visibilité enregistrée.')).toBeInTheDocument();
    expect(saved).toEqual({ profile_visibility: 'members' });
    const [patch] = callsTo(fetchMock, 'PATCH /me/profile');
    expect(patch?.headers.get('X-CSRF-Token')).toBe('csrf-member');
  });

  it('opts into the Connect directory with the field violation on the control', async () => {
    setLanguages(['fr-FR']);
    let saved: unknown;
    member({
      'GET /me': () => jsonResponse({ account: memberAccount, profile }),
      'PATCH /me/profile': async (request: Request) => {
        saved = await request.json();
        if (saved && (saved as { open_to_collab?: boolean }).open_to_collab) {
          return problemResponse(422, 'VALIDATION_ERROR', {
            violations: [{ field: '/open_to_collab', code: 'invalid_value' }],
          });
        }
        return jsonResponse(profile);
      },
    });
    await renderApp('/parametres');
    const optIn = await screen.findByRole('checkbox', { name: 'Visible dans l’annuaire' });
    await userEvent.click(optIn);
    expect(saved).toEqual({ open_to_collab: true });
    expect(optIn).toHaveAccessibleDescription(/n’a pas été enregistrée/);
  });

  it('keeps the choice and explains when saving fails', async () => {
    setLanguages(['fr-FR']);
    member({
      'GET /me': () => jsonResponse({ account: memberAccount, profile }),
      'PATCH /me/profile': () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 20 }),
    });
    await renderApp('/parametres');
    await userEvent.click(await screen.findByRole('radio', { name: /Privé/ }));
    await userEvent.click(screen.getByRole('button', { name: 'Enregistrer la visibilité' }));
    await waitFor(() => {
      expect(screen.getByText(/Réessayez dans 20/)).toBeVisible();
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
