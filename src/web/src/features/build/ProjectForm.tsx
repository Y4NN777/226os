import { Button } from '@astryxdesign/core/Button';
import { HStack } from '@astryxdesign/core/HStack';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { TextArea } from '@astryxdesign/core/TextArea';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import { Banner } from '@astryxdesign/core/Banner';
import { ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FailureNotice, FocusRegion } from '../identity/notices';
import { useFailure } from '../identity/useFailure';
import { useInvalidFocus } from '../identity/useInvalidFocus';
import { errorStatus } from '../identity/validation';
import { violationName } from '../identity/violations';
import type { ProjectCreate } from './queries';
import { checkDescription, checkGithubUrl, checkTechnologies, parseTags } from './validate';

type Lang = components['schemas']['Lang'];
type Field = 'github_url' | 'description' | 'technologies';
type ServerErrors = Partial<Record<Field, MessageKey>>;

const keyOfField: Record<Field, MessageKey> = {
  github_url: 'build.err.github',
  description: 'build.err.description',
  technologies: 'build.err.technologies',
};

function fieldOf(pointer: string): Field | undefined {
  const head = violationName(pointer).split('/')[0];
  return head === 'github_url' || head === 'description' || head === 'technologies'
    ? head
    : undefined;
}

/** Field-level outcome of a failed submit; anything else is shown as a notice. */
export function serverFieldErrors(error: unknown): ServerErrors {
  if (!(error instanceof ApiError)) return {};
  if (error.code === 'INVALID_GITHUB_URL') return { github_url: 'build.err.github.invalid' };
  // Duplicate repository: 409 STATE_CONFLICT, answered on the URL field.
  if (error.code === 'STATE_CONFLICT') return { github_url: 'build.err.github.duplicate' };
  const result: ServerErrors = {};
  for (const violation of error.violations) {
    const field = fieldOf(violation.field);
    if (field && !result[field]) result[field] = keyOfField[field];
  }
  return result;
}

function UpstreamNotice({ error }: { error: ApiError }) {
  const { t } = useI18n();
  return (
    <FocusRegion role="alert">
      <Banner
        status="warning"
        title={t('build.upstream.title')}
        description={
          error.retryAfter
            ? t('build.upstream.wait', { seconds: error.retryAfter })
            : t('build.upstream.body')
        }
        collapsible={false}
      />
    </FocusRegion>
  );
}

/** The project form: the GitHub address plus the optional fields of the contract. */
export function ProjectForm({
  onSubmit,
  onCancel,
}: {
  onSubmit: (body: ProjectCreate) => Promise<void>;
  onCancel: () => void;
}) {
  const { t, locale } = useI18n();
  const { failure, report, clear } = useFailure();
  const { formRef, focusFirstInvalid } = useInvalidFocus();
  const [serverErrors, setServerErrors] = useState<ServerErrors>({});

  const defaults: {
    github_url: string;
    description: string;
    technologies: string;
    lang: Lang;
    visibility: NonNullable<ProjectCreate['visibility']>;
  } = { github_url: '', description: '', technologies: '', lang: locale, visibility: 'public' };
  const form = useForm({
    defaultValues: defaults,
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      clear();
      setServerErrors({});
      const technologies = parseTags(value.technologies);
      const description = value.description.trim();
      try {
        await onSubmit({
          github_url: value.github_url.trim(),
          lang: value.lang,
          visibility: value.visibility,
          ...(description ? { description } : {}),
          ...(technologies.length > 0 ? { technologies } : {}),
        });
      } catch (error) {
        const fields = serverFieldErrors(error);
        if (Object.keys(fields).length > 0) {
          setServerErrors(fields);
          focusFirstInvalid();
        } else {
          report(error);
        }
      }
    },
  });

  const forget = (field: Field) => {
    setServerErrors((current) => ({ ...current, [field]: undefined }));
  };

  return (
    <VStack gap={4}>
      {failure ? (
        failure.error instanceof ApiError && failure.error.code === 'UPSTREAM_UNAVAILABLE' ? (
          <UpstreamNotice key={failure.n} error={failure.error} />
        ) : (
          <FailureNotice key={failure.n} error={failure.error} />
        )
      ) : null}
      <form
        method="post"
        ref={formRef}
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          void form.handleSubmit();
        }}
      >
        <VStack gap={4}>
          <form.Field
            name="github_url"
            validators={{
              onBlur: ({ value }) => checkGithubUrl(value),
              onSubmit: ({ value }) => checkGithubUrl(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.github_url;
              return (
                <TextInput
                  label={t('build.form.github')}
                  description={t('build.form.github.hint')}
                  type="url"
                  value={field.state.value}
                  htmlName={field.name}
                  autoCapitalize="none"
                  hasAutoCorrect={false}
                  hasSpellCheck={false}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('github_url');
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="description"
            validators={{ onSubmit: ({ value }) => checkDescription(value) }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.description;
              return (
                <TextArea
                  label={t('build.form.description')}
                  description={t('build.form.description.hint')}
                  isOptional
                  rows={4}
                  value={field.state.value}
                  htmlName={field.name}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('description');
                    field.handleChange(value);
                  }}
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
                  isOptional
                  value={field.state.value}
                  htmlName={field.name}
                  autoCapitalize="none"
                  hasAutoCorrect={false}
                  hasSpellCheck={false}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('technologies');
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field name="lang">
            {(field) => (
              <SegmentedControl
                label={t('build.form.lang')}
                value={field.state.value}
                onChange={(value) => {
                  field.handleChange(value as Lang);
                }}
              >
                <SegmentedControlItem value="fr" label={t('build.lang.fr')} />
                <SegmentedControlItem value="en" label={t('build.lang.en')} />
              </SegmentedControl>
            )}
          </form.Field>
          <form.Field name="visibility">
            {(field) => (
              <SegmentedControl
                label={t('build.form.visibility')}
                value={field.state.value}
                onChange={(value) => {
                  field.handleChange(value === 'private' ? 'private' : 'public');
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
                  label={t('build.form.submit')}
                  isLoading={isSubmitting}
                />
                <Button variant="ghost" label={t('build.form.cancel')} onClick={onCancel} />
              </HStack>
            )}
          </form.Subscribe>
        </VStack>
      </form>
    </VStack>
  );
}
