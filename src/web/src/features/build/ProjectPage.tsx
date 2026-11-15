import { Avatar } from '@astryxdesign/core/Avatar';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Link } from '@astryxdesign/core/Link';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { Text } from '@astryxdesign/core/Text';
import { TextArea } from '@astryxdesign/core/TextArea';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useParams } from '@tanstack/react-router';
import { useState } from 'react';
import { ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { relatedPath } from '../related';
import type { MessageKey } from '../../i18n/messages';
import { FailureNotice } from '../identity/notices';
import { useFailure } from '../identity/useFailure';
import { errorStatus } from '../identity/validation';
import { violationName } from '../identity/violations';
import { Applications } from './Applications';
import './build.css';
import { Needs } from './Needs';
import {
  BuildBreadcrumbs,
  countryName,
  DateText,
  MetaBadge,
  Monogram,
  needsLabelKey,
  PersonLink,
  StatePill,
  Technologies,
} from './present';
import { useProject, useUpdateProject, type Project, type ProjectUpdate } from './queries';
import { checkDescription, checkTechnologies, parseTags } from './validate';

type Visibility = components['schemas']['ProjectVisibility'];

/** The merge patch holds only what changed. Links are never part of it: they are added, not removed. */
function diffOf(
  project: Project,
  draft: { description: string; technologies: string[]; visibility: Visibility },
): ProjectUpdate {
  const patch: ProjectUpdate = {};
  if (draft.description !== project.description) patch.description = draft.description;
  if (draft.technologies.join(',') !== project.technologies.join(',')) {
    patch.technologies = draft.technologies;
  }
  if (draft.visibility !== project.visibility) patch.visibility = draft.visibility;
  return patch;
}

function EditProject({ project, onDone }: { project: Project; onDone: () => void }) {
  const { t } = useI18n();
  const update = useUpdateProject(project.id);
  const { failure, report, clear } = useFailure();
  const [serverErrors, setServerErrors] = useState<
    Partial<Record<'description' | 'technologies', MessageKey>>
  >({});

  const form = useForm({
    defaultValues: {
      description: project.description,
      technologies: project.technologies.join(', '),
      visibility: project.visibility,
    },
    onSubmit: async ({ value }) => {
      clear();
      setServerErrors({});
      const patch = diffOf(project, {
        description: value.description.trim(),
        technologies: parseTags(value.technologies),
        visibility: value.visibility,
      });
      if (Object.keys(patch).length === 0) {
        onDone();
        return;
      }
      try {
        await update.mutateAsync(patch);
        onDone();
      } catch (error) {
        const fields: typeof serverErrors = {};
        if (error instanceof ApiError) {
          for (const violation of error.violations) {
            const head = violationName(violation.field).split('/')[0];
            if (head === 'description') fields.description = 'build.err.description';
            if (head === 'technologies') fields.technologies = 'build.err.technologies';
          }
        }
        if (Object.keys(fields).length > 0) setServerErrors(fields);
        else report(error);
      }
    },
  });

  return (
    <VStack gap={3}>
      {failure ? <FailureNotice key={failure.n} error={failure.error} /> : null}
      <form
        method="post"
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          void form.handleSubmit();
        }}
      >
        <VStack gap={3}>
          <form.Field
            name="description"
            validators={{ onSubmit: ({ value }) => checkDescription(value) }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.description;
              return (
                <TextArea
                  label={t('build.form.description')}
                  rows={4}
                  value={field.state.value}
                  htmlName={field.name}
                  onChange={field.handleChange}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="technologies"
            validators={{ onSubmit: ({ value }) => checkTechnologies(value) }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.technologies;
              return (
                <TextInput
                  label={t('build.form.technologies')}
                  description={t('build.form.technologies.hint')}
                  value={field.state.value}
                  htmlName={field.name}
                  autoCapitalize="none"
                  hasAutoCorrect={false}
                  hasSpellCheck={false}
                  onChange={field.handleChange}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field name="visibility">
            {(field) => (
              <SegmentedControl
                label={t('build.form.visibility')}
                value={field.state.value}
                onChange={(value) => {
                  field.handleChange(value as Visibility);
                }}
              >
                <SegmentedControlItem value="public" label={t('build.visibility.public')} />
                <SegmentedControlItem value="private" label={t('build.visibility.private')} />
              </SegmentedControl>
            )}
          </form.Field>
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <HStack gap={2}>
                <Button
                  type="submit"
                  variant="primary"
                  label={t('project.save')}
                  isLoading={isSubmitting}
                />
                <Button variant="ghost" label={t('project.cancel')} onClick={onDone} />
              </HStack>
            )}
          </form.Subscribe>
        </VStack>
      </form>
    </VStack>
  );
}

const kindKey = {
  question: 'project.kind.question',
  resource: 'project.kind.resource',
  project: 'project.kind.project',
  article: 'project.kind.article',
  series: 'project.kind.series',
} as const satisfies Record<Project['related'][number]['type'], MessageKey>;

/**
 * Linked context (prototype `.linked-items`): each card shows its kind above the title. The title is
 * the link and covers the card, so the link's name is the title alone.
 */
function Related({ links }: { links: Project['related'] }) {
  const { t } = useI18n();
  if (links.length === 0) return null;
  return (
    <section className="build-section" aria-labelledby="project-related-heading">
      <Heading level={2} id="project-related-heading">
        {t('project.related')}
      </Heading>
      <ul className="build-linked">
        {links.map((link) => (
          <li key={`${link.type}-${link.id}`} className="build-linked-item">
            <span className="build-linked-kind">{t(kindKey[link.type])}</span>
            <span className="build-linked-title" lang={link.lang}>
              <Link href={relatedPath(link)}>{link.title}</Link>
            </span>
          </li>
        ))}
      </ul>
    </section>
  );
}

/** Side facts: repository link (github.com only), language, visibility and dates. */
function About({ project }: { project: Project }) {
  const { t } = useI18n();
  return (
    <section className="build-panel build-section" aria-labelledby="project-about-heading">
      <Heading level={2} id="project-about-heading">
        {t('project.about')}
      </Heading>
      {project.github_url.startsWith('https://github.com/') ? (
        <Link href={project.github_url} isExternalLink newTabLabel={t('discover.newTab')}>
          {t('project.github')}
        </Link>
      ) : null}
      <dl className="build-facts">
        <dt>{t('project.about.lang')}</dt>
        <dd>{t(project.lang === 'en' ? 'build.lang.en' : 'build.lang.fr')}</dd>
        <dt>{t('project.about.visibility')}</dt>
        <dd>{t(project.visibility === 'private' ? 'build.private' : 'build.visibility.public')}</dd>
        <dt>{t('project.about.updated')}</dt>
        <dd>
          <DateText iso={project.updated_at} />
        </dd>
        <dt>{t('project.about.created')}</dt>
        <dd>
          <DateText iso={project.created_at} />
        </dd>
      </dl>
    </section>
  );
}

function Maintainer({ owner }: { owner: Project['owner'] }) {
  const { t, locale } = useI18n();
  return (
    <section className="build-panel build-section" aria-labelledby="project-maintainer-heading">
      <Heading level={2} id="project-maintainer-heading">
        {t('project.maintainer')}
      </Heading>
      <div className="build-maintainer">
        <span aria-hidden="true">
          <Avatar name={owner.display_name} size="md" />
        </span>
        <div className="build-maintainer-text">
          <PersonLink username={owner.username} name={owner.display_name} />
          {owner.country ? (
            <Text type="supporting">{countryName(owner.country, locale)}</Text>
          ) : null}
        </div>
      </div>
    </section>
  );
}

export function ProjectPage() {
  const { t } = useI18n();
  const { projectId } = useParams({ from: '/app/build/$projectId' });
  const session = useSession();
  const query = useProject(projectId);
  const [editing, setEditing] = useState(false);
  const user = session.data?.authenticated ? session.data.user : null;

  if (query.isPending) return <LoadingState />;
  if (query.isError) {
    if (query.error instanceof ApiError && query.error.status === 404) {
      return (
        <EmptyState
          headingLevel={1}
          title={t('project.notFound.title')}
          description={t('project.notFound.body')}
          actions={<Button label={t('project.back')} href="/build" />}
        />
      );
    }
    return <ErrorState error={query.error} onRetry={() => void query.refetch()} />;
  }

  const project = query.data;
  const isOwner = user !== null && user.id === project.owner.id;
  const viewer = isOwner ? 'owner' : user ? 'member' : session.isPending ? 'none' : 'anonymous';
  const done = () => {
    setEditing(false);
  };

  return (
    <div className="build-project">
      <BuildBreadcrumbs
        items={[
          { label: t('build.crumb.home'), href: '/' },
          { label: t('build.title'), href: '/build' },
          { label: project.name },
        ]}
      />
      <header className="build-project-head">
        <div className="build-header-row">
          <div className="build-project-id">
            <Monogram name={project.name} size="lg" />
            <div className="build-project-id-text">
              <div lang={project.lang}>
                <Heading level={1}>{project.name}</Heading>
              </div>
              <div className="build-meta">
                <StatePill
                  tone={project.open_needs_count > 0 ? 'open' : 'muted'}
                  label={t(needsLabelKey(project.open_needs_count), {
                    n: project.open_needs_count,
                  })}
                />
                {project.visibility === 'private' ? <MetaBadge label={t('build.private')} /> : null}
              </div>
            </div>
          </div>
          <div className="build-project-actions">
            {isOwner && !editing ? (
              <IconButton
                variant="ghost"
                label={t('project.edit')}
                icon={<Icon icon="wrench" />}
                onClick={() => {
                  setEditing(true);
                }}
              />
            ) : null}
            {viewer === 'none' ? null : (
              <Button variant="primary" label={t('project.jump')} href="#besoins" />
            )}
          </div>
        </div>
      </header>
      <div className="build-columns">
        <div className="build-main">
          <Related links={project.related} />
          {editing ? (
            <div className="build-panel">
              <EditProject project={project} onDone={done} />
            </div>
          ) : project.description ? (
            <section className="build-panel build-section" aria-labelledby="project-desc-heading">
              <div className="build-title-cluster">
                <Heading level={2} id="project-desc-heading">
                  {t('project.description')}
                </Heading>
                <MetaBadge
                  label={project.lang.toUpperCase()}
                  title={t('project.lang.label', {
                    lang: t(project.lang === 'en' ? 'build.lang.en' : 'build.lang.fr'),
                  })}
                />
              </div>
              <p className="build-prose" lang={project.lang}>
                {project.description}
              </p>
            </section>
          ) : null}
          {viewer === 'none' ? null : <Needs projectId={project.id} viewer={viewer} />}
          {isOwner ? <Applications projectId={project.id} /> : null}
        </div>
        <aside className="build-aside">
          <About project={project} />
          {project.technologies.length > 0 ? (
            <section className="build-panel build-section" aria-labelledby="project-tech-heading">
              <Heading level={2} id="project-tech-heading">
                {t('project.technologies')}
              </Heading>
              <Technologies items={project.technologies} />
            </section>
          ) : null}
          <Maintainer owner={project.owner} />
        </aside>
      </div>
    </div>
  );
}
