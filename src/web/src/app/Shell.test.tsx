import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, memberSession, mockApi, problemResponse } from '../test/api';
import { jsonResponse, renderApp, setLanguages } from '../test/render';

function primaryNav() {
  return screen.getByRole('navigation', { name: 'Navigation principale' });
}

describe('Shell', () => {
  it('offers sign-in and sign-up to an anonymous visitor', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    expect(await screen.findByRole('link', { name: 'Se connecter' })).toHaveAttribute(
      'href',
      '/connexion',
    );
    expect(screen.queryByRole('button', { name: /Compte/ })).not.toBeInTheDocument();
  });

  it('lists Apprendre and Connect before the other pillars, all in one navigation', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    const nav = await screen.findByRole('navigation', { name: 'Navigation principale' });
    const pillars = ['Apprendre', 'Connect', 'Résoudre', 'Découvrir', 'Construire'];
    const inDocumentOrder = within(nav)
      .getAllByText(new RegExp(`^(${pillars.join('|')})$`))
      .map((el) => el.textContent);
    const order = pillars.map((name) => inDocumentOrder.indexOf(name));
    expect(order.every((index) => index >= 0)).toBe(true);
    expect([...order].sort((a, b) => a - b)).toEqual(order);
  });

  it('labels areas that are not built yet as coming soon, never as live links', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    const solve = within(primaryNav()).getByText('Résoudre').closest('a, button, [aria-disabled]');
    expect(solve).toHaveAttribute('aria-disabled', 'true');
    expect(within(primaryNav()).getAllByText('Bientôt').length).toBeGreaterThan(0);
  });

  it('links Discover as a live pillar and offers search from the top bar', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    expect(within(primaryNav()).getByRole('link', { name: 'Découvrir' })).toHaveAttribute(
      'href',
      '/discover',
    );
    expect(screen.getByRole('link', { name: 'Rechercher' })).toHaveAttribute('href', '/recherche');
  });

  it('shows the signed-in member and signs out with the session CSRF token', async () => {
    setLanguages(['fr-FR']);
    let signedIn = true;
    const fetchMock = mockApi({
      'GET /auth/session': () =>
        jsonResponse(
          signedIn ? memberSession : { authenticated: false, user: null, csrf_token: 'csrf-anon' },
        ),
      'POST /auth/sign-out': () => {
        signedIn = false;
        return new Response(null, { status: 204 });
      },
    });
    await renderApp('/');
    await userEvent.click(await screen.findByRole('button', { name: /Awa Traoré/ }));
    await userEvent.click(await screen.findByRole('menuitem', { name: 'Se déconnecter' }));
    expect(await screen.findByRole('link', { name: 'Se connecter' })).toBeInTheDocument();
    const [signOut] = callsTo(fetchMock, 'POST /auth/sign-out');
    expect(signOut?.headers.get('X-CSRF-Token')).toBe('csrf-member');
  });

  it('treats signing out of an already expired session as signed out', async () => {
    setLanguages(['fr-FR']);
    let signedIn = true;
    mockApi({
      'GET /auth/session': () =>
        jsonResponse(
          signedIn ? memberSession : { authenticated: false, user: null, csrf_token: 'csrf-anon' },
        ),
      'POST /auth/sign-out': () => {
        signedIn = false;
        return problemResponse(401, 'UNAUTHENTICATED');
      },
    });
    await renderApp('/');
    await userEvent.click(await screen.findByRole('button', { name: /Awa Traoré/ }));
    await userEvent.click(await screen.findByRole('menuitem', { name: 'Se déconnecter' }));
    expect(await screen.findByRole('link', { name: 'Se connecter' })).toBeInTheDocument();
    expect(screen.queryByText('La déconnexion a échoué. Réessayez.')).not.toBeInTheDocument();
  });

  it('switches the interface language from the account menu', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await userEvent.click(await screen.findByRole('button', { name: 'Plus d’options' }));
    await userEvent.click(await screen.findByRole('menuitem', { name: 'English' }));
    await waitFor(() => {
      expect(screen.getByRole('link', { name: 'Sign in' })).toBeInTheDocument();
    });
    expect(screen.getByRole('navigation', { name: 'Main navigation' })).toBeInTheDocument();
  });

  it('still offers sign-in when the session cannot be read', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /auth/session': () => new Response('down', { status: 503 }) });
    await renderApp('/');
    expect(await screen.findByRole('link', { name: 'Sign in' })).toBeInTheDocument();
  });
});
