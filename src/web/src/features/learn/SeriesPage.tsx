import { Breadcrumbs, BreadcrumbItem } from '@astryxdesign/core/Breadcrumbs';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { VStack } from '@astryxdesign/core/VStack';
import { useParams } from '@tanstack/react-router';
import { ApiError } from '../../api/http';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { ArrowRightIcon, BookIcon, HelpIcon } from './icons';
import './learn.css';
import { AuthorLine, formatMinutes, KindBadge, LangCode } from './present';
import { useSeries, type TrainingSeries } from './queries';

type Lesson = TrainingSeries['lessons'][number];

/** The course length, only when every lesson says how long it takes: never a partial sum. */
function totalMinutes(lessons: readonly Lesson[]): number | null {
  if (lessons.length === 0) return null;
  let total = 0;
  for (const lesson of lessons) {
    if (!lesson.duration_minutes) return null;
    total += lesson.duration_minutes;
  }
  return total;
}

function Detail({ series }: { series: TrainingSeries }) {
  const { t, locale } = useI18n();
  const lessons = [...series.lessons].sort((a, b) => a.position - b.position);
  const firstLesson = lessons[0];
  const total = totalMinutes(lessons);
  const count =
    lessons.length === 1 ? t('learn.lessons.one') : t('learn.lessons.many', { n: lessons.length });
  return (
    <VStack gap={6}>
      <Breadcrumbs label={t('learn.title')} variant="supporting">
        <BreadcrumbItem href="/apprendre">{t('learn.title')}</BreadcrumbItem>
        <BreadcrumbItem isCurrent>{series.title}</BreadcrumbItem>
      </Breadcrumbs>
      <header className="learn-course-hero">
        <KindBadge kind="series" />
        <div lang={series.lang}>
          <VStack gap={3}>
            <h1 className="learn-course-title">{series.title}</h1>
            <p className="learn-feature-sum">{series.summary}</p>
          </VStack>
        </div>
        <p className="learn-meta">
          <span className="learn-meta-item">
            <Icon icon={BookIcon} size="sm" />
            {count}
          </span>
          {total !== null ? (
            <span className="learn-meta-item">
              <Icon icon="clock" size="sm" />
              {formatMinutes(total)}
            </span>
          ) : null}
          <span className="learn-meta-item">{t(`learn.level.${series.level}`)}</span>
          <LangCode lang={series.lang} />
        </p>
        <AuthorLine author={series.author} />
        {firstLesson ? (
          <div className="learn-actions" lang={locale}>
            <Button
              variant="primary"
              label={t('learn.feature.start')}
              href={`/apprendre/articles/${firstLesson.article_id}`}
              endContent={<Icon icon={ArrowRightIcon} />}
            />
            {/* A question can only link to an article (QuestionLinkRef has no series), so the
                overview points at its first lesson and says so. */}
            <Button
              variant="secondary"
              label={t('learn.series.askFirst')}
              href={`/solve?link=article:${firstLesson.article_id}`}
              icon={<Icon icon={HelpIcon} />}
            />
          </div>
        ) : null}
      </header>
      <section className="learn-section">
        <h2 className="learn-section-title" id="learn-lessons">
          {t('learn.series.lessonsTitle')}
        </h2>
        <ol className="learn-lessons" aria-labelledby="learn-lessons">
          {lessons.map((lesson, index) => (
            <li key={lesson.id} className="learn-lesson">
              <span className="learn-step-n" aria-hidden="true">
                {index + 1}
              </span>
              <div className="learn-lesson-body">
                <h3 lang={series.lang}>
                  <Link
                    href={`/apprendre/articles/${lesson.article_id}`}
                    type="inherit"
                    color="inherit"
                  >
                    {lesson.title}
                  </Link>
                </h3>
                {lesson.duration_minutes ? (
                  <p className="learn-meta">
                    <span className="learn-meta-item">
                      <Icon icon="clock" size="sm" />
                      {formatMinutes(lesson.duration_minutes)}
                    </span>
                  </p>
                ) : null}
              </div>
            </li>
          ))}
        </ol>
      </section>
    </VStack>
  );
}

export function SeriesPage() {
  const { t } = useI18n();
  const { seriesId } = useParams({ from: '/app/apprendre/formations/$seriesId' });
  const series = useSeries(seriesId);

  if (series.isPending) return <LoadingState />;
  if (series.isError) {
    if (series.error instanceof ApiError && series.error.status === 404) {
      return (
        <EmptyState
          headingLevel={1}
          title={t('learn.series.notFound.title')}
          description={t('learn.series.notFound.body')}
          actions={<Button label={t('learn.series.back')} href="/apprendre" />}
        />
      );
    }
    return <ErrorState error={series.error} onRetry={() => void series.refetch()} />;
  }
  return <Detail series={series.data} />;
}
