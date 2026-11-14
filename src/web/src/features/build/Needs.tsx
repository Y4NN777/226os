import { Badge } from '@astryxdesign/core/Badge';
import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { Text } from '@astryxdesign/core/Text';
import { TextArea } from '@astryxdesign/core/TextArea';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FailureNotice, FocusRegion } from '../identity/notices';
import { useFailure } from '../identity/useFailure';
import { useInvalidFocus } from '../identity/useInvalidFocus';
import { errorStatus } from '../identity/validation';
import { violationName } from '../identity/violations';
import { signInHref } from '../solve/signIn';
import { Technologies } from './present';
import {
  useCreateApplication,
  useCreateNeed,
  useDecideApplication,
  useNeeds,
  type Need,
} from './queries';
import {
  checkApplicationMessage,
  checkNeedDescription,
  checkNeedSkills,
  checkNeedTitle,
  parseTags,
} from './validate';

type Lang = components['schemas']['Lang'];

const statusKey = {
  open: 'project.need.status.open',
  filled: 'project.need.status.filled',
  closed: 'project.need.status.closed',
} as const satisfies Record<Need['status'], MessageKey>;

/** A refusal the page can explain precisely, shown as one focused notice. */
function MessageNotice({ message }: { message: string }) {
  return (
    <FocusRegion role="alert">
      <Banner status="error" title={message} collapsible={false} />
    </FocusRegion>
  );
}

/** What this visitor did on a need: applied (with the id to withdraw by, if known) or withdrawn. */
type Sent = { id: string | null; withdrawn: boolean };

/**
 * Kept in the query cache, not in component state: leaving the project page and coming back must
 * not offer the apply button again. Sign-out clears the cache (app/session.ts), so it never leaks
 * to another account; after a reload the server answers DUPLICATE_APPLICATION, which is explained.
 */
function useSent(needId: string) {
  const queryClient = useQueryClient();
  const key = ['applications', 'sent', needId] as const;
  const { data } = useQuery<Sent | null>({
    queryKey: key,
    queryFn: () => null,
    staleTime: Infinity,
    gcTime: Infinity,
    enabled: false,
    initialData: null,
  });
  return [
    data,
    (next: Sent) => {
      queryClient.setQueryData(key, next);
    },
  ] as const;
}

function ApplyForm({ need, projectId }: { need: Need; projectId: string }) {
  const { t } = useI18n();
  const create = useCreateApplication(need.id);
  const decide = useDecideApplication(projectId);
  const { failure, report, clear } = useFailure();
  const { formRef, focusFirstInvalid } = useInvalidFocus();
  const [serverInvalid, setServerInvalid] = useState(false);
  const [sent, setSent] = useSent(need.id);

  const form = useForm({
    defaultValues: { message: '' },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      clear();
      setServerInvalid(false);
      try {
        const created = await create.mutateAsync(value.message.trim());
        setSent({ id: created.id, withdrawn: false });
      } catch (error) {
        if (
          error instanceof ApiError &&
          error.violations.some((v) => violationName(v.field) === 'message')
        ) {
          setServerInvalid(true);
          focusFirstInvalid();
        } else {
          report(error);
        }
      }
    },
  });

  if (sent?.withdrawn)
    return (
      <Text as="p" role="status">
        {t('project.apply.withdrawn')}
      </Text>
    );
  if (sent) {
    return (
      <VStack gap={2} hAlign="start">
        <Text as="p" role="status">
          {t('project.apply.sent')}
        </Text>
        {sent.id !== null ? (
          <Button
            variant="ghost"
            label={t('project.apply.withdraw')}
            isLoading={decide.isPending}
            onClick={() => {
              decide.mutate(
                { id: sent.id ?? '', status: 'withdrawn' },
                {
                  onSuccess: () => {
                    setSent({ id: sent.id, withdrawn: true });
                  },
                },
              );
            }}
          />
        ) : null}
        {decide.isError ? (
          decide.error instanceof ApiError && decide.error.code === 'STATE_CONFLICT' ? (
            <MessageNotice message={t('project.err.decided')} />
          ) : (
            <ErrorState error={decide.error} />
          )
        ) : null}
      </VStack>
    );
  }

  const failureNotice = (error: unknown) => {
    if (error instanceof ApiError && error.code === 'DUPLICATE_APPLICATION') {
      return <MessageNotice message={t('project.err.duplicate')} />;
    }
    if (error instanceof ApiError && error.code === 'STATE_CONFLICT') {
      return <MessageNotice message={t('project.err.closed')} />;
    }
    return <FailureNotice error={error} />;
  };

  return (
    <VStack gap={3}>
      {failure ? failureNotice(failure.error) : null}
      <form
        method="post"
        ref={formRef}
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          void form.handleSubmit();
        }}
      >
        <VStack gap={3} hAlign="start">
          <form.Field
            name="message"
            validators={{
              onBlur: ({ value }) => checkApplicationMessage(value),
              onSubmit: ({ value }) => checkApplicationMessage(value),
            }}
          >
            {(field) => {
              const key =
                field.state.meta.errors[0] ?? (serverInvalid ? 'project.err.message' : undefined);
              return (
                <TextArea
                  label={t('project.apply.field')}
                  description={t('project.apply.hint')}
                  rows={5}
                  value={field.state.value}
                  htmlName={field.name}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    setServerInvalid(false);
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <Button
                type="submit"
                variant="primary"
                label={t('project.apply.submit')}
                isLoading={isSubmitting}
              />
            )}
          </form.Subscribe>
        </VStack>
      </form>
    </VStack>
  );
}

