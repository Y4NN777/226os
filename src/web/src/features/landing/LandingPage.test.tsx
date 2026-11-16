import { act, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { components } from '../../api/schema';
import { catalogs } from '../../i18n/messages';
import { callsTo, memberSession, mockApi, problemResponse, type Routes } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

type Question = components['schemas']['Question'];
type Article = components['schemas']['ArticleSummary'];
type Project = components['schemas']['Project'];

const author = {
  id: '6f1c2a34-0000-4000-8000-000000000001',
  username: 'awa',
  display_name: 'Awa Traoré',
};
const demoAuthor = {
  id: '6f1c2a34-0000-4000-8000-000000000099',
  username: 'demo-kofi',
  display_name: 'Kofi Démo',
};

function question(n: number, tags: string[] = ['go'], by = author): Question {
  return {
    id: `7a1c2a34-0000-4000-8000-00000000000${String(n)}`,
    title: `Erreur CORS numéro ${String(n)} sur Go`,
    body_md: 'Corps',
    lang: 'fr',
    tags,
    snippets: [],
    state: 'open',
    author: by,
    answer_count: 0,
    vote_score: 0,
    viewer_vote: 0,
    accepted_answer_id: null,
    related: [],
    created_at: '2026-11-10T10:00:00Z',
    updated_at: '2026-11-10T10:00:00Z',
  };
}

const article: Article = {
  id: '8a1c2a34-0000-4000-8000-000000000001',
  type: 'tutorial',
  title: 'Brancher Orange Money pas à pas',
  summary: 'Résumé',
  lang: 'fr',
  level: 'beginner',
  tags: ['paiement'],
  author,
  published_at: '2026-11-12T10:00:00Z',
};

const project = {
  id: '9a1c2a34-0000-4000-8000-000000000001',
  name: 'bugsecure-mcp',
  github_url: 'https://github.com/y4nn777/bugsecure-mcp',
  description: 'Serveur MCP',
  lang: 'fr',
  technologies: ['security'],
  visibility: 'public',
  owner: author,
  open_needs_count: 0,
  stage: 'active',
  license_spdx: null,
  country_code: 'BF',
  related: [],
  created_at: '2026-11-14T10:00:00Z',
  updated_at: '2026-11-14T10:00:00Z',
} as Project;

/** `more`: the list goes on after this page, so the server total is the only exact count. */
const page = (items: unknown[], total?: number, more = false) =>
  jsonResponse({
    items,
    next_cursor: more ? 'next' : null,
    ...(total === undefined ? {} : { total }),
  });

type External = components['schemas']['ExternalDeveloper'];

function developer(n: number): External {
  return {
    login: `dev${String(n)}`,
    name: `Développeur ${String(n)}`,
    avatar_url: `https://avatars.githubusercontent.com/u/${String(n)}?v=4`,
    country: 'BF',
    rank: n,
    contributions: 1000 - n,
    organizations: [],
    generated_at: '2026-11-15T10:00:00Z',
    source: 'committers.top',
  };
}

const talents = (external: External[]) => () =>
  jsonResponse({ items: [], next_cursor: null, external });

/** A full, healthy API; each test overrides what it needs. */
function healthy(overrides: Routes = {}): Routes {
  return {
    'GET /profiles': talents([1, 2, 3, 4, 5, 6].map(developer)),
    'GET /questions': () =>
      page([question(1, ['go', 'cors']), question(2), question(3), question(4)], 1280, true),
    'GET /learn/articles': () => page([article], 9),
    'GET /projects': () => page([project], 8),
    'GET /resources': () => page([], 4),
    ...overrides,
  };
}

async function open(routes: Routes = healthy()) {
  mockApi(routes);
  const view = await renderApp('/');
  await screen.findByRole('heading', { level: 1 });
  const main = screen.getByRole('main');
  return { ...view, main: within(main), mainElement: main };
}

class FakeObserver {
  static instances: FakeObserver[] = [];
  targets: Element[] = [];
  constructor(readonly callback: IntersectionObserverCallback) {
    FakeObserver.instances.push(this);
  }
  observe(target: Element) {
    this.targets.push(target);
  }
  unobserve() {}
  disconnect() {
    this.targets = [];
  }
  takeRecords() {
    return [];
  }
  /** Reports every observed target as entering (true) or leaving above the viewport (false). */
  fire(isIntersecting: boolean, only?: Element[]) {
    const targets = only ?? this.targets;
    this.callback(
      targets.map((target) => ({
        isIntersecting,
        target,
        boundingClientRect: { top: isIntersecting ? 100 : -300 },
      })) as never,
      this as never,
    );
  }
  static fireAll(isIntersecting: boolean) {
    act(() => {
      for (const observer of FakeObserver.instances) observer.fire(isIntersecting);
    });
  }
}

function stubMotion(reduced: boolean) {
  FakeObserver.instances = [];
  vi.stubGlobal('IntersectionObserver', FakeObserver);
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: reduced && query.includes('prefers-reduced-motion'),
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
    })),
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('hero', () => {
  const h1Fr = 'Les projets avancent mieux à plusieurs';

  it('states one promise and offers the sign-up call to action', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open();
    const heading = main.getByRole('heading', { level: 1, name: h1Fr });
    expect(heading).toBeInTheDocument();
    const hero = within(main.getByRole('region', { name: h1Fr }));
    expect(hero.getByRole('link', { name: 'Créer un compte' })).toHaveAttribute(
      'href',
      '/inscription',
    );
    expect(
      hero.getByText(
        'Trouvez une API de paiement mobile ou de SMS, posez votre question avec votre code, repérez les développeurs de votre pays et publiez votre projet open source.',
      ),
    ).toBeInTheDocument();
    // The accent is on the last words only.
    expect(heading.querySelector('.lp-accent')).toHaveTextContent('à plusieurs');
  });

  it('searches the site from the hero, without an account', async () => {
    setLanguages(['fr-FR']);
    const { main, router } = await open();
    const hero = within(main.getByRole('region', { name: h1Fr }));
    const form = hero.getByRole('search', { name: 'Rechercher sur 226os' });
    expect(form).toHaveAttribute('action', '/recherche');
    expect(form).toHaveAttribute('method', 'get');
    const input = within(form).getByRole('searchbox', { name: 'Rechercher sur 226os' });
    expect(input).toHaveAttribute('name', 'q');
    expect(input).toHaveAttribute('placeholder', 'Une API, une question, un développeur…');
    const user = userEvent.setup({ delay: null });
    await user.type(input, ' orange money ');
    await user.click(within(form).getByRole('button', { name: 'Rechercher' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/recherche');
    });
    expect(router.state.location.search).toEqual({ q: 'orange money' });
  });

  it('links to the five spaces from the hero, Découvrir and Connect first', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open();
    const hero = within(main.getByRole('region', { name: h1Fr }));
    const quick = within(hero.getByRole('navigation', { name: 'Accès direct aux espaces' }));
    expect(quick.getAllByRole('link').map((a) => [a.textContent, a.getAttribute('href')])).toEqual([
      ['Découvrir', '/discover'],
      ['Connect', '/connect'],
      ['Résoudre', '/solve'],
      ['Apprendre', '/apprendre'],
      ['Construire', '/build'],
    ]);
  });

  it('draws a decorative globe, hidden from assistive technology, with arcs across the continent', async () => {
    setLanguages(['fr-FR']);
    const { container } = await open();
    const globe = container.querySelector('svg.lp-globe');
    expect(globe).toHaveAttribute('aria-hidden', 'true');
    const arcs = globe?.querySelectorAll('.lp-arc').length ?? 0;
    expect(arcs).toBeGreaterThanOrEqual(25);
    expect(arcs).toBeLessThanOrEqual(35);
    // A free map of dots: no sphere, no outline, no disc.
    expect(globe?.querySelector('.lp-limb')).toBeNull();
    expect(globe?.querySelector('circle[r="468"]')).toBeNull();
  });
});

