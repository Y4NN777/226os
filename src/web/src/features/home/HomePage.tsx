import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { TextInput } from '@astryxdesign/core/TextInput';
import { useNavigate } from '@tanstack/react-router';
import { useId, useState, type SyntheticEvent, type ReactNode } from 'react';
import { useSession } from '../../app/session';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import {
  BookIcon,
  CompassIcon,
  HammerIcon,
  HelpIcon,
  NetworkIcon,
  type SvgIcon,
} from '../../components/icons';
import { List } from '../../components/List';
import { LoadingState } from '../../components/LoadingState';
import { PageHeader } from '../../components/PageHeader';
import { useI18n } from '../../i18n/I18nProvider';
import { useHomeArticles, useHomeProjects, useHomeQuestions, useHomeResources } from './queries';
import { ArticleRow, ProjectRow, QuestionRow, ResourceRow } from './rows';
import './home.css';

const SEARCH_MAX_LENGTH = 100;

interface LiveQuery<T> {
  data: { items: readonly T[] } | undefined;
  error: unknown;
  isPending: boolean;
  refetch: () => unknown;
}

/** One section on live data: head with its "see all" link, then loading, error, empty or rows. */
function LiveSection<T>({
  icon,
  title,
  more,
  result,
  empty,
  render,
}: {
  icon: SvgIcon;
  title: string;
  more: { href: string; label: string };
  result: LiveQuery<T>;
  empty: string;
  render: (item: T) => ReactNode;
}) {
  const headingId = useId();
  let body: ReactNode;
  if (result.isPending) {
    body = <LoadingState />;
  } else if (result.error || !result.data) {
    body = <ErrorState error={result.error} onRetry={() => void result.refetch()} />;
  } else if (result.data.items.length === 0) {
    body = <EmptyState title={empty} icon={icon} headingLevel={3} />;
  } else {
    body = <List label={title}>{result.data.items.map(render)}</List>;
  }
  return (
    <section className="feed-section" aria-labelledby={headingId}>
      <div className="feed-head">
        <Icon icon={icon} color="secondary" />
        <Heading level={2} id={headingId}>
          {title}
        </Heading>
        <span className="feed-head-more">
          <Link href={more.href}>{more.label}</Link>
        </span>
      </div>
      {body}
    </section>
  );
}

/** Connect: the prototype's lead band, one sentence and one action. */
function ConnectLead() {
  const { t } = useI18n();
  const headingId = useId();
  return (
    <section className="feed-lead" aria-labelledby={headingId}>
      <span className="feed-lead-icon" aria-hidden="true">
        <Icon icon={NetworkIcon} color="accent" size="lg" />
      </span>
      <div className="feed-lead-text">
        <Heading level={2} id={headingId}>
          {t('feed.connect.title')}
        </Heading>
        <p>{t('feed.connect.body')}</p>
      </div>
      <Button variant="primary" href="/connect" label={t('feed.connect.link')} />
    </section>
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
    <form onSubmit={submit} role="search" className="feed-search">
      <span className="feed-search-field">
        <TextInput
          type="search"
          label={t('feed.search.label')}
          isLabelHidden
          startIcon={<Icon icon="search" />}
          placeholder={t('feed.search.placeholder')}
          value={value}
          onChange={setValue}
          enterKeyHint="search"
          width="100%"
        />
      </span>
      <Button type="submit" label={t('feed.search.submit')} />
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
    <div className="feed-page">
      <header className="feed-hero">
        <PageHeader title={t('feed.hello', { name })} />
        <SearchForm />
      </header>

      <LiveSection
        icon={BookIcon}
        title={t('feed.learn.title')}
        more={{ href: '/apprendre', label: t('feed.learn.more') }}
        result={articles}
        empty={t('feed.learn.empty')}
        render={(article) => <ArticleRow key={article.id} article={article} />}
      />

      <ConnectLead />

      <LiveSection
        icon={HelpIcon}
        title={t('feed.questions.title')}
        more={{ href: '/solve', label: t('feed.questions.more') }}
        result={questions}
        empty={t('feed.questions.empty')}
        render={(question) => <QuestionRow key={question.id} question={question} />}
      />

      <LiveSection
        icon={CompassIcon}
        title={t('feed.catalog.title')}
        more={{ href: '/discover', label: t('feed.catalog.more') }}
        result={resources}
        empty={t('feed.catalog.empty')}
        render={(item) => <ResourceRow key={item.id} resource={item} />}
      />

      <LiveSection
        icon={HammerIcon}
        title={t('feed.projects.title')}
        more={{ href: '/build', label: t('feed.projects.more') }}
        result={projects}
        empty={t('feed.projects.empty')}
        render={(project) => <ProjectRow key={project.id} project={project} />}
      />
    </div>
  );
}
