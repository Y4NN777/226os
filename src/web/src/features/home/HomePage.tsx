import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { Grid } from '@astryxdesign/core/Grid';
import { HStack } from '@astryxdesign/core/HStack';
import { Heading } from '@astryxdesign/core/Heading';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { useId, useState, type SyntheticEvent, type ReactNode } from 'react';
import { api } from '../../api/http';
import { useSession } from '../../app/session';
import { DemoBadge } from '../../components/DemoBadge';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { articleFixture, projectFixture, questionFixture, type Fixture } from './fixtures';

const SEARCH_MAX_LENGTH = 100;

const resourcesQueryKey = ['resources', 'list', { limit: 4 }] as const;

function Section({
  title,
  isDemo = false,
  children,
}: {
  title: string;
  isDemo?: boolean;
  children: ReactNode;
}) {
  const { t } = useI18n();
  const headingId = useId();
  return (
    <VStack as="section" aria-labelledby={headingId} gap={3}>
      <HStack gap={3} align="center" wrap="wrap">
        <Heading level={2} id={headingId}>
          {title}
        </Heading>
        {isDemo ? <DemoBadge /> : null}
      </HStack>
      {isDemo ? (
        <Text as="p" type="supporting">
          {t('feed.demo.note')}
        </Text>
      ) : null}
      {children}
    </VStack>
  );
}

/** A fixture entry: plain text, never a link (its page is not built yet). */
function SampleCard({
  title,
  text,
  meta,
  author,
  lang,
}: {
  title: string;
  text?: string;
  meta?: string;
  author: string;
  lang: string;
}) {
  const { t } = useI18n();
  return (
    <Card variant="muted">
      <div lang={lang}>
        <VStack gap={2}>
          <Text as="p" weight="semibold">
            {title}
          </Text>
          {text ? (
            <Text as="p" type="supporting">
              {text}
            </Text>
          ) : null}
          {meta ? (
            <Text as="p" type="supporting">
              {meta}
            </Text>
          ) : null}
          <Text as="p" type="supporting">
            {t('feed.by', { name: author })}
          </Text>
        </VStack>
      </div>
    </Card>
  );
}

function SampleGrid({ children }: { children: ReactNode }) {
  return (
    <Grid columns={{ minWidth: 240 }} gap={4}>
      {children}
    </Grid>
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

function Catalog() {
  const { t } = useI18n();
  const { data, error, isPending, refetch } = useQuery({
    queryKey: resourcesQueryKey,
    retry: false, // the error state offers an explicit retry
    queryFn: () =>
      api.call(() => api.client.GET('/resources', { params: { query: { limit: 4 } } })),
  });

  let body: ReactNode;
  if (isPending) {
    body = <LoadingState />;
  } else if (error) {
    body = <ErrorState error={error} onRetry={() => void refetch()} />;
  } else if (data.items.length === 0) {
    body = (
      <Text as="p" type="supporting">
        {t('feed.catalog.empty')}
      </Text>
    );
  } else {
    body = (
      <SampleGrid>
        {data.items.map((item) => (
          <Card key={item.id}>
            <VStack gap={2}>
              <Link href={`/discover/${item.id}`}>{item.name}</Link>
              <Text as="p" type="supporting">
                {item.summary}
              </Text>
              <Text as="p" type="supporting">
                {t('feed.catalog.provider', { name: item.provider.name })}
              </Text>
            </VStack>
          </Card>
        ))}
      </SampleGrid>
    );
  }

  return (
    <Section title={t('feed.catalog.title')}>
      {body}
      <HStack>
        <Button href="/discover" label={t('feed.catalog.more')} />
      </HStack>
    </Section>
  );
}

function fixtureItems<T>(fixture: Fixture<T>): readonly T[] {
  return fixture.items;
}

export function HomePage() {
  const { t } = useI18n();
  const session = useSession();
  const name = session.data?.user?.display_name ?? '';

  return (
    <VStack gap={8}>
      <VStack gap={4}>
        <Heading level={1}>{t('feed.hello', { name })}</Heading>
        <SearchForm />
      </VStack>

      <Section title={t('feed.learn.title')} isDemo>
        <SampleGrid>
          {fixtureItems(articleFixture).map((article) => (
            <SampleCard
              key={article.id}
              title={article.title}
              text={article.summary}
              meta={t(`feed.level.${article.level}`)}
              author={article.author.username}
              lang={article.lang}
            />
          ))}
        </SampleGrid>
      </Section>

      <Section title={t('feed.connect.title')}>
        <Text as="p" type="supporting">
          {t('feed.connect.soon')}
        </Text>
      </Section>

      <Section title={t('feed.questions.title')} isDemo>
        <SampleGrid>
          {fixtureItems(questionFixture).map((question) => (
            <SampleCard
              key={question.id}
              title={question.title}
              meta={question.tags.join(', ')}
              author={question.author.username}
              lang={question.lang}
            />
          ))}
        </SampleGrid>
      </Section>

      <Catalog />

      <Section title={t('feed.projects.title')} isDemo>
        <SampleGrid>
          {fixtureItems(projectFixture).map((project) => (
            <SampleCard
              key={project.id}
              title={project.name}
              text={project.description}
              meta={project.technologies.join(', ')}
              author={project.owner.username}
              lang={project.lang}
            />
          ))}
        </SampleGrid>
      </Section>
    </VStack>
  );
}
