import { Button } from '@astryxdesign/core/Button';
import { CheckboxInput } from '@astryxdesign/core/CheckboxInput';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Link } from '@astryxdesign/core/Link';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { Selector } from '@astryxdesign/core/Selector';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import { useSession } from '../../app/session';
import { countryOptions } from '../../components/countries';
import { CountryCode } from '../../components/CountryCode';
import { ResultsMeta } from '../../components/ResultsMeta';
import { SortSelect } from '../../components/SortSelect';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { LinkIcon, PlusIcon } from '../../components/icons';
import { List, ListRow } from '../../components/List';
import { LoadingState } from '../../components/LoadingState';
import { Meta, MetaItem } from '../../components/Meta';
import { PageHeader } from '../../components/PageHeader';
import { Section } from '../../components/Section';
import { StatusBadge } from '../../components/StatusBadge';
import { TagList } from '../../components/Tag';
import { useI18n } from '../../i18n/I18nProvider';
import { relatedPath } from '../related';
import { signInHref } from '../solve/signIn';
import './build.css';
import { DateText, needsKind, needsLabelKey, stageKey } from './present';
import { ProjectForm } from './ProjectForm';
import { ProjectThumb } from './ProjectThumb';
import { useCreateProject, useProjectList, type Project, type ProjectFilters } from './queries';
import { projectSorts, projectStages, validateBuildSearch, type BuildSearch } from './routeSearch';

/** How many technology tags a row shows before it says "+n". */
const ROW_TAGS = 3;

/**
 * One dense project row (prototype `build.html`, `.list > .row`) on the shared ListRow: the
 * repository preview (monogram tile when there is none), title link, excerpt, technologies, a meta line, and the open-needs pill at the end. The
 * owner and origin-question links sit above the row link and keep their own names.
 */
export function ProjectRow({ project }: { project: Project }) {
  const { t } = useI18n();
  const shown = project.technologies.slice(0, ROW_TAGS);
  const more = project.technologies.length - shown.length;
  const origin = project.related.find((link) => link.type === 'question');
  return (
    <ListRow
      title={project.name}
      href={`/build/${project.id}`}
      titleLang={project.lang}
      lead={<ProjectThumb name={project.name} githubUrl={project.github_url} />}
      {...(project.description ? { excerpt: project.description } : {})}
      {...(shown.length > 0
        ? {
            tags: (
              <TagList
                tags={shown}
                more={{ count: more, label: t('build.row.moreTags', { n: more }) }}
              />
            ),
          }
        : {})}
      meta={
        <Meta>
          <MetaItem>
            <Link href={`/profils/${project.owner.username}`}>{project.owner.display_name}</Link>
          </MetaItem>
          <MetaItem>{t(stageKey(project.stage))}</MetaItem>
          {project.country_code ? (
            <MetaItem>
              <CountryCode code={project.country_code} withName />
            </MetaItem>
          ) : null}
          <MetaItem>
            <VisuallyHidden>{`${t('build.row.updated')} `}</VisuallyHidden>
            <DateText iso={project.updated_at} />
          </MetaItem>
          {origin ? (
            <MetaItem icon={LinkIcon}>
              <VisuallyHidden>{`${t('build.row.origin')} `}</VisuallyHidden>
              <span lang={origin.lang} className="build-origin">
                <Link href={relatedPath(origin)}>{origin.title}</Link>
              </span>
            </MetaItem>
          ) : null}
        </Meta>
      }
      aside={
        <>
          <StatusBadge
            kind={needsKind(project.open_needs_count)}
            label={t(needsLabelKey(project.open_needs_count), { n: project.open_needs_count })}
          />
          {project.visibility === 'private' ? (
            <StatusBadge kind="private" label={t('build.private')} />
          ) : null}
        </>
      }
    />
  );
}

function ProposeForm({ onClose }: { onClose: () => void }) {
  const navigate = useNavigate();
  const create = useCreateProject();
  return (
    <ProjectForm
      onCancel={onClose}
      onSubmit={async (body) => {
        const created = await create.mutateAsync(body);
        await navigate({ to: '/build/$projectId', params: { projectId: created.id } });
      }}
    />
  );
}

/** A copy of the search without one key, so a cleared filter leaves the address. */
function without(search: BuildSearch, key: keyof BuildSearch): BuildSearch {
  return Object.fromEntries(Object.entries(search).filter(([name]) => name !== key));
}

