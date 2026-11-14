import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

const SIGN_IN = 'POST /auth/sign-in';

async function fill(user: ReturnType<typeof userEvent.setup>, email: string, password: string) {
  if (email) await user.type(screen.getByLabelText('Adresse e-mail'), email);
  if (password) await user.type(screen.getByLabelText('Mot de passe'), password);
}

/** The alert region that contains `text` (Astryx keeps its own live region in the body). */
async function findAlert(text: string | RegExp) {
  const element = await screen.findByText(text);
  const alert = element.closest<HTMLElement>('[role="alert"]');
  if (!alert) throw new Error('not inside an alert');
  return alert;
}

afterEach(() => {
  vi.useRealTimers();
});

describe('SignInPage', () => {
  it('shows the form with one h1 and the links', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/connexion?next=/recherche');
    expect(await screen.findByRole('heading', { level: 1, name: 'Connexion' })).toBeInTheDocument();
    expect(screen.getByLabelText('Adresse e-mail')).toHaveAttribute('autocomplete', 'username');
    expect(screen.getByLabelText('Mot de passe')).toHaveAttribute(
      'autocomplete',
      'current-password',
    );
    expect(screen.getByRole('link', { name: /Mot de passe oublié/ })).toHaveAttribute(
      'href',
      '/mot-de-passe-oublie',
    );
    expect(screen.getByRole('link', { name: 'Créer un compte' })).toHaveAttribute(
      'href',
      '/inscription?next=%2Frecherche',
    );
  });

  it('validates on submit, announces the errors and focuses the first invalid field', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi();
    const user = userEvent.setup();
    await renderApp('/connexion');
    await user.click(await screen.findByRole('button', { name: 'Se connecter' }));
    expect(await screen.findAllByText('Saisissez votre adresse e-mail.')).not.toHaveLength(0);
    expect(screen.getAllByText('Saisissez votre mot de passe.')).not.toHaveLength(0);
    await waitFor(() => expect(screen.getByLabelText('Adresse e-mail')).toHaveFocus());
    expect(screen.getByLabelText('Adresse e-mail')).toBeInvalid();
    expect(callsTo(fetchMock, SIGN_IN)).toHaveLength(0);
  });

  it('rejects an implausible address', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi();
    const user = userEvent.setup();
    await renderApp('/connexion');
    await screen.findByRole('heading', { level: 1 });
    await fill(user, 'awa', 'x');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));
    expect(await screen.findAllByText('Cette adresse e-mail semble incomplète.')).not.toHaveLength(
      0,
    );
    expect(callsTo(fetchMock, SIGN_IN)).toHaveLength(0);
  });

  it('signs in with the CSRF header, stores the session and goes to next', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ [SIGN_IN]: () => jsonResponse(memberSession) });
    const user = userEvent.setup();
    const { router } = await renderApp('/connexion?next=/recherche');
    await screen.findByRole('heading', { level: 1 });
    await fill(user, ' awa@example.org ', 'a long passphrase');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/recherche');
    });
    const [request] = callsTo(fetchMock, SIGN_IN);
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(await request?.clone().json()).toEqual({
      email: 'awa@example.org',
      password: 'a long passphrase',
    });
  });

  it('goes home when there is no next', async () => {
    setLanguages(['fr-FR']);
    mockApi({ [SIGN_IN]: () => jsonResponse(memberSession) });
    const user = userEvent.setup();
    const { router } = await renderApp('/connexion');
    await screen.findByRole('heading', { level: 1 });
    await fill(user, 'awa@example.org', 'a long passphrase');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/');
    });
  });

  it.each(['INVALID_CREDENTIALS', 'UNAUTHENTICATED'] as const)(
    'shows one generic banner on 401 %s and focuses it',
    async (code) => {
      setLanguages(['fr-FR']);
      mockApi({ [SIGN_IN]: () => problemResponse(401, code) });
      const user = userEvent.setup();
      await renderApp('/connexion');
      await screen.findByRole('heading', { level: 1 });
      await fill(user, 'awa@example.org', 'wrong password!');
      await user.click(screen.getByRole('button', { name: 'Se connecter' }));
      const alert = await findAlert('E-mail ou mot de passe incorrect.');
      await waitFor(() => expect(alert).toHaveFocus());
      expect(screen.getByLabelText('Mot de passe')).toHaveValue('');
    },
  );

  it.each([
    '/connexion',
    '/inscription',
    '/mot-de-passe-oublie',
    '/reinitialiser',
    '/verifier-email',
  ])('never lets a native submit put credentials in the URL (%s)', async (path) => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp(path);
    await screen.findByRole('heading', { level: 1 });
    const form = document.querySelector('main form');
    expect(form).toHaveAttribute('method', 'post');
  });

  it('counts down after a 429 and keeps submit disabled until it reaches 0', async () => {
    setLanguages(['fr-FR']);
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const fetchMock = mockApi({
      [SIGN_IN]: () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 3 }),
    });
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    await renderApp('/connexion');
    await screen.findByRole('heading', { level: 1 });
    await fill(user, 'awa@example.org', 'a long passphrase');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));
    const waiting = await screen.findByRole('button', { name: 'Réessayer dans 3 s' });
    expect(waiting).toHaveAttribute('aria-disabled', 'true');
    await user.click(waiting);
    expect(callsTo(fetchMock, SIGN_IN)).toHaveLength(1);
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(screen.getByRole('button', { name: 'Réessayer dans 2 s' })).toBeInTheDocument();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(screen.getByRole('button', { name: 'Se connecter' })).not.toHaveAttribute(
      'aria-disabled',
      'true',
    );
  });

  it('is calm on a 429 without Retry-After', async () => {
    setLanguages(['fr-FR']);
    mockApi({ [SIGN_IN]: () => problemResponse(429, 'RATE_LIMITED') });
    const user = userEvent.setup();
    await renderApp('/connexion');
    await screen.findByRole('heading', { level: 1 });
    await fill(user, 'awa@example.org', 'a long passphrase');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));
    expect(await findAlert('Trop de tentatives')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Se connecter' })).toBeInTheDocument();
  });

  it('shows a generic error with its reference on a server failure', async () => {
    setLanguages(['fr-FR']);
    mockApi({ [SIGN_IN]: () => problemResponse(500, 'INTERNAL_ERROR') });
    const user = userEvent.setup();
    await renderApp('/connexion');
    await screen.findByRole('heading', { level: 1 });
    await fill(user, 'awa@example.org', 'a long passphrase');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));
    const alert = await findAlert(/Une erreur est survenue/);
    expect(alert).toHaveTextContent('req-test-1');
  });

  it('shows the network message when the server cannot be reached', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      [SIGN_IN]: () => {
        throw new TypeError('failed to fetch');
      },
    });
    const user = userEvent.setup();
    await renderApp('/connexion');
    await screen.findByRole('heading', { level: 1 });
    await fill(user, 'awa@example.org', 'a long passphrase');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));
    expect(await findAlert(/Le serveur ne répond pas/)).toBeInTheDocument();
  });

  it('sends one request on a double click', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      [SIGN_IN]: () =>
        new Promise<Response>((resolve) => {
          setTimeout(() => {
            resolve(jsonResponse(memberSession));
          }, 50);
        }),
    });
    const user = userEvent.setup();
    await renderApp('/connexion');
    await screen.findByRole('heading', { level: 1 });
    await fill(user, 'awa@example.org', 'a long passphrase');
    const submit = screen.getByRole('button', { name: 'Se connecter' });
    await user.dblClick(submit);
    await waitFor(() => {
      expect(callsTo(fetchMock, SIGN_IN)).toHaveLength(1);
    });
    await new Promise((resolve) => setTimeout(resolve, 100));
    expect(callsTo(fetchMock, SIGN_IN)).toHaveLength(1);
  });

  it('toggles the password visibility with an accessible button', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    const user = userEvent.setup();
    await renderApp('/connexion');
    await screen.findByRole('heading', { level: 1 });
    const toggle = screen.getByRole('button', { name: 'Afficher le mot de passe' });
    expect(toggle).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByLabelText('Mot de passe')).toHaveAttribute('type', 'password');
    await user.click(toggle);
    expect(toggle).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByLabelText('Mot de passe')).toHaveAttribute('type', 'text');
    await user.click(toggle);
    expect(screen.getByLabelText('Mot de passe')).toHaveAttribute('type', 'password');
  });

  it('allows pasting a password', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    const user = userEvent.setup();
    await renderApp('/connexion');
    await screen.findByRole('heading', { level: 1 });
    await user.click(screen.getByLabelText('Mot de passe'));
    await user.paste('pasted passphrase');
    expect(screen.getByLabelText('Mot de passe')).toHaveValue('pasted passphrase');
  });

  it('is available in English', async () => {
    setLanguages(['en-US']);
    mockApi({ [SIGN_IN]: () => problemResponse(401, 'INVALID_CREDENTIALS') });
    const user = userEvent.setup();
    await renderApp('/connexion');
    expect(await screen.findByRole('heading', { level: 1, name: 'Sign in' })).toBeInTheDocument();
    await user.type(screen.getByLabelText('Email address'), 'awa@example.org');
    await user.type(screen.getByLabelText('Password'), 'wrong password!');
    await user.click(screen.getByRole('button', { name: 'Sign in' }));
    expect(await findAlert('Incorrect email or password.')).toBeInTheDocument();
  });
});
