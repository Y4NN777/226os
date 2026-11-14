import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it } from 'vitest';
import { callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { api } from '../../api/http';
import { sessionQueryKey } from '../../app/session';
import { I18nProvider } from '../../i18n/I18nProvider';
import { renderApp, setLanguages } from '../../test/render';
import { ResetPasswordPage } from './ResetPasswordPage';
import { findVisibleText, getVisibleText } from '../../test/visibleText';

const CONFIRM = 'POST /auth/password-reset/confirm';
const TOKEN = 'abcdefghijklmnopqrstuvwxyz012345';
const noContent = () => new Response(null, { status: 204 });

async function open(language = 'fr-FR', routes: Parameters<typeof mockApi>[0] = {}) {
  setLanguages([language]);
  const fetchMock = mockApi(routes);
  const user = userEvent.setup();
  const app = await renderApp('/reinitialiser');
  await screen.findByRole('heading', { level: 1 });
  return { fetchMock, user, ...app };
}

async function fillAndSubmit(
  user: ReturnType<typeof userEvent.setup>,
  password = 'a brand new passphrase',
) {
  await user.type(screen.getByLabelText('Code de confirmation'), TOKEN);
  await user.type(screen.getByLabelText('Nouveau mot de passe'), password);
  await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
}

afterEach(() => {
  window.history.replaceState(null, '', '/');
});

describe('ResetPasswordPage', () => {
  it('shows the code and the new password with the live rule', async () => {
    const { user } = await open();
    expect(
      screen.getByRole('heading', { level: 1, name: 'Choisir un nouveau mot de passe' }),
    ).toBeVisible();
    expect(screen.getByLabelText('Nouveau mot de passe')).toHaveAttribute(
      'autocomplete',
      'new-password',
    );
    expect(screen.getByText('12 caractères minimum : pas encore atteint')).toBeInTheDocument();
    await user.type(screen.getByLabelText('Nouveau mot de passe'), 'a brand new passphrase');
    expect(screen.getByText('12 caractères minimum : respecté')).toBeInTheDocument();
  });

  it('toggles the password visibility', async () => {
    const { user } = await open();
    const toggle = screen.getByRole('button', { name: 'Afficher le mot de passe' });
    await user.click(toggle);
    expect(toggle).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByLabelText('Nouveau mot de passe')).toHaveAttribute('type', 'text');
  });

  it('validates the code and the password, focusing the first invalid field', async () => {
    const { user, fetchMock } = await open();
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(await screen.findAllByText('Saisissez le code reçu par e-mail.')).not.toHaveLength(0);
    expect(screen.getAllByText('Saisissez votre mot de passe.')).not.toHaveLength(0);
    await waitFor(() => expect(screen.getByLabelText('Code de confirmation')).toHaveFocus());
    await user.type(screen.getByLabelText('Code de confirmation'), TOKEN);
    await user.type(screen.getByLabelText('Nouveau mot de passe'), 'short');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(
      await screen.findAllByText('Le mot de passe doit contenir au moins 12 caractères.'),
    ).not.toHaveLength(0);
    await waitFor(() => expect(screen.getByLabelText('Nouveau mot de passe')).toHaveFocus());
    expect(callsTo(fetchMock, CONFIRM)).toHaveLength(0);
  });

  it('changes the password and offers to sign in', async () => {
    const { user, fetchMock } = await open('fr-FR', { [CONFIRM]: noContent });
    await fillAndSubmit(user);
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Mot de passe modifié' }),
    ).toBeVisible();
    expect(screen.getByText(/Tous vos appareils ont été déconnectés/)).toBeVisible();
    expect(screen.getByRole('link', { name: 'Se connecter' })).toHaveAttribute(
      'href',
      '/connexion',
    );
    const [request] = callsTo(fetchMock, CONFIRM);
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(await request?.clone().json()).toEqual({
      token: TOKEN,
      new_password: 'a brand new passphrase',
    });
  });

  it('forgets the CSRF token and invalidates the session after a reset', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ [CONFIRM]: noContent });
    const queryClient = new QueryClient();
    queryClient.setQueryData(sessionQueryKey, memberSession);
    api.csrf.set('csrf-member');
    const user = userEvent.setup();
    render(
      <QueryClientProvider client={queryClient}>
        <I18nProvider>
          <ResetPasswordPage />
        </I18nProvider>
      </QueryClientProvider>,
    );
    await fillAndSubmit(user);
    await screen.findByRole('heading', { level: 1, name: 'Mot de passe modifié' });
    expect(queryClient.getQueryState(sessionQueryKey)?.isInvalidated).toBe(true);
    expect(callsTo(fetchMock, 'GET /auth/csrf')).toHaveLength(0);
    await api.csrf.get();
    expect(callsTo(fetchMock, 'GET /auth/csrf')).toHaveLength(1);
  });

  it.each([
    ['INVALID_TOKEN', () => problemResponse(422, 'INVALID_TOKEN')],
    [
      'a violation on token',
      () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/token', code: 'too_short' }],
        }),
    ],
  ] as const)(
    'shows the invalid-code state with a link to request a new code on %s',
    async (_name, respond) => {
      const { user } = await open('fr-FR', { [CONFIRM]: respond });
      await fillAndSubmit(user);
      expect(
        await screen.findByRole('heading', {
          level: 1,
          name: 'Ce code n’est pas valide ou a expiré',
        }),
      ).toBeVisible();
      expect(screen.getByRole('link', { name: 'Demander un nouveau code' })).toHaveAttribute(
        'href',
        '/mot-de-passe-oublie',
      );
    },
  );

  it('maps a new_password violation to the password field and focuses it', async () => {
    const { user } = await open('fr-FR', {
      [CONFIRM]: () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: 'new_password', code: 'invalid_value' }],
        }),
    });
    await fillAndSubmit(user);
    expect(
      await screen.findAllByText('Le mot de passe doit contenir au moins 12 caractères.'),
    ).not.toHaveLength(0);
    await waitFor(() => expect(screen.getByLabelText('Nouveau mot de passe')).toHaveFocus());
    expect(screen.getByLabelText('Code de confirmation')).toHaveValue(TOKEN);
  });

  it('prefills the code from the fragment and removes it from the address bar', async () => {
    window.history.replaceState(null, '', `/reinitialiser#token=${TOKEN}`);
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ [CONFIRM]: noContent });
    const user = userEvent.setup();
    await renderApp('/reinitialiser');
    await screen.findByRole('heading', { level: 1 });
    expect(screen.getByLabelText('Code de confirmation')).toHaveValue(TOKEN);
    expect(window.location.hash).toBe('');
    expect(window.location.href).not.toContain(TOKEN);
    expect(callsTo(fetchMock, CONFIRM)).toHaveLength(0);
    await user.type(screen.getByLabelText('Nouveau mot de passe'), 'a brand new passphrase');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Mot de passe modifié' }),
    ).toBeVisible();
  });

  it('is calm on a 429 and counts down on the button', async () => {
    const { user } = await open('fr-FR', {
      [CONFIRM]: () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 10 }),
    });
    await fillAndSubmit(user);
    expect(await screen.findByRole('button', { name: 'Réessayer dans 10 s' })).toBeInTheDocument();
    expect(getVisibleText('Trop de tentatives')).toBeInTheDocument();
  });

  it('shows a generic error with a reference on a server failure', async () => {
    const { user } = await open('fr-FR', {
      [CONFIRM]: () => problemResponse(500, 'INTERNAL_ERROR'),
    });
    await fillAndSubmit(user);
    expect(await findVisibleText(/Une erreur est survenue/)).toBeInTheDocument();
    expect(getVisibleText(/req-test-1/)).toBeInTheDocument();
  });

  it('sends one request on a double click', async () => {
    const { user, fetchMock } = await open('fr-FR', {
      [CONFIRM]: () =>
        new Promise<Response>((resolve) => {
          setTimeout(() => {
            resolve(noContent());
          }, 50);
        }),
    });
    await user.type(screen.getByLabelText('Code de confirmation'), TOKEN);
    await user.type(screen.getByLabelText('Nouveau mot de passe'), 'a brand new passphrase');
    await user.dblClick(screen.getByRole('button', { name: 'Enregistrer' }));
    await new Promise((resolve) => setTimeout(resolve, 120));
    expect(callsTo(fetchMock, CONFIRM)).toHaveLength(1);
  });

  it('is available in English', async () => {
    const { user } = await open('en-US', { [CONFIRM]: noContent });
    expect(screen.getByRole('heading', { level: 1, name: 'Choose a new password' })).toBeVisible();
    await user.type(screen.getByLabelText('Confirmation code'), TOKEN);
    await user.type(screen.getByLabelText('New password'), 'a brand new passphrase');
    await user.click(screen.getByRole('button', { name: 'Save' }));
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Password changed' }),
    ).toBeVisible();
  });
});
