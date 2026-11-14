import { Card } from '@astryxdesign/core/Card';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { Outlet } from '@tanstack/react-router';
import { useI18n } from '../i18n/I18nProvider';

/** Standalone account pages: no app chrome, one centred card, a way home. */
export function AuthLayout() {
  const { t } = useI18n();
  return (
    <main className="auth-layout">
      <VStack gap={6} width="100%" maxWidth={440}>
        <Link href="/" isStandalone>
          <Text weight="bold" size="xl">
            {t('app.name')}
          </Text>
        </Link>
        <Card>
          <Outlet />
        </Card>
      </VStack>
    </main>
  );
}