function NeedItem({
  need,
  projectId,
  viewer,
}: {
  need: Need;
  projectId: string;
  viewer: 'owner' | 'member' | 'anonymous';
}) {
  const { t } = useI18n();
  const [applying, setApplying] = useState(false);
  const [sent] = useSent(need.id);
  return (
    <Card variant="muted">
      <VStack gap={2}>
        <HStack gap={2} wrap="wrap" vAlign="center">
          <Heading level={3}>{need.title}</Heading>
          <Badge
            variant={need.status === 'open' ? 'success' : 'neutral'}
            label={t(statusKey[need.status])}
          />
        </HStack>
        <Text as="p">{need.description}</Text>
        <Technologies items={need.skills} />
        {need.status === 'open' && viewer === 'member' ? (
          applying || sent ? (
            <ApplyForm need={need} projectId={projectId} />
          ) : (
            <HStack>
              <Button
                label={t('project.apply')}
                onClick={() => {
                  setApplying(true);
                }}
              />
            </HStack>
          )
        ) : null}
        {need.status === 'open' && viewer === 'anonymous' ? (
          <HStack>
            <Button label={t('project.apply.signIn')} href={signInHref(`/build/${projectId}`)} />
          </HStack>
        ) : null}
      </VStack>
    </Card>
  );
}

type NeedField = 'title' | 'description' | 'skills';
const needKeys: Record<NeedField, MessageKey> = {
  title: 'project.err.needTitle',
  description: 'project.err.needDescription',
  skills: 'project.err.needSkills',
};

