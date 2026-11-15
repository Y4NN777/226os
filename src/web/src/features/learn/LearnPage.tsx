import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { Selector } from '@astryxdesign/core/Selector';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useId } from 'react';
import type { components } from '../../api/schema';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { RouterAnchor } from '../../components/RouterAnchor';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { formatDate } from '../discover/present';
import { ArrowRightIcon, BookIcon, EditIcon, HelpIcon, SparklesIcon } from './icons';
import './learn.css';
import { formatMinutes, ItemRow, KindBadge, KindThumb, LangCode } from './present';
import {
  useArticleList,
  useSeries,
  useSeriesList,
  useTopicQuestions,
  type ArticleSummary,
  type SeriesSummary,
} from './queries';
import { validateLearnSearch, type LearnSearch } from './routeSearch';
import { topicCodes, topicKey, type TopicCode } from './topics';

type QuestionState = components['schemas']['QuestionState'];

const stateKey = {
  open: 'solve.state.open',
  answered: 'solve.state.answered',
  resolved: 'solve.state.resolved',
} as const satisfies Record<QuestionState, MessageKey>;

const stateVariant = { open: 'neutral', answered: 'blue', resolved: 'success' } as const;

function useCount() {
  const { t } = useI18n();
  return {
    lessons: (n: number) => (n === 1 ? t('learn.lessons.one') : t('learn.lessons.many', { n })),
    answers: (n: number) => (n === 1 ? t('learn.answers.one') : t('learn.answers.many', { n })),
  };
}

function Filter({ search }: { search: LearnSearch }) {
  const { t } = useI18n();
  const navigate = useNavigate();
  return (
    <div className="learn-toolbar">
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
    </div>
  );
}

function ArticleRow({ article }: { article: ArticleSummary }) {
  const { t, locale } = useI18n();
  return (
    <ItemRow
      lead={<KindThumb kind={article.type} id={article.id} />}
      top={<KindBadge kind={article.type} />}
      title={article.title}
      href={`/apprendre/articles/${article.id}`}
      lang={article.lang}
      excerpt={article.summary}
      meta={
        <>
          <span className="learn-meta-item">{article.author.display_name}</span>
          <span className="learn-meta-item">{t(`learn.level.${article.level}`)}</span>
          <LangCode lang={article.lang} />
        </>
      }
      aside={
        <time dateTime={article.published_at}>{formatDate(article.published_at, locale)}</time>
      }
    />
  );
}

function SeriesRow({ series }: { series: SeriesSummary }) {
  const { t } = useI18n();
  const count = useCount();
  return (
    <ItemRow
      lead={<KindThumb kind="series" id={series.id} />}
      top={<KindBadge kind="series" />}
      title={series.title}
      href={`/apprendre/formations/${series.id}`}
      lang={series.lang}
      excerpt={series.summary}
      meta={
        <>
          <span className="learn-meta-item">{series.author.display_name}</span>
          <span className="learn-meta-item">{t(`learn.level.${series.level}`)}</span>
          <LangCode lang={series.lang} />
        </>
      }
      aside={<span>{count.lessons(series.lesson_count)}</span>}
    />
  );
}

/**
 * The first series, shown as the prototype's featured training. Its lessons come from
 * getTrainingSeries; while they load, or if that call fails, the panel simply has no lesson column.
 */
function FeaturedSeries({ series }: { series: SeriesSummary }) {
  const { t } = useI18n();
  const count = useCount();
  const detail = useSeries(series.id);
  const titleId = useId();
  const stepsId = useId();
  const lessons = [...(detail.data?.lessons ?? [])].sort((a, b) => a.position - b.position);
  const first = lessons[0];
  const href = `/apprendre/formations/${series.id}`;
  return (
    <div className={lessons.length > 0 ? 'learn-feature learn-has-steps' : 'learn-feature'}>
      <div className="learn-feature-main">
        <p className="learn-kicker">
          <Icon icon={SparklesIcon} size="sm" />
          <span>{t('learn.feature.kicker', { name: series.author.display_name })}</span>
        </p>
        <h2 className="learn-feature-title" id={titleId} lang={series.lang}>
          <Link href={href} type="inherit" color="inherit">
            {series.title}
          </Link>
        </h2>
        <p className="learn-feature-sum" lang={series.lang}>
          {series.summary}
        </p>
        <p className="learn-meta">
          <span className="learn-meta-item">
            <Icon icon={BookIcon} size="sm" />
            {count.lessons(series.lesson_count)}
          </span>
          <span className="learn-meta-item">{t(`learn.level.${series.level}`)}</span>
          <LangCode lang={series.lang} />
        </p>
        <div className="learn-actions">
          {first ? (
            <Button
              variant="primary"
              label={t('learn.feature.start')}
              href={`/apprendre/articles/${first.article_id}`}
              endContent={<Icon icon={ArrowRightIcon} />}
            />
          ) : null}
          <Link href={href} isStandalone>
            {t('learn.feature.program')}
          </Link>
        </div>
      </div>
      {lessons.length > 0 ? (
        <div className="learn-steps" role="group" aria-labelledby={stepsId}>
          <p className="learn-steps-title" id={stepsId}>
            {t('learn.feature.steps')}
          </p>
          <ol aria-describedby={titleId}>
            {lessons.map((lesson, index) => (
              <li key={lesson.id}>
                <RouterAnchor
                  className="learn-step"
                  href={`/apprendre/articles/${lesson.article_id}`}
                >
                  <span className="learn-step-n" aria-hidden="true">
                    {index + 1}
                  </span>
                  <span className="learn-step-t" lang={series.lang}>
                    {lesson.title}
                  </span>
                  {lesson.duration_minutes ? (
                    <span className="learn-step-m">{formatMinutes(lesson.duration_minutes)}</span>
                  ) : null}
                </RouterAnchor>
              </li>
            ))}
          </ol>
        </div>
      ) : null}
    </div>
  );
}

