import { Badge } from '@astryxdesign/core/Badge';
import { HStack } from '@astryxdesign/core/HStack';
import { Text } from '@astryxdesign/core/Text';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import type { Question } from './queries';

const stateKey = {
  open: 'solve.state.open',
  answered: 'solve.state.answered',
  resolved: 'solve.state.resolved',
} as const satisfies Record<Question['state'], MessageKey>;

const stateVariant = { open: 'neutral', answered: 'info', resolved: 'success' } as const;

export function StateBadge({ state }: { state: Question['state'] }) {
  const { t } = useI18n();
  return <Badge variant={stateVariant[state]} label={t(stateKey[state])} />;
}

export function TagList({ tags }: { tags: readonly string[] }) {
  return (
    <HStack gap={1} wrap="wrap">
      {tags.map((tag) => (
        <Badge key={tag} variant="blue" label={tag} />
      ))}
    </HStack>
  );
}

export function Byline({ name }: { name: string }) {
  const { t } = useI18n();
  return <Text type="supporting">{t('solve.by', { name })}</Text>;
}
