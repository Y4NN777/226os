import { Button } from '@astryxdesign/core/Button';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { sessionQueryKey, type SessionState } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { useI18n } from '../../i18n/I18nProvider';
import { violationsByName } from '../identity/violations';
import {
  DISPLAY_NAME_MAX_LENGTH,
  DISPLAY_NAME_MIN_LENGTH,
  displayNameViolation,
  validateDisplayName,
  type DisplayNameError,
} from './useDisplayName';
import { meQueryKey } from './useMe';

type Me = components['schemas']['Me'];

/**
 * The display name, saved on the account (PATCH /me/profile). On success the `me` and session
 * caches take the server's answer, so the header shows the new name without a reload.
 */
export function DisplayNameForm({ initialName }: { initialName: string }) {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const [serverError, setServerError] = useState<DisplayNameError | null>(null);

  const save = useMutation({
    mutationFn: (displayName: string) =>
      api.call(() => api.client.PATCH('/me/profile', { body: { display_name: displayName } })),
    onSuccess: (profile) => {
      queryClient.setQueryData<Me>(meQueryKey, (old) =>
        old ? { account: { ...old.account, display_name: profile.display_name }, profile } : old,
      );
      queryClient.setQueryData<SessionState>(sessionQueryKey, (old) =>
        old?.user ? { ...old, user: { ...old.user, display_name: profile.display_name } } : old,
      );
    },
    onError: (error) => {
      const code = error instanceof ApiError ? violationsByName(error).get('display_name') : null;
      setServerError((code && displayNameViolation(code)) || null);
    },
  });

  const form = useForm({
    defaultValues: { displayName: initialName },
    onSubmit: async ({ value }) => {
      if (save.isPending) return;
      setServerError(null);
      await save.mutateAsync(value.displayName.trim()).catch(() => undefined);
    },
  });

  const reset = () => {
    setServerError(null);
    save.reset();
  };

  return (
    <form
      noValidate
      onSubmit={(event) => {
        event.preventDefault();
        reset();
        void form.handleSubmit();
      }}
    >
      <VStack gap={3} hAlign="start">
        <form.Field
          name="displayName"
          validators={{
            onBlur: ({ value }) => validateDisplayName(value),
            onSubmit: ({ value }) => validateDisplayName(value),
          }}
        >
          {(field) => {
            const errorKey = field.state.meta.errors[0] ?? serverError;
            return (
              <TextInput
                label={t('form.name.label')}
                description={t('form.name.description')}
                value={field.state.value}
                htmlName={field.name}
                autoComplete="nickname"
                onBlur={field.handleBlur}
                onChange={(value) => {
                  reset();
                  field.handleChange(value);
                }}
                {...(errorKey
                  ? {
                      status: {
                        type: 'error' as const,
                        message: t(errorKey, {
                          min: DISPLAY_NAME_MIN_LENGTH,
                          max: DISPLAY_NAME_MAX_LENGTH,
                        }),
                      },
                    }
                  : {})}
              />
            );
          }}
        </form.Field>
        <Button type="submit" variant="primary" label={t('form.save')} isLoading={save.isPending} />
        <div role="status">{save.isSuccess ? <Text>{t('form.saved')}</Text> : null}</div>
        {save.isError && !serverError ? (
          <ErrorState error={save.error} title="form.name.failed" />
        ) : null}
      </VStack>
    </form>
  );
}
