import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

const SIGN_UP = 'POST /auth/sign-up';
const accepted = () => jsonResponse({ status: 'accepted' }, { status: 202 });

async function open(language = 'fr-FR', routes: Parameters<typeof mockApi>[0] = {}) {
  setLanguages([language]);
  const fetchMock = mockApi(routes);
  const user = userEvent.setup();
  const app = await renderApp('/inscription');
  await screen.findByRole('heading', { level: 1 });
  return { fetchMock, user, ...app };
}

async function fillAll(user: ReturnType<typeof userEvent.setup>, labels: string[]) {
  const [username, email, password] = labels;
  await user.type(screen.getByLabelText(/^Nom d’affichage/), 'Awa Traoré');
  await user.type(screen.getByLabelText(username ?? ''), 'awa_t');
  await user.type(screen.getByLabelText(email ?? ''), 'awa@example.org');
  await user.type(screen.getByLabelText(password ?? ''), 'a long passphrase');
}

const FR = ['Nom d’utilisateur', 'Adresse e-mail', 'Mot de passe'];

describe('SignUpPage', () => {
  it('shows the form with an h1 and the username rule', async () => {
    await open();
    expect(screen.getByRole('heading', { level: 1, name: 'Créer votre compte' })).toBeVisible();
    expect(screen.getByText(/Identifiant public/)).toBeVisible();
    expect(screen.getByLabelText('Mot de passe')).toHaveAttribute('autocomplete', 'new-password');
    expect(screen.getByLabelText('Adresse e-mail')).toHaveAttribute('autocomplete', 'email');
    expect(screen.getByRole('link', { name: 'Se connecter' })).toHaveAttribute(
      'href',
      '/connexion',
    );
  });

  it('validates on submit and focuses the first invalid field', async () => {
    const { user, fetchMock } = await open();
    await user.click(screen.getByRole('button', { name: 'Créer mon compte' }));
    expect(await screen.findAllByText('Choisissez un nom d’utilisateur.')).not.toHaveLength(0);
    expect(screen.getAllByText('Saisissez votre adresse e-mail.')).not.toHaveLength(0);
    await waitFor(() => expect(screen.getByLabelText('Nom d’utilisateur')).toHaveFocus());
    expect(callsTo(fetchMock, SIGN_UP)).toHaveLength(0);
  });

  it('rejects a malformed username and a short password', async () => {
    const { user, fetchMock } = await open();
    await user.type(screen.getByLabelText('Nom d’utilisateur'), 'Awa!');
    await user.type(screen.getByLabelText('Adresse e-mail'), 'awa@example.org');
    await user.type(screen.getByLabelText('Mot de passe'), 'short');
    await user.click(screen.getByRole('button', { name: 'Créer mon compte' }));
    expect(
      await screen.findAllByText(/lettres minuscules, chiffres, - et _, en commençant/),
    ).not.toHaveLength(0);
    expect(
      screen.getAllByText('Le mot de passe doit contenir au moins 12 caractères.'),
    ).not.toHaveLength(0);
    expect(callsTo(fetchMock, SIGN_UP)).toHaveLength(0);
  });

  it('rejects a display name over 80 characters', async () => {
    const { user, fetchMock } = await open();
    await user.click(screen.getByLabelText(/^Nom d’affichage/));
    await user.paste('x'.repeat(81));
    await user.type(screen.getByLabelText('Nom d’utilisateur'), 'awa_t');
    await user.type(screen.getByLabelText('Adresse e-mail'), 'awa@example.org');
    await user.type(screen.getByLabelText('Mot de passe'), 'a long passphrase');
    await user.click(screen.getByRole('button', { name: 'Créer mon compte' }));
    expect(
      await screen.findAllByText('Le nom ne peut pas dépasser 80 caractères.'),
    ).not.toHaveLength(0);
    expect(callsTo(fetchMock, SIGN_UP)).toHaveLength(0);
  });

  it('announces the password rule as text, met or not', async () => {
    const { user } = await open();
    expect(screen.getByText('12 caractères minimum : pas encore atteint')).toBeInTheDocument();
    await user.type(screen.getByLabelText('Mot de passe'), 'a long passphrase');
    expect(screen.getByText('12 caractères minimum : respecté')).toBeInTheDocument();
    expect(
      screen.getByText('12 caractères minimum : respecté').closest('[role="status"]'),
    ).not.toBeNull();
  });

  it('toggles the password visibility', async () => {
    const { user } = await open();
    const toggle = screen.getByRole('button', { name: 'Afficher le mot de passe' });
    await user.click(toggle);
    expect(toggle).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByLabelText('Mot de passe')).toHaveAttribute('type', 'text');
  });

  it('sends the sign-up, then moves to the verification page with nothing in the URL or storage', async () => {
    const { user, fetchMock, router } = await open('fr-FR', { [SIGN_UP]: accepted });
    await fillAll(user, FR);
    await user.click(screen.getByRole('button', { name: 'Créer mon compte' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/verifier-email');
    });
    const [request] = callsTo(fetchMock, SIGN_UP);
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(await request?.clone().json()).toEqual({
      email: 'awa@example.org',
      password: 'a long passphrase',
      username: 'awa_t',
      display_name: 'Awa Traoré',
      ui_lang: 'fr',
    });
    expect(router.state.location.searchStr).toBe('');
    expect(router.state.location.hash).toBe('');
    expect(localStorage.length).toBe(0);
    expect(sessionStorage.length).toBe(0);
  });

  it('omits an empty display name and sends the English locale', async () => {
    const { user, fetchMock, router } = await open('en-US', { [SIGN_UP]: accepted });
    await user.type(screen.getByLabelText('Username'), 'awa_t');
    await user.type(screen.getByLabelText('Email address'), 'awa@example.org');
    await user.type(screen.getByLabelText('Password'), 'a long passphrase');
    await user.click(screen.getByRole('button', { name: 'Create my account' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/verifier-email');
    });
    const [request] = callsTo(fetchMock, SIGN_UP);
    expect(await request?.clone().json()).toEqual({
      email: 'awa@example.org',
      password: 'a long passphrase',
      username: 'awa_t',
      ui_lang: 'en',
    });
  });

  it('maps 422 violations to the fields and focuses the first one', async () => {
    const { user } = await open('fr-FR', {
      [SIGN_UP]: () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [
            { field: '/username', code: 'invalid_format' },
            { field: '/password', code: 'too_short' },
            { field: '/email', code: 'invalid_format' },
            { field: '/display_name', code: 'too_long' },
          ],
        }),
    });
    await fillAll(user, FR);
    await user.click(screen.getByRole('button', { name: 'Créer mon compte' }));
    expect(
      await screen.findAllByText(/lettres minuscules, chiffres, - et _, en commençant/),
    ).not.toHaveLength(0);
    expect(
      screen.getAllByText('Le mot de passe doit contenir au moins 12 caractères.'),
    ).not.toHaveLength(0);
    expect(screen.getAllByText('Cette adresse e-mail semble incomplète.')).not.toHaveLength(0);
    expect(screen.getAllByText('Le nom ne peut pas dépasser 80 caractères.')).not.toHaveLength(0);
    await waitFor(() => expect(screen.getByLabelText(/^Nom d’affichage/)).toHaveFocus());
  });

  it('also maps violations named without a pointer', async () => {
    const { user } = await open('fr-FR', {
      [SIGN_UP]: () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: 'password', code: 'too_long' }],
        }),
    });
    await fillAll(user, FR);
    await user.click(screen.getByRole('button', { name: 'Créer mon compte' }));
    expect(
      await screen.findAllByText('Le mot de passe ne peut pas dépasser 128 caractères.'),
    ).not.toHaveLength(0);
  });

  it('shows a generic notice when a 422 names no known field', async () => {
    const { user } = await open('fr-FR', {
      [SIGN_UP]: () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/unknown', code: 'invalid_value' }],
        }),
    });
    await fillAll(user, FR);
    await user.click(screen.getByRole('button', { name: 'Créer mon compte' }));
    expect(await screen.findByText(/Certains champs sont à corriger/)).toBeInTheDocument();
  });

  it('is calm on a 429 and counts down on the button', async () => {
    const { user } = await open('fr-FR', {
      [SIGN_UP]: () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 30 }),
    });
    await fillAll(user, FR);
    await user.click(screen.getByRole('button', { name: 'Créer mon compte' }));
    expect(await screen.findByRole('button', { name: 'Réessayer dans 30 s' })).toBeInTheDocument();
    expect(screen.getByText('Trop de tentatives')).toBeInTheDocument();
  });

  it('shows a generic error with a reference on a server failure', async () => {
    const { user } = await open('fr-FR', {
      [SIGN_UP]: () => problemResponse(500, 'INTERNAL_ERROR'),
    });
    await fillAll(user, FR);
    await user.click(screen.getByRole('button', { name: 'Créer mon compte' }));
    expect(await screen.findByText(/Une erreur est survenue/)).toBeInTheDocument();
    expect(screen.getByText(/req-test-1/)).toBeInTheDocument();
  });

  it('sends one request on a double click', async () => {
    const { user, fetchMock } = await open('fr-FR', {
      [SIGN_UP]: () =>
        new Promise<Response>((resolve) => {
          setTimeout(() => {
            resolve(accepted());
          }, 50);
        }),
    });
    await fillAll(user, FR);
    await user.dblClick(screen.getByRole('button', { name: 'Créer mon compte' }));
    await new Promise((resolve) => setTimeout(resolve, 120));
    expect(callsTo(fetchMock, SIGN_UP)).toHaveLength(1);
  });
});
