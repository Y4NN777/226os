import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useParams } from '@tanstack/react-router';
import type { ReactNode } from 'react';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale } from '../../i18n/messages';
import { USERNAME_RE } from '../identity/validation';
import { Checklist } from './Checklist';
import { githubProfileUrl } from './github';
import { isProfileMissing, useProfile, type PublicProfile } from './queries';

function displayName(type: 'region' | 'language', code: string, locale: Locale): string {
  try {
    return new Intl.DisplayNames([locale], { type }).of(code) ?? code;
  } catch {
    return code;
  }
}

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <VStack gap={0.5}>
      <Text type="label" color="secondary">
        {label}
      </Text>
      {children}
    </VStack>
  );
}

function Details({ profile, isOwn }: { profile: PublicProfile; isOwn: boolean }) {
  const { t, locale } = useI18n();
  const githubUrl = profile.github ? githubProfileUrl(profile.github) : null;
  const isEmpty =
    !profile.bio &&
    !profile.country &&
    !githubUrl &&
    !(profile.languages && profile.languages.length > 0) &&
    !(profile.skills && profile.skills.length > 0);
  const joined = new Intl.DateTimeFormat(locale, { year: 'numeric', month: 'long' }).format(
    new Date(profile.joined_at),
  );
  return (
    <VStack gap={4} hAlign="start">
      <VStack gap={1}>
        <Heading level={1}>{profile.display_name}</Heading>
        <HStack gap={2} wrap="wrap" vAlign="center">
          <Text type="supporting">@{profile.username}</Text>
          <Text type="supporting">{t('profile.joined', { date: joined })}</Text>
          {profile.open_to_collab ? <Badge variant="success" label={t('profile.collab')} /> : null}
        </HStack>
      </VStack>
      {/* A visitor sees why the page is bare; the member sees the checklist instead. */}
      {isEmpty && !isOwn ? (
        <Text as="p" type="supporting">
          {t('profile.empty', { name: profile.display_name })}
        </Text>
      ) : null}
      {profile.bio ? <Text as="p">{profile.bio}</Text> : null}
      {profile.country ? (
        <Row label={t('profile.country')}>
          <Text>{displayName('region', profile.country, locale)}</Text>
        </Row>
      ) : null}
      {profile.languages && profile.languages.length > 0 ? (
        <Row label={t('profile.languages')}>
          <Text>
            {profile.languages.map((code) => displayName('language', code, locale)).join(', ')}
          </Text>
        </Row>
      ) : null}
      {profile.skills && profile.skills.length > 0 ? (
        <Row label={t('profile.skills')}>
          <HStack gap={1} wrap="wrap">
            {profile.skills.map((skill) => (
              <Badge key={skill} variant="blue" label={skill} />
            ))}
          </HStack>
        </Row>
      ) : null}
      {githubUrl ? (
        <Row label={t('profile.github')}>
          <Link href={githubUrl} isExternalLink newTabLabel={t('discover.newTab')}>
            {profile.github}
          </Link>
        </Row>
      ) : null}
    </VStack>
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
    <VStack gap={6} maxWidth={640}>
      {/* Who the page is about first; the member's own checklist follows. */}
      <Details profile={query.data} isOwn={isOwn} />
      {isOwn ? <Checklist username={username} /> : null}
    </VStack>
  );
}
