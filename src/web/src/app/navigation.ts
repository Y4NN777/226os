import type { MessageKey } from '../i18n/messages';

export interface NavItem {
  label: MessageKey;
  /** Route path when live; absent while the area is not built ("Bientôt", never a dead link). */
  to?: string;
}

export interface NavSection {
  title: MessageKey;
  items: NavItem[];
}

/**
 * Pillar emphasis (O7): Apprendre and Connect come first; Solve, Discover and Build stay
 * full pillars in the same navigation, on desktop and in the mobile drawer.
 */
export const navSections: NavSection[] = [
  { title: 'nav.section.grow', items: [{ label: 'nav.learn' }, { label: 'nav.connect' }] },
  {
    title: 'nav.section.pillars',
    items: [
      { label: 'nav.solve', to: '/solve' },
      { label: 'nav.discover', to: '/discover' },
      { label: 'nav.build', to: '/build' },
    ],
  },
];
