import { screen, waitFor } from '@testing-library/react';
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
  return mockApi({
    'GET /auth/session': () => jsonResponse(memberSession),
    'GET /me': () => jsonResponse({ account: memberAccount, profile }),
    ...routes,
  });
}

describe('SettingsPage display-name form (live profile)', () => {
  it('loads the display name from the API', async () => {
    setLanguages(['fr-FR']);
    member();
    await renderApp('/parametres');
    expect(await screen.findByLabelText(/Nom d’affichage/)).toHaveValue('Awa Traoré');
  });

  it('invites an anonymous visitor to sign in instead of showing the form', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/parametres');
    expect(
      await screen.findByRole('link', { name: 'Connectez-vous pour modifier votre nom' }),
    ).toHaveAttribute('href', '/connexion?next=%2Fparametres');
    expect(screen.queryByLabelText(/Nom d’affichage/)).toBeNull();
  });

  it('rejects an empty name on submit and sends nothing', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member();
    await renderApp('/parametres');
    await userEvent.clear(await screen.findByLabelText(/Nom d’affichage/));
    await userEvent.click(screen.getByRole('button', { name: 'Enregistrer' }));
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Saisissez un nom d’affichage.');
    });
    expect(screen.getByLabelText(/Nom d’affichage/)).toBeInvalid();
    expect(callsTo(fetchMock, 'PATCH /me/profile')).toHaveLength(0);
  });

  // Owner decision 2026-11-14 (#146): the client follows the contract and the server, a 1-character name is valid.
  it('accepts a one-character name, as the contract does', async () => {
    setLanguages(['en-US']);
    const fetchMock = member();
    await renderApp('/parametres');
    const input = await screen.findByLabelText(/Display name/);
    await userEvent.clear(input);
    await userEvent.type(input, 'A');
    await userEvent.tab();
    // Astryx keeps an aria-live region with role "alert" between tests, so check the field itself.
    expect(input).toBeValid();
    expect(screen.queryByText(/at least|Enter a display name/)).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, 'PATCH /me/profile')).toHaveLength(1);
    });
  });

  it('saves a trimmed name with PATCH /me/profile, the CSRF token and a JSON body', async () => {
    setLanguages(['fr-FR']);
    let sent: unknown;
    const fetchMock = member({
      'PATCH /me/profile': async (request: Request) => {
        sent = await request.json();
        return jsonResponse({ ...profile, display_name: 'Awa K.' });
      },
    });
    await renderApp('/parametres');
    const input = await screen.findByLabelText(/Nom d’affichage/);
    await userEvent.clear(input);
    await userEvent.type(input, '  Awa K.  ');
    await userEvent.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(await screen.findByText('Nom enregistré.')).toBeInTheDocument();
    expect(sent).toEqual({ display_name: 'Awa K.' });
    const [patch] = callsTo(fetchMock, 'PATCH /me/profile');
    expect(patch?.headers.get('X-CSRF-Token')).toBe('csrf-member');
    expect(patch?.headers.get('Content-Type')).toContain('application/json');
  });

  it('shows the new name in the header and keeps no copy in the browser storage', async () => {
    setLanguages(['fr-FR']);
    member({
      'PATCH /me/profile': () => jsonResponse({ ...profile, display_name: 'Awa K.' }),
    });
    await renderApp('/parametres');
    const input = await screen.findByLabelText(/Nom d’affichage/);
    await userEvent.clear(input);
    await userEvent.type(input, 'Awa K.');
    await userEvent.click(screen.getByRole('button', { name: 'Enregistrer' }));
    await screen.findByText('Nom enregistré.');
    expect(await screen.findByRole('button', { name: 'Awa K.' })).toBeInTheDocument();
    expect(localStorage.getItem('226os.displayName')).toBeNull();
  });

  it('shows the server violation on the name field', async () => {
    setLanguages(['fr-FR']);
    member({
      'PATCH /me/profile': () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/display_name', code: 'too_long' }],
        }),
    });
    await renderApp('/parametres');
    const input = await screen.findByLabelText(/Nom d’affichage/);
    await userEvent.clear(input);
    await userEvent.type(input, 'Awa');
    await userEvent.click(screen.getByRole('button', { name: 'Enregistrer' }));
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Le nom doit contenir 80 caractères au plus.',
      );
    });
    expect(screen.getByLabelText(/Nom d’affichage/)).toBeInvalid();
    expect(screen.queryByText('Nom enregistré.')).toBeNull();
  });

  it('keeps unsaved edits when the language changes, and only relabels the UI', async () => {
    setLanguages(['fr-FR']);
    member();
    await renderApp('/parametres');
    const input = await screen.findByLabelText(/Nom d’affichage/);
    await userEvent.clear(input);
    await userEvent.type(input, 'Moussa');
    await userEvent.click(screen.getByRole('radio', { name: 'English' }));
    expect(await screen.findByLabelText(/Display name/)).toHaveValue('Moussa');
    expect(localStorage.getItem('226os.locale')).toBe('en');
  });

  it('lets the user choose a theme preference', async () => {
    setLanguages(['fr-FR']);
    member();
    await renderApp('/parametres');
    expect(await screen.findByRole('radio', { name: 'Système' })).toBeChecked();
    await userEvent.click(screen.getByRole('radio', { name: 'Sombre' }));
    expect(screen.getByRole('radio', { name: 'Sombre' })).toBeChecked();
    expect(document.documentElement).toHaveAttribute('data-theme', 'dark');
    expect(localStorage.getItem('226os.theme')).toBe('dark');
  });
});

describe('SettingsPage preferences', () => {
  it('shows a visible label above the theme and language switches', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/parametres');
    expect(await screen.findByText('Thème')).toBeVisible();
    expect(screen.getByText('Langue')).toBeVisible();
  });
});
