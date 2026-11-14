import { Badge } from '@astryxdesign/core/Badge';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import type { components } from '../../api/schema';
import { useI18n } from '../../i18n/I18nProvider';

type Author = components['schemas']['AuthorRef'];

export function TagList({ tags }: { tags: readonly string[] }) {
  return (
    <HStack as="ul" gap={1} wrap="wrap">
      {tags.map((tag) => (
        <li key={tag}>
          <Badge variant="blue" label={tag} />
        </li>
      ))}
    </HStack>
  );
}

/** Display name only, as in lists. */
export function Byline({ author }: { author: Author }) {
  const { t } = useI18n();
  return <Text type="supporting">{t('learn.by', { name: author.display_name })}</Text>;
}

/** Display name and a link to the profile by username. */
export function AuthorLine({ author }: { author: Author }) {
  const { t } = useI18n();
  return (
    <HStack gap={2} wrap="wrap" vAlign="center">
      <Text type="supporting">{t('learn.by', { name: author.display_name })}</Text>
      <Link href={`/profils/${author.username}`}>@{author.username}</Link>
    </HStack>
  );
}
