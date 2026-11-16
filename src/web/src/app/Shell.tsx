import { AppShell, useAppShellMobile } from '@astryxdesign/core/AppShell';
import { Avatar } from '@astryxdesign/core/Avatar';
import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import {
  DropdownMenu,
  DropdownMenuDivider,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
} from '@astryxdesign/core/DropdownMenu';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Link } from '@astryxdesign/core/Link';
import { MobileNav } from '@astryxdesign/core/MobileNav';
import {
  SideNav,
  SideNavCollapseButton,
  SideNavItem,
  SideNavRenderContext,
  SideNavSection,
} from '@astryxdesign/core/SideNav';
import { TextInput } from '@astryxdesign/core/TextInput';
import { TopNav } from '@astryxdesign/core/TopNav';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from 'react';
import { Brand } from '../components/Brand';
import { RouterAnchor } from '../components/RouterAnchor';
import {
  EditIcon,
  GlobeIcon,
  LogInIcon,
  LogOutIcon,
  MenuIcon,
  PanelLeftIcon,
  MoonIcon,
  PlusIcon,
  SearchIcon,
  SettingsIcon,
  SunIcon,
  UserIcon,
  UserPlusIcon,
  XIcon,
} from '../components/icons';
import { LandingLayout } from '../features/landing/LandingChrome';
import { useI18n } from '../i18n/I18nProvider';
import { locales, type Locale } from '../i18n/messages';
import { useResolvedTheme, useThemePreference, type ThemeMode } from '../theme/ThemeProvider';
import {
  bottomNavItems,
  homeItem,
  isCurrent,
  navSections,
  projectLinks,
  railQuery,
  type NavItem,
  type NavSection,
} from './navigation';
import { readStorage, writeStorage } from './storage';
import { useSession, useSignOut } from './session';

const endonyms: Record<Locale, string> = { fr: 'Français', en: 'English' };
const themeModes: readonly ThemeMode[] = ['light', 'dark', 'system'];

/** Site-wide search: submitting goes to the search page with the query (`/recherche?q=`). */
function HeaderSearch({
  onSubmit,
  onEscape,
}: { onSubmit?: () => void; onEscape?: () => void } = {}) {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [value, setValue] = useState('');
  return (
    <form
      role="search"
      aria-label={t('shell.search')}
      className="shell-search"
      onSubmit={(event) => {
        event.preventDefault();
        const q = value.trim();
        onSubmit?.();
        void navigate({ to: '/recherche', search: q ? { q } : {} });
      }}
    >
      <TextInput
        type="search"
        label={t('shell.search')}
        isLabelHidden
        placeholder={t('shell.search')}
        startIcon={SearchIcon}
        value={value}
        onChange={setValue}
        htmlName="q"
        enterKeyHint="search"
        autoComplete="off"
        width="100%"
        onKeyDown={(event) => {
          if (event.key === 'Escape' && onEscape) {
            event.preventDefault();
            onEscape();
          }
        }}
      />
    </form>
  );
}

/** The full search field stays in the desktop top bar. */
function DesktopSearch() {
  const { isMobile } = useAppShellMobile();
  return isMobile ? null : <HeaderSearch />;
}

/**
 * On a phone, one icon opens a search overlay in the single-row header. Escape closes it and
 * returns focus to the icon, matching the landing header's compact rhythm without hiding search.
 */