describe('landing header and footer', () => {
  it('shows the landing header on / for a visitor, with sections, language, theme and both account entries', async () => {
    setLanguages(['fr-FR']);
    await open();
    const banner = within(screen.getByRole('banner'));
    expect(banner.getByRole('link', { name: 'Accueil 226os' })).toHaveAttribute('href', '#accueil');
    const nav = within(await banner.findByRole('navigation', { name: 'Sections de la page' }));
    expect(nav.getAllByRole('link').map((a) => [a.textContent, a.getAttribute('href')])).toEqual([
      ['Espaces', '#espaces'],
      ['Avantages', '#avantages'],
      ['Activité', '#activite'],
      ['Rejoindre', '#rejoindre'],
    ]);
    expect(banner.getByRole('button', { name: 'Changer de langue' })).toBeInTheDocument();
    expect(banner.getByRole('button', { name: /^Thème (clair|sombre)$/ })).toBeInTheDocument();
    expect(banner.getByRole('link', { name: 'Se connecter' })).toHaveAttribute(
      'href',
      '/connexion',
    );
    expect(banner.getByRole('link', { name: 'Créer un compte' })).toHaveAttribute(
      'href',
      '/inscription',
    );
    // No app search bar in the header for a visitor: the search is in the hero.
    expect(banner.queryByRole('search')).not.toBeInTheDocument();
    expect(screen.queryByRole('navigation', { name: 'Barre du site' })).not.toBeInTheDocument();
  });

  it('turns to glass once the page is scrolled', async () => {
    setLanguages(['fr-FR']);
    await open();
    const banner = screen.getByRole('banner');
    expect(banner).toHaveAttribute('data-scrolled', 'false');
    Object.defineProperty(window, 'scrollY', { value: 120, configurable: true });
    act(() => {
      window.dispatchEvent(new Event('scroll'));
    });
    await waitFor(() => {
      expect(banner).toHaveAttribute('data-scrolled', 'true');
    });
    Object.defineProperty(window, 'scrollY', { value: 0, configurable: true });
  });

  it('opens a compact menu with the sections, and closes it again', async () => {
    setLanguages(['fr-FR']);
    await open();
    const user = userEvent.setup({ delay: null });
    const toggle = within(screen.getByRole('banner')).getByRole('button', {
      name: 'Ouvrir le menu',
    });
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    await user.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'true');
    const panel = document.getElementById(toggle.getAttribute('aria-controls') ?? '');
    expect(panel).not.toBeNull();
    expect(within(panel as HTMLElement).getAllByRole('link', { name: 'Espaces' })).toHaveLength(1);
    await user.keyboard('{Escape}');
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
  });

  it('switches the page language from the globe button', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open();
    const user = userEvent.setup({ delay: null });
    await user.click(
      within(screen.getByRole('banner')).getByRole('button', { name: 'Changer de langue' }),
    );
    expect(
      await screen.findByRole('menuitemradio', { name: 'Français', checked: true }),
    ).toBeInTheDocument();
    await user.click(screen.getByRole('menuitemradio', { name: 'English' }));
    expect(
      await main.findByRole('heading', {
        level: 1,
        name: 'Projects move faster together',
      }),
    ).toBeInTheDocument();
  });

  it('has an informative footer with no private detail', async () => {
    setLanguages(['fr-FR']);
    await open();
    const footer = within(screen.getByRole('contentinfo'));
    expect(
      footer.getByText('Entraide, tutoriels, API et projets entre développeurs.'),
    ).toBeInTheDocument();
    const spaces = within(footer.getByRole('navigation', { name: 'Espaces' }));
    expect(spaces.getAllByRole('link').map((a) => a.textContent)).toEqual([
      'Apprendre',
      'Résoudre',
      'Découvrir',
      'Construire',
      'Connect',
    ]);
    expect(spaces.getAllByRole('link').map((a) => a.getAttribute('href'))).toEqual([
      '/apprendre',
      '/solve',
      '/discover',
      '/build',
      '/connect',
    ]);
    const project = within(footer.getByRole('navigation', { name: 'Projet' }));
    expect(project.getByRole('link', { name: /Code source/ })).toHaveAttribute(
      'href',
      'https://github.com/Y4NN777/226os',
    );
    expect(project.getByRole('link', { name: /Contribuer/ })).toBeInTheDocument();
    expect(project.getByRole('link', { name: /Licence Apache 2\.0/ })).toBeInTheDocument();
    expect(footer.getByText('© 2026 226os')).toBeInTheDocument();
    const text = screen.getByRole('contentinfo').textContent;
    expect(text).not.toMatch(/@|localhost|https?:\/\/|v\d+\.\d+|docker|traefik/i);
    for (const link of screen.getByRole('contentinfo').querySelectorAll('a')) {
      expect(link.getAttribute('href')).not.toMatch(/^mailto:/);
    }
  });

  it('is only on / for a visitor: a member and other pages keep the app chrome', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /auth/session': () => jsonResponse(memberSession) });
    const member = await renderApp('/');
    expect(
      await screen.findByRole('navigation', { name: 'Navigation principale' }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('navigation', { name: 'Sections de la page' }),
    ).not.toBeInTheDocument();
    member.unmount();

    mockApi();
    await renderApp('/discover');
    expect(await screen.findByRole('navigation', { name: 'Barre du site' })).toBeInTheDocument();
    expect(
      screen.queryByRole('navigation', { name: 'Sections de la page' }),
    ).not.toBeInTheDocument();
  });
});

