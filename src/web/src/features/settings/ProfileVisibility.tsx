import { Button } from '@astryxdesign/core/Button';
import { CheckboxInput } from '@astryxdesign/core/CheckboxInput';
import { Heading } from '@astryxdesign/core/Heading';
import { Link } from '@astryxdesign/core/Link';
import { RadioList, RadioListItem } from '@astryxdesign/core/RadioList';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { violationName } from '../identity/violations';
import { meQueryKey, useMe } from './useMe';

type Visibility = components['schemas']['ProfileVisibility'];

const options = ['public', 'members', 'private'] as const satisfies readonly Visibility[];

function isVisibility(value: string): value is Visibility {
  return (options as readonly string[]).includes(value);
}

/**
 * Who may see the member's profile (C14b). Bylines always keep the name. getMe and
 * updateMyProfile may still answer 501 on this environment: the setting is then shown
 * disabled with a clear label, and turns on by itself once profiles are live.
 */
export function ProfileVisibility() {
  const { t } = useI18n();
  const session = useSession();
  const signedIn = session.data?.authenticated === true;
  const queryClient = useQueryClient();
  const me = useMe();
  const [choice, setChoice] = useState<Visibility | null>(null);
  const save = useMutation({
    mutationFn: (visibility: Visibility) =>
      api.call(() => api.client.PATCH('/me/profile', { body: { profile_visibility: visibility } })),
    onSuccess: (profile) => {
      queryClient.setQueryData(meQueryKey, (old: typeof me.data) =>
        old ? { ...old, profile } : old,
      );
    },
  });
  const saveDirectory = useMutation({
    mutationFn: (open: boolean) =>
      api.call(() => api.client.PATCH('/me/profile', { body: { open_to_collab: open } })),
    onSuccess: (profile) => {
      queryClient.setQueryData(meQueryKey, (old: typeof me.data) =>
        old ? { ...old, profile } : old,
      );
    },
  });

  const heading = (
    <VStack gap={1}>
      <Heading level={2}>{t('settings.visibility.title')}</Heading>
      <Text as="p" type="supporting">
        {t('settings.visibility.bylines')}
      </Text>
    </VStack>
  );

  if (session.isPending) return null;
  if (!signedIn) {
    return (
      <VStack gap={3} hAlign="start">
        {heading}
        <Link href="/connexion?next=%2Fparametres" isStandalone>
          {t('settings.visibility.signIn')}
        </Link>
      </VStack>
    );
  }
  if (me.isPending) return <LoadingState />;

  const notLive = me.error instanceof ApiError && me.error.status === 501;
  if (me.isError && !notLive) {
    return (
      <VStack gap={3}>
        {heading}
        <ErrorState error={me.error} onRetry={() => void me.refetch()} />
      </VStack>
    );
  }

  const current = me.data?.profile.profile_visibility ?? 'public';
  const value = choice ?? current;
  return (
    <form
      noValidate
      onSubmit={(event) => {
        event.preventDefault();
        if (!save.isPending) save.mutate(value);
      }}
    >
      <VStack gap={4} hAlign="start">
        {heading}
        <RadioList
          label={t('settings.visibility.label')}
          value={value}
          isDisabled={notLive}
          {...(notLive ? { description: t('settings.visibility.soon') } : {})}
          onChange={(next) => {
            if (isVisibility(next)) {
              setChoice(next);
              save.reset();
            }
          }}
        >
          {options.map((option) => (
            <RadioListItem
              key={option}
              value={option}
              label={t(`settings.visibility.${option}`)}
              description={t(`settings.visibility.${option}.hint`)}
            />
          ))}
        </RadioList>
        <CheckboxInput
          label={t('settings.directory.label')}
          description={t('settings.directory.description')}
          value={me.data?.profile.open_to_collab ?? false}
          isLoading={saveDirectory.isPending}
          isDisabled={notLive}
          onChange={(open) => {
            saveDirectory.mutate(open);
          }}
          {...(saveDirectory.error instanceof ApiError &&
          saveDirectory.error.violations.some(
            (violation) => violationName(violation.field) === 'open_to_collab',
          )
            ? {
                status: {
                  type: 'error' as const,
                  message: t('settings.directory.failed'),
                },
              }
            : {})}
        />
        {notLive ? null : (
          <Button
            type="submit"
            variant="primary"
            label={t('settings.visibility.save')}
            isLoading={save.isPending}
          />
        )}
        <div role="status">
          {save.isSuccess ? <Text>{t('settings.visibility.saved')}</Text> : null}
        </div>
        {save.isError ? <ErrorState error={save.error} title="settings.visibility.failed" /> : null}
        {saveDirectory.isError &&
        !(
          saveDirectory.error instanceof ApiError &&
          saveDirectory.error.violations.some(
            (violation) => violationName(violation.field) === 'open_to_collab',
          )
        ) ? (
          <ErrorState error={saveDirectory.error} title="settings.directory.failed" />
        ) : null}
      </VStack>
    </form>
  );
}
