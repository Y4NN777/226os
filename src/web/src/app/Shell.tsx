import { AppShell } from '@astryxdesign/core/AppShell';
import { Badge } from '@astryxdesign/core/Badge';
import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { DropdownMenu, type DropdownMenuOption } from '@astryxdesign/core/DropdownMenu';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { SideNav, SideNavItem, SideNavSection } from '@astryxdesign/core/SideNav';
import { TopNav, TopNavHeading } from '@astryxdesign/core/TopNav';
import { Outlet, useRouterState } from '@tanstack/react-router';
import { useI18n } from '../i18n/I18nProvider';
import { locales, type Locale } from '../i18n/messages';
import { navSections } from './navigation';
import { useSession, useSignOut } from './session';

const endonyms: Record<Locale, string> = { fr: 'Français', en: 'English' };

function useLanguageItems(): DropdownMenuOption {
  const { t, locale, setLocale } = useI18n();
  return {
    type: 'section',
    title: t('account.language'),
    items: locales.map((option) => ({
      label: endonyms[option],
      ...(option === locale ? { icon: <Icon icon="check" /> } : {}),
      onClick: () => {
        setLocale(option);
      },
    })),
  };
}

function AccountArea({ onSignOut }: { onSignOut: () => void }) {
  const { t } = useI18n();
  const session = useSession();
  const languages = useLanguageItems();
  if (session.isPending) return null;
  const user = session.data?.authenticated ? session.data.user : null;
  if (user) {
    return (
      <DropdownMenu
        button={{ label: user.display_name, variant: 'ghost' }}
        alignment="end"
        items={[
          { label: t('account.settings'), href: '/parametres' },
          { type: 'divider' },
          languages,
          { type: 'divider' },
          { label: t('account.signOut'), onClick: onSignOut },
        ]}
      />
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
        items={[
          { label: t('account.signUp'), href: '/inscription' },
          { label: t('account.settings'), href: '/parametres' },
          { type: 'divider' },
          languages,
        ]}
      />
    </HStack>
  );
}

function PrimaryNav() {
  const { t } = useI18n();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  return (
    <SideNav aria-label={t('nav.label')}>
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
    </SideNav>
  );
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
      mobileNav={{ content: <PrimaryNav /> }}
    >
      <Outlet />
    </AppShell>
  );
}