function NeedForm({ projectId, onClose }: { projectId: string; onClose: () => void }) {
  const { t, locale } = useI18n();
  const create = useCreateNeed(projectId);
  const { failure, report, clear } = useFailure();
  const { formRef, focusFirstInvalid } = useInvalidFocus();
  const [serverErrors, setServerErrors] = useState<Partial<Record<NeedField, MessageKey>>>({});

  const form = useForm({
    defaultValues: { title: '', description: '', skills: '', lang: locale },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      clear();
      setServerErrors({});
      try {
        await create.mutateAsync({
          title: value.title.trim(),
          description: value.description.trim(),
          lang: value.lang,
          skills: parseTags(value.skills),
        });
        onClose();
      } catch (error) {
        const fields: Partial<Record<NeedField, MessageKey>> = {};
        if (error instanceof ApiError) {
          for (const violation of error.violations) {
            const head = violationName(violation.field).split('/')[0];
            if (head === 'title' || head === 'description' || head === 'skills') {
              fields[head] ??= needKeys[head];
            }
          }
        }
        if (Object.keys(fields).length > 0) {
          setServerErrors(fields);
          focusFirstInvalid();
        } else {
          report(error);
        }
      }
    },
  });

  const forget = (field: NeedField) => {
    setServerErrors((current) => ({ ...current, [field]: undefined }));
  };

  return (
    <VStack gap={3}>
      {failure ? <FailureNotice key={failure.n} error={failure.error} /> : null}
      <form
        method="post"
        ref={formRef}
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          void form.handleSubmit();
        }}
      >
        <VStack gap={3}>
          <form.Field
            name="title"
            validators={{
              onBlur: ({ value }) => checkNeedTitle(value),
              onSubmit: ({ value }) => checkNeedTitle(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.title;
              return (
                <TextInput
                  label={t('project.need.title')}
                  value={field.state.value}
                  htmlName={field.name}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('title');
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="description"
            validators={{
              onBlur: ({ value }) => checkNeedDescription(value),
              onSubmit: ({ value }) => checkNeedDescription(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.description;
              return (
                <TextArea
                  label={t('project.need.description')}
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
            name="skills"
            validators={{
              onBlur: ({ value }) => checkNeedSkills(value),
              onSubmit: ({ value }) => checkNeedSkills(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.skills;
              return (
                <TextInput
                  label={t('project.need.skills')}
                  description={t('project.need.skills.hint')}
                  value={field.state.value}
                  htmlName={field.name}
                  autoCapitalize="none"
                  hasAutoCorrect={false}
                  hasSpellCheck={false}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('skills');
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
                label={t('project.need.lang')}
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
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <HStack gap={2}>
                <Button
                  type="submit"
                  variant="primary"
                  label={t('project.need.submit')}
                  isLoading={isSubmitting}
                />
                <Button variant="ghost" label={t('project.cancel')} onClick={onClose} />
              </HStack>
            )}
          </form.Subscribe>
        </VStack>
      </form>
    </VStack>
  );
}

export function Needs({
  projectId,
  viewer,
}: {
  projectId: string;
  viewer: 'owner' | 'member' | 'anonymous';
}) {
  const { t } = useI18n();
  const needs = useNeeds(projectId);
  const [adding, setAdding] = useState(false);
  const items = needs.data?.pages.flatMap((page) => page.items) ?? [];
  return (
    <section aria-label={t('project.needs.title')}>
      <VStack gap={3}>
        <HStack gap={3} vAlign="center" wrap="wrap">
          <Heading level={2}>{t('project.needs.title')}</Heading>
          {viewer === 'owner' && !adding ? (
            <Button
              variant="ghost"
              label={t('project.need.add')}
              onClick={() => {
                setAdding(true);
              }}
            />
          ) : null}
        </HStack>
        {viewer === 'owner' && adding ? (
          <NeedForm
            projectId={projectId}
            onClose={() => {
              setAdding(false);
            }}
          />
        ) : null}
        {needs.isPending ? <LoadingState /> : null}
        {needs.isError && !needs.data ? (
          <ErrorState error={needs.error} onRetry={() => void needs.refetch()} />
        ) : null}
        {needs.data && items.length === 0 ? <Text as="p">{t('project.needs.none')}</Text> : null}
        {items.length > 0 ? (
          <VStack as="ul" gap={3}>
            {items.map((need) => (
              <li key={need.id}>
                <NeedItem need={need} projectId={projectId} viewer={viewer} />
              </li>
            ))}
          </VStack>
        ) : null}
        {needs.hasNextPage ? (
          <HStack hAlign="center">
            <Button
              label={t('build.more')}
              isLoading={needs.isFetchingNextPage}
              onClick={() => void needs.fetchNextPage()}
            />
          </HStack>
        ) : null}
      </VStack>
    </section>
  );
}
