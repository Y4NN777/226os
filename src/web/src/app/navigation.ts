import {
  BookIcon,
  CompassIcon,
  HammerIcon,
  HelpIcon,
  HomeIcon,
  NetworkIcon,
  type SvgIcon,
} from '../components/icons';
import type { MessageKey } from '../i18n/messages';

export interface NavItem {
  label: MessageKey;
  /** Drawn before the label; decorative, the label names the link. */
  icon: SvgIcon;
  /** Route path when live; absent while the area is not built ("Bientôt", never a dead link). */
  to?: string;
}

export interface NavSection {
  title: MessageKey;
  items: NavItem[];
}

/** A live navigation entry: one with a route. */
export type LiveNavItem = NavItem & { to: string };

export const homeItem: LiveNavItem = { label: 'nav.home', icon: HomeIcon, to: '/' };

/** Whether `to` is the current area: the home only on `/`, any other area on its subpaths too. */
export function isCurrent(pathname: string, to: string) {
  return to === '/' ? pathname === '/' : pathname === to || pathname.startsWith(`${to}/`);
}

/**
 * Pillar emphasis (O7): Apprendre and Connect come first; Solve, Discover and Build stay
 * full pillars in the same navigation, on desktop and in the mobile drawer.
 */
export const navSections: NavSection[] = [
  {
    title: 'nav.section.grow',
    items: [
      { label: 'nav.learn', icon: BookIcon, to: '/apprendre' },
      { label: 'nav.connect', icon: NetworkIcon, to: '/connect' },
    ],
  },
  {
    title: 'nav.section.pillars',
    items: [
      { label: 'nav.solve', icon: HelpIcon, to: '/solve' },
      { label: 'nav.discover', icon: CompassIcon, to: '/discover' },
      { label: 'nav.build', icon: HammerIcon, to: '/build' },
    ],
  },
];

/** The prototype's tablet range, where the sidebar becomes an icon rail (92px). */
export const railQuery = '(min-width: 768px) and (max-width: 1023px)';

/**
 * The phone bottom bar (prototype `.bottom-nav`: four areas, then « Plus »). Mapped to our
 * navigation: Accueil, Apprendre, Résoudre, Découvrir; Connect and Construire stay in the drawer.
 */
export const bottomNavItems: readonly LiveNavItem[] = [
  homeItem,
  { label: 'nav.learn', icon: BookIcon, to: '/apprendre' },
  { label: 'nav.solve', icon: HelpIcon, to: '/solve' },
  { label: 'nav.discover', icon: CompassIcon, to: '/discover' },
];

const repository = 'https://github.com/Y4NN777/226os/blob/main';

/**
 * Footer links (prototype `footer.*`) that point at a document that exists in the public
 * repository. The prototype's code of conduct and privacy links have no document yet: left out.
 */
export const projectLinks: readonly { label: MessageKey; href: string }[] = [
  { label: 'shell.footer.licence', href: `${repository}/LICENSE` },
  { label: 'shell.footer.contribute', href: `${repository}/CONTRIBUTING.md` },
];
