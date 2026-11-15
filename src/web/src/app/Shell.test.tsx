import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { SideNav } from '@astryxdesign/core/SideNav';
import { HammerIcon, HelpIcon } from '../components/icons';
import { callsTo, memberAccount, memberSession, mockApi, problemResponse } from '../test/api';
import { jsonResponse, renderApp, renderWithProviders, setLanguages } from '../test/render';
import { NavSections } from './Shell';

function primaryNav() {
  return screen.getByRole('navigation', { name: 'Navigation principale' });
}

/** Below the phone breakpoint: every media query matches, and the drawer can open in jsdom. */
function stubPhone() {
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: true,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  );
  // jsdom has no <dialog>.showModal(); a stand-in lets the drawer open so its tree can be read.
  HTMLDialogElement.prototype.showModal = function showModal(this: HTMLDialogElement) {
    this.setAttribute('open', '');
  };
  HTMLDialogElement.prototype.close = function close(this: HTMLDialogElement) {
    this.removeAttribute('open');
  };
}

/**
 * The site bar. Found by its own label: page sections that use `<header>` (the landing page) are
 * also `banner` landmarks for Testing Library inside AppShell's `role="main"` div.
 */
function topBar() {
  return screen.getByRole('navigation', { name: 'Barre du site' });
}

function bottomBar() {
  return screen.getByRole('navigation', { name: 'Navigation rapide' });
}