describe('scrollbar', () => {
  it('keeps the native scrollbar on the landing while clipping horizontal overflow', async () => {
    setLanguages(['fr-FR']);
    const { router } = await open();
    expect(document.documentElement).not.toHaveClass('lp-noscroll');
    expect(document.querySelector('.lp')).toBeInTheDocument();
    await act(async () => {
      await router.navigate({ to: '/discover' });
    });
    await waitFor(() => {
      expect(document.documentElement).not.toHaveClass('lp-noscroll');
    });
  });

  it('does not change the document scrollbar class when the page is unmounted', async () => {
    setLanguages(['fr-FR']);
    const view = await open();
    expect(document.documentElement).not.toHaveClass('lp-noscroll');
    view.unmount();
    expect(document.documentElement).not.toHaveClass('lp-noscroll');
  });
});

describe('every link leads somewhere real', () => {
  it('sends each link to sign-up, a section, a public space or a GitHub profile', async () => {
    setLanguages(['fr-FR']);
    const { mainElement, main } = await open();
    await main.findByRole('list', { name: 'Dernières publications' });
    await main.findByRole('region', { name: 'Les talents d’Afrique de l’Ouest' });
    const links = [...mainElement.querySelectorAll('a')];
    expect(links.length).toBeGreaterThanOrEqual(2);
    for (const link of links) {
      expect(link.getAttribute('href'), link.textContent).toMatch(
        /^(\/inscription$|#|\/(discover|connect|solve|apprendre|build)(\?|$)|https:\/\/github\.com\/[\w-]+$)/,
      );
    }
    for (const link of main.getAllByRole('link', { name: 'Créer un compte' })) {
      expect(link).toHaveAttribute('href', '/inscription');
    }
  });
});

describe('sections', () => {
  it('answers a question per section, in order, each with its anchor', async () => {
    setLanguages(['fr-FR']);
    const { mainElement, main } = await open();
    await main.findByRole('list', { name: 'Dernières publications' });
    await main.findByRole('region', { name: 'Les talents d’Afrique de l’Ouest' });
    const titles = [...mainElement.querySelectorAll('section[id]')].map((s) => [
      s.id,
      s.querySelector('h1, h2')?.textContent,
    ]);
    expect(titles).toEqual([
      ['accueil', 'Les projets avancent mieux à plusieurs'],
      ['espaces', 'Cinq espaces, un seul compte'],
      ['talents', 'Les talents d’Afrique de l’Ouest'],
      ['avantages', 'Des réponses qui tiennent la route'],
      ['activite', 'Ce qui s’écrit en ce moment'],
      ['rejoindre', 'Entrez en trois étapes'],
    ]);
  });

  it('lists at most four advantages and five spaces as links, Découvrir and Connect first', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open();
    const why = within(main.getByRole('region', { name: 'Des réponses qui tiennent la route' }));
    expect(why.getAllByRole('heading', { level: 3 }).map((h) => h.textContent)).toEqual([
      'Vérifiable',
      'Vos réalités',
      'Deux langues',
      'Travail visible',
    ]);
    const spaces = within(main.getByRole('region', { name: 'Cinq espaces, un seul compte' }));
    expect(spaces.getAllByRole('heading', { level: 3 }).map((h) => h.textContent)).toEqual([
      'Découvrir',
      'Connect',
      'Résoudre',
      'Apprendre',
      'Construire',
    ]);
    expect(spaces.getAllByRole('listitem')).toHaveLength(5);
    // Every card is an entry into its space, with a visible affordance.
    const links = spaces.getAllByRole('link');
    expect(links.map((a) => a.getAttribute('href'))).toEqual([
      '/discover',
      '/connect',
      '/solve',
      '/apprendre',
      '/build',
    ]);
    for (const link of links) expect(link).toHaveTextContent('Explorer →');
    // Découvrir and Connect carry more weight.
    expect(links.slice(0, 2).every((a) => a.closest('.lp-space-wide'))).toBe(true);
    expect(links.slice(2).some((a) => a.closest('.lp-space-wide'))).toBe(false);
    // The section follows the hero straight away.
    expect(document.getElementById('accueil')?.nextElementSibling?.id).toBe('espaces');
  });

  it('explains the three steps and ends on the sign-up button', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open();
    const join = within(main.getByRole('region', { name: 'Entrez en trois étapes' }));
    const steps = within(join.getByRole('list', { name: 'Les trois étapes' }));
    expect(steps.getAllByRole('listitem')).toHaveLength(3);
    expect(join.getByRole('link', { name: 'Créer un compte' })).toHaveAttribute(
      'href',
      '/inscription',
    );
    expect(join.queryByText(/Compte gratuit/)).not.toBeInTheDocument();
  });

  it('uses decorative photos only: empty alt, lazy, sized, and no portrait named or captioned', async () => {
    setLanguages(['fr-FR']);
    const { container, main } = await open();
    await main.findByRole('list', { name: 'Dernières publications' });
    const images = [...container.querySelectorAll<HTMLImageElement>('img[src^="/landing/"]')];
    expect(images.length).toBeGreaterThanOrEqual(10);
    for (const image of images) {
      expect(image.getAttribute('alt')).toBe('');
      expect(image.getAttribute('loading')).toBe('lazy');
      expect(Number(image.getAttribute('width'))).toBeGreaterThan(0);
    }
    const sources = images.map((image) => image.getAttribute('src'));
    expect(new Set(sources).size).toBe(sources.length);
    expect(sources).not.toContain('/landing/portrait-7.jpg');
    // The big banner and the coffee cup are gone.
    expect(sources.join(' ')).not.toMatch(/scene-desk|scene-collaborate/);
    expect(container.querySelector('figcaption')).toBeNull();
  });

  it('shows device mockups of the real interface, sharp, at their natural size, with no demo content', async () => {
    setLanguages(['fr-FR']);
    const { container, main } = await open();
    await main.findByRole('list', { name: 'Dernières publications' });
    const shots = [...container.querySelectorAll<HTMLImageElement>('img[src^="/landing/mock-"]')];
    expect(shots.map((shot) => shot.getAttribute('src'))).toEqual([
      '/landing/mock-phone-question-1x.jpg',
      '/landing/mock-desktop-article-1x.jpg',
      '/landing/mock-phone-article-1x.jpg',
    ]);
    for (const shot of shots) {
      const src = shot.getAttribute('src') ?? '';
      // 1x and 2x of the same capture, never enlarged: the 1x file is the displayed size.
      expect(shot.getAttribute('srcset')).toBe(`${src} 1x, ${src.replace('-1x', '-2x')} 2x`);
      expect(shot.getAttribute('width')).not.toBeNull();
      expect(shot.getAttribute('height')).not.toBeNull();
      expect(`${src} ${shot.getAttribute('alt') ?? ''}`).not.toMatch(/demo/i);
    }
    // The device frames are plain CSS around the picture.
    expect(container.querySelectorAll('.lp-device').length).toBe(3);
  });

  it('keeps the mockup of the advantages when there is no activity', async () => {
    setLanguages(['fr-FR']);
    const empty = () => page([], 0);
    const { container } = await open(
      healthy({
        'GET /questions': empty,
        'GET /learn/articles': empty,
        'GET /projects': empty,
        'GET /resources': empty,
      }),
    );
    expect(container.querySelectorAll('img[src^="/landing/mock-"]')).toHaveLength(1);
  });
});

