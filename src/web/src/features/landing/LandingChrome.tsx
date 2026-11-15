import {
  DropdownMenu,
  DropdownMenuGroup,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
} from '@astryxdesign/core/DropdownMenu';
import { useEffect, useId, useState, type MouseEvent, type ReactNode } from 'react';
import logoWhite from '../../assets/brand/226os-logo-white.png';
import { GlobeIcon, MenuIcon, MoonIcon, SunIcon, XIcon } from '../../components/icons';
import { RouterAnchor } from '../../components/RouterAnchor';
import { useI18n } from '../../i18n/I18nProvider';
import { locales, type Locale, type MessageKey } from '../../i18n/messages';
import { useResolvedTheme, useThemePreference } from '../../theme/ThemeProvider';
import { useLandingActivity } from './useLandingData';
import './landing.css';

const endonyms: Record<Locale, string> = { fr: 'Français', en: 'English' };
const REPOSITORY = 'https://github.com/Y4NN777/226os';

interface Section {
  id: string;
  label: MessageKey;
}
const sections: readonly Section[] = [
  { id: 'avantages', label: 'landing.nav.why' },
  { id: 'espaces', label: 'landing.nav.spaces' },
  { id: 'activite', label: 'landing.nav.activity' },
  { id: 'rejoindre', label: 'landing.nav.join' },
];

/** The anchors of the sections that are on the page: the activity one only when it has data. */
function useSections(enabled: boolean) {
  const { hasActivity } = useLandingActivity(enabled);
  return sections.filter((section) => section.id !== 'activite' || hasActivity);
}

function prefersReducedMotion() {
  return typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/** Smooth scroll to a section of the page; the link still works as a plain `#anchor` without JS. */
function goTo(event: MouseEvent<HTMLAnchorElement>, id: string, after?: () => void) {
  const target = document.getElementById(id);
  if (!target) return;
  event.preventDefault();
  // jsdom has no scrollIntoView, real browsers always do.
  // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
  target.scrollIntoView?.({ behavior: prefersReducedMotion() ? 'auto' : 'smooth', block: 'start' });
  after?.();
}

function useScrolled() {
  const [scrolled, setScrolled] = useState(false);
  useEffect(() => {
    const update = () => {
      setScrolled(window.scrollY > 8);
    };
    update();
    window.addEventListener('scroll', update, { passive: true });
    return () => {
      window.removeEventListener('scroll', update);
    };
  }, []);
  return scrolled;
}

/** The id of the section being read: the one crossing the middle band of the viewport. */
function useActiveSection(ids: readonly string[]) {
  const [active, setActive] = useState<string | null>(null);
  const joined = ids.join(' ');
  useEffect(() => {
    if (typeof IntersectionObserver === 'undefined') return;
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) setActive(entry.target.id);
        }
      },
      { rootMargin: '-45% 0px -50% 0px', threshold: 0 },
    );
    for (const id of joined.split(' ')) {
      const node = document.getElementById(id);
      if (node) observer.observe(node);
    }
    return () => {
      observer.disconnect();
    };
  }, [joined]);
  return active;
}

function ThemeButton() {
  const { t } = useI18n();
  const { setMode } = useThemePreference();
  const isDark = useResolvedTheme() === 'dark';
  const Glyph = isDark ? SunIcon : MoonIcon;
  return (
    <button
      type="button"
      className="lh-icon"
      aria-label={t(isDark ? 'landing.nav.themeLight' : 'landing.nav.themeDark')}
      onClick={() => {
        setMode(isDark ? 'light' : 'dark');
      }}
    >
      <Glyph aria-hidden="true" />
    </button>
  );
}