function MobileHeaderSearch() {
  const { t } = useI18n();
  const { isMobile } = useAppShellMobile();
  const [isOpen, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const panel = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isOpen) {
      panel.current?.querySelector<HTMLInputElement>('input')?.focus();
    }
  }, [isOpen]);

  const close = () => {
    setOpen(false);
    requestAnimationFrame(() => {
      trigger.current?.focus();
    });
  };

  if (!isMobile) return null;
  return (
    <span className="shell-mobile-search">
      <IconButton
        ref={trigger}
        variant="ghost"
        label={t('shell.search')}
        icon={<Icon icon={SearchIcon} size="md" />}
        aria-hidden={isOpen || undefined}
        aria-expanded={isOpen}
        aria-controls="shell-mobile-search-panel"
        tabIndex={isOpen ? -1 : undefined}
        onClick={() => {
          setOpen(true);
        }}
      />
      {isOpen ? (
        <div ref={panel} id="shell-mobile-search-panel" className="shell-mobile-search-panel">
          <HeaderSearch
            onSubmit={() => {
              setOpen(false);
            }}
            onEscape={close}
          />
          <IconButton
            variant="ghost"
            label={t('shell.search.close')}
            icon={<Icon icon={XIcon} size="md" />}
            aria-expanded="true"
            aria-controls="shell-mobile-search-panel"
            onClick={close}
          />
        </div>
      ) : null}
    </span>
  );
}

function ThemeToggle() {
  const { t } = useI18n();
  const { setMode } = useThemePreference();
  const isDark = useResolvedTheme() === 'dark';
  return (
    <IconButton
      variant="ghost"
      className="shell-theme-toggle"
      label={t('shell.darkTheme')}
      aria-pressed={isDark}
      icon={<Icon icon={isDark ? MoonIcon : SunIcon} size="md" />}
      onClick={() => {
        setMode(isDark ? 'light' : 'dark');
      }}
    />
  );
}

/** The language choice: radios, so the current one is exposed as checked, not only drawn. */
function LanguageGroup() {
  const { t, locale, setLocale } = useI18n();
  return (
    <DropdownMenuGroup title={t('account.language')}>
      <DropdownMenuRadioGroup
        label={t('account.language')}
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
  );
}

function ThemeGroup() {
  const { t } = useI18n();
  const { mode, setMode } = useThemePreference();
  return (
    <DropdownMenuGroup title={t('account.theme')}>
      <DropdownMenuRadioGroup
        label={t('account.theme')}
        value={mode}
        onChange={(value) => {
          const next = themeModes.find((option) => option === value);
          if (next) setMode(next);
        }}
      >
        {themeModes.map((option) => (
          <DropdownMenuRadioItem
            key={option}
            value={option}
            label={t(`settings.theme.${option}`)}
          />
        ))}
      </DropdownMenuRadioGroup>
    </DropdownMenuGroup>
  );
}

function PreferencesMenu() {
  const { t } = useI18n();
  return (
    <DropdownMenu
      alignment="end"
      hasChevron={false}
      renderTrigger={(props) => (
        <IconButton
          {...props}
          variant="ghost"
          label={t('account.prefs')}
          icon={<Icon icon={GlobeIcon} size="md" />}
        />
      )}
    >
      <LanguageGroup />
      <DropdownMenuDivider />
      <ThemeGroup />
      <DropdownMenuDivider />
      <DropdownMenuItem label={t('account.settings')} href="/parametres" />
    </DropdownMenu>
  );
}

function AccountArea({ onSignOut }: { onSignOut: () => void }) {
  const { t } = useI18n();
  const { isMobile } = useAppShellMobile();
  const session = useSession();
  if (session.isPending) return null;
  const user = session.data?.authenticated ? session.data.user : null;
  if (user) {
    return (
      <DropdownMenu
        alignment="end"
        hasChevron={false}
        renderTrigger={(props) => (
          <IconButton
            {...props}
            variant="ghost"
            label={user.display_name}
            icon={<Avatar name={user.display_name} size="sm" tooltip={false} />}
          />
        )}
      >
        <DropdownMenuItem label={t('account.profile')} href={`/profils/${user.username}`} />
        <DropdownMenuItem label={t('account.write')} href="/ecrire" />
        <DropdownMenuItem label={t('account.settings')} href="/parametres" />
        <DropdownMenuDivider />
        <LanguageGroup />
        <DropdownMenuDivider />
        <ThemeGroup />
        <DropdownMenuDivider />
        <DropdownMenuItem label={t('account.signOut')} onClick={onSignOut} />
      </DropdownMenu>
    );
  }
  if (isMobile) {
    return (
      <Button
        className="shell-sign-up"
        label={t('account.signUp')}
        href="/inscription"
        variant="primary"
      />
    );
  }
  return (
    <>
      <PreferencesMenu />
      <span className="shell-auth">
        <Button label={t('account.signIn')} href="/connexion" variant="ghost" />
        <Button label={t('account.signUp')} href="/inscription" variant="primary" />
      </span>
    </>
  );
}