describe('activity: real data only', () => {
  it('shows the latest question, article and project as plain cards, not links', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open();
    const cards = within(await main.findByRole('list', { name: 'Dernières publications' }));
    expect(cards.getAllByRole('listitem')).toHaveLength(3);
    expect(cards.getByText('Erreur CORS numéro 1 sur Go')).toBeInTheDocument();
    expect(cards.getByText('Brancher Orange Money pas à pas')).toBeInTheDocument();
    expect(cards.getByText('bugsecure-mcp')).toBeInTheDocument();
    expect(cards.getAllByText(/^Par Awa Traoré/)).toHaveLength(3);
    expect(cards.queryAllByRole('link')).toHaveLength(0);
  });

  it('never shows content whose author is a demo account', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open(
      healthy({
        'GET /questions': () =>
          page([question(1, ['go'], demoAuthor), question(2), question(3)], 3),
        'GET /learn/articles': () => page([{ ...article, author: demoAuthor }], 1),
        'GET /projects': () => page([{ ...project, owner: demoAuthor }], 1),
      }),
    );
    const cards = within(await main.findByRole('list', { name: 'Dernières publications' }));
    // The first question is a demo one: the next real one takes its place.
    expect(cards.getByText('Erreur CORS numéro 2 sur Go')).toBeInTheDocument();
    expect(cards.queryByText('Erreur CORS numéro 1 sur Go')).not.toBeInTheDocument();
    expect(cards.queryByText('Brancher Orange Money pas à pas')).not.toBeInTheDocument();
    expect(cards.queryByText('bugsecure-mcp')).not.toBeInTheDocument();
    expect(screen.queryByText(/Kofi Démo|demo-kofi/)).not.toBeInTheDocument();
  });

  it('renders no card, no number and no section when the API has nothing, and drops the nav entry', async () => {
    setLanguages(['fr-FR']);
    const empty = () => page([], 0);
    const { main } = await open(
      healthy({
        'GET /questions': empty,
        'GET /learn/articles': empty,
        'GET /projects': empty,
        'GET /resources': empty,
      }),
    );
    await waitFor(() => {
      expect(
        screen.queryByRole('heading', { name: 'Ce qui s’écrit en ce moment' }),
      ).not.toBeInTheDocument();
    });
    expect(main.queryByRole('list', { name: 'Dernières publications' })).not.toBeInTheDocument();
    expect(main.queryByRole('list', { name: 'En chiffres' })).not.toBeInTheDocument();
    expect(document.getElementById('activite')).toBeNull();
    expect(
      within(screen.getByRole('banner')).queryByRole('link', { name: 'Activité' }),
    ).not.toBeInTheDocument();
    // What is not data stays.
    expect(main.getAllByRole('heading', { level: 3 }).length).toBeGreaterThanOrEqual(9);
  });

  it('drops only the card whose request failed, without any error banner', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open(
      healthy({ 'GET /learn/articles': () => problemResponse(503, 'SERVICE_UNAVAILABLE') }),
    );
    const cards = within(await main.findByRole('list', { name: 'Dernières publications' }));
    expect(cards.getAllByRole('listitem')).toHaveLength(2);
    expect(cards.queryByText('Dernier article')).not.toBeInTheDocument();
    expect(main.queryByRole('alert')).not.toBeInTheDocument();
  });
});

