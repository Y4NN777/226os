import { Avatar } from '@astryxdesign/core/Avatar';
import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { StackItem } from '@astryxdesign/core/Stack';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useParams } from '@tanstack/react-router';
import { useId } from 'react';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { CountryMeta, MetaItem, SkillTags } from '../connect/PersonRow';
import { languageName } from '../discover/present';
import { USERNAME_RE } from '../identity/validation';
import { Checklist } from './Checklist';
import { githubProfileUrl } from './github';
import { EditIcon, GitBranchIcon, LanguagesIcon, LockIcon } from './icons';
import { isProfileMissing, useProfile, type PublicProfile } from './queries';

/**
 * The prototype's profile hero: avatar, name and handle, one meta line (country, languages, join
 * date), the bio, skills, then the collaboration pill and GitHub. Only what the server returned
 * for this caller is drawn; a hidden field leaves no label or placeholder.
 */
function Hero({ profile, isOwn }: { profile: PublicProfile; isOwn: boolean }) {
  const { t, locale } = useI18n();
  const headingId = useId();
  const githubUrl = profile.github ? githubProfileUrl(profile.github) : null;
  const languages =
    profile.languages && profile.languages.length > 0
      ? profile.languages.map((code) => languageName(code, locale)).join(', ')
      : null;
  const isEmpty =
    !profile.bio &&
    !profile.country &&
    !githubUrl &&
    languages === null &&
    !(profile.skills && profile.skills.length > 0);
  const joined = new Intl.DateTimeFormat(locale, { year: 'numeric', month: 'long' }).format(
    new Date(profile.joined_at),
  );
  return (
    <section aria-labelledby={headingId}>
      <Card variant="muted" padding={5}>
        <HStack gap={4} vAlign="start" wrap="nowrap">
          <Avatar name={profile.display_name} size={72} />
          {/* fill: takes the rest of the row and may shrink, so a phone never scrolls sideways. */}
          <StackItem size="fill">
            <VStack gap={3}>
              <HStack gap={3} vAlign="start" justify="between" wrap="nowrap">
                <VStack gap={0.5}>
                  <Heading level={1} id={headingId}>
                    {profile.display_name}
                  </Heading>
                  <Text type="supporting">@{profile.username}</Text>
                </VStack>
                {isOwn ? (
                  <Button
                    href="/parametres"
                    variant="ghost"
                    label={t('profile.edit')}
                    icon={<Icon icon={EditIcon} />}
                    isIconOnly
                    tooltip={t('profile.edit')}
                  />
                ) : null}
              </HStack>
              <HStack gap={4} wrap="wrap" vAlign="center">
                {profile.country ? <CountryMeta code={profile.country} /> : null}
                {languages ? (
                  <MetaItem icon={LanguagesIcon}>
                    <Text type="supporting">{languages}</Text>
                  </MetaItem>
                ) : null}
                <MetaItem icon="clock">
                  <Text type="supporting">{t('profile.joined', { date: joined })}</Text>
                </MetaItem>
              </HStack>
              {/* A visitor sees why the page is bare; the member sees the checklist instead. */}
              {isEmpty && !isOwn ? (
                <Text as="p" type="supporting">
                  {t('profile.empty', { name: profile.display_name })}
                </Text>
              ) : null}
              {profile.bio ? <Text as="p">{profile.bio}</Text> : null}
              {profile.skills && profile.skills.length > 0 ? (
                <SkillTags skills={profile.skills} />
              ) : null}
              {profile.open_to_collab || githubUrl ? (
                <HStack gap={4} wrap="wrap" vAlign="center">
                  {profile.open_to_collab ? (
                    // The prototype's state pill: a word and an icon on a neutral fill.
                    <Badge
                      variant="neutral"
                      icon={<Icon icon="success" color="success" />}
                      label={t('profile.collab')}
                    />
                  ) : null}
                  {githubUrl ? (
                    <Link href={githubUrl} isExternalLink newTabLabel={t('discover.newTab')}>
                      <MetaItem icon={GitBranchIcon}>{profile.github}</MetaItem>
                    </Link>
                  ) : null}
                </HStack>
              ) : null}
            </VStack>
          </StackItem>
        </HStack>
      </Card>
    </section>
  );
}

/** Own profile only: the email is never public, and the visibility settings are one link away. */
function PrivacyNote() {
  const { t } = useI18n();
  return (
    <HStack gap={2} vAlign="center" wrap="wrap">
      <Icon icon={LockIcon} size="sm" color="secondary" />
      <Text type="supporting">{t('profile.privacy')}</Text>
      <Link href="/parametres#confidentialite">{t('profile.privacy.link')}</Link>
    </HStack>
  );
}

export function ProfilePage() {
  const { t } = useI18n();
  const { username } = useParams({ from: '/app/profils/$username' });
  const session = useSession();
  const wellFormed = USERNAME_RE.test(username);
  const query = useProfile(username, wellFormed);
  const user = session.data?.authenticated ? session.data.user : null;
  const isOwn = user !== null && user.username === username;

  if (!wellFormed || (query.isError && isProfileMissing(query.error))) {
    return (
      <EmptyState
        headingLevel={1}
        title={t('profile.notFound.title')}
        description={t('profile.notFound.body')}
        actions={<Button label={t('profile.back')} href="/" />}
      />
    );
  }
  if (!query.isError && !query.data) return <LoadingState />;
  if (query.isError) return <ErrorState error={query.error} onRetry={() => void query.refetch()} />;

  return (
    <VStack gap={4} maxWidth={960}>
      {/* Who the page is about first; the member's own checklist follows. */}
      <Hero profile={query.data} isOwn={isOwn} />
      {isOwn ? <PrivacyNote /> : null}
      {isOwn ? <Checklist username={username} /> : null}
    </VStack>
  );
}