/** The language: a globe icon button (named for assistive technology) opening Français / English. */
function LanguageMenu() {
  const { t, locale, setLocale } = useI18n();
  return (
    <DropdownMenu
      alignment="end"
      hasChevron={false}
      renderTrigger={(props) => (
        <button {...props} type="button" className="lh-icon" aria-label={t('landing.nav.language')}>
          <GlobeIcon aria-hidden="true" />
        </button>
      )}
    >
      <DropdownMenuGroup title={t('landing.footer.language')}>
        <DropdownMenuRadioGroup
          label={t('landing.footer.language')}
          value={locale}
          onChange={(value) => {
            const next = locales.find((option) => option === value);
            if (next) setLocale(next);
          }}
        >
          {locales.map((option) => (
            <DropdownMenuRadioItem key={option} value={option} label={endonyms[option]} />
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuGroup>
    </DropdownMenu>
  );
}

/** « Français | English » as two toggle buttons, for the menu panel and the footer. */
function LanguageSwitch() {
  const { t, locale, setLocale } = useI18n();
  return (
    <div role="group" aria-label={t('landing.footer.language')} className="lh-switch">
      {locales.map((option) => (
        <button
          key={option}
          type="button"
          lang={option}
          aria-pressed={locale === option}
          onClick={() => {
            setLocale(option);
          }}
        >
          {endonyms[option]}
        </button>
      ))}
    </div>
  );
}

/** Landing-only header: sticky, glass once the page scrolls, a compact menu on small screens. */
export function LandingHeader({ ready }: { ready: boolean }) {
  const { t } = useI18n();
  const items = useSections(ready);
  const active = useActiveSection(items.map((item) => item.id));
  const scrolled = useScrolled();
  const [open, setOpen] = useState(false);
  const panelId = useId();
  const { setMode } = useThemePreference();
  const isDark = useResolvedTheme() === 'dark';

  // Escape closes the menu and gives the focus back to its button.
  useEffect(() => {
    if (!open) return;
    const onKeyDown = (event: globalThis.KeyboardEvent) => {
      if (event.key !== 'Escape') return;
      setOpen(false);
      document.querySelector<HTMLElement>('.lh-menu')?.focus();
    };
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [open]);

  return (
    <header className="lh lp-dark" data-scrolled={String(scrolled)}>
      <div className="lh-bar">
        <a
          className="lh-logo"
          href="#accueil"
          aria-label={t('landing.nav.home')}
          onClick={(event) => {
            goTo(event, 'accueil');
          }}
        >
          <img src={logoWhite} alt="" width={92} height={32} />
        </a>
        <nav className="lh-nav" aria-label={t('landing.nav.label')}>
          <ul>
            {items.map((item) => (
              <li key={item.id}>
                <a
                  href={`#${item.id}`}
                  aria-current={active === item.id ? 'location' : undefined}
                  onClick={(event) => {
                    goTo(event, item.id);
                  }}
                >
                  {t(item.label)}
                </a>
              </li>
            ))}
          </ul>
        </nav>
        <div className="lh-actions">
          <span className="lh-quiet">
            <LanguageMenu />
            <ThemeButton />
            <RouterAnchor href="/connexion" className="lh-signin">
              {t('landing.nav.signin')}
            </RouterAnchor>
          </span>
          <RouterAnchor href="/inscription" className="lp-btn lh-signup">
            {t('landing.nav.signup')}
          </RouterAnchor>
          <button
            type="button"
            className="lh-icon lh-menu"
            aria-expanded={open}
            aria-controls={panelId}
            aria-label={t(open ? 'landing.nav.menuClose' : 'landing.nav.menu')}
            onClick={() => {
              setOpen(!open);
            }}
          >
            {open ? <XIcon aria-hidden="true" /> : <MenuIcon aria-hidden="true" />}
          </button>
        </div>
      </div>
      <div id={panelId} className="lh-panel" hidden={!open}>
        <ul>
          {items.map((item) => (
            <li key={item.id}>
              <a
                href={`#${item.id}`}
                onClick={(event) => {
                  goTo(event, item.id, () => {
                    setOpen(false);
                  });
                }}
              >
                {t(item.label)}
              </a>
            </li>
          ))}
        </ul>
        <div className="lh-panel-tools">
          <LanguageSwitch />
          <button
            type="button"
            className="lp-btn lp-btn-quiet"
            onClick={() => {
              setMode(isDark ? 'light' : 'dark');
            }}
          >
            {isDark ? <SunIcon aria-hidden="true" /> : <MoonIcon aria-hidden="true" />}
            {t(isDark ? 'landing.nav.themeLight' : 'landing.nav.themeDark')}
          </button>
          <RouterAnchor href="/connexion" className="lp-btn lp-btn-quiet">
            {t('landing.nav.signin')}
          </RouterAnchor>
        </div>
      </div>
    </header>
  );
}

const footerSpaces: readonly { id: string; label: MessageKey }[] = [
  { id: 'learn', label: 'landing.pillar.learn.title' },
  { id: 'solve', label: 'landing.pillar.solve.title' },
  { id: 'discover', label: 'landing.pillar.discover.title' },
  { id: 'build', label: 'landing.pillar.build.title' },
  { id: 'connect', label: 'landing.pillar.connect.title' },
];

function External({ href, children }: { href: string; children: ReactNode }) {
  const { t } = useI18n();
  return (
    <a href={href} target="_blank" rel="noopener noreferrer">
      {children}
      <span className="lp-sr"> {t('shell.newTab')}</span>
    </a>
  );
}

/** Landing-only footer: what the project is and where to find it. Nothing private. */
export function LandingFooter() {
  const { t } = useI18n();
  const spacesTitle = useId();
  const projectTitle = useId();
  return (
    <footer className="lf lp-dark" aria-label={t('landing.footer.label')}>
      <div className="lp-wrap lf-grid">
        <div className="lf-brand">
          <img src={logoWhite} alt="226os" width={92} height={32} />
          <p>{t('landing.footer.about')}</p>
        </div>
        <nav aria-labelledby={spacesTitle}>
          <p id={spacesTitle} className="lf-title">
            {t('landing.footer.spaces')}
          </p>
          <ul>
            {footerSpaces.map((space) => (
              <li key={space.id}>
                <a
                  href="#espaces"
                  onClick={(event) => {
                    goTo(event, 'espaces');
                  }}
                >
                  {t(space.label)}
                </a>
              </li>
            ))}
          </ul>
        </nav>
        <nav aria-labelledby={projectTitle}>
          <p id={projectTitle} className="lf-title">
            {t('landing.footer.project')}
          </p>
          <ul>
            <li>
              <External href={REPOSITORY}>{t('landing.footer.source')}</External>
            </li>
            <li>
              <External href={`${REPOSITORY}/blob/main/CONTRIBUTING.md`}>
                {t('landing.footer.contribute')}
              </External>
            </li>
            <li>
              <External href={`${REPOSITORY}/blob/main/LICENSE`}>
                {t('landing.footer.licence')}
              </External>
            </li>
          </ul>
        </nav>
      </div>
      <div className="lp-wrap lf-bottom">
        <p>{t('landing.footer.copyright')}</p>
        <LanguageSwitch />
      </div>
    </footer>
  );
}

/** The page frame of `/` for a visitor: its own header and footer around the content. */
export function LandingLayout({ ready, children }: { ready: boolean; children: ReactNode }) {
  // No visible scrollbar on the landing; scrolling itself is untouched. Removed when leaving.
  useEffect(() => {
    document.documentElement.classList.add('lp-noscroll');
    return () => {
      document.documentElement.classList.remove('lp-noscroll');
    };
  }, []);
  return (
    <div className="lp">
      <LandingHeader ready={ready} />
      <main id="contenu">{children}</main>
      <LandingFooter />
    </div>
  );
}