describe('totals', () => {
  it('shows each real total with its noun, and nothing for a zero or a failure', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open(
      healthy({
        'GET /questions': () => page([question(1)], 1280, true),
        'GET /learn/articles': () => page([article], 10, true),
        'GET /projects': () => page([], 0),
        'GET /resources': () => problemResponse(500, 'INTERNAL_ERROR'),
      }),
    );
    const proof = within(await main.findByRole('list', { name: 'En chiffres' }));
    const items = proof.getAllByRole('listitem');
    expect(items).toHaveLength(2);
    expect(items[0]?.textContent).toMatch(/1[\s\u202f\u00a0]280.*questions$/u);
    expect(items[1]?.textContent).toMatch(/10.*articles$/u);
  });

  it('shows a total from 10 and hides it below, so a small number never reads as a weak community', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open(
      healthy({
        'GET /questions': () => page([question(1)], 9, true),
        'GET /learn/articles': () => page([article], 10, true),
        'GET /projects': () => page([project], 12, true),
        'GET /resources': () => page([], 0),
      }),
    );
    const proof = within(await main.findByRole('list', { name: 'En chiffres' }));
    const text = proof.getAllByRole('listitem').map((li) => li.textContent);
    expect(text).toHaveLength(2);
    expect(text.join(' ')).toMatch(/10.*articles/u);
    expect(text.join(' ')).toMatch(/12.*projets/u);
    expect(text.join(' ')).not.toMatch(/questions/u);
  });

  it('hides the whole list when every total is under 10, keeping the latest posts', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open(
      healthy({
        'GET /questions': () => page([question(1)], 3, true),
        'GET /learn/articles': () => page([article], 9, true),
        'GET /projects': () => page([project], 8, true),
        'GET /resources': () => page([], 4),
      }),
    );
    await main.findByRole('list', { name: 'Dernières publications' });
    expect(main.queryByRole('list', { name: 'En chiffres' })).not.toBeInTheDocument();
  });

  it('exposes the final value to screen readers at once and counts up when it enters the screen, once', async () => {
    setLanguages(['fr-FR']);
    stubMotion(false);
    const frames: FrameRequestCallback[] = [];
    vi.stubGlobal('requestAnimationFrame', (callback: FrameRequestCallback) => {
      frames.push(callback);
      return frames.length;
    });
    vi.stubGlobal('cancelAnimationFrame', vi.fn());
    const { main } = await open();
    const proof = within(await main.findByRole('list', { name: 'En chiffres' }));
    const first = proof.getAllByRole('listitem')[0] as HTMLElement;
    const final = first.querySelector('.lp-sr');
    const visible = first.querySelector('[aria-hidden="true"]');
    expect(final?.textContent).toMatch(/^1[\s\u202f\u00a0]280$/u);
    expect(visible?.textContent).toBe('0');

    FakeObserver.fireAll(true);
    await act(async () => {
      await Promise.resolve();
    });
    // Run the animation to its end: the visible number reaches the real total.
    act(() => {
      let time = 0;
      for (let guard = 0; guard < 400 && frames.length > 0; guard += 1) {
        const batch = frames.splice(0);
        time += 100;
        for (const frame of batch) frame(time);
      }
    });
    expect(visible?.textContent).toMatch(/^1[\s\u202f\u00a0]280$/u);

    // Entering again does not start it again.
    FakeObserver.fireAll(false);
    FakeObserver.fireAll(true);
    expect(frames).toHaveLength(0);
    expect(visible?.textContent).toMatch(/^1[\s\u202f\u00a0]280$/u);
  });

  it('shows the final value straight away under reduced motion', async () => {
    setLanguages(['fr-FR']);
    stubMotion(true);
    const { main } = await open();
    const proof = within(await main.findByRole('list', { name: 'En chiffres' }));
    const first = proof.getAllByRole('listitem')[0] as HTMLElement;
    expect(first.querySelector('[aria-hidden="true"]')?.textContent).toMatch(
      /^1[\s\u202f\u00a0]280$/u,
    );
  });
});

