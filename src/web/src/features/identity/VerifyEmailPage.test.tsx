import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { renderApp, setLanguages } from '../../test/render';

const VERIFY = 'POST /auth/verify-email';
const TOKEN = 'abcdefghijklmnopqrstuvwxyz012345';
const noContent = () => new Response(null, { status: 204 });

async function open(language = 'fr-FR', routes: Parameters<typeof mockApi>[0] = {}) {
  setLanguages([language]);
  const fetchMock = mockApi(routes);
  const user = userEvent.setup();
  const app = await renderApp('/verifier-email');
  await screen.findByRole('heading', { level: 1 });
  return { fetchMock, user, ...app };
}

afterEach(() => {
  window.history.replaceState(null, '', '/');
});

describe('VerifyEmailPage', () => {
  it('explains where the code is and offers no resend', async () => {
    await open();
    expect(
      screen.getByRole('heading', { level: 1, name: 'Vérifiez votre boîte mail' }),
    ).toBeVisible();
    expect(screen.getByText(/contenant un code de confirmation/)).toBeVisible();
    expect(screen.getByLabelText('Code de confirmation')).toBeVisible();
    expect(screen.queryByRole('button', { name: /renvoyer/i })).not.toBeInTheDocument();
    expect(screen.queryByText(/renvoyer/i)).not.toBeInTheDocument();
  });

  it('validates the code on submit and focuses the field', async () => {
    const { user, fetchMock } = await open();
    await user.click(screen.getByRole('button', { name: 'Confirmer' }));
    expect(await screen.findAllByText('Saisissez le code reçu par e-mail.')).not.toHaveLength(0);
    await waitFor(() => expect(screen.getByLabelText('Code de confirmation')).toHaveFocus());
    await user.type(screen.getByLabelText('Code de confirmation'), 'short');
    await user.click(screen.getByRole('button', { name: 'Confirmer' }));
    expect(
      await screen.findAllByText('Le code doit contenir entre 16 et 256 caractères.'),
    ).not.toHaveLength(0);
    expect(callsTo(fetchMock, VERIFY)).toHaveLength(0);
  });

  it('confirms a pasted code and offers to sign in', async () => {
    const { user, fetchMock } = await open('fr-FR', { [VERIFY]: noContent });
    await user.click(screen.getByLabelText('Code de confirmation'));
    await user.paste(`  ${TOKEN}  `);
    await user.click(screen.getByRole('button', { name: 'Confirmer' }));
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Adresse confirmée' }),
    ).toBeVisible();
    const [request] = callsTo(fetchMock, VERIFY);
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(await request?.clone().json()).toEqual({ token: TOKEN });
    expect(screen.getByRole('link', { name: 'Se connecter' })).toHaveAttribute(
      'href',
      '/connexion?next=/bienvenue',
    );
  });

  it.each([
    ['INVALID_TOKEN', problemResponse(422, 'INVALID_TOKEN')],
    [
      'a violation on token',
      problemResponse(422, 'VALIDATION_ERROR', {
        violations: [{ field: '/token', code: 'too_short' }],
      }),
    ],
  ] as const)(
    'shows the invalid-code state on %s, then lets the user retry',
    async (_name, response) => {
      const { user } = await open('fr-FR', { [VERIFY]: () => response.clone() });
      await user.type(screen.getByLabelText('Code de confirmation'), TOKEN);
      await user.click(screen.getByRole('button', { name: 'Confirmer' }));
      expect(
        await screen.findByRole('heading', {
          level: 1,
          name: 'Ce code n’est pas valide ou a expiré',
        }),
      ).toBeVisible();
      expect(screen.getByRole('link', { name: 'Créer un compte' })).toHaveAttribute(
        'href',
        '/inscription',
      );
      expect(screen.queryByRole('button', { name: /renvoyer/i })).not.toBeInTheDocument();
      await user.click(screen.getByRole('button', { name: 'Saisir un autre code' }));
      expect(await screen.findByLabelText('Code de confirmation')).toBeVisible();
    },
  );

  it('keeps an h1 on screen while a fragment token is being checked', async () => {
    window.history.replaceState(null, '', `/verifier-email#token=${TOKEN}`);
    setLanguages(['fr-FR']);
    mockApi({
      [VERIFY]: () =>
        new Promise<Response>((resolve) => {
          setTimeout(() => {
            resolve(noContent());
          }, 80);
        }),
    });
    await renderApp('/verifier-email');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Vérifiez votre boîte mail' }),
    ).toBeVisible();
    expect(screen.getByText('Vérification du code…')).toBeInTheDocument();
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Adresse confirmée' }),
    ).toBeVisible();
  });

  it('reads a token from the fragment, removes it from the address bar and submits it', async () => {
    window.history.replaceState(null, '', `/verifier-email#token=${TOKEN}`);
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ [VERIFY]: noContent });
    await renderApp('/verifier-email');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Adresse confirmée' }),
    ).toBeVisible();
    expect(window.location.hash).toBe('');
    expect(window.location.href).not.toContain(TOKEN);
    expect(callsTo(fetchMock, VERIFY)).toHaveLength(1);
    expect(await callsTo(fetchMock, VERIFY)[0]?.clone().json()).toEqual({ token: TOKEN });
  });

  it('shows the invalid state when the fragment token is refused', async () => {
    window.history.replaceState(null, '', `/verifier-email#token=${TOKEN}`);
    setLanguages(['fr-FR']);
    mockApi({ [VERIFY]: () => problemResponse(422, 'INVALID_TOKEN') });
    await renderApp('/verifier-email');
    expect(
      await screen.findByRole('heading', {
        level: 1,
        name: 'Ce code n’est pas valide ou a expiré',
      }),
    ).toBeVisible();
    expect(window.location.hash).toBe('');
  });

  it('keeps the fragment token in the field after a transient failure', async () => {
    window.history.replaceState(null, '', `/verifier-email#token=${TOKEN}`);
    setLanguages(['fr-FR']);
    mockApi({ [VERIFY]: () => problemResponse(500, 'INTERNAL_ERROR') });
    await renderApp('/verifier-email');
    expect(await screen.findByText(/Une erreur est survenue/)).toBeInTheDocument();
    expect(screen.getByLabelText('Code de confirmation')).toHaveValue(TOKEN);
  });

  it('is calm on a 429', async () => {
    const { user } = await open('fr-FR', {
      [VERIFY]: () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 20 }),
    });
    await user.type(screen.getByLabelText('Code de confirmation'), TOKEN);
    await user.click(screen.getByRole('button', { name: 'Confirmer' }));
    expect(await screen.findByRole('button', { name: 'Réessayer dans 20 s' })).toBeInTheDocument();
    expect(screen.getByText('Trop de tentatives')).toBeInTheDocument();
  });

  it('shows the network message when the server cannot be reached', async () => {
    const { user } = await open('fr-FR', {
      [VERIFY]: () => {
        throw new TypeError('failed to fetch');
      },
    });
    await user.type(screen.getByLabelText('Code de confirmation'), TOKEN);
    await user.click(screen.getByRole('button', { name: 'Confirmer' }));
    expect(await screen.findByText(/Le serveur ne répond pas/)).toBeInTheDocument();
  });

  it('sends one request on a double click', async () => {
    const { user, fetchMock } = await open('fr-FR', {
      [VERIFY]: () =>
        new Promise<Response>((resolve) => {
          setTimeout(() => {
            resolve(noContent());
          }, 50);
        }),
    });
    await user.type(screen.getByLabelText('Code de confirmation'), TOKEN);
    await user.dblClick(screen.getByRole('button', { name: 'Confirmer' }));
    await new Promise((resolve) => setTimeout(resolve, 120));
    expect(callsTo(fetchMock, VERIFY)).toHaveLength(1);
  });

  it('is available in English', async () => {
    const { user } = await open('en-US', { [VERIFY]: noContent });
    expect(screen.getByRole('heading', { level: 1, name: 'Check your inbox' })).toBeVisible();
    await user.type(screen.getByLabelText('Confirmation code'), TOKEN);
    await user.click(screen.getByRole('button', { name: 'Confirm' }));
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Address confirmed' }),
    ).toBeVisible();
  });
});
