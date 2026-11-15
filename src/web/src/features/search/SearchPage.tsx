import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { TextInput } from '@astryxdesign/core/TextInput';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import type { components } from '../../api/schema';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import {
  BookIcon,
  CompassIcon,
  HammerIcon,
  HelpIcon,
  SearchIcon,
  type SvgIcon,
} from '../../components/icons';
import { List, ListRow } from '../../components/List';
import { LoadingState } from '../../components/LoadingState';
import { PageHeader } from '../../components/PageHeader';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { AssistantPanel } from './AssistantPanel';
import { isSearchable, normalizeQuery, searchQuery } from './searchQuery';
import './search.css';

type SearchResults = components['schemas']['SearchResults'];
type SearchHit = components['schemas']['SearchHit'];
type GroupName = keyof SearchResults['groups'];

const groupOrder = [
  ['resources', 'search.group.resources', CompassIcon],
  ['questions', 'search.group.questions', HelpIcon],
  ['learn', 'search.group.learn', BookIcon],
  ['projects', 'search.group.projects', HammerIcon],
] as const satisfies readonly (readonly [GroupName, MessageKey, SvgIcon])[];

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

/** One hit as a shared dense row: the title link and the plain-text snippet. */
function Hit({ hit, group }: { hit: SearchHit; group: GroupName }) {
  const { locale } = useI18n();
  const lang = hit.lang && hit.lang !== locale ? hit.lang : undefined;
  return (
    <ListRow
      title={hit.title}
      href={pathByGroup[group](hit.id)}
      lang={lang}
      excerpt={hit.snippet || undefined}
    />
  );
}

function Results({ q, data }: { q: string; data: SearchResults }) {
  const { t } = useI18n();
  const groups = groupOrder.flatMap(([name, label, icon]) => {
    const group = data.groups[name];
    return group && group.items.length > 0 ? [{ name, label, icon, group }] : [];
  });
  if (groups.length === 0) {
    return (
      <EmptyState
        icon={SearchIcon}
        title={t('search.none.title', { q })}
        description={t('search.none.body')}
      />
    );
  }
  return (
    <div className="srch-groups">
      {groups.map(({ name, label, icon, group }) => (
        <section key={name} className="srch-section" aria-labelledby={`search-group-${name}`}>
          <div className="srch-head">
            <Icon icon={icon} color="secondary" />
            <Heading level={2} id={`search-group-${name}`}>
              {t(label)}
            </Heading>
            {/* A count only when it is the whole group: with more results it would be false. */}
            {group.has_more ? null : (
              <span className="srch-count">{String(group.items.length)}</span>
            )}
          </div>
          <List label={t(label)}>
            {group.items.map((hit) => (
              <Hit key={hit.id} hit={hit} group={name} />
            ))}
          </List>
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
      <PageHeader title={searchable ? t('search.results.title', { q }) : t('search.title')} />
      <SearchForm key={q} initial={q} />
      {q === '' ? (
        <EmptyState
          icon={SearchIcon}
          title={t('search.start.title')}
          description={t('search.start.body')}
        />
      ) : !searchable ? (
        <EmptyState
          icon={SearchIcon}
          title={t('search.short.title')}
          description={t('search.short.body')}
        />
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