/** The filters the contract serves (q, open_needs, stage, country), all held in the address. */
function Filters({
  value,
  seenCountries,
  onChange,
}: {
  value: ProjectFilters;
  seenCountries: readonly string[];
  onChange: (next: ProjectFilters) => void;
}) {
  const { t, locale } = useI18n();
  const [q, setQ] = useState(value.q ?? '');
  const active = Object.keys(value).length > 0;
  const apply = (next: ProjectFilters) => {
    onChange(next);
  };
  const submit = () => {
    const text = q.trim();
    apply(text ? { ...value, q: text } : without(value, 'q'));
  };
  // The raw value goes through the same validation as the address, so only a known one stays.
  const withKey = (key: 'stage' | 'country', picked: string | null): BuildSearch =>
    picked ? validateBuildSearch({ ...value, [key]: picked }) : without(value, key);
  return (
    <form
      role="search"
      aria-label={t('build.filters.label')}
      className="build-filters"
      onSubmit={(event) => {
        event.preventDefault();
        submit();
      }}
    >
      <div className="build-filters-search">
        <TextInput
          type="search"
          label={t('build.search.label')}
          isLabelHidden
          placeholder={t('build.search.placeholder')}
          startIcon={<Icon icon="search" />}
          value={q}
          onChange={setQ}
          onEnter={submit}
          autoComplete="off"
          width="100%"
          size="lg"
        />
      </div>
      <div className="build-filters-bar">
        <HStack gap={3} vAlign="center" wrap="wrap">
          <Selector
            label={t('build.filter.stage')}
            placeholder={t('build.filter.stage.any')}
            hasClear
            size="sm"
            options={projectStages.map((stage) => ({ value: stage, label: t(stageKey(stage)) }))}
            value={value.stage ?? null}
            onChange={(picked) => {
              apply(withKey('stage', picked));
            }}
          />
          <Selector
            label={t('build.filter.country')}
            placeholder={t('build.filter.country.any')}
            hasClear
            hasSearch
            size="sm"
            options={countryOptions(
              [...seenCountries, ...(value.country ? [value.country] : [])],
              locale,
            )}
            value={value.country ?? null}
            onChange={(picked) => {
              apply(withKey('country', picked));
            }}
          />
          <CheckboxInput
            label={t('build.filter.openOnly')}
            value={value.open === true}
            onChange={(checked) => {
              apply(checked ? { ...value, open: true } : without(value, 'open'));
            }}
          />
        </HStack>
        {active ? (
          <IconButton
            variant="ghost"
            label={t('build.filter.resetIcon')}
            icon={<Icon icon="close" />}
            onClick={() => {
              setQ('');
              apply({});
            }}
          />
        ) : null}
      </div>
    </form>
  );
}

export function BuildPage() {
  const { t } = useI18n();
  const session = useSession();
  const navigate = useNavigate();
  // Router merges the raw parent search over the validated one: validate again here.
  const filters = validateBuildSearch(useSearch({ from: '/app/build' }));
  const [proposing, setProposing] = useState(false);
  const list = useProjectList(filters);
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];
  const total = list.data?.pages.at(-1)?.total;
  const signedIn = session.data?.authenticated === true;
  const filtered =
    filters.q !== undefined ||
    filters.open !== undefined ||
    filters.stage !== undefined ||
    filters.country !== undefined;
  const go = (next: BuildSearch) => {
    void navigate({ to: '/build', search: next });
  };
  const clearFilters = () => {
    go({});
  };
  const seenCountries = items.flatMap((project) =>
    project.country_code ? [project.country_code] : [],
  );
  const sortOptions = projectSorts.map((sort) => ({ value: sort, label: t(`build.sort.${sort}`) }));

  return (
    <div className="ui-page">
      <PageHeader
        title={t('build.title')}
        breadcrumbs={[{ label: t('nav.home'), href: '/' }, { label: t('build.title') }]}
        lead={t('build.lede')}
        actions={
          proposing && signedIn ? undefined : signedIn ? (
            <Button
              variant="primary"
              label={t('build.propose')}
              icon={<Icon icon={PlusIcon} />}
              onClick={() => {
                setProposing(true);
              }}
            />
          ) : session.isPending ? undefined : (
            <Button
              variant="primary"
              label={t('build.propose.signIn')}
              href={signInHref('/build')}
            />
          )
        }
      />
      {proposing && signedIn ? (
        <div className="ui-panel">
          <ProposeForm
            onClose={() => {
              setProposing(false);
            }}
          />
        </div>
      ) : null}
      <Filters key={filters.q ?? ''} value={filters} seenCountries={seenCountries} onChange={go} />
      {/* The prototype shows « Projets » above the list, as Connect shows « Membres ». */}
      <Section title={t('build.results.label')} titleId="build-results-heading">
        {/* Always mounted, so a count that appears later is announced. */}
        <ResultsMeta
          summary={
            total === undefined || items.length === 0
              ? null
              : t(total === 1 ? 'build.count.one' : 'build.count.many', { n: total })
          }
        >
          <SortSelect
            label={t('build.sort.label')}
            options={sortOptions}
            value={filters.sort ?? 'updated'}
            onChange={(sort) => {
              go(sort === 'updated' ? without(filters, 'sort') : { ...filters, sort });
            }}
          />
        </ResultsMeta>
        {list.isPending ? <LoadingState rows={4} /> : null}
        {list.isError && !list.data ? (
          <ErrorState error={list.error} onRetry={() => void list.refetch()} />
        ) : null}
        {list.data && items.length === 0 ? (
          filtered ? (
            <EmptyState
              headingLevel={3}
              title={t('build.filtered.title')}
              description={t('build.filtered.body')}
              actions={<Button label={t('build.filter.reset')} onClick={clearFilters} />}
            />
          ) : (
            <EmptyState
              headingLevel={3}
              title={t('build.empty.title')}
              description={t('build.empty.body')}
            />
          )
        ) : null}
        {items.length > 0 ? (
          <>
            <List label={t('build.results.label')}>
              {items.map((project) => (
                <ProjectRow key={project.id} project={project} />
              ))}
            </List>
            {list.isError ? (
              <ErrorState error={list.error} onRetry={() => void list.fetchNextPage()} />
            ) : null}
            {list.hasNextPage ? (
              <HStack hAlign="center">
                <Button
                  label={t('build.more')}
                  isLoading={list.isFetchingNextPage}
                  onClick={() => void list.fetchNextPage()}
                />
              </HStack>
            ) : null}
          </>
        ) : null}
      </Section>
    </div>
  );
}
