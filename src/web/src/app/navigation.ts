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

export const homeItem: NavItem = { label: 'nav.home', icon: HomeIcon, to: '/' };

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
