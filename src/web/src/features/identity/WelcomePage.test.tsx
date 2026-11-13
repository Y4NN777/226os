import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { memberAccount, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

describe('WelcomePage', () => {
  it('greets a signed-in member by display name with next actions and no profile form', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /auth/session': () => jsonResponse(memberSession) });
    await renderApp('/bienvenue');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Bienvenue, Awa Traoré' }),
    ).toBeVisible();
    expect(screen.getByRole('link', { name: 'Découvrir le catalogue' })).toHaveAttribute(
      'href',
      '/discover',
    );
    expect(screen.getByRole('link', { name: 'Rechercher' })).toHaveAttribute('href', '/recherche');
    expect(screen.getByRole('link', { name: 'Réglages' })).toHaveAttribute('href', '/settings');
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
  });

  it('falls back to the username when the display name is empty', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      'GET /auth/session': () =>
        jsonResponse({
          ...memberSession,
          user: { ...memberAccount, display_name: '' },
        }),
    });
    await renderApp('/bienvenue');
    expect(await screen.findByRole('heading', { level: 1, name: 'Bienvenue, awa' })).toBeVisible();
  });

  it('invites an anonymous visitor to sign in', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/bienvenue');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Bienvenue sur 226os' }),
    ).toBeVisible();
    expect(screen.getByRole('link', { name: 'Se connecter' })).toHaveAttribute(
      'href',
      '/connexion?next=/bienvenue',
    );
    expect(screen.getByRole('link', { name: 'Créer un compte' })).toHaveAttribute(
      'href',
      '/inscription',
    );
  });

  it('shows an error with a retry when the session cannot be read', async () => {
    setLanguages(['fr-FR']);
    let healthy = false;
    mockApi({
      'GET /auth/session': () =>
        healthy ? jsonResponse(memberSession) : problemResponse(500, 'INTERNAL_ERROR'),
    });
    const user = userEvent.setup();
    await renderApp('/bienvenue');
    expect(await screen.findByText(/Une erreur est survenue/)).toBeInTheDocument();
    healthy = true;
    await user.click(screen.getByRole('button', { name: 'Réessayer' }));
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Bienvenue, Awa Traoré' }),
    ).toBeVisible();
  });

  it('is available in English', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /auth/session': () => jsonResponse(memberSession) });
    await renderApp('/bienvenue');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Welcome, Awa Traoré' }),
    ).toBeVisible();
    expect(screen.getByRole('link', { name: 'Browse the catalogue' })).toBeVisible();
  });
});
