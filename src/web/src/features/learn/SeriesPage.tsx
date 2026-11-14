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
import { AuthorLine } from './present';
import { useSeries, type TrainingSeries } from './queries';

function Detail({ series }: { series: TrainingSeries }) {
  const { t } = useI18n();
  const lessons = [...series.lessons].sort((a, b) => a.position - b.position);
  return (
    <VStack gap={5}>
      <VStack gap={2}>
        <Breadcrumbs label={t('learn.title')} variant="supporting">
          <BreadcrumbItem href="/apprendre">{t('learn.title')}</BreadcrumbItem>
          <BreadcrumbItem isCurrent>{series.title}</BreadcrumbItem>
        </Breadcrumbs>
        <div lang={series.lang}>
          <VStack gap={2}>
            <Heading level={1}>{series.title}</Heading>
            <Text as="p">{series.summary}</Text>
          </VStack>
        </div>
        <AuthorLine author={series.author} />
      </VStack>
      <VStack gap={3}>
        <Heading level={2} id="learn-lessons">
          {t('learn.series.lessonsTitle')}
        </Heading>
        <VStack as="ol" gap={3} aria-labelledby="learn-lessons">
          {lessons.map((lesson) => (
            <li key={lesson.id}>
              <HStack gap={3} wrap="wrap" vAlign="center">
                <Link
                  href={`/apprendre/articles/${lesson.article_id}`}
                  type="large"
                  weight="semibold"
                >
                  {lesson.title}
                </Link>
                {lesson.duration_minutes ? (
                  <Text type="supporting">
                    {t('learn.series.minutes', { n: lesson.duration_minutes })}
                  </Text>
                ) : null}
              </HStack>
            </li>
          ))}
        </VStack>
      </VStack>
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
