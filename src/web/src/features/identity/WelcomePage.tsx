import { Button } from '@astryxdesign/core/Button';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { VStack } from '@astryxdesign/core/VStack';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { FeedIcon, type FeedIconName } from '../home/FeedIcon';
import { CardHead } from './AuthCard';

/** The prototype's choice tile (bienvenue.html): icon, title, one line; the title link covers it. */
function WelcomeTile({
  icon,
  href,
  title,
  text,
}: {
  icon: FeedIconName;
  href: string;
  title: string;
  text: string;
}) {
  return (
    <li className="acc-tile">
      <FeedIcon name={icon} />
      <Link href={href} weight="semibold" color="primary">
        {title}
      </Link>
      <span className="acc-tile-text">{text}</span>
    </li>
  );
}

export function WelcomePage() {
  const { t } = useI18n();
  const session = useSession();

  if (session.isPending) return <LoadingState />;
  if (session.isError) {
    return (
      <ErrorState
        error={session.error}
        title="identity.error.title"
        onRetry={() => void session.refetch()}
      />
    );
  }

  const user = session.data.user;
  if (!session.data.authenticated || !user) {
    return (
      <VStack gap={4}>
        <CardHead
          title={t('identity.welcome.anonymous.title')}
          lede={t('identity.welcome.anonymous.body')}
        />
        <HStack gap={3} wrap="wrap">
          <Button
            variant="primary"
            href="/connexion?next=/bienvenue"
            label={t('identity.signIn.submit')}
          />
          <Button href="/inscription" label={t('identity.signIn.signUp')} />
        </HStack>
      </VStack>
    );
  }

  return (
    <VStack gap={4}>
      <CardHead
        title={t('identity.welcome.title', { name: user.display_name || user.username })}
        lede={t('identity.welcome.intro')}
      />
      <ul className="acc-tiles">
        <WelcomeTile
          icon="book"
          href="/apprendre"
          title={t('identity.welcome.learn')}
          text={t('identity.welcome.learn.text')}
        />
        <WelcomeTile
          icon="users"
          href="/connect"
          title={t('identity.welcome.connect')}
          text={t('identity.welcome.connect.text')}
        />
        <WelcomeTile
          icon="user"
          href={`/profils/${user.username}`}
          title={t('identity.welcome.profile')}
          text={t('identity.welcome.profile.text')}
        />
      </ul>
    </VStack>
  );
}
