import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { Grid } from '@astryxdesign/core/Grid';
import { HStack } from '@astryxdesign/core/HStack';
import { Heading } from '@astryxdesign/core/Heading';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate } from '@tanstack/react-router';
import { useId, useState, type SyntheticEvent, type ReactNode } from 'react';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { useHomeArticles, useHomeProjects, useHomeQuestions, useHomeResources } from './queries';

const SEARCH_MAX_LENGTH = 100;

function Section({ title, children }: { title: string; children: ReactNode }) {
  const headingId = useId();
  return (
    <VStack as="section" aria-labelledby={headingId} gap={3}>
      <Heading level={2} id={headingId}>
        {title}
      </Heading>
      {children}
    </VStack>
  );
}

interface LiveQuery<T> {
  data: { items: readonly T[] } | undefined;
  error: unknown;
  isPending: boolean;
  refetch: () => unknown;
}

/** One section on live data: loading, error with its own retry, empty, then the entries. */
function LiveSection<T>({
  title,
  result,
  empty,
  more,
  render,
}: {
  title: string;
  result: LiveQuery<T>;
  empty: string;
  more?: ReactNode;
  render: (item: T) => ReactNode;
}) {
  let body: ReactNode;
  if (result.isPending) {
    body = <LoadingState />;
  } else if (result.error || !result.data) {
    body = <ErrorState error={result.error} onRetry={() => void result.refetch()} />;
  } else if (result.data.items.length === 0) {
    body = (
      <Text as="p" type="supporting">
        {empty}
      </Text>
    );
  } else {
    body = (
      <Grid columns={{ minWidth: 240 }} gap={4}>
        {result.data.items.map(render)}
      </Grid>
    );
  }
  return (
    <Section title={title}>
      {body}
      {more}
    </Section>
  );
}

function EntryCard({
  href,
  title,
  lines,
  lang,
}: {
  href: string;
  title: string;
  lines: readonly (string | undefined)[];
  lang?: string;
}) {
  return (
    <Card>
      <div {...(lang ? { lang } : {})}>
        <VStack gap={2}>
          <Link href={href}>{title}</Link>
          {lines.map((line) =>
            line ? (
              <Text key={line} as="p" type="supporting">
                {line}
              </Text>
            ) : null,
          )}
        </VStack>
      </div>
    </Card>
  );
}

function SearchForm() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [value, setValue] = useState('');

  function submit(event: SyntheticEvent) {
    event.preventDefault();
    const q = value.trim().slice(0, SEARCH_MAX_LENGTH);
    if (q) void navigate({ to: '/recherche', search: { q } });
  }

  return (
    <form onSubmit={submit} role="search">
      <HStack gap={2} align="end" wrap="wrap">
        <TextInput
          type="search"
          label={t('feed.search.label')}
          value={value}
          onChange={setValue}
          enterKeyHint="search"
          width="min(100%, 28rem)"
        />
        <Button type="submit" label={t('feed.search.submit')} />
      </HStack>
    </form>
  );
}

export function HomePage() {
  const { t } = useI18n();
  const session = useSession();
  const name = session.data?.user?.display_name ?? '';
  const articles = useHomeArticles();
  const questions = useHomeQuestions();
  const resources = useHomeResources();
  const projects = useHomeProjects();

  return (
    <VStack gap={8}>
      <VStack gap={4}>
        <Heading level={1}>{t('feed.hello', { name })}</Heading>
        <SearchForm />
      </VStack>

      <LiveSection
        title={t('feed.learn.title')}
        result={articles}
        empty={t('feed.learn.empty')}
        render={(article) => (
          <EntryCard
            key={article.id}
            href={`/apprendre/articles/${article.id}`}
            title={article.title}
            lang={article.lang}
            lines={[
              article.summary,
              t(`feed.level.${article.level}`),
              t('feed.by', { name: article.author.display_name }),
            ]}
          />
        )}
      />

      <Section title={t('feed.connect.title')}>
        <Text as="p" type="supporting">
          {t('feed.connect.soon')}
        </Text>
      </Section>

      <LiveSection
        title={t('feed.questions.title')}
        result={questions}
        empty={t('feed.questions.empty')}
        render={(question) => (
          <EntryCard
            key={question.id}
            href={`/questions/${question.id}`}
            title={question.title}
            lang={question.lang}
            lines={[question.tags.join(', '), t('feed.by', { name: question.author.display_name })]}
          />
        )}
      />

      <LiveSection
        title={t('feed.catalog.title')}
        result={resources}
        empty={t('feed.catalog.empty')}
        more={
          <HStack>
            <Button href="/discover" label={t('feed.catalog.more')} />
          </HStack>
        }
        render={(item) => (
          <EntryCard
            key={item.id}
            href={`/discover/${item.id}`}
            title={item.name}
            lines={[item.summary, t('feed.catalog.provider', { name: item.provider.name })]}
          />
        )}
      />

      <LiveSection
        title={t('feed.projects.title')}
        result={projects}
        empty={t('feed.projects.empty')}
        render={(project) => (
          <EntryCard
            key={project.id}
            href={`/build/${project.id}`}
            title={project.name}
            lang={project.lang}
            lines={[
              project.description,
              project.technologies.join(', '),
              t('feed.by', { name: project.owner.display_name }),
            ]}
          />
        )}
      />
    </VStack>
  );
}
