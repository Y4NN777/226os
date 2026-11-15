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
import { MobileNav } from '@astryxdesign/core/MobileNav';
import {
  SideNav,
  SideNavItem,
  SideNavRenderContext,
  SideNavSection,
} from '@astryxdesign/core/SideNav';
import { TextInput } from '@astryxdesign/core/TextInput';
import { TopNav } from '@astryxdesign/core/TopNav';
import { Link, Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useState } from 'react';
import logoDark from '../assets/brand/226os-logo-white.png';
import logoLight from '../assets/brand/226os-logo.png';
import { GlobeIcon, MoonIcon, SearchIcon, SunIcon } from '../components/icons';
import { useI18n } from '../i18n/I18nProvider';
import { locales, type Locale } from '../i18n/messages';
import { useResolvedTheme, useThemePreference, type ThemeMode } from '../theme/ThemeProvider';
import { homeItem, navSections, type NavItem, type NavSection } from './navigation';
import { useSession, useSignOut } from './session';

const endonyms: Record<Locale, string> = { fr: 'Français', en: 'English' };
const themeModes: readonly ThemeMode[] = ['light', 'dark', 'system'];

/** The 226os logo, linking home; the white variant on the dark theme. */
function Brand() {
  const theme = useResolvedTheme();
  return (
    <Link to="/" className="shell-brand">
      <img
        src={theme === 'dark' ? logoDark : logoLight}
        alt="226os"
        width={92}
        height={32}
        className="shell-brand-logo"
      />
    </Link>
  );
}

/** Site-wide search: submitting goes to the search page with the query (`/recherche?q=`). */
function HeaderSearch() {
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
      />
    </form>
  );
}

/** On a phone the top bar has no room for the field: it opens the page, above the content. */
function MobileSearch() {
  const { isMobile } = useAppShellMobile();
  return isMobile ? <HeaderSearch /> : null;
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
      icon={<Icon icon={isDark ? MoonIcon : SunIcon} />}
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

function AccountArea({ onSignOut }: { onSignOut: () => void }) {
  const { t } = useI18n();
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
            label={t('account.menu', { name: user.display_name })}
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
  return (
    <>
      <DropdownMenu
        alignment="end"
        hasChevron={false}
        renderTrigger={(props) => (
          <IconButton
            {...props}
            variant="ghost"
            label={t('account.prefs')}
            icon={<Icon icon={GlobeIcon} />}
          />
        )}
      >
        <LanguageGroup />
        <DropdownMenuDivider />
        <ThemeGroup />
      </DropdownMenu>
      <span className="shell-auth">
        <Button label={t('account.signIn')} href="/connexion" variant="ghost" />
        <Button label={t('account.signUp')} href="/inscription" variant="primary" />
      </span>
    </>
  );
}

function isCurrent(pathname: string, to: string) {
  return to === '/' ? pathname === '/' : pathname === to || pathname.startsWith(`${to}/`);
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

function PrimaryNav({ isDrawer = false }: { isDrawer?: boolean }) {
  const { t } = useI18n();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
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
    <SideNav aria-label={t('nav.label')} className="shell-sidenav">
      {sections}
    </SideNav>
  );
}

/** App chrome for every page except the standalone account pages (prototype shell). */
export function Shell() {
  const { t } = useI18n();
  const signOut = useSignOut();
  return (
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
          centerContent={<HeaderSearch />}
          endContent={
            <span className="shell-actions">
              <ThemeToggle />
              <AccountArea
                onSignOut={() => {
                  signOut.mutate();
                }}
              />
            </span>
          }
        />
      }
      sideNav={<PrimaryNav />}
      mobileNav={{
        // The same links in a drawer that is a modal dialog for assistive technology as well.
        content: (
          <MobileNav aria-modal="true" header={<Brand />}>
            <PrimaryNav isDrawer />
          </MobileNav>
        ),
      }}
    >
      <div className="shell-page">
        <MobileSearch />
        <Outlet />
      </div>
    </AppShell>
  );
}