/**
 * « Poser une question » for members (prototype `.header-create`): the solve page opens its ask
 * view on `#ask`. Below 1024px only the plus icon shows; the label stays the accessible name.
 */
function AskButton() {
  const { t } = useI18n();
  const session = useSession();
  if (!session.data?.authenticated) return null;
  return (
    <Button
      label={t('shell.ask')}
      href="/solve#ask"
      variant="primary"
      className="shell-ask"
      icon={<Icon icon={PlusIcon} size="md" />}
    >
      <span className="shell-ask-label">{t('shell.ask')}</span>
    </Button>
  );
}

function NavEntry({ item, pathname }: { item: NavItem; pathname: string }) {
  const { t } = useI18n();
  if (!item.to) {
    // Not built yet: named, marked « Bientôt », never a dead link.
    return (
      <SideNavItem
        label={t(item.label)}
        icon={item.icon}
        isDisabled
        endContent={<span className="nav-soon">{t('nav.soon')}</span>}
      />
    );
  }
  return (
    <SideNavItem
      label={t(item.label)}
      icon={item.icon}
      href={item.to}
      isSelected={isCurrent(pathname, item.to)}
    />
  );
}

/** The grouped navigation items; the sidebar and the phone drawer both render it. */
export function NavSections({
  sections,
  pathname,
}: {
  sections: readonly NavSection[];
  pathname: string;
}) {
  const { t } = useI18n();
  return (
    <>
      {sections.map((section) => (
        <SideNavSection key={section.title} title={t(section.title)}>
          {section.items.map((item) => (
            <NavEntry key={item.label} item={item} pathname={pathname} />
          ))}
        </SideNavSection>
      ))}
    </>
  );
}

function subscribeRail(notify: () => void) {
  const query = window.matchMedia(railQuery);
  query.addEventListener('change', notify);
  return () => {
    query.removeEventListener('change', notify);
  };
}

/** Whether the viewport is in the tablet range, where the navigation is an icon rail. */
function useIsRail() {
  return useSyncExternalStore(
    subscribeRail,
    () => window.matchMedia(railQuery).matches,
    () => false,
  );
}

/** Tablet rail (prototype `.nav-item` from 768px): the areas as icons over a short label. */
function RailNav({ pathname }: { pathname: string }) {
  const { t } = useI18n();
  return (
    <nav aria-label={t('nav.label')} className="shell-rail">
      {[homeItem, ...navSections.flatMap((section) => section.items)].map((item) => {
        const Glyph = item.icon;
        const label = <span>{t(item.label)}</span>;
        if (!item.to) {
          return (
            <span key={item.label} className="shell-rail-item" aria-disabled="true">
              <Glyph />
              {label}
            </span>
          );
        }
        return (
          <RouterAnchor
            key={item.label}
            href={item.to}
            className="shell-rail-item"
            aria-current={isCurrent(pathname, item.to) ? 'page' : undefined}
          >
            <Glyph />
            {label}
          </RouterAnchor>
        );
      })}
    </nav>
  );
}

const sideNavStorageKey = '226os.sideNavCollapsed';

/**
 * Whether the desktop sidebar is collapsed to its icons, kept in this browser. Astryx 0.7.14 has
 * no `autoSaveId` on `collapsible` (only on `resizable`, which adds a resize handle we do not
 * want), so the state is controlled here.
 */
