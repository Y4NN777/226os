import { Breadcrumbs, BreadcrumbItem } from '@astryxdesign/core/Breadcrumbs';
import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useLocation, useSearch } from '@tanstack/react-router';
import { useId, useState } from 'react';
import type { MessageKey } from '../../i18n/messages';
import { CheckCircleIcon, EditIcon, EyeIcon } from './icons';
import './learn.css';
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
      <div className="learn-done-card">
        <span className="learn-icon-badge" aria-hidden="true">
          <Icon icon={CheckCircleIcon} size="lg" />
        </span>
        <Heading level={1}>{t('learn.done.title')}</Heading>
        <p>{t('learn.done.body', { title })}</p>
        <div className="learn-actions">
          <Button
            variant="primary"
            label={t('learn.done.view')}
            href={`/apprendre/articles/${id}`}
            icon={<Icon icon={EyeIcon} />}
          />
          <Button
            variant="secondary"
            label={t('learn.done.another')}
            onClick={onAnother}
            icon={<Icon icon={EditIcon} />}
          />
          <Link href="/apprendre" isStandalone>
            {t('learn.series.back')}
          </Link>
        </div>
      </div>
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
    <VStack gap={6}>
      <Crumbs />
      <header className="learn-head-text">
        <Heading level={1}>{t('learn.write.title')}</Heading>
        <p className="learn-lede">{t('learn.write.lede')}</p>
      </header>
      <div className="learn-write">
        <div className="learn-form">
          <ArticleForm key={round} initialLink={parseLink(search.link)} onSubmit={submit} />
        </div>
        <Tips />
      </div>
    </VStack>
  );
}

function Crumbs() {
  const { t } = useI18n();
  return (
    <Breadcrumbs label={t('learn.title')} variant="supporting">
      <BreadcrumbItem href="/apprendre">{t('learn.title')}</BreadcrumbItem>
      <BreadcrumbItem isCurrent>{t('learn.write.title')}</BreadcrumbItem>
    </Breadcrumbs>
  );
}

const tipKeys = [
  'learn.write.tip.1',
  'learn.write.tip.2',
  'learn.write.tip.3',
  'learn.write.tip.4',
] as const satisfies readonly MessageKey[];

/** The prototype's writing advice, beside the form on a wide screen. */
function Tips() {
  const { t } = useI18n();
  const titleId = useId();
  return (
    <aside className="learn-write-aside learn-aside-card" aria-labelledby={titleId}>
      <h2 id={titleId}>{t('learn.write.tips.title')}</h2>
      <ul className="learn-checks">
        {tipKeys.map((key) => (
          <li key={key}>
            <Icon icon="check" size="sm" />
            <span>{t(key)}</span>
          </li>
        ))}
      </ul>
    </aside>
  );
}
