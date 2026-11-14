import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';

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
        <Heading level={1}>{t('identity.welcome.anonymous.title')}</Heading>
        <Text as="p">{t('identity.welcome.anonymous.body')}</Text>
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
      <Heading level={1}>
        {t('identity.welcome.title', { name: user.display_name || user.username })}
      </Heading>
      <Text as="p">{t('identity.welcome.intro')}</Text>
      <HStack gap={3} wrap="wrap">
        <Button variant="primary" href="/discover" label={t('identity.welcome.discover')} />
        <Button href="/recherche" label={t('identity.welcome.search')} />
        <Button href="/settings" label={t('identity.welcome.settings')} />
      </HStack>
    </VStack>
  );
}
