import { Badge } from '@astryxdesign/core/Badge';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { useI18n } from '../../i18n/I18nProvider';

export function Technologies({ items }: { items: readonly string[] }) {
  return (
    <HStack gap={1} wrap="wrap">
      {items.map((item) => (
        <Badge key={item} variant="blue" label={item} />
      ))}
    </HStack>
  );
}

/** Person link to the public profile; the profile itself decides what the caller may see. */
export function PersonLink({ username, name }: { username: string; name: string }) {
  return <Link href={`/profils/${username}`}>{name}</Link>;
}

export function OwnerByline({ username, name }: { username: string; name: string }) {
  const { t } = useI18n();
  return (
    <HStack gap={1} vAlign="center">
      <Text type="supporting">{t('project.owner')}</Text>
      <PersonLink username={username} name={name} />
    </HStack>
  );
}
