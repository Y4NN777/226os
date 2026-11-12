import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useI18n } from '../../i18n/I18nProvider';
import { useHealth } from './useHealth';

export function HealthStatus() {
  const { t } = useI18n();
  const health = useHealth();

  return (
    <VStack gap={3} hAlign="start">
      <Heading level={2}>{t('health.title')}</Heading>
      {/* Polite live region: state changes are announced, never signalled by colour alone. */}
      <div role="status">
        {health.isPending ? <Text>{t('health.loading')}</Text> : null}
        {health.isError ? (
          <VStack gap={2} hAlign="start">
            <Badge variant="error" label={t('health.error')} />
            <Text type="supporting">{t('health.errorHint')}</Text>
          </VStack>
        ) : null}
        {health.isSuccess ? (
          <VStack gap={2} hAlign="start">
            <Badge variant="success" label={t('health.ok')} />
            <Text type="supporting">{t('health.status', { status: health.data.status })}</Text>
          </VStack>
        ) : null}
      </div>
      <Button
        label={t('health.retry')}
        variant="secondary"
        isLoading={health.isFetching}
        onClick={() => {
          void health.refetch();
        }}
      />
    </VStack>
  );
}
