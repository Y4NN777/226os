import { AppShell } from '@astryxdesign/core/AppShell';
import { Badge } from '@astryxdesign/core/Badge';
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
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { MobileNav } from '@astryxdesign/core/MobileNav';
import {
  SideNav,
  SideNavItem,
  SideNavRenderContext,
  SideNavSection,
} from '@astryxdesign/core/SideNav';
import { TopNav, TopNavHeading } from '@astryxdesign/core/TopNav';
import { Outlet, useRouterState } from '@tanstack/react-router';
import { useI18n } from '../i18n/I18nProvider';
import { locales, type Locale } from '../i18n/messages';
import { navSections } from './navigation';
import { useSession, useSignOut } from './session';

const endonyms: Record<Locale, string> = { fr: 'Français', en: 'English' };

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

function AccountArea({ onSignOut }: { onSignOut: () => void }) {
  const { t } = useI18n();
  const session = useSession();
  if (session.isPending) return null;
  const user = session.data?.authenticated ? session.data.user : null;
  if (user) {
    return (
      <DropdownMenu button={{ label: user.display_name, variant: 'ghost' }} alignment="end">
        <DropdownMenuItem label={t('account.profile')} href={`/profils/${user.username}`} />
        <DropdownMenuItem label={t('account.write')} href="/ecrire" />
        <DropdownMenuItem label={t('account.settings')} href="/parametres" />
        <DropdownMenuDivider />
        <LanguageGroup />
        <DropdownMenuDivider />
        <DropdownMenuItem label={t('account.signOut')} onClick={onSignOut} />
      </DropdownMenu>
    );
  }
  return (
    <HStack gap={2} vAlign="center">
      <Button label={t('account.signIn')} href="/connexion" variant="primary" />
      <DropdownMenu
        alignment="end"
        hasChevron={false}
        renderTrigger={(props) => (
          <IconButton
            {...props}
            variant="ghost"
            label={t('account.more')}
            icon={<Icon icon="moreHorizontal" />}
          />
        )}
      >
        <DropdownMenuItem label={t('account.signUp')} href="/inscription" />
        <DropdownMenuItem label={t('account.settings')} href="/parametres" />
        <DropdownMenuDivider />
        <LanguageGroup />
      </DropdownMenu>
    </HStack>
  );
}

function PrimaryNav({ isDrawer = false }: { isDrawer?: boolean }) {
  const { t } = useI18n();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const sections = (
    <>
      <SideNavSection title={t('nav.home')} isHeaderHidden>
        <SideNavItem label={t('nav.home')} href="/" isSelected={pathname === '/'} />
      </SideNavSection>
      {navSections.map((section) => (
        <SideNavSection key={section.title} title={t(section.title)}>
          {section.items.map((item) =>
            item.to ? (
              <SideNavItem
                key={item.label}
                label={t(item.label)}
                href={item.to}
                isSelected={pathname === item.to || pathname.startsWith(`${item.to}/`)}
              />
            ) : (
              <SideNavItem
                key={item.label}
                label={t(item.label)}
                isDisabled
                endContent={<Badge label={t('nav.soon')} />}
              />
            ),
          )}
        </SideNavSection>
      ))}
    </>
  );
  if (isDrawer) {
    return <SideNavRenderContext value="drawer-content">{sections}</SideNavRenderContext>;
  }
  return <SideNav aria-label={t('nav.label')}>{sections}</SideNav>;
}

/** App chrome for every page except the standalone account pages. */
export function Shell() {
  const { t } = useI18n();
  const signOut = useSignOut();
  return (
    <AppShell
      height="auto"
      contentPadding={6}
      banner={
        signOut.isError ? (
          <Banner status="error" title={t('account.signOutFailed')} container="section" />
        ) : undefined
      }
      topNav={
        <TopNav
          label={t('nav.top')}
          heading={<TopNavHeading heading={t('app.name')} headingHref="/" />}
          endContent={
            <HStack gap={2} vAlign="center">
              <IconButton
                href="/recherche"
                variant="ghost"
                label={t('nav.search')}
                icon={<Icon icon="search" />}
              />
              <AccountArea
                onSignOut={() => {
                  signOut.mutate();
                }}
              />
            </HStack>
          }
        />
      }
      sideNav={<PrimaryNav />}
      mobileNav={{
        // The same links in a drawer that is a modal dialog for assistive technology as well.
        content: (
          <MobileNav
            aria-modal="true"
            header={<TopNavHeading heading={t('app.name')} headingHref="/" />}
          >
            <PrimaryNav isDrawer />
          </MobileNav>
        ),
      }}
    >
      <Outlet />
    </AppShell>
  );
}
