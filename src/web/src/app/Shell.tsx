import { Layout, LayoutContent, LayoutHeader } from '@astryxdesign/core/Layout';
import { HStack } from '@astryxdesign/core/HStack';
import { Text } from '@astryxdesign/core/Text';
import { Link, Outlet } from '@tanstack/react-router';
import { useI18n } from '../i18n/I18nProvider';

export function Shell() {
  const { t } = useI18n();
  return (
    <Layout
      header={
        <LayoutHeader hasDivider>
          <HStack gap={6} vAlign="center">
            <Text weight="bold">{t('app.name')}</Text>
            <nav aria-label={t('nav.label')}>
              <HStack gap={4}>
                <Link to="/" className="app-nav-link" activeProps={{ 'aria-current': 'page' }}>
                  {t('nav.home')}
                </Link>
                <Link
                  to="/settings"
                  className="app-nav-link"
                  activeProps={{ 'aria-current': 'page' }}
                >
                  {t('nav.settings')}
                </Link>
              </HStack>
            </nav>
          </HStack>
        </LayoutHeader>
      }
    >
      <LayoutContent padding={6} role="main">
        <Outlet />
      </LayoutContent>
    </Layout>
  );
}
