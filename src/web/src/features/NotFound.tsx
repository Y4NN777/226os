import { Heading } from '@astryxdesign/core/Heading';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { Link } from '@tanstack/react-router';
import { useI18n } from '../i18n/I18nProvider';

export function NotFound() {
  const { t } = useI18n();
  return (
    // No route matched, so no layout route rendered: this page brings its own frame.
    <main className="auth-layout">
      <VStack gap={3} hAlign="start" width="100%" maxWidth={440}>
        <Heading level={1}>{t('notFound.title')}</Heading>
        <Text as="p">{t('notFound.body')}</Text>
        <Link to="/" className="app-nav-link">
          {t('notFound.back')}
        </Link>
      </VStack>
    </main>
  );
}
