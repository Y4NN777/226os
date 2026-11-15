import { Button } from '@astryxdesign/core/Button';
import { CheckboxInput } from '@astryxdesign/core/CheckboxInput';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { relatedPath } from '../related';
import { signInHref } from '../solve/signIn';
import './build.css';
import {
  BuildBreadcrumbs,
  DateText,
  MetaBadge,
  Monogram,
  needsLabelKey,
  StatePill,
} from './present';
import { ProjectForm } from './ProjectForm';
import { useCreateProject, useProjectList, type Project, type ProjectFilters } from './queries';

/** How many technology tags a row shows before it says "+n". */
const ROW_TAGS = 3;

/**
 * One dense project row (prototype `build.html`, `.list > .row`): monogram, title link, excerpt,
 * technologies, a meta line, and the open-needs pill at the end. Only the title is a link, so the
 * owner and origin-question links keep their own names.
 */
export function ProjectRow({ project }: { project: Project }) {
  const { t } = useI18n();
  const shown = project.technologies.slice(0, ROW_TAGS);
  const more = project.technologies.length - shown.length;
  const origin = project.related.find((link) => link.type === 'question');
  return (
    <li className="build-row">
      <Monogram name={project.name} size="lg" />
      <div className="build-row-main">
        <h3 className="build-row-title">
          <span lang={project.lang}>
            <Link href={`/build/${project.id}`} weight="semibold" color="primary" size="lg">
              {project.name}
            </Link>
          </span>
        </h3>
        {project.description ? (
          <p className="build-row-excerpt" lang={project.lang}>
            {project.description}
          </p>
        ) : null}
        {shown.length > 0 ? (
          <ul className="build-tags">
            {shown.map((item) => (
              <li key={item} className="build-tag">
                {item}
              </li>
            ))}
            {more > 0 ? (
              <li className="build-tag-more">
                <span aria-hidden="true">+{more}</span>
                <span className="build-visually-hidden">
                  {t('build.row.moreTags', { n: more })}
                </span>
              </li>
            ) : null}
          </ul>
        ) : null}
        <div className="build-meta">
          <span className="build-meta-item">
            <Link href={`/profils/${project.owner.username}`} size="sm">
              {project.owner.display_name}
            </Link>
          </span>
          <span className="build-meta-item">
            <span className="build-visually-hidden">{t('build.row.updated')} </span>
            <DateText iso={project.updated_at} />
          </span>
          {project.visibility === 'private' ? <MetaBadge label={t('build.private')} /> : null}
          {origin ? (
            <span className="build-meta-item build-meta-origin">
              <Icon icon="info" size="xsm" color="secondary" />
              <span className="build-visually-hidden">{t('build.row.origin')} </span>
              <span lang={origin.lang}>
                <Link href={relatedPath(origin)} size="sm" color="secondary" hasUnderline>
                  {origin.title}
                </Link>
              </span>
            </span>
          ) : null}
        </div>
      </div>
      <div className="build-row-aside">
        <StatePill
          tone={project.open_needs_count > 0 ? 'open' : 'muted'}
          label={t(needsLabelKey(project.open_needs_count), { n: project.open_needs_count })}
        />
      </div>
    </li>
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
    <div className="build-page">
      <BuildBreadcrumbs
        items={[{ label: t('build.crumb.home'), href: '/' }, { label: t('build.title') }]}
      />
      <header className="build-header">
        <div className="build-header-row">
          <Heading level={1}>{t('build.title')}</Heading>
          {proposing && signedIn ? null : signedIn ? (
            <Button
              variant="primary"
              label={t('build.propose')}
              onClick={() => {
                setProposing(true);
              }}
            />
          ) : session.isPending ? null : (
            <Button
              variant="primary"
              label={t('build.propose.signIn')}
              href={signInHref('/build')}
            />
          )}
        </div>
        <Text as="p" color="secondary">
          {t('build.lede')}
        </Text>
      </header>
      {proposing && signedIn ? (
        <div className="build-panel">
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
      <section aria-labelledby="build-results-heading" className="build-results">
        <div className="build-results-head">
          <Heading level={2} id="build-results-heading">
            {t('build.results.label')}
          </Heading>
        </div>
        {list.isPending ? <LoadingState /> : null}
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
            <ul className="build-list">
              {items.map((project) => (
                <ProjectRow key={project.id} project={project} />
              ))}
            </ul>
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
      </section>
    </div>
  );
}
