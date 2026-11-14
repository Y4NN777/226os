import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Selector } from '@astryxdesign/core/Selector';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { Byline, TagList } from './present';
import {
  useArticleList,
  useSeriesList,
  useTopicQuestions,
  type ArticleSummary,
  type SeriesSummary,
} from './queries';
import { validateLearnSearch, type LearnSearch } from './routeSearch';
import { topicCodes, topicKey, type TopicCode } from './topics';

function Filter({ search }: { search: LearnSearch }) {
  const { t } = useI18n();
  const navigate = useNavigate();
  return (
    <HStack gap={2} vAlign="end" wrap="wrap">
      <Selector
        label={t('learn.topic.label')}
        isLabelHidden
        hasClear
        placeholder={t('learn.form.topic.placeholder')}
        options={topicCodes.map((code) => ({ value: code, label: t(topicKey(code)) }))}
        value={search.topic ?? null}
        onChange={(value) => {
          const topic = topicCodes.find((code) => code === value);
          void navigate({ to: '/apprendre', search: topic ? { topic } : {} });
        }}
      />
      {search.topic ? (
        <Button
          variant="ghost"
          label={t('learn.topic.reset')}
          onClick={() => void navigate({ to: '/apprendre', search: {} })}
        />
      ) : null}
    </HStack>
  );
}

function ArticleCard({ article }: { article: ArticleSummary }) {
  return (
    <Card variant="muted">
      <VStack gap={2}>
        <div lang={article.lang}>
          <VStack gap={2}>
            <Link href={`/apprendre/articles/${article.id}`} type="large" weight="semibold">
              {article.title}
            </Link>
            {article.summary ? <Text as="p">{article.summary}</Text> : null}
          </VStack>
        </div>
        {article.tags.length > 0 ? <TagList tags={article.tags} /> : null}
        <Byline author={article.author} />
      </VStack>
    </Card>
  );
}

function SeriesCard({ series }: { series: SeriesSummary }) {
  const { t } = useI18n();
  return (
    <Card variant="muted">
      <VStack gap={2}>
        <div lang={series.lang}>
          <VStack gap={2}>
            <Link href={`/apprendre/formations/${series.id}`} type="large" weight="semibold">
              {series.title}
            </Link>
            <Text as="p">{series.summary}</Text>
          </VStack>
        </div>
        <HStack gap={3} wrap="wrap" vAlign="center">
          <Byline author={series.author} />
          <Text type="supporting">{t('learn.series.lessons', { n: series.lesson_count })}</Text>
        </HStack>
      </VStack>
    </Card>
  );
}

/** Questions tagged with the topic code, next to its articles: a read-only feed. */
function TopicQuestions({ topic }: { topic: TopicCode }) {
  const { t } = useI18n();
  const questions = useTopicQuestions(topic);
  return (
    <section aria-label={t('learn.topic.questions')}>
      <VStack gap={3}>
        <Heading level={2}>{t('learn.topic.questions')}</Heading>
        {questions.isPending ? <LoadingState /> : null}
        {questions.isError ? (
          <ErrorState error={questions.error} onRetry={() => void questions.refetch()} />
        ) : null}
        {questions.data && questions.data.items.length === 0 ? (
          <Text as="p" type="supporting">
            {t('learn.topic.questions.empty')}
          </Text>
        ) : null}
        {questions.data && questions.data.items.length > 0 ? (
          <VStack as="ul" gap={2}>
            {questions.data.items.map((item) => (
              <li key={item.id}>
                <span lang={item.lang}>
                  <Link href={`/questions/${item.id}`}>{item.title}</Link>
                </span>
              </li>
            ))}
          </VStack>
        ) : null}
        <HStack>
          <Link href={`/solve?tag=${topic}`}>{t('learn.topic.questions.all')}</Link>
        </HStack>
      </VStack>
    </section>
  );
}

/** `apprendre`: the published articles and the training series, filtered by topic. */
export function LearnPage() {
  const { t } = useI18n();
  // Router merges the raw parent search over the validated one, so an unknown topic is dropped again here.
  const search = validateLearnSearch(useSearch({ from: '/app/apprendre' }));
  const session = useSession();
  const articles = useArticleList(search);
  const series = useSeriesList(search);
  const articleItems = articles.data?.pages.flatMap((page) => page.items) ?? [];
  const seriesItems = series.data?.pages.flatMap((page) => page.items) ?? [];

  return (
    <VStack gap={5}>
      <VStack gap={1} hAlign="start">
        <Heading level={1}>{t('learn.title')}</Heading>
        <Text as="p">{t('learn.lede')}</Text>
        {session.data?.authenticated ? (
          <Button variant="primary" label={t('learn.write.link')} href="/ecrire" />
        ) : null}
      </VStack>
      <Filter key={search.topic ?? ''} search={search} />

      {series.isError && !series.data ? (
        <ErrorState error={series.error} onRetry={() => void series.refetch()} />
      ) : null}
      {seriesItems.length > 0 ? (
        <section aria-label={t('learn.series.label')}>
          <VStack gap={3}>
            <Heading level={2}>{t('learn.series.label')}</Heading>
            <VStack as="ul" gap={3}>
              {seriesItems.map((item) => (
                <li key={item.id}>
                  <SeriesCard series={item} />
                </li>
              ))}
            </VStack>
            {series.hasNextPage ? (
              <HStack hAlign="center">
                <Button
                  label={t('learn.series.more')}
                  isLoading={series.isFetchingNextPage}
                  onClick={() => void series.fetchNextPage()}
                />
              </HStack>
            ) : null}
          </VStack>
        </section>
      ) : null}

      <section aria-label={t('learn.articles.label')}>
        <VStack gap={3}>
          <Heading level={2}>{t('learn.articles.label')}</Heading>
          {articles.isPending ? <LoadingState /> : null}
          {articles.isError && !articles.data ? (
            <ErrorState error={articles.error} onRetry={() => void articles.refetch()} />
          ) : null}
          {articles.data && articleItems.length === 0 ? (
            <EmptyState
              headingLevel={3}
              title={t('learn.articles.empty.title')}
              description={t('learn.articles.empty.body')}
            />
          ) : null}
          {articleItems.length > 0 ? (
            <VStack as="ul" gap={3}>
              {articleItems.map((item) => (
                <li key={item.id}>
                  <ArticleCard article={item} />
                </li>
              ))}
            </VStack>
          ) : null}
          {articles.isError && articles.data ? (
            <ErrorState error={articles.error} onRetry={() => void articles.fetchNextPage()} />
          ) : null}
          {articles.hasNextPage ? (
            <HStack hAlign="center">
              <Button
                label={t('learn.articles.more')}
                isLoading={articles.isFetchingNextPage}
                onClick={() => void articles.fetchNextPage()}
              />
            </HStack>
          ) : null}
        </VStack>
      </section>

      {search.topic ? <TopicQuestions topic={search.topic} /> : null}
    </VStack>
  );
}
