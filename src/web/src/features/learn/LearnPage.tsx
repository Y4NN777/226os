import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { Selector } from '@astryxdesign/core/Selector';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useId, useState } from 'react';
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
import { formatMinutes, isSingular, ItemRow, KindBadge, KindThumb, LangCode } from './present';
import {
  articleFilters,
  seriesFilters,
  useArticleList,
  useSeries,
  useSeriesList,
  useTopicQuestions,
  type ArticleSummary,
  type SeriesSummary,
} from './queries';
import {
  learnLangs,
  learnLevels,
  learnTypes,
  validateLearnSearch,
  type LearnSearch,
  type LearnType,
} from './routeSearch';
import { topicCodes, topicKey, type TopicCode } from './topics';

type QuestionState = components['schemas']['QuestionState'];

// The solve pass replaced its solve.state.* words with the state pills and status words.
const stateKey = {
  open: 'solve.pill.open',
  answered: 'status.answered',
  resolved: 'solve.pill.resolved',
} as const satisfies Record<QuestionState, MessageKey>;

const stateVariant = { open: 'neutral', answered: 'blue', resolved: 'success' } as const;

function useCount() {
  const { t, locale } = useI18n();
  return {
    lessons: (n: number) =>
      t(isSingular(n, locale) ? 'learn.lessons.one' : 'learn.lessons.many', { n }),
    answers: (n: number) =>
      t(isSingular(n, locale) ? 'learn.answers.one' : 'learn.answers.many', { n }),
  };
}

type SearchPatch = { [K in keyof LearnSearch]?: LearnSearch[K] | undefined };

const ALL = 'all';

/** The prototype's filter area: type chips, then the text search and the Sujet, Niveau, Langue selects. */
function Filters({
  search,
  onChange,
}: {
  search: LearnSearch;
  onChange: (patch: SearchPatch) => void;
}) {
  const { t } = useI18n();
  const [draft, setDraft] = useState(search.q ?? '');
  const searchable = search.type !== 'series';
  return (
    <div className="learn-toolbar">
      {/* Prototype chips wrap; segments truncate, so the row scrolls on its own instead at 375px. */}
      <div className="learn-seg">
        <SegmentedControl
          label={t('learn.filter.type')}
          value={search.type ?? ALL}
          onChange={(value) => {
            const type = learnTypes.find((candidate) => candidate === value);
            // The series endpoint has no text search, so the Formations chip leaves q behind.
            onChange(type === 'series' ? { type, q: undefined } : { type });
          }}
        >
          <SegmentedControlItem value={ALL} label={t('learn.filter.all')} />
          {learnTypes.map((type) => (
            <SegmentedControlItem key={type} value={type} label={t(`learn.filter.${type}`)} />
          ))}
        </SegmentedControl>
      </div>
      <div className={searchable ? 'learn-toolbar-main' : 'learn-toolbar-main learn-no-search'}>
        {searchable ? (
          <form
            role="search"
            className="learn-search"
            onSubmit={(event) => {
              event.preventDefault();
              onChange({ q: draft.trim().slice(0, 200) || undefined });
            }}
          >
            <TextInput
              type="search"
              startIcon={<Icon icon="search" />}
              label={t('learn.filter.q')}
              placeholder={t('learn.filter.q.placeholder')}
              value={draft}
              onChange={setDraft}
              width="100%"
            />
          </form>
        ) : null}
        <div className="learn-selects">
          <Selector
            label={t('learn.topic.label')}
            hasClear
            width="100%"
            placeholder={t('learn.form.topic.placeholder')}
            options={topicCodes.map((code) => ({ value: code, label: t(topicKey(code)) }))}
            value={search.topic ?? null}
            onChange={(value) => {
              onChange({ topic: topicCodes.find((code) => code === value) });
            }}
          />
          <Selector
            label={t('learn.filter.level')}
            hasClear
            width="100%"
            placeholder={t('learn.filter.level.any')}
            options={learnLevels.map((level) => ({
              value: level,
              label: t(`learn.level.${level}`),
            }))}
            value={search.level ?? null}
            onChange={(value) => {
              onChange({ level: learnLevels.find((level) => level === value) });
            }}
          />
          <Selector
            label={t('learn.filter.lang')}
            hasClear
            width="100%"
            placeholder={t('learn.filter.lang.any')}
            options={learnLangs.map((lang) => ({ value: lang, label: t(`learn.lang.${lang}`) }))}
            value={search.lang ?? null}
            onChange={(value) => {
              onChange({ lang: learnLangs.find((lang) => lang === value) });
            }}
          />
        </div>
      </div>
    </div>
  );
}

type CountT = (key: MessageKey, params?: { n: number }) => string;

/** Results count: a total only once every shown list is fully loaded, since pages are cursor-based. */
function countText(n: number, hasMore: boolean, singular: boolean, t: CountT) {
  if (hasMore) return t('learn.count.first', { n });
  return t(singular ? 'learn.count.one' : 'learn.count.many', { n });
}

const emptyKey = {
  all: 'learn.empty.all',
  article: 'learn.empty.article',
  tutorial: 'learn.empty.tutorial',
  series: 'learn.empty.series',
} as const satisfies Record<LearnType | typeof ALL, MessageKey>;

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
    <section
      className={lessons.length > 0 ? 'learn-feature learn-has-steps' : 'learn-feature'}
      aria-labelledby={titleId}
    >
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
    </section>
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