describe('Shell', () => {
  it('keeps the rest of the navigation in a drawer on a phone, opened by « Plus »', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    stubPhone();
    await renderApp('/');
    // The prototype's phone shell has no top-bar hamburger: « Plus » in the bottom bar opens it.
    await screen.findByRole('navigation', { name: 'Navigation rapide' });
    expect(screen.queryByRole('button', { name: 'Ouvrir la navigation' })).not.toBeInTheDocument();
    const toggle = within(bottomBar()).getByRole('button', { name: 'Plus' });
    // Closed: the drawer-only links are not laid out in the page under the content.
    expect(screen.queryByRole('link', { name: 'Construire' })).not.toBeInTheDocument();
    await userEvent.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'true');
    // The drawer is a modal dialog in the accessibility tree, not only a native top-layer one.
    const drawer = screen.getByRole('dialog', { name: 'Navigation', hidden: true });
    expect(drawer).toHaveAttribute('aria-modal', 'true');
    expect(toggle).toHaveAttribute('aria-controls', drawer.id);
    // As in the prototype, the drawer slides in from the left, whatever side the toggle is on.
    expect(drawer).toHaveAttribute('data-side', 'start');
    for (const name of ['Résoudre', 'Connect', 'Construire']) {
      expect(within(drawer).getByRole('link', { name, hidden: true })).toBeInTheDocument();
    }
  });

  it('closes the drawer when one of its links is followed', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    stubPhone();
    const { router } = await renderApp('/');
    await screen.findByRole('navigation', { name: 'Navigation rapide' });
    const toggle = within(bottomBar()).getByRole('button', { name: 'Plus' });
    const user = userEvent.setup({ delay: null });
    await user.click(toggle);
    const drawer = screen.getByRole('dialog', { name: 'Navigation', hidden: true });
    await user.click(within(drawer).getByRole('link', { name: 'Construire', hidden: true }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/build');
    });
    await waitFor(() => {
      expect(toggle).toHaveAttribute('aria-expanded', 'false');
    });
  });

  it('closes the drawer on Escape (the dialog cancel)', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    stubPhone();
    await renderApp('/');
    await screen.findByRole('navigation', { name: 'Navigation rapide' });
    const toggle = within(bottomBar()).getByRole('button', { name: 'Plus' });
    const user = userEvent.setup({ delay: null });
    await user.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'true');
    await user.keyboard('{Escape}');
    await waitFor(() => {
      expect(toggle).toHaveAttribute('aria-expanded', 'false');
    });
  });

  it('shows the phone bottom bar: Accueil, Apprendre, Résoudre, Découvrir and Plus', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    stubPhone();
    await renderApp('/discover');
    await screen.findByRole('navigation', { name: 'Navigation rapide' });
    const links = within(bottomBar()).getAllByRole('link');
    expect(links.map((link) => [link.textContent, link.getAttribute('href')])).toEqual([
      ['Accueil', '/'],
      ['Apprendre', '/apprendre'],
      ['Résoudre', '/solve'],
      ['Découvrir', '/discover'],
    ]);
    // The current area is exposed, not only drawn; the icons are decorative.
    expect(within(bottomBar()).getByRole('link', { name: 'Découvrir' })).toHaveAttribute(
      'aria-current',
      'page',
    );
    expect(within(bottomBar()).getByRole('link', { name: 'Accueil' })).not.toHaveAttribute(
      'aria-current',
    );
    for (const link of links) {
      expect(link.querySelector('svg')?.closest('[aria-hidden="true"]')).not.toBeNull();
    }
    expect(within(bottomBar()).getByRole('button', { name: 'Plus' })).toBeInTheDocument();
  });

  it('has no bottom bar above the phone breakpoint', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    expect(screen.queryByRole('navigation', { name: 'Navigation rapide' })).not.toBeInTheDocument();
  });

  it('puts the visitor account links in the phone drawer', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    stubPhone();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    await userEvent.click(within(bottomBar()).getByRole('button', { name: 'Plus' }));
    const drawer = screen.getByRole('dialog', { name: 'Navigation', hidden: true });
    expect(
      within(drawer).getByRole('link', { name: 'Se connecter', hidden: true }),
    ).toHaveAttribute('href', '/connexion');
    expect(
      within(drawer).getByRole('link', { name: 'Créer un compte', hidden: true }),
    ).toHaveAttribute('href', '/inscription');
  });

  it('puts the member account links and sign-out in the phone drawer', async () => {
    setLanguages(['fr-FR']);
    let signedIn = true;
    mockApi({
      'GET /auth/session': () =>
        jsonResponse(
          signedIn ? memberSession : { authenticated: false, user: null, csrf_token: 'csrf-anon' },
        ),
      'POST /auth/sign-out': () => {
        signedIn = false;
        return new Response(null, { status: 204 });
      },
    });
    stubPhone();
    await renderApp('/');
    await screen.findByRole('button', { name: /Awa Traoré/ });
    const user = userEvent.setup({ delay: null });
    await user.click(within(bottomBar()).getByRole('button', { name: 'Plus' }));
    const drawer = screen.getByRole('dialog', { name: 'Navigation', hidden: true });
    // The account entries form their own group, named for what it holds.
    const account = within(drawer).getByRole('group', { name: 'Compte', hidden: true });
    expect(
      within(account).getByRole('link', { name: 'Mon profil', hidden: true }),
    ).toBeInTheDocument();
    expect(within(drawer).getByRole('link', { name: 'Mon profil', hidden: true })).toHaveAttribute(
      'href',
      `/profils/${memberAccount.username}`,
    );
    expect(within(drawer).getByRole('link', { name: 'Paramètres', hidden: true })).toHaveAttribute(
      'href',
      '/parametres',
    );
    await user.click(within(drawer).getByRole('button', { name: 'Se déconnecter', hidden: true }));
    expect(await screen.findByRole('link', { name: 'Se connecter' })).toBeInTheDocument();
  });

  it('offers « Poser une question » in the top bar to a signed-in member', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /auth/session': () => jsonResponse(memberSession) });
    await renderApp('/');
    await screen.findByRole('button', { name: /Awa Traoré/ });
    const ask = within(topBar()).getByRole('link', {
      name: 'Poser une question',
    });
    // The solve page opens its ask view on this fragment.
    expect(ask).toHaveAttribute('href', '/solve#ask');
    expect(ask.querySelector('svg')?.closest('[aria-hidden="true"]')).not.toBeNull();
  });

  it('does not offer « Poser une question » to a visitor', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    expect(
      within(topBar()).queryByRole('link', { name: 'Poser une question' }),
    ).not.toBeInTheDocument();
  });

  it('ends every page with the 226os footer and links only to documents that exist', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    const footer = screen.getByRole('contentinfo');
    expect(within(footer).getByText('226os')).toBeInTheDocument();
    expect(within(footer).getByRole('link', { name: /^Licence/ })).toHaveAttribute(
      'href',
      'https://github.com/Y4NN777/226os/blob/main/LICENSE',
    );
    expect(within(footer).getByRole('link', { name: /^Contribuer/ })).toHaveAttribute(
      'href',
      'https://github.com/Y4NN777/226os/blob/main/CONTRIBUTING.md',
    );
    // Like every external link of the app: a new tab, said in the link text, no opener.
    for (const link of within(footer).getAllByRole('link')) {
      expect(link).toHaveAttribute('target', '_blank');
      expect(link.getAttribute('rel')).toMatch(/noopener/);
      expect(link.getAttribute('rel')).toMatch(/noreferrer/);
      expect(link).toHaveTextContent(/nouvel onglet/);
    }
    // No code of conduct or privacy page exists yet: the prototype's links to them are left out.
    expect(within(footer).queryByRole('link', { name: /conduite|Confidentialité/ })).toBeNull();
  });

  it('names the settings page « Paramètres » in the menu and the document title', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/parametres');
    const user = userEvent.setup({ delay: null });
    await user.click(await screen.findByRole('button', { name: 'Préférences' }));
    expect(await screen.findByRole('menuitem', { name: 'Paramètres' })).toHaveAttribute(
      'href',
      '/parametres',
    );
    await waitFor(() => {
      expect(document.title).toBe('Paramètres · 226os');
    });
  });

  it('offers sign-in and sign-up to an anonymous visitor', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    expect(await screen.findByRole('link', { name: 'Se connecter' })).toHaveAttribute(
      'href',
      '/connexion',
    );
    // The landing page has its own call to action; this one is the top bar's.
    expect(within(topBar()).getByRole('link', { name: 'Créer un compte' })).toHaveAttribute(
      'href',
      '/inscription',
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

  it('links every pillar, Connect included, and shows no area as coming soon', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    // Connect went live with the E5 directory: it is a link now, and nothing is left disabled.
    expect(within(primaryNav()).getByRole('link', { name: 'Connect' })).toHaveAttribute(
      'href',
      '/connect',
    );
    expect(within(primaryNav()).queryByText('Bientôt')).not.toBeInTheDocument();
  });

  it('links Résoudre as a live pillar', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    expect(within(primaryNav()).getByRole('link', { name: 'Résoudre' })).toHaveAttribute(
      'href',
      '/solve',
    );
  });

  it('links Discover as a live pillar and offers search from the top bar', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    const { router } = await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    expect(within(primaryNav()).getByRole('link', { name: 'Découvrir' })).toHaveAttribute(
      'href',
      '/discover',
    );
    // The top bar holds the search field itself (prototype shell), not a link to the page.
    const search = screen.getByRole('search', { name: 'Rechercher sur 226os' });
    await userEvent
      .setup({ delay: null })
      .type(
        within(search).getByRole('searchbox', { name: 'Rechercher sur 226os' }),
        ' webhook {Enter}',
      );
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/recherche');
    });
    expect(router.state.location.search).toEqual({ q: 'webhook' });
  });

  it('opens the search page without a query when the top bar field is empty', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    const { router } = await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    await userEvent
      .setup({ delay: null })
      .type(screen.getByRole('searchbox', { name: 'Rechercher sur 226os' }), '{Enter}');
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/recherche');
    });
    expect(router.state.location.search).toEqual({});
  });

  it('shows the 226os logo as the link home', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    const brand = screen.getAllByRole('link', { name: '226os' })[0];
    expect(brand).toHaveAttribute('href', '/');
    expect(within(brand as HTMLElement).getByRole('img', { name: '226os' })).toBeInTheDocument();
  });

  it('switches between light and dark from the top bar', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    const toggle = await screen.findByRole('button', { name: 'Thème sombre' });
    expect(toggle).toHaveAttribute('aria-pressed', 'false');
    await userEvent.setup({ delay: null }).click(toggle);
    expect(screen.getByRole('button', { name: 'Thème sombre' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
    expect(localStorage.getItem('226os.theme')).toBe('dark');
  });

  it('draws an icon before each navigation item, hidden from assistive technology', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    const links = within(primaryNav()).getAllByRole('link');
    expect(links.map((link) => link.textContent)).toEqual([
      'Accueil',
      'Apprendre',
      'Connect',
      'Résoudre',
      'Découvrir',
      'Construire',
    ]);
    for (const link of links) {
      const icon = link.querySelector('svg');
      expect(icon).not.toBeNull();
      expect(icon?.closest('[aria-hidden="true"]')).not.toBeNull();
    }
  });

  it('marks the current area in the sidebar', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await screen.findByRole('link', { name: 'Se connecter' });
    expect(within(primaryNav()).getByRole('link', { name: 'Accueil' })).toHaveAttribute(
      'aria-current',
      'page',
    );
    expect(within(primaryNav()).getByRole('link', { name: 'Découvrir' })).not.toHaveAttribute(
      'aria-current',
    );
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
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('button', { name: /Awa Traoré/ }));
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('menuitem', { name: 'Se déconnecter' }));
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
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('button', { name: /Awa Traoré/ }));
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('menuitem', { name: 'Se déconnecter' }));
    expect(await screen.findByRole('link', { name: 'Se connecter' })).toBeInTheDocument();
    expect(screen.queryByText('La déconnexion a échoué. Réessayez.')).not.toBeInTheDocument();
  });

  it('switches the interface language from the account menu', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('button', { name: 'Préférences' }));
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('menuitemradio', { name: 'English' }));
    await waitFor(() => {
      expect(screen.getByRole('link', { name: 'Sign in' })).toBeInTheDocument();
    });
    expect(screen.getByRole('navigation', { name: 'Main navigation' })).toBeInTheDocument();
  });

  it('exposes the current language as a checked radio in the anonymous menu', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('button', { name: 'Préférences' }));
    expect(
      await screen.findByRole('menuitemradio', { name: 'Français', checked: true }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('menuitemradio', { name: 'English', checked: false }),
    ).toBeInTheDocument();
  });

  it('offers the theme as radios in the preferences menu', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/');
    const user = userEvent.setup({ delay: null });
    await user.click(await screen.findByRole('button', { name: 'Préférences' }));
    expect(
      await screen.findByRole('menuitemradio', { name: 'Système', checked: true }),
    ).toBeInTheDocument();
    await user.click(screen.getByRole('menuitemradio', { name: 'Sombre' }));
    expect(localStorage.getItem('226os.theme')).toBe('dark');
  });

  it('exposes the current language as a checked radio in the account menu', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /auth/session': () => jsonResponse(memberSession) });
    await renderApp('/');
    await userEvent
      .setup({ delay: null })
      .click(await screen.findByRole('button', { name: /Awa Traoré/ }));
    expect(
      await screen.findByRole('menuitemradio', { name: 'English', checked: true }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('menuitemradio', { name: 'Français', checked: false }),
    ).toBeInTheDocument();
  });

  it('still offers sign-in when the session cannot be read', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /auth/session': () => new Response('down', { status: 503 }) });
    await renderApp('/');
    expect(await screen.findByRole('link', { name: 'Sign in' })).toBeInTheDocument();
  });
});

describe('NavSections', () => {
  it('shows an area that is not built yet as « Bientôt », never as a dead link', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(
      <SideNav aria-label="Navigation principale">
        <NavSections
          pathname="/"
          sections={[
            {
              title: 'nav.section.pillars',
              items: [
                { label: 'nav.solve', icon: HelpIcon, to: '/solve' },
                { label: 'nav.build', icon: HammerIcon },
              ],
            },
          ]}
        />
      </SideNav>,
    );
    expect(screen.getByRole('link', { name: 'Résoudre' })).toHaveAttribute('href', '/solve');
    expect(screen.queryByRole('link', { name: /Construire/ })).not.toBeInTheDocument();
    expect(screen.getByText('Construire')).toBeInTheDocument();
    expect(screen.getByText('Bientôt')).toBeInTheDocument();
  });
});