describe('scroll reveals', () => {
  it('names a direction on revealed items, among up, down, left, right, zoom-in and zoom-out', async () => {
    setLanguages(['fr-FR']);
    const { container, main } = await open();
    await main.findByRole('list', { name: 'Dernières publications' });
    const kinds = new Set(
      [...container.querySelectorAll('[data-reveal]')].map((n) => n.getAttribute('data-reveal')),
    );
    for (const kind of ['up', 'down', 'left', 'right', 'zoom-in', 'zoom-out']) {
      expect(kinds, kind).toContain(kind);
    }
  });

  it('hides items below the fold, shows them on entering, and moves them out on leaving', async () => {
    setLanguages(['fr-FR']);
    stubMotion(false);
    const { container } = await open();
    const items = [...container.querySelectorAll('[data-reveal]')];
    expect(items.length).toBeGreaterThan(8);
    expect(items.every((n) => n.getAttribute('data-reveal-state') === 'before')).toBe(true);
    FakeObserver.fireAll(true);
    expect(items.every((n) => n.getAttribute('data-reveal-state') === 'in')).toBe(true);
    FakeObserver.fireAll(false);
    expect(items.every((n) => n.getAttribute('data-reveal-state') === 'after')).toBe(true);
    FakeObserver.fireAll(true);
    expect(items.every((n) => n.getAttribute('data-reveal-state') === 'in')).toBe(true);
  });

  it('staggers the items of a group', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open();
    const why = main.getByRole('region', { name: 'Des réponses qui tiennent la route' });
    const delays = [...why.querySelectorAll<HTMLElement>('li[data-reveal]')].map((n) =>
      n.style.getPropertyValue('--rd'),
    );
    expect(delays).toEqual(['0', '1', '2', '3']);
  });

  it('never hides anything under reduced motion', async () => {
    setLanguages(['fr-FR']);
    stubMotion(true);
    const { container } = await open();
    expect(container.querySelector('.lp-hero')).toHaveAttribute('data-motion', 'reduced');
    expect(document.querySelector('[data-reveal-state]')).toBeNull();
  });

  it('keeps everything visible without IntersectionObserver', async () => {
    setLanguages(['fr-FR']);
    vi.stubGlobal('IntersectionObserver', undefined);
    const { main } = await open();
    expect(document.querySelector('[data-reveal-state]')).toBeNull();
    expect(main.getByRole('heading', { level: 1 })).toBeVisible();
  });

  it('runs the hero loops only while it is on screen', async () => {
    setLanguages(['fr-FR']);
    stubMotion(false);
    const { container } = await open();
    const hero = container.querySelector('.lp-hero') as HTMLElement;
    expect(hero).toHaveAttribute('data-motion', 'full');
    const watcher = FakeObserver.instances.find((o) => o.targets.includes(hero));
    expect(watcher).toBeDefined();
    act(() => {
      watcher?.fire(false);
    });
    expect(hero).toHaveAttribute('data-live', 'false');
    act(() => {
      watcher?.fire(true);
    });
    expect(hero).toHaveAttribute('data-live', 'true');
  });
});