/** Applies a patch to the URL filters; an undefined value removes the filter. */
function patched(base: LearnSearch, patch: SearchPatch): LearnSearch {
  const entries = Object.entries({ ...base, ...patch }).filter(([, value]) => value !== undefined);
  return Object.fromEntries(entries);
}

/**
 * `apprendre`: the featured training, then one filter area over the series and the articles.
 * Each list keeps its own cursor (two endpoints), so the series rows come first, then the
 * articles, as the prototype orders formations before the rest; nothing is merged or re-sorted.
 */
export function LearnPage() {
  const { t, locale } = useI18n();
  // Router merges the raw parent search over the validated one, so unknown values are dropped again here.
  const search = validateLearnSearch(useSearch({ from: '/app/apprendre' }));
  const navigate = useNavigate();
  const browseId = useId();
  const session = useSession();
  const member = session.data?.authenticated === true;
  const filtered = Object.keys(search).length > 0;
  const showArticles = search.type !== 'series';
  // listTrainingSeries has no type and no text search: series rows only under Tout or Formations, without q.
  const showSeries = (search.type === undefined || search.type === 'series') && !search.q;
  const articles = useArticleList(articleFilters(search), showArticles);
  const series = useSeriesList(seriesFilters(search), showSeries);
  // The featured training ignores the filters, as in the prototype; unfiltered, it shares the list's cache entry.
  const [featured] = useSeriesList({}).data?.pages.flatMap((page) => page.items) ?? [];
  const seriesItems = showSeries ? (series.data?.pages.flatMap((page) => page.items) ?? []) : [];
  const articleItems = showArticles
    ? (articles.data?.pages.flatMap((page) => page.items) ?? [])
    : [];
  const shown = seriesItems.length + articleItems.length;
  const pending = (showArticles && articles.isPending) || (showSeries && series.isPending);
  const articlesFailed = showArticles && articles.isError && !articles.data;
  const seriesFailed = showSeries && series.isError && !series.data;
  const settled = !pending && !articlesFailed && !seriesFailed;
  const hasMore = (showArticles && articles.hasNextPage) || (showSeries && series.hasNextPage);

  const change = (patch: SearchPatch) =>
    void navigate({ to: '/apprendre', search: patched(search, patch) });
  const reset = () => void navigate({ to: '/apprendre', search: {} });

  return (
    <VStack gap={6}>
      <header className="learn-head">
        <div className="learn-head-text">
          <Heading level={1}>{t('learn.title')}</Heading>
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

      {featured ? <FeaturedSeries key={featured.id} series={featured} /> : null}

      <section className="learn-section learn-browse" aria-labelledby={browseId}>
        {/* Keeps the outline h1 > h2 > row titles, as the prototype's labelled browse section. */}
        <VisuallyHidden as="h2" id={browseId}>
          {t('learn.browse.label')}
        </VisuallyHidden>
        <Filters key={search.q ?? ''} search={search} onChange={change} />
        <div className="learn-resultsbar">
          {/* Always mounted, so a count that appears later is announced. */}
          <p className="learn-count" role="status">
            {settled && shown > 0 ? countText(shown, hasMore, isSingular(shown, locale), t) : null}
          </p>
          {filtered && shown > 0 ? (
            <Button variant="ghost" size="sm" label={t('learn.reset')} onClick={reset} />
          ) : null}
        </div>
        {search.q && !search.type ? (
          <Text as="p" type="supporting">
            {t('learn.filter.q.note')}
          </Text>
        ) : null}
        {pending ? <LoadingState /> : null}
        {seriesFailed ? (
          <ErrorState error={series.error} onRetry={() => void series.refetch()} />
        ) : null}
        {articlesFailed ? (
          <ErrorState error={articles.error} onRetry={() => void articles.refetch()} />
        ) : null}
        {settled && shown === 0 ? (
          <EmptyState
            headingLevel={3}
            icon={<Icon icon="search" />}
            title={t(filtered ? emptyKey[search.type ?? ALL] : 'learn.empty.none.title')}
            description={
              filtered
                ? search.q
                  ? t('learn.empty.q')
                  : t('learn.empty.body')
                : t('learn.empty.none.body')
            }
            actions={
              filtered ? (
                <Button variant="secondary" label={t('learn.reset.all')} onClick={reset} />
              ) : undefined
            }
          />
        ) : null}
        {shown > 0 ? (
          <ul className="learn-list">
            {seriesItems.map((item) => (
              <SeriesRow key={item.id} series={item} />
            ))}
            {articleItems.map((item) => (
              <ArticleRow key={item.id} article={item} />
            ))}
          </ul>
        ) : null}
        {showArticles && articles.isError && articles.data ? (
          <ErrorState error={articles.error} onRetry={() => void articles.fetchNextPage()} />
        ) : null}
        {showSeries && series.hasNextPage ? (
          <div className="learn-more">
            <Button
              label={t('learn.series.more')}
              isLoading={series.isFetchingNextPage}
              onClick={() => void series.fetchNextPage()}
            />
          </div>
        ) : null}
        {showArticles && articles.hasNextPage ? (
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
