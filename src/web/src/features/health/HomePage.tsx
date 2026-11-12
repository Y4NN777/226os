import { Heading } from '@astryxdesign/core/Heading';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useI18n } from '../../i18n/I18nProvider';
import { HealthStatus } from './HealthStatus';

export function HomePage() {
  const { t } = useI18n();
  return (
    <VStack gap={6} hAlign="start" maxWidth={720}>
      <Heading level={1}>{t('home.title')}</Heading>
      <Text as="p">{t('home.intro')}</Text>
      <HealthStatus />
    </VStack>
  );
}
