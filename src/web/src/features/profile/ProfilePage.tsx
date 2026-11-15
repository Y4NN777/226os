import { Button } from '@astryxdesign/core/Button';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { useParams } from '@tanstack/react-router';
import { useId } from 'react';
import { useSession } from '../../app/session';
import { ProfilePhoto } from '../../components/ProfilePhoto';
import { Breadcrumbs } from '../../components/Breadcrumbs';
import { CountryCode } from '../../components/CountryCode';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import {
  ClockIcon,
  EditIcon,
  GitBranchIcon,
  LanguagesIcon,
  LockIcon,
  MapPinIcon,
} from '../../components/icons';
import { LoadingState } from '../../components/LoadingState';
import { Meta, MetaItem } from '../../components/Meta';
import { PageHeader } from '../../components/PageHeader';
import { StatusBadge } from '../../components/StatusBadge';
import { TagList } from '../../components/Tag';
import { useI18n } from '../../i18n/I18nProvider';
import { languageName } from '../discover/present';
import { USERNAME_RE } from '../identity/validation';
import { Activity } from './Activity';
import { Checklist } from './Checklist';
import { githubProfileUrl } from './github';
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
    // The prototype's hero card (`.id-hero`): the shared header on the panel surface.
    <section aria-labelledby={headingId} className="ui-panel profile-hero">
      <PageHeader
        title={profile.display_name}
        headingId={headingId}
        media={<ProfilePhoto name={profile.display_name} github={profile.github} size="xl" />}
        subtitle={`@${profile.username}`}
        meta={
          <>
            {profile.country ? (
              <MetaItem icon={MapPinIcon}>
                <CountryCode code={profile.country} withName />
              </MetaItem>
            ) : null}
            {languages ? <MetaItem icon={LanguagesIcon}>{languages}</MetaItem> : null}
            <MetaItem icon={ClockIcon}>{t('profile.joined', { date: joined })}</MetaItem>
          </>
        }
        actions={
          isOwn ? (
            <Button
              href="/parametres"
              variant="ghost"
              label={t('profile.edit')}
              icon={<Icon icon={EditIcon} />}
              isIconOnly
              tooltip={t('profile.edit')}
            />
          ) : undefined
        }
      >
        {/* A visitor sees why the page is bare; the member sees the checklist instead. */}
        {isEmpty && !isOwn ? (
          <Text as="p" type="supporting">
            {t('profile.empty', { name: profile.display_name })}
          </Text>
        ) : null}
        {profile.bio ? <p>{profile.bio}</p> : null}
        {profile.skills && profile.skills.length > 0 ? <TagList tags={profile.skills} /> : null}
        {profile.open_to_collab || githubUrl ? (
          <Meta>
            {profile.open_to_collab ? (
              <StatusBadge kind="open" label={t('profile.collab')} />
            ) : null}
            {githubUrl ? (
              <Link href={githubUrl} isExternalLink newTabLabel={t('discover.newTab')}>
                <MetaItem icon={GitBranchIcon}>{profile.github}</MetaItem>
              </Link>
            ) : null}
          </Meta>
        ) : null}
      </PageHeader>
    </section>
  );
}

/** Own profile only: the email is never public, and the visibility settings are one link away. */
function PrivacyNote() {
  const { t } = useI18n();
  return (
    <Meta>
      <MetaItem icon={LockIcon}>{t('profile.privacy')}</MetaItem>
      <Link href="/parametres#confidentialite">{t('profile.privacy.link')}</Link>
    </Meta>
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
    <div className="ui-page">
      <Breadcrumbs
        items={[
          { label: t('nav.home'), href: '/' },
          { label: t('connect.title'), href: '/connect' },
          { label: query.data.display_name },
        ]}
      />
      {/* Who the page is about first; the member's own checklist follows. */}
      <Hero profile={query.data} isOwn={isOwn} />
      {isOwn ? <PrivacyNote /> : null}
      {isOwn ? <Checklist username={username} /> : null}
      <Activity username={username} />
    </div>
  );
}