function useSideNavCollapsed() {
  const [isCollapsed, setCollapsed] = useState(() => readStorage(sideNavStorageKey) === 'true');
  const change = useCallback((next: boolean) => {
    setCollapsed(next);
    writeStorage(sideNavStorageKey, String(next));
  }, []);
  return [isCollapsed, change] as const;
}

/** The public landing (`/` for a visitor) has its own header and footer, and no app navigation. */
function useIsLanding() {
  const session = useSession();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  if (pathname !== '/') return false;
  if (session.isPending) return null;
  return !session.data?.authenticated;
}

function PrimaryNav({
  isDrawer = false,
  isCollapsed = false,
  onCollapsedChange = () => undefined,
}: {
  isDrawer?: boolean;
  isCollapsed?: boolean;
  onCollapsedChange?: (next: boolean) => void;
}) {
  const { t } = useI18n();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const isRail = useIsRail();
  if (isRail && !isDrawer) return <RailNav pathname={pathname} />;
  const sections = (
    <>
      <SideNavSection title={t(homeItem.label)} isHeaderHidden>
        <NavEntry item={homeItem} pathname={pathname} />
      </SideNavSection>
      <NavSections sections={navSections} pathname={pathname} />
    </>
  );
  if (isDrawer) {
    return <SideNavRenderContext value="drawer-content">{sections}</SideNavRenderContext>;
  }
  return (
    <SideNav
      aria-label={t('nav.label')}
      className={isCollapsed ? 'shell-sidenav shell-sidenav-collapsed' : 'shell-sidenav'}
      collapsible={{ isCollapsed, onCollapsedChange, hasButton: false }}
      footerIcons={
        <SideNavCollapseButton
          label={t(isCollapsed ? 'shell.sideNavExpand' : 'shell.sideNavCollapse')}
          aria-expanded={!isCollapsed}
        >
          <Icon icon={PanelLeftIcon} size="md" />
        </SideNavCollapseButton>
      }
    >
      {sections}
    </SideNav>
  );
}

/** The account entries at the end of the phone drawer: the top bar has no room for them. */
function DrawerAccount({ onSignOut }: { onSignOut: () => void }) {
  const { t } = useI18n();
  const session = useSession();
  if (session.isPending) return null;
  const user = session.data?.authenticated ? session.data.user : null;
  return (
    <>
      <SideNavSection title={t('shell.account')} isHeaderHidden>
        {user ? (
          <>
            <SideNavItem
              label={t('account.profile')}
              icon={UserIcon}
              href={`/profils/${user.username}`}
            />
            <SideNavItem label={t('account.write')} icon={EditIcon} href="/ecrire" />
            <SideNavItem label={t('account.settings')} icon={SettingsIcon} href="/parametres" />
            <SideNavItem label={t('account.signOut')} icon={LogOutIcon} onClick={onSignOut} />
          </>
        ) : (
          <>
            <SideNavItem label={t('account.signIn')} icon={LogInIcon} href="/connexion" />
            <SideNavItem label={t('account.signUp')} icon={UserPlusIcon} href="/inscription" />
            <SideNavItem label={t('account.settings')} icon={SettingsIcon} href="/parametres" />
          </>
        )}
      </SideNavSection>
      {!user ? (
        <div className="shell-drawer-preferences">
          <PreferencesMenu />
        </div>
      ) : null}
    </>
  );
}

/**
 * Phone bottom bar (prototype `.bottom-nav`): Accueil, Apprendre, Résoudre, Découvrir, then
 * « Plus », which opens the drawer with the rest. Only below the shell's phone breakpoint.
 */
