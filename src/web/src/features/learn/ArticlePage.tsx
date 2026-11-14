import { Badge } from '@astryxdesign/core/Badge';
import { Breadcrumbs, BreadcrumbItem } from '@astryxdesign/core/Breadcrumbs';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useParams } from '@tanstack/react-router';
import { ApiError } from '../../api/http';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { RichText } from '../solve/RichText';
import { AuthorLine, TagList } from './present';
import { useArticle, type Article } from './queries';

type Related = Article['related'][number];

const relatedKey = {
  question: 'learn.related.question',
  resource: 'learn.related.resource',
  project: 'learn.related.project',
  article: 'learn.related.article',
  series: 'learn.related.series',
} as const satisfies Record<Related['type'], MessageKey>;

function pathOf(link: Related): string {
  switch (link.type) {
    case 'question':
      return `/questions/${link.id}`;
    case 'resource':
      return `/discover/${link.id}`;
    case 'project':
      return `/build/${link.id}`;
    case 'article':
      return `/apprendre/articles/${link.id}`;
    case 'series':
      return `/apprendre/formations/${link.id}`;
  }
}

function Detail({ article }: { article: Article }) {
  const { t } = useI18n();
  return (
    <VStack gap={5}>
      <VStack gap={2}>
        <Breadcrumbs label={t('learn.title')} variant="supporting">
          <BreadcrumbItem href="/apprendre">{t('learn.title')}</BreadcrumbItem>
          <BreadcrumbItem isCurrent>{article.title}</BreadcrumbItem>
        </Breadcrumbs>
        {/* The content keeps its own language, whatever the interface language is. */}
        <article lang={article.lang}>
          <VStack gap={4}>
            <VStack gap={2}>
              <Heading level={1}>{article.title}</Heading>
              {article.summary ? <Text as="p">{article.summary}</Text> : null}
            </VStack>
            <RichText markdown={article.body_md} />
          </VStack>
        </article>
        <HStack gap={3} wrap="wrap" vAlign="center">
          <AuthorLine author={article.author} />
          <Badge label={t(`learn.type.${article.type}`)} />
          <Badge label={t(`learn.level.${article.level}`)} />
        </HStack>
        {article.tags.length > 0 ? <TagList tags={article.tags} /> : null}
      </VStack>

      <HStack>
        <Button
          variant="primary"
          label={t('learn.article.askLesson')}
          href={`/solve?link=article:${article.id}`}
        />
      </HStack>

      {article.related.length > 0 ? (
        <VStack gap={2}>
          <Heading level={2} id="learn-related">
            {t('learn.article.related')}
          </Heading>
          <VStack as="ul" gap={2} aria-labelledby="learn-related">
            {article.related.map((link) => (
              <li key={`${link.type}-${link.id}`}>
                <HStack gap={2} wrap="wrap">
                  <Text type="supporting">{t(relatedKey[link.type])}</Text>
                  <span lang={link.lang}>
                    <Link href={pathOf(link)}>{link.title}</Link>
                  </span>
                </HStack>
              </li>
            ))}
          </VStack>
        </VStack>
      ) : null}
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
