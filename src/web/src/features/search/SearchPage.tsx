import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { TextInput } from '@astryxdesign/core/TextInput';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import type { components } from '../../api/schema';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FeedIcon, type FeedIconName } from '../home/FeedIcon';
import { RowList, Thumb } from '../home/rows';
import { AssistantPanel } from './AssistantPanel';
import { isSearchable, normalizeQuery, searchQuery } from './searchQuery';
import './search.css';

type SearchResults = components['schemas']['SearchResults'];
type SearchHit = components['schemas']['SearchHit'];
type GroupName = keyof SearchResults['groups'];

const groupOrder = [
  ['resources', 'search.group.resources', 'compass'],
  ['questions', 'search.group.questions', 'help'],
  ['learn', 'search.group.learn', 'book'],
  ['projects', 'search.group.projects', 'hammer'],
] as const satisfies readonly (readonly [GroupName, MessageKey, FeedIconName])[];

function SearchForm({ initial }: { initial: string }) {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [value, setValue] = useState(initial);
  const submit = () => {
    const q = normalizeQuery(value);
    void navigate({ to: '/recherche', search: q ? { q } : {} });
  };
  return (
    <form
      role="search"
      className="srch-form"
      onSubmit={(event) => {
        event.preventDefault();
        submit();
      }}
    >
      <span className="srch-field">
        <TextInput
          type="search"
          label={t('search.label')}
          isLabelHidden
          startIcon={<Icon icon="search" />}
          size="lg"
          value={value}
          onChange={setValue}
          htmlName="q"
          enterKeyHint="search"
          autoComplete="off"
          width="100%"
        />
      </span>
      {/* The field takes the free width; the button keeps its full label. */}
      <span className="no-shrink">
        <Button type="submit" variant="primary" size="lg" label={t('search.submit')} />
      </span>
    </form>
  );
}

const pathByGroup: Record<GroupName, (id: string) => string> = {
  resources: (id) => `/discover/${id}`,
  questions: (id) => `/questions/${id}`,
  learn: (id) => `/apprendre/articles/${id}`,
  projects: (id) => `/build/${id}`,
};

/** One hit as a dense row: the group glyph, the title link, the plain-text snippet. */
function Hit({ hit, group, icon }: { hit: SearchHit; group: GroupName; icon: FeedIconName }) {
  const { locale } = useI18n();
  const lang = hit.lang && hit.lang !== locale ? hit.lang : undefined;
  return (
    <li className="feed-row" lang={lang}>
      <div className="feed-row-lead">
        <Thumb icon={icon} />
      </div>
      <div className="feed-row-main">
        <h3 className="feed-row-title">
          <Link href={pathByGroup[group](hit.id)} color="primary" weight="semibold">
            {hit.title}
          </Link>
        </h3>
        {hit.snippet ? <p className="feed-row-excerpt">{hit.snippet}</p> : null}
      </div>
    </li>
  );
}

function Results({ q, data }: { q: string; data: SearchResults }) {
  const { t } = useI18n();
  const groups = groupOrder.flatMap(([name, label, icon]) => {
    const group = data.groups[name];
    return group && group.items.length > 0 ? [{ name, label, icon, group }] : [];
  });
  if (groups.length === 0) {
    return <EmptyState title={t('search.none.title', { q })} description={t('search.none.body')} />;
  }
  return (
    <div className="srch-groups">
      {groups.map(({ name, label, icon, group }) => (
        <section key={name} className="feed-section" aria-labelledby={`search-group-${name}`}>
          <div className="feed-head">
            <FeedIcon name={icon} />
            <Heading level={2} id={`search-group-${name}`}>
              {t(label)}
            </Heading>
            {/* A count only when it is the whole group: with more results it would be false. */}
            {group.has_more ? null : (
              <span className="srch-count">{String(group.items.length)}</span>
            )}
          </div>
          <RowList>
            {group.items.map((hit) => (
              <Hit key={hit.id} hit={hit} group={name} icon={icon} />
            ))}
          </RowList>
          {group.has_more ? <p className="srch-more">{t('search.more')}</p> : null}
        </section>
      ))}
    </div>
  );
}

function Outcome({ q }: { q: string }) {
  const query = useQuery(searchQuery(q));
  if (query.isPending) return <LoadingState />;
  if (query.isError) {
    return (
      <ErrorState
        error={query.error}
        title="search.error.title"
        onRetry={() => void query.refetch()}
      />
    );
  }
  return <Results q={q} data={query.data} />;
}

export function SearchPage() {
  const { t } = useI18n();
  const { q: rawQ } = useSearch({ strict: false });
  const q = normalizeQuery(rawQ ?? '');
  const searchable = q !== '' && isSearchable(q);
  return (
    <div className="srch-page">
      <Heading level={1}>
        {searchable ? t('search.results.title', { q }) : t('search.title')}
      </Heading>
      <SearchForm key={q} initial={q} />
      {q === '' ? (
        <EmptyState title={t('search.start.title')} description={t('search.start.body')} />
      ) : !searchable ? (
        <EmptyState title={t('search.short.title')} description={t('search.short.body')} />
      ) : (
        <>
          <Outcome q={q} />
          <div className="srch-assistant">
            <AssistantPanel key={q} q={q} />
          </div>
        </>
      )}
    </div>
  );
}
