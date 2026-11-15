import { Button } from '@astryxdesign/core/Button';
import { CheckboxInput } from '@astryxdesign/core/CheckboxInput';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Link } from '@astryxdesign/core/Link';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import { useSession } from '../../app/session';
import { Avatar } from '../../components/Avatar';
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
import { DateText, needsKind, needsLabelKey } from './present';
import { ProjectForm } from './ProjectForm';
import { useCreateProject, useProjectList, type Project, type ProjectFilters } from './queries';

/** How many technology tags a row shows before it says "+n". */
const ROW_TAGS = 3;

/**
 * One dense project row (prototype `build.html`, `.list > .row`) on the shared ListRow: monogram
 * tile, title link, excerpt, technologies, a meta line, and the open-needs pill at the end. The
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
      lead={<Avatar name={project.name} variant="tile" />}
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

/** Only the filters the contract serves: free text (q) and "has open needs" (open_needs). */
function Filters({
  value,
  onChange,
}: {
  value: ProjectFilters;
  onChange: (next: ProjectFilters) => void;
}) {
  const { t } = useI18n();
  const [q, setQ] = useState(value.q ?? '');
  const active = Boolean(value.q) || Boolean(value.openNeeds);
  const apply = (next: ProjectFilters) => {
    onChange(next);
  };
  const submit = () => {
    apply({ ...value, q: q.trim() });
  };
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
        <CheckboxInput
          label={t('build.filter.openOnly')}
          value={value.openNeeds === true}
          onChange={(checked) => {
            apply({ ...value, q: q.trim(), openNeeds: checked });
          }}
        />
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

/** Drops empty values so equal filters always give the same query key. */
function normalize(filters: ProjectFilters): ProjectFilters {
  return {
    ...(filters.q ? { q: filters.q } : {}),
    ...(filters.openNeeds ? { openNeeds: true } : {}),
  };
}

export function BuildPage() {
  const { t } = useI18n();
  const session = useSession();
  const [proposing, setProposing] = useState(false);
  const [filters, setFilters] = useState<ProjectFilters>({});
  // The search field keeps its own draft; a reset from the empty state remounts it cleared.
  const [filtersKey, setFiltersKey] = useState(0);
  const list = useProjectList(filters);
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];
  const signedIn = session.data?.authenticated === true;
  const filtered = filters.q !== undefined || filters.openNeeds !== undefined;
  const clearFilters = () => {
    setFilters({});
    setFiltersKey((n) => n + 1);
  };

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
      <Filters
        key={filtersKey}
        value={filters}
        onChange={(next) => {
          setFilters(normalize(next));
        }}
      />
      {/* The prototype shows « Projets » above the list, as Connect shows « Membres ». */}
      <Section title={t('build.results.label')} titleId="build-results-heading">
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