describe('talents: public GitHub ranking', () => {
  it('asks the API for six GitHub developers from Burkina Faso', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(healthy());
    await renderApp('/');
    await screen.findByRole('region', { name: 'Les talents d’Afrique de l’Ouest' });
    const url = new URL(callsTo(fetchMock, 'GET /profiles')[0]?.url ?? 'http://x');
    expect(url.searchParams.get('source')).toBe('github');
    expect(url.searchParams.get('country')).toBe('BF');
    expect(url.searchParams.get('limit')).toBe('6');
  });

  it('shows six developers with rank, name, contributions, avatar and GitHub link', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open(
      healthy({ 'GET /profiles': talents([1, 2, 3, 4, 5, 6, 7].map(developer)) }),
    );
    const region = within(
      await main.findByRole('region', { name: 'Les talents d’Afrique de l’Ouest' }),
    );
    const list = within(region.getByRole('list', { name: 'Classement GitHub' }));
    const entries = list.getAllByRole('listitem');
    expect(entries).toHaveLength(6);
    expect(within(entries[0] as HTMLElement).getByText('n° 1')).toBeInTheDocument();
    expect(within(entries[0] as HTMLElement).getByText('Développeur 1')).toBeInTheDocument();
    expect(within(entries[0] as HTMLElement).getByText('999 contributions')).toBeInTheDocument();
    const links = list.getAllByRole('link');
    expect(links).toHaveLength(6);
    expect(links[0]).toHaveAttribute('href', 'https://github.com/dev1');
    for (const link of links) {
      expect(link).toHaveAttribute('target', '_blank');
      expect(link.getAttribute('rel')).toMatch(/noopener/);
      const img = link.querySelector('img');
      expect(img?.getAttribute('src')).toMatch(/^https:\/\/avatars\.githubusercontent\.com\//);
      expect(img).toHaveAttribute('referrerpolicy', 'no-referrer');
      expect(img).toHaveAttribute('loading', 'lazy');
    }
    expect(region.getByRole('link', { name: 'Voir tout le classement' })).toHaveAttribute(
      'href',
      '/connect?source=github&country=BF',
    );
    expect(region.getByRole('link', { name: 'C’est vous ? Rejoignez 226os' })).toHaveAttribute(
      'href',
      '/inscription',
    );
    expect(
      region.getByText('Source : committers.top · données publiques GitHub'),
    ).toBeInTheDocument();
  });

  it('never loads an avatar from another host', async () => {
    setLanguages(['fr-FR']);
    const evil = { ...developer(1), avatar_url: 'https://tracker.example/a.png' };
    const { main } = await open(healthy({ 'GET /profiles': talents([evil, developer(2)]) }));
    const region = await main.findByRole('region', { name: 'Les talents d’Afrique de l’Ouest' });
    expect(region.querySelector('img[src^="https://tracker.example"]')).toBeNull();
    expect(within(region).getByText('Développeur 1')).toBeInTheDocument();
  });

  it.each([
    ['fails', () => problemResponse(503, 'SERVICE_UNAVAILABLE')],
    ['has no external entry', talents([])],
  ])('renders nothing when the request %s', async (_, handler) => {
    setLanguages(['fr-FR']);
    const { main } = await open(healthy({ 'GET /profiles': handler }));
    await main.findByRole('list', { name: 'Dernières publications' });
    expect(document.getElementById('talents')).toBeNull();
    expect(main.queryByText(/committers\.top/)).not.toBeInTheDocument();
  });
});

