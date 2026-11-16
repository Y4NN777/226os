import { Button } from '@astryxdesign/core/Button';
import { HStack } from '@astryxdesign/core/HStack';
import { Tab, TabList } from '@astryxdesign/core/TabList';
import { VStack } from '@astryxdesign/core/VStack';
import { useState, type ReactNode } from 'react';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { List } from '../../components/List';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { ProjectRow } from '../build/BuildPage';
import { useOwnedProjects } from '../build/queries';
import { QuestionRow } from '../solve/QuestionRow';
import { useMemberQuestions } from '../solve/queries';

const tabs = ['questions', 'answers', 'projects'] as const;
type TabName = (typeof tabs)[number];

/** The few fields of an infinite list a panel needs, whatever the item type. */
interface Feed<T> {
  isPending: boolean;
  isError: boolean;
  error: unknown;
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  refetch: () => unknown;
  fetchNextPage: () => unknown;
  data: { pages: { items: T[]; total: number }[] } | undefined;
}

function Panel<T extends { id: string }>({
  name,
  feed,
  row,
  label,
  hidden,
}: {
  name: TabName;
  feed: Feed<T>;
  row: (item: T) => ReactNode;
  label: string;
  hidden: boolean;
}) {
  const { t } = useI18n();
  const items = feed.data?.pages.flatMap((page) => page.items) ?? [];
  return (
    <div role="tabpanel" id={`profile-panel-${name}`} aria-label={label} hidden={hidden}>
      {feed.isPending ? <LoadingState rows={2} /> : null}
      {feed.isError && !feed.data ? (
        <ErrorState error={feed.error} onRetry={() => void feed.refetch()} />
      ) : null}
      {feed.data && items.length === 0 ? (
        <EmptyState headingLevel={3} title={t(`profile.empty.${name}`)} />
      ) : null}
      {items.length > 0 ? (
        <VStack gap={4}>
          <List label={label}>{items.map((item) => row(item))}</List>
          {feed.hasNextPage ? (
            <HStack hAlign="center">
              <Button
                label={t('profile.more')}
                isLoading={feed.isFetchingNextPage}
                onClick={() => void feed.fetchNextPage()}
              />
            </HStack>
          ) : null}
        </VStack>
      ) : null}
    </div>
  );
}

/**
 * The prototype's activity tabs (profile.html): Questions asked, Réponses (questions answered) and
 * Projets (owned), each a short list with its total. A hidden or private profile answers empty
 * pages, so it reads as « Aucune … », never as an error.
 */
export function Activity({ username }: { username: string }) {
  const { t } = useI18n();
  const [selected, setSelected] = useState<TabName>('questions');
  const questions = useMemberQuestions('author', username, true);
  const answers = useMemberQuestions('answers_by', username, true);
  const projects = useOwnedProjects(username, true);
  const totals: Record<TabName, number | undefined> = {
    questions: questions.data?.pages.at(-1)?.total,
    answers: answers.data?.pages.at(-1)?.total,
    projects: projects.data?.pages.at(-1)?.total,
  };
  const labelOf = (name: TabName) => t(`profile.tab.${name}` satisfies MessageKey);

  return (
    <section className="profile-activity" aria-label={t('profile.tabs')}>
      <TabList
        role="tablist"
        aria-label={t('profile.tabs')}
        value={selected}
        onChange={(value) => {
          const next = tabs.find((name) => name === value);
          if (next) setSelected(next);
        }}
      >
        {tabs.map((name) => (
          <Tab
            key={name}
            value={name}
            label={labelOf(name)}
            panelId={`profile-panel-${name}`}
            endContent={totals[name] === undefined ? undefined : <span>{totals[name]}</span>}
          />
        ))}
      </TabList>
      <Panel
        name="questions"
        feed={questions}
        label={labelOf('questions')}
        hidden={selected !== 'questions'}
        row={(question) => <QuestionRow key={question.id} question={question} />}
      />
      <Panel
        name="answers"
        feed={answers}
        label={labelOf('answers')}
        hidden={selected !== 'answers'}
        row={(question) => <QuestionRow key={question.id} question={question} />}
      />
      <Panel
        name="projects"
        feed={projects}
        label={labelOf('projects')}
        hidden={selected !== 'projects'}
        row={(project) => <ProjectRow key={project.id} project={project} />}
      />
    </section>
  );
}
