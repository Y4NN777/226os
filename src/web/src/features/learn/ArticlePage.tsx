import { Breadcrumbs, BreadcrumbItem } from '@astryxdesign/core/Breadcrumbs';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { Markdown } from '@astryxdesign/core/Markdown';
import { Outline, useOutlineFromMarkdown } from '@astryxdesign/core/Outline';
import { VStack } from '@astryxdesign/core/VStack';
import { useParams } from '@tanstack/react-router';
import { useId } from 'react';
import { ApiError } from '../../api/http';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { relatedPath } from '../related';
import {
  BookIcon,
  CompassIcon,
  FileTextIcon,
  HammerIcon,
  HelpIcon,
  LinkIcon,
  type SvgIcon,
} from './icons';
import './learn.css';
import { AuthorByline, CodeTags, KindBadge, LangCode, LevelBadge } from './present';
import { useArticle, type Article } from './queries';

type Related = Article['related'][number];

const relatedKey = {
  question: 'learn.related.question',
  resource: 'learn.related.resource',
  project: 'learn.related.project',
  article: 'learn.related.article',
  series: 'learn.related.series',
} as const satisfies Record<Related['type'], MessageKey>;

const relatedIcon: Record<Related['type'], SvgIcon> = {
  question: HelpIcon,
  resource: CompassIcon,
  project: HammerIcon,
  article: FileTextIcon,
  series: BookIcon,
};

function RelatedLinks({ links }: { links: Related[] }) {
  const { t } = useI18n();
  const titleId = useId();
  return (
    <section className="learn-section">
      <h2 className="learn-section-title" id={titleId}>
        <Icon icon={LinkIcon} size="sm" />
        <span>{t('learn.article.related')}</span>
      </h2>
      <ul className="learn-related" aria-labelledby={titleId}>
        {links.map((link) => (
          <li key={`${link.type}-${link.id}`} className="learn-linked">
            <span className="learn-kind">
              <Icon icon={relatedIcon[link.type]} size="xsm" />
              {t(relatedKey[link.type])}
            </span>
            <span lang={link.lang}>
              <Link href={relatedPath(link)}>{link.title}</Link>
            </span>
          </li>
        ))}
      </ul>
    </section>
  );
}

function Detail({ article }: { article: Article }) {
  const { t, locale } = useI18n();
  // The outline reads the same source as the body, so its links match the heading ids.
  const outline = useOutlineFromMarkdown(article.body_md);
  const hasOutline = outline.length > 1;
  return (
    <VStack gap={6}>
      <Breadcrumbs label={t('learn.title')} variant="supporting">
        <BreadcrumbItem href="/apprendre">{t('learn.title')}</BreadcrumbItem>
        <BreadcrumbItem isCurrent>{article.title}</BreadcrumbItem>
      </Breadcrumbs>
      {/* The article keeps its own language whatever the interface language is; the interface
          parts inside it say theirs. */}
      <article lang={article.lang}>
        <VStack gap={6}>
          <header className="learn-art-head">
            <div className="learn-pills" lang={locale}>
              <KindBadge kind={article.type} />
              <LevelBadge level={article.level} />
              <LangCode lang={article.lang} />
            </div>
            <VStack gap={3}>
              <h1 className="learn-art-title">{article.title}</h1>
              {article.summary ? <p className="learn-lede">{article.summary}</p> : null}
            </VStack>
            <div lang={locale}>
              <AuthorByline author={article.author} date={article.published_at} />
            </div>
          </header>
          <div className={hasOutline ? 'learn-art-grid learn-has-toc' : 'learn-art-grid'}>
            <div className="learn-body">
              {/* Astryx Markdown: raw HTML never becomes elements, unsafe links stay text,
                  fenced code is a CodeBlock with its copy button. */}
              <div className="learn-prose">
                {/* "## Section" is an h2 under the page h1, as the writing hint teaches. */}
                <Markdown headingLevelStart={1} contentWidth="100%">
                  {article.body_md}
                </Markdown>
              </div>
              <div className="learn-body" lang={locale}>
                {article.tags.length > 0 ? (
                  <CodeTags tags={article.tags} label={t('learn.form.tags')} />
                ) : null}
                <div>
                  <Button
                    variant="primary"
                    label={t('learn.article.askLesson')}
                    href={`/solve?link=article:${article.id}`}
                    icon={<Icon icon={HelpIcon} />}
                  />
                </div>
                {article.related.length > 0 ? <RelatedLinks links={article.related} /> : null}
              </div>
            </div>
            {hasOutline ? (
              <aside className="learn-toc" lang={locale}>
                <Outline items={outline} label={t('learn.article.toc')} density="compact" />
              </aside>
            ) : null}
          </div>
        </VStack>
      </article>
    </VStack>
  );
}

export function ArticlePage() {
  const { t } = useI18n();
  const { articleId } = useParams({ from: '/app/apprendre/articles/$articleId' });
  const article = useArticle(articleId);

  if (article.isPending) return <LoadingState />;
  if (article.isError) {
    if (article.error instanceof ApiError && article.error.status === 404) {
      return (
        <EmptyState
          headingLevel={1}
          title={t('learn.article.notFound.title')}
          description={t('learn.article.notFound.body')}
          actions={<Button label={t('learn.article.back')} href="/apprendre" />}
        />
      );
    }
    return <ErrorState error={article.error} onRetry={() => void article.refetch()} />;
  }
  return <Detail article={article.data} />;
}
