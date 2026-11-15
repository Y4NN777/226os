import { Avatar } from '@astryxdesign/core/Avatar';
import { Badge } from '@astryxdesign/core/Badge';
import { Divider } from '@astryxdesign/core/Divider';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { StackItem } from '@astryxdesign/core/Stack';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import type { ComponentProps, ReactNode } from 'react';
import { useI18n } from '../../i18n/I18nProvider';
import { countryName, languageName } from '../discover/present';
import { githubProfileUrl } from '../profile/github';
import { GitBranchIcon, LanguagesIcon, MapPinIcon } from '../profile/icons';
import type { ProfileSummary } from './queries';

/** One fact of a meta line: a decorative glyph, then the words. */
export function MetaItem({
  icon,
  children,
}: {
  icon: ComponentProps<typeof Icon>['icon'];
  children: ReactNode;
}) {
  return (
    <HStack as="span" gap={1} vAlign="center" wrap="nowrap">
      <Icon icon={icon} size="sm" color="secondary" />
      {children}
    </HStack>
  );
}

/**
 * A country as the prototype shows it: the two-letter code as a small chip (decoration, hidden
 * from screen readers), then the name, which is what is read.
 */
export function CountryMeta({ code }: { code: string }) {
  const { locale } = useI18n();
  return (
    <MetaItem icon={MapPinIcon}>
      <span aria-hidden="true">
        <Badge variant="neutral" label={code} />
      </span>
      <Text type="supporting">{countryName(code, locale)}</Text>
    </MetaItem>
  );
}

/** Skills as the prototype's tags: small neutral chips, no colour meaning. */
export function SkillTags({ skills }: { skills: readonly string[] }) {
  return (
    <HStack as="ul" gap={1} wrap="wrap">
      {skills.map((skill) => (
        <li key={skill}>
          <Badge variant="neutral" label={skill} />
        </li>
      ))}
    </HStack>
  );
}

/**
 * A directory row (prototype `list > row`): avatar, name linking to the profile, handle, a meta
 * line with what the caller may see, skills, and the GitHub link at the end. Fields the server
 * hid are absent, so nothing is drawn for them.
 */
export function PersonRow({ person, isLast }: { person: ProfileSummary; isLast: boolean }) {
  const { t, locale } = useI18n();
  const githubUrl = person.github ? githubProfileUrl(person.github) : null;
  const languages =
    person.languages && person.languages.length > 0
      ? person.languages.map((code) => languageName(code, locale)).join(', ')
      : null;
  const hasMeta = Boolean(person.country) || languages !== null;
  return (
    <li>
      <HStack gap={3} vAlign="start" wrap="nowrap" paddingBlock={3}>
        <Avatar name={person.display_name} size="md" />
        <StackItem size="fill">
          <VStack gap={1.5}>
            <HStack gap={2} wrap="wrap" vAlign="center">
              <Link href={`/profils/${person.username}`} type="large" weight="semibold">
                {person.display_name}
              </Link>
              <Text type="supporting">@{person.username}</Text>
            </HStack>
            {hasMeta ? (
              <HStack gap={4} wrap="wrap" vAlign="center">
                {person.country ? <CountryMeta code={person.country} /> : null}
                {languages ? (
                  <MetaItem icon={LanguagesIcon}>
                    <Text type="supporting">{languages}</Text>
                  </MetaItem>
                ) : null}
              </HStack>
            ) : null}
            {person.skills && person.skills.length > 0 ? (
              <SkillTags skills={person.skills} />
            ) : null}
          </VStack>
        </StackItem>
        {githubUrl ? (
          <Link
            href={githubUrl}
            aria-label={t('connect.github', { name: person.display_name })}
            isStandalone
          >
            <MetaItem icon={GitBranchIcon}>GitHub</MetaItem>
          </Link>
        ) : null}
      </HStack>
      {isLast ? null : <Divider variant="subtle" />}
    </li>
  );
}