describe('one language per page', () => {
  /** Everything a reader or a screen reader gets from the page: text and accessible names. */
  function pageText() {
    const names = [...document.querySelectorAll('[aria-label]')].map((n) =>
      n.getAttribute('aria-label'),
    );
    return `${document.body.textContent}\n${names.join('\n')}`;
  }
  const keys = Object.keys(catalogs.fr).filter((key) =>
    key.startsWith('landing.'),
  ) as (keyof typeof catalogs.fr)[];

  it('is fully French in French, with no English label', async () => {
    setLanguages(['fr-FR']);
    const { main } = await open();
    await main.findByRole('list', { name: 'Dernières publications' });
    const text = pageText();
    for (const key of keys) {
      const english = catalogs.en[key] ?? '';
      const french = catalogs.fr[key] ?? '';
      if (english === french || french.includes(english) || key === 'landing.live.by') continue;
      expect(text, key).not.toContain(english);
    }
  });

  it('is fully English in English, with no French label', async () => {
    setLanguages(['en-US']);
    const { main } = await open();
    await main.findByRole('list', { name: 'Latest posts' });
    const text = pageText();
    for (const key of keys) {
      const english = catalogs.en[key] ?? '';
      const french = catalogs.fr[key] ?? '';
      if (english === french || english.includes(french) || key === 'landing.live.by') continue;
      expect(text, key).not.toContain(french);
    }
    expect(
      main.getByRole('heading', {
        level: 1,
        name: 'Projects move faster together',
      }),
    ).toBeInTheDocument();
    expect(main.getByRole('search', { name: 'Search 226os' })).toBeInTheDocument();
    const quick = within(main.getByRole('navigation', { name: 'Go straight to a space' }));
    expect(quick.getAllByRole('link').map((a) => a.textContent)).toEqual([
      'Discover',
      'Connect',
      'Solve',
      'Learn',
      'Build',
    ]);
    const spaces = within(main.getByRole('region', { name: 'Five spaces, one account' }));
    for (const link of spaces.getAllByRole('link')) expect(link).toHaveTextContent('Explore →');
    const talentsRegion = within(await main.findByRole('region', { name: 'West Africa’s talent' }));
    expect(talentsRegion.getByText('No. 1')).toBeInTheDocument();
    expect(talentsRegion.getByRole('link', { name: 'See the full ranking' })).toBeInTheDocument();
    expect(
      talentsRegion.getByRole('link', { name: 'Is this you? Join 226os' }),
    ).toBeInTheDocument();
    expect(
      talentsRegion.getByText('Source: committers.top · public GitHub data'),
    ).toBeInTheDocument();
    expect(main.getAllByRole('link', { name: 'Create an account' })[0]).toHaveAttribute(
      'href',
      '/inscription',
    );
  });
});
