import { Heading } from '@astryxdesign/core/Heading';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { Link } from '@tanstack/react-router';
import { useI18n } from '../i18n/I18nProvider';

export function NotFound() {
  const { t } = useI18n();
  return (
    <VStack gap={3} hAlign="start">
      <Heading level={1}>{t('notFound.title')}</Heading>
      <Text as="p">{t('notFound.body')}</Text>
      <Link to="/" className="app-nav-link">
        {t('notFound.back')}
      </Link>
    </VStack>
  );
}
