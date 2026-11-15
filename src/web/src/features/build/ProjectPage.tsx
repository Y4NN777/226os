import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
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
import { Avatar } from '../../components/Avatar';
import { CountryCode } from '../../components/CountryCode';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { EditIcon, UsersIcon } from '../../components/icons';
import { LangCode } from '../../components/LangCode';
import { LinkedItems } from '../../components/LinkedItems';
import { LoadingState } from '../../components/LoadingState';
import { PageHeader } from '../../components/PageHeader';
import { Section } from '../../components/Section';
import { StatusBadge } from '../../components/StatusBadge';
import { TagList } from '../../components/Tag';
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
import { DateText, needsKind, needsLabelKey, PersonLink, stageKey } from './present';
import { ProjectThumb } from './ProjectThumb';
import { CountryField, LicenseField, StageField } from './ProjectFields';
import { useProject, useUpdateProject, type Project, type ProjectUpdate } from './queries';
import { checkDescription, checkLicense, checkTechnologies, parseTags } from './validate';

type Visibility = components['schemas']['ProjectVisibility'];

/** The merge patch holds only what changed. Links are never part of it: they are added, not removed. */
function diffOf(
  project: Project,
  draft: {
    description: string;
    technologies: string[];
    visibility: Visibility;
    stage: Project['stage'];
    license: string;
    country: string | null;
  },
): ProjectUpdate {
  const patch: ProjectUpdate = {};
  if (draft.stage !== project.stage) patch.stage = draft.stage;
  // An emptied field clears the value (null); an untouched empty one stays out of the patch.
  const license = draft.license === '' ? null : draft.license;
  if (license !== project.license_spdx) patch.license_spdx = license;
  if (draft.country !== project.country_code) patch.country_code = draft.country;
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
    Partial<Record<'description' | 'technologies' | 'stage' | 'country_code', MessageKey>>
  >({});

  const forget = (field: keyof typeof serverErrors) => {
    setServerErrors((current) =>
      Object.fromEntries(Object.entries(current).filter(([name]) => name !== field)),
    );
  };

  const form = useForm({
    defaultValues: {
      description: project.description,
      technologies: project.technologies.join(', '),
      visibility: project.visibility,
      stage: project.stage,
      license: project.license_spdx ?? '',
      country: project.country_code,
    },
    onSubmit: async ({ value }) => {
      clear();
      setServerErrors({});
      const patch = diffOf(project, {
        description: value.description.trim(),
        technologies: parseTags(value.technologies),
        visibility: value.visibility,
        stage: value.stage,
        license: value.license.trim(),
        country: value.country,
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
            if (head === 'stage') fields.stage = 'build.err.stage';
            if (head === 'country_code') fields.country_code = 'build.err.country';
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
          <form.Field name="stage">
            {(field) => (
              <StageField
                value={field.state.value}
                error={serverErrors.stage}
                onChange={(stage) => {
                  field.handleChange(stage);
                  forget('stage');
                }}
              />
            )}
          </form.Field>
          <form.Field name="license" validators={{ onSubmit: ({ value }) => checkLicense(value) }}>
            {(field) => (
              <LicenseField
                value={field.state.value}
                error={field.state.meta.errors[0]}
                onChange={field.handleChange}
              />
            )}
          </form.Field>
          <form.Field name="country">
            {(field) => (
              <CountryField
                value={field.state.value}
                error={serverErrors.country_code}
                onChange={(country) => {
                  field.handleChange(country);
                  forget('country_code');
                }}
              />
            )}
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
    <Section title={t('project.related')}>
      <LinkedItems
        items={links.map((link) => ({
          key: `${link.type}-${link.id}`,
          kind: t(kindKey[link.type]),
          title: link.title,
          href: relatedPath(link),
          lang: link.lang,
        }))}
      />
    </Section>
  );
}

/** Side facts: repository link (github.com only), language, visibility and dates. */
function About({ project }: { project: Project }) {
  const { t } = useI18n();
  return (
    <Section as="section" title={t('project.about')} isPanel>
      <ProjectThumb name={project.name} githubUrl={project.github_url} size="lg" fallback={false} />
      {project.github_url.startsWith('https://github.com/') ? (
        <Link href={project.github_url} isExternalLink newTabLabel={t('discover.newTab')}>
          {t('project.github')}
        </Link>
      ) : null}
      <dl className="ui-facts">
        <dt>{t('project.about.lang')}</dt>
        <dd>{t(project.lang === 'en' ? 'build.lang.en' : 'build.lang.fr')}</dd>
        {project.license_spdx ? (
          <>
            <dt>{t('project.about.license')}</dt>
            <dd>{project.license_spdx}</dd>
          </>
        ) : null}
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
    </Section>
  );
}

function Maintainer({ owner }: { owner: Project['owner'] }) {
  const { t } = useI18n();
  return (
    <Section title={t('project.maintainer')} isPanel>
      <div className="build-maintainer">
        <Avatar name={owner.display_name} />
        <div className="build-maintainer-text">
          <PersonLink username={owner.username} name={owner.display_name} />
          {owner.country ? (
            <Text type="supporting">
              <CountryCode code={owner.country} withName />
            </Text>
          ) : null}
        </div>
      </div>
    </Section>
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
    <div className="ui-page">
      <PageHeader
        title={project.name}
        titleLang={project.lang}
        breadcrumbs={[
          { label: t('nav.home'), href: '/' },
          { label: t('build.title'), href: '/build' },
          { label: project.name },
        ]}
        media={<Avatar name={project.name} variant="tile" size="lg" />}
        meta={
          <>
            <Badge variant="neutral" label={t(stageKey(project.stage))} />
            {project.country_code ? (
              <Text type="supporting">
                <CountryCode code={project.country_code} withName />
              </Text>
            ) : null}
            <StatusBadge
              kind={needsKind(project.open_needs_count)}
              label={t(needsLabelKey(project.open_needs_count), { n: project.open_needs_count })}
            />
            {project.visibility === 'private' ? (
              <StatusBadge kind="private" label={t('build.private')} />
            ) : null}
          </>
        }
        actions={
          <>
            {isOwner && !editing ? (
              <IconButton
                variant="ghost"
                label={t('project.edit')}
                icon={<Icon icon={EditIcon} />}
                onClick={() => {
                  setEditing(true);
                }}
              />
            ) : null}
            {viewer === 'none' ? null : (
              <Button
                variant="primary"
                label={t('project.jump')}
                icon={<Icon icon={UsersIcon} />}
                href="#besoins"
              />
            )}
          </>
        }
      />
      <div className="ui-columns">
        <div className="ui-columns-main">
          <Related links={project.related} />
          {editing ? (
            <div className="ui-panel">
              <EditProject project={project} onDone={done} />
            </div>
          ) : project.description ? (
            <Section
              title={t('project.description')}
              isPanel
              titleAside={
                <LangCode
                  lang={project.lang}
                  title={t('project.lang.label', {
                    lang: t(project.lang === 'en' ? 'build.lang.en' : 'build.lang.fr'),
                  })}
                />
              }
            >
              <p className="build-prose" lang={project.lang}>
                {project.description}
              </p>
            </Section>
          ) : null}
          {viewer === 'none' ? null : <Needs projectId={project.id} viewer={viewer} />}
          {isOwner ? <Applications projectId={project.id} /> : null}
        </div>
        <aside className="ui-columns-aside">
          <About project={project} />
          {project.technologies.length > 0 ? (
            <Section title={t('project.technologies')} isPanel>
              <TagList tags={project.technologies} />
            </Section>
          ) : null}
          <Maintainer owner={project.owner} />
        </aside>
      </div>
    </div>
  );
}