/** Questions tagged with the topic code, next to its articles: a read-only feed. */
function TopicQuestions({ topic }: { topic: TopicCode }) {
  const { t, locale } = useI18n();
  const count = useCount();
  const questions = useTopicQuestions(topic);
  return (
    <section className="learn-section" aria-label={t('learn.topic.questions')}>
      <div className="learn-section-head">
        <Heading level={2}>{t('learn.topic.questions')}</Heading>
        <Link href={`/solve?tag=${topic}`}>{t('learn.topic.questions.all')}</Link>
      </div>
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
        <ul className="learn-list">
          {questions.data.items.map((item) => (
            <ItemRow
              key={item.id}
              lead={
                <span className="learn-thumb learn-tone-2" aria-hidden="true">
                  <Icon icon={HelpIcon} size="lg" />
                </span>
              }
              top={<Badge variant={stateVariant[item.state]} label={t(stateKey[item.state])} />}
              title={item.title}
              href={`/questions/${item.id}`}
              lang={item.lang}
              meta={
                item.answer_count > 0 ? (
                  <span className="learn-meta-item">{count.answers(item.answer_count)}</span>
                ) : null
              }
              aside={<time dateTime={item.created_at}>{formatDate(item.created_at, locale)}</time>}
            />
          ))}
        </ul>
      ) : null}
    </section>
  );
}

function Invite() {
  const { t } = useI18n();
  const titleId = useId();
  return (
    <section className="learn-invite" aria-labelledby={titleId}>
      <span className="learn-icon-badge" aria-hidden="true">
        <Icon icon={EditIcon} size="lg" />
      </span>
      <div>
        <h2 id={titleId}>{t('learn.invite.title')}</h2>
        <p>{t('learn.invite.body')}</p>
      </div>
      <div>
        <Button
          variant="primary"
          label={t('learn.invite.cta')}
          href="/ecrire"
          icon={<Icon icon={EditIcon} />}
        />
      </div>
    </section>
  );
}

/** `apprendre`: the published articles and the training series, filtered by topic. */
export function LearnPage() {
  const { t } = useI18n();
  // Router merges the raw parent search over the validated one, so an unknown topic is dropped again here.
  const search = validateLearnSearch(useSearch({ from: '/app/apprendre' }));
  const session = useSession();
  const member = session.data?.authenticated === true;
  const articles = useArticleList(search);
  const series = useSeriesList(search);
  const articleItems = articles.data?.pages.flatMap((page) => page.items) ?? [];
  const [featured, ...otherSeries] = series.data?.pages.flatMap((page) => page.items) ?? [];

  return (
    <VStack gap={6}>
      <header className="learn-head">
        <div className="learn-head-text">
          <Heading level={1}>{t('learn.title')}</Heading>
          <p className="learn-lede">{t('learn.lede')}</p>
        </div>
        {member ? (
          <Button
            variant="primary"
            label={t('learn.write.link')}
            href="/ecrire"
            icon={<Icon icon={EditIcon} />}
          />
        ) : null}
      </header>

      {series.isError && !series.data ? (
        <ErrorState error={series.error} onRetry={() => void series.refetch()} />
      ) : null}
      {featured ? (
        <section className="learn-section" aria-label={t('learn.series.label')}>
          <FeaturedSeries key={featured.id} series={featured} />
          {otherSeries.length > 0 ? (
            <ul className="learn-list">
              {otherSeries.map((item) => (
                <SeriesRow key={item.id} series={item} />
              ))}
            </ul>
          ) : null}
          {series.hasNextPage ? (
            <div className="learn-more">
              <Button
                label={t('learn.series.more')}
                isLoading={series.isFetchingNextPage}
                onClick={() => void series.fetchNextPage()}
              />
            </div>
          ) : null}
        </section>
      ) : null}

      <Filter key={search.topic ?? ''} search={search} />

      <section className="learn-section" aria-label={t('learn.articles.label')}>
        <div className="learn-section-head">
          <Heading level={2}>{t('learn.articles.label')}</Heading>
        </div>
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
          <ul className="learn-list">
            {articleItems.map((item) => (
              <ArticleRow key={item.id} article={item} />
            ))}
          </ul>
        ) : null}
        {articles.isError && articles.data ? (
          <ErrorState error={articles.error} onRetry={() => void articles.fetchNextPage()} />
        ) : null}
        {articles.hasNextPage ? (
          <div className="learn-more">
            <Button
              label={t('learn.articles.more')}
              isLoading={articles.isFetchingNextPage}
              onClick={() => void articles.fetchNextPage()}
            />
          </div>
        ) : null}
      </section>

      {search.topic ? <TopicQuestions topic={search.topic} /> : null}

      {member ? <Invite /> : null}
    </VStack>
  );
}
