import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
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
import type { MessageKey } from '../../i18n/messages';
import { FailureNotice } from '../identity/notices';
import { useFailure } from '../identity/useFailure';
import { errorStatus } from '../identity/validation';
import { violationName } from '../identity/violations';
import { Applications } from './Applications';
import { Needs } from './Needs';
import { OwnerByline, Technologies } from './present';
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

function Related({ links }: { links: Project['related'] }) {
  const { t } = useI18n();
  if (links.length === 0) return null;
  const hrefOf = (link: Project['related'][number]) =>
    link.type === 'resource'
      ? `/discover/${link.id}`
      : link.type === 'question'
        ? `/questions/${link.id}`
        : link.type === 'project'
          ? `/build/${link.id}`
          : undefined;
  return (
    <VStack gap={1}>
      <Text type="label" color="secondary">
        {t('project.related')}
      </Text>
      <VStack as="ul" gap={1}>
        {links.map((link) => {
          const href = hrefOf(link);
          return (
            <li key={`${link.type}-${link.id}`}>
              {href ? <Link href={href}>{link.title}</Link> : <Text>{link.title}</Text>}
            </li>
          );
        })}
      </VStack>
    </VStack>
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

  return (
    <VStack gap={6}>
      <VStack gap={3} hAlign="start">
        <Link href="/build">{t('project.back')}</Link>
        <HStack gap={2} vAlign="center" wrap="wrap">
          <Heading level={1}>{project.name}</Heading>
          {project.visibility === 'private' ? (
            <Badge variant="neutral" label={t('build.private')} />
          ) : null}
          {isOwner && !editing ? (
            <Button
              variant="ghost"
              label={t('project.edit')}
              onClick={() => {
                setEditing(true);
              }}
            />
          ) : null}
        </HStack>
        {editing ? (
          <EditProject
            project={project}
            onDone={() => {
              setEditing(false);
            }}
          />
        ) : (
          <>
            {project.description ? <Text as="p">{project.description}</Text> : null}
            <Technologies items={project.technologies} />
          </>
        )}
        <HStack gap={3} wrap="wrap" vAlign="center">
          <OwnerByline username={project.owner.username} name={project.owner.display_name} />
          <Link href={project.github_url} isExternalLink newTabLabel={t('discover.newTab')}>
            {t('project.github')}
          </Link>
        </HStack>
        <Related links={project.related} />
      </VStack>
      {viewer === 'none' ? null : <Needs projectId={project.id} viewer={viewer} />}
      {isOwner ? <Applications projectId={project.id} /> : null}
    </VStack>
  );
}