function BottomNav() {
  const { t } = useI18n();
  const { isMobile, isMobileNavOpen, mobileNavId, toggleMobileNav } = useAppShellMobile();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  if (!isMobile) return null;
  return (
    <nav aria-label={t('shell.bottomNav')} className="shell-bottom-nav">
      {bottomNavItems.map((item) => {
        const Glyph = item.icon;
        return (
          <RouterAnchor
            key={item.to}
            href={item.to}
            className="shell-bottom-item"
            aria-current={isCurrent(pathname, item.to) ? 'page' : undefined}
          >
            <Glyph />
            <span>{t(item.label)}</span>
          </RouterAnchor>
        );
      })}
      <button
        type="button"
        className="shell-bottom-item"
        aria-expanded={isMobileNavOpen}
        aria-controls={mobileNavId}
        onClick={toggleMobileNav}
      >
        <MenuIcon />
        <span>{t('shell.more')}</span>
      </button>
    </nav>
  );
}

/** « 226os » and links to the project documents that exist (prototype `.app-footer`). */
function SiteFooter({ isCollapsed }: { isCollapsed: boolean }) {
  const { t } = useI18n();
  const className = ['shell-footer', isCollapsed && 'shell-footer-collapsed']
    .filter(Boolean)
    .join(' ');
  return (
    <footer className={className}>
      <span className="shell-footer-name">{t('app.name')}</span>
      <ul className="shell-footer-links" aria-label={t('shell.footer')}>
        {projectLinks.map((link) => (
          <li key={link.href}>
            <Link href={link.href} isExternalLink newTabLabel={t('shell.newTab')}>
              {t(link.label)}
            </Link>
          </li>
        ))}
      </ul>
    </footer>
  );
}

/** App chrome for every page except the standalone account pages (prototype shell). */
export function Shell() {
  const { t } = useI18n();
  const signOut = useSignOut();
  const onSignOut = () => {
    signOut.mutate();
  };
  const session = useSession();
  const isLanding = useIsLanding();
  const [isCollapsed, setCollapsed] = useSideNavCollapsed();
  // The identity of `/` is intentionally undecided until the session is known. Rendering only
  // the route's loading state inside this neutral frame prevents either chrome from flashing.
  if (isLanding === null) {
    return (
      <main className="shell-neutral-frame">
        <Outlet />
      </main>
    );
  }
  // The public landing has its own header and footer, not the app's.
  if (isLanding) {
    return (
      <LandingLayout ready={!session.isPending}>
        <Outlet />
      </LandingLayout>
    );
  }
  return (
    <>
      <AppShell
        height="auto"
        variant="wash"
        contentPadding={0}
        className="shell"
        banner={
          signOut.isError ? (
            <Banner status="error" title={t('account.signOutFailed')} container="section" />
          ) : undefined
        }
        topNav={
          <TopNav
            label={t('nav.top')}
            className="shell-topnav"
            heading={<Brand />}
            centerContent={<DesktopSearch />}
            endContent={
              <span className="shell-actions">
                <MobileHeaderSearch />
                <ThemeToggle />
                <AccountArea onSignOut={onSignOut} />
                <AskButton />
              </span>
            }
          />
        }
        sideNav={<PrimaryNav isCollapsed={isCollapsed} onCollapsedChange={setCollapsed} />}
        mobileNav={{
          // No top-bar hamburger on a phone: « Plus » in the bottom bar opens the drawer.
          hasToggle: false,
          // The same links in a drawer that is a modal dialog for assistive technology as well,
          // sliding in from the left as in the prototype.
          content: (
            <MobileNav aria-modal="true" side="start" header={<Brand />}>
              <PrimaryNav isDrawer />
              <SideNavRenderContext value="drawer-content">
                <DrawerAccount onSignOut={onSignOut} />
              </SideNavRenderContext>
            </MobileNav>
          ),
        }}
      >
        <div className="shell-page">
          <Outlet />
        </div>
        <BottomNav />
      </AppShell>
      <SiteFooter isCollapsed={isCollapsed} />
    </>
  );
}
