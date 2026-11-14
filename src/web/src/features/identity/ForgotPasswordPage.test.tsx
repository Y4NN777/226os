import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

const REQUEST = 'POST /auth/password-reset/request';
const accepted = () => jsonResponse({ status: 'accepted' }, { status: 202 });

async function open(language = 'fr-FR', routes: Parameters<typeof mockApi>[0] = {}) {
  setLanguages([language]);
  const fetchMock = mockApi(routes);
  const user = userEvent.setup();
  const app = await renderApp('/mot-de-passe-oublie');
  await screen.findByRole('heading', { level: 1 });
  return { fetchMock, user, ...app };
}

describe('ForgotPasswordPage', () => {
  it('shows the form and a way back to sign in', async () => {
    await open();
    expect(screen.getByRole('heading', { level: 1, name: 'Mot de passe oublié' })).toBeVisible();
    expect(screen.getByLabelText('Adresse e-mail')).toHaveAttribute('autocomplete', 'email');
    expect(screen.getByRole('link', { name: 'Retour à la connexion' })).toHaveAttribute(
      'href',
      '/connexion',
    );
  });

  it('validates the address and focuses the field', async () => {
    const { user, fetchMock } = await open();
    await user.click(screen.getByRole('button', { name: 'Envoyer le code' }));
    expect(await screen.findAllByText('Saisissez votre adresse e-mail.')).not.toHaveLength(0);
    await waitFor(() => expect(screen.getByLabelText('Adresse e-mail')).toHaveFocus());
    await user.type(screen.getByLabelText('Adresse e-mail'), 'awa');
    await user.click(screen.getByRole('button', { name: 'Envoyer le code' }));
    expect(await screen.findAllByText('Cette adresse e-mail semble incomplète.')).not.toHaveLength(
      0,
    );
    expect(callsTo(fetchMock, REQUEST)).toHaveLength(0);
  });

  it('always ends on the same neutral confirmation, with a link to the reset page', async () => {
    const { user, fetchMock, router } = await open('fr-FR', { [REQUEST]: accepted });
    await user.type(screen.getByLabelText('Adresse e-mail'), ' awa@example.org ');
    await user.click(screen.getByRole('button', { name: 'Envoyer le code' }));
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Consultez votre boîte mail' }),
    ).toBeVisible();
    expect(
      screen.getByText(
        'Si un compte existe pour cette adresse, un code vient de partir. Il expire dans 30 minutes.',
      ),
    ).toBeVisible();
    expect(screen.getByRole('link', { name: 'Saisir le code' })).toHaveAttribute(
      'href',
      '/reinitialiser',
    );
    const [request] = callsTo(fetchMock, REQUEST);
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(await request?.clone().json()).toEqual({ email: 'awa@example.org' });
    expect(router.state.location.searchStr).toBe('');
    expect(sessionStorage.length + localStorage.length).toBe(0);
  });

  it('is calm on a 429 and counts down on the button', async () => {
    const { user } = await open('fr-FR', {
      [REQUEST]: () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 45 }),
    });
    await user.type(screen.getByLabelText('Adresse e-mail'), 'awa@example.org');
    await user.click(screen.getByRole('button', { name: 'Envoyer le code' }));
    expect(await screen.findByRole('button', { name: 'Réessayer dans 45 s' })).toBeInTheDocument();
    expect(screen.getByText('Trop de tentatives')).toBeInTheDocument();
  });

  it('shows a generic error with a reference on a server failure', async () => {
    const { user } = await open('fr-FR', {
      [REQUEST]: () => problemResponse(500, 'INTERNAL_ERROR'),
    });
    await user.type(screen.getByLabelText('Adresse e-mail'), 'awa@example.org');
    await user.click(screen.getByRole('button', { name: 'Envoyer le code' }));
    expect(await screen.findByText(/Une erreur est survenue/)).toBeInTheDocument();
    expect(screen.getByText(/req-test-1/)).toBeInTheDocument();
  });

  it('sends one request on a double click', async () => {
    const { user, fetchMock } = await open('fr-FR', {
      [REQUEST]: () =>
        new Promise<Response>((resolve) => {
          setTimeout(() => {
            resolve(accepted());
          }, 50);
        }),
    });
    await user.type(screen.getByLabelText('Adresse e-mail'), 'awa@example.org');
    await user.dblClick(screen.getByRole('button', { name: 'Envoyer le code' }));
    await new Promise((resolve) => setTimeout(resolve, 120));
    expect(callsTo(fetchMock, REQUEST)).toHaveLength(1);
  });

  it('is available in English', async () => {
    const { user } = await open('en-US', { [REQUEST]: accepted });
    expect(screen.getByRole('heading', { level: 1, name: 'Forgot your password?' })).toBeVisible();
    await user.type(screen.getByLabelText('Email address'), 'awa@example.org');
    await user.click(screen.getByRole('button', { name: 'Send the code' }));
    expect(await screen.findByText(/expires in 30 minutes/)).toBeVisible();
  });
});
