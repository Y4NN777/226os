import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useLocation, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import type { components } from '../../api/schema';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { isSafeInternalPath } from '../../components/safePath';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { FocusRegion } from '../identity/notices';
import { ArticleForm } from './ArticleForm';
import { parseLink } from './routeSearch';
import { useCreateArticle } from './useCreateArticle';

type ArticleCreate = components['schemas']['ArticleCreate'];

/** Sign-in address that comes back here, with the pre-filled link kept. */
function signInHref(path: string) {
  return isSafeInternalPath(path) ? `/connexion?next=${encodeURIComponent(path)}` : '/connexion';
}

function Published({ id, title, onAnother }: { id: string; title: string; onAnother: () => void }) {
  const { t } = useI18n();
  return (
    <FocusRegion role="status">
      <VStack gap={3} hAlign="start">
        <Heading level={1}>{t('learn.done.title')}</Heading>
        <Text as="p">{t('learn.done.body', { title })}</Text>
        <Button variant="primary" label={t('learn.done.another')} onClick={onAnother} />
        <Link href={`/apprendre/articles/${id}`} isStandalone>
          {t('learn.done.view')}
        </Link>
      </VStack>
    </FocusRegion>
  );
}

/**
 * `ecrire`: a member writes a Learn article (createLearnArticle). Anonymous visitors are invited
 * to sign in and come back; the API decides everything else (unverified email is a 403).
 */
export function WritePage() {
  const { t } = useI18n();
  const session = useSession();
  const search: { link?: string } = useSearch({ strict: false });
  const location = useLocation();
  const create = useCreateArticle();
  const [published, setPublished] = useState<{ id: string; title: string } | null>(null);
  // A new key per "write another": changing it remounts the form with empty fields.
  const [round, setRound] = useState(0);

  if (session.isPending) return <LoadingState />;
  if (session.isError)
    return <ErrorState error={session.error} onRetry={() => void session.refetch()} />;

  if (!session.data.authenticated) {
    return (
      <VStack gap={3} hAlign="start" maxWidth={560}>
        <Heading level={1}>{t('learn.signIn.title')}</Heading>
        <Text as="p">{t('learn.signIn.body')}</Text>
        <Button
          variant="primary"
          label={t('account.signIn')}
          href={signInHref(`${location.pathname}${location.searchStr}`)}
        />
      </VStack>
    );
  }

  if (published !== null) {
    return (
      <VStack gap={3} hAlign="start" maxWidth={560}>
        <Published
          id={published.id}
          title={published.title}
          onAnother={() => {
            setPublished(null);
            setRound((n) => n + 1);
          }}
        />
      </VStack>
    );
  }

  const submit = async (draft: ArticleCreate) => {
    const { id } = await create(draft);
    setPublished({ id, title: draft.title });
  };

  return (
    <VStack gap={4} hAlign="stretch" maxWidth={720}>
      <VStack gap={1} hAlign="start">
        <Heading level={1}>{t('learn.write.title')}</Heading>
        <Text as="p" color="secondary">
          {t('learn.write.lede')}
        </Text>
      </VStack>
      <ArticleForm key={round} initialLink={parseLink(search.link)} onSubmit={submit} />
    </VStack>
  );
}
