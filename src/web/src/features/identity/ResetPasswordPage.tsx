import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { api, ApiError } from '../../api/http';
import { sessionQueryKey } from '../../app/session';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FailureNotice, FocusRegion, RateLimitNotice } from './notices';
import { PasswordField } from './PasswordField';
import { PasswordRule } from './PasswordRule';
import { useCountdown } from './useCountdown';
import { useFailure } from './useFailure';
import { useHashToken } from './useHashToken';
import { useInvalidFocus } from './useInvalidFocus';
import { isRefusedToken, violationsByName } from './violations';
import { checkNewPassword, checkToken, errorStatus } from './validation';

type Phase = 'form' | 'done' | 'invalid';

export function ResetPasswordPage() {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const hashToken = useHashToken();
  const [phase, setPhase] = useState<Phase>('form');
  const [passwordError, setPasswordError] = useState<MessageKey | undefined>();
  const { failure, report, clear } = useFailure();
  const countdown = useCountdown();
  const { formRef, focusFirstInvalid } = useInvalidFocus();

  const form = useForm({
    defaultValues: { token: hashToken ?? '', password: '' },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      if (countdown.seconds > 0) return;
      clear();
      setPasswordError(undefined);
      try {
        await api.call(() =>
          api.client.POST('/auth/password-reset/confirm', {
            body: { token: value.token.trim(), new_password: value.password },
          }),
        );
        // Every session was revoked server-side: drop the CSRF token and read the session again.
        api.csrf.set(null);
        await queryClient.invalidateQueries({ queryKey: sessionQueryKey });
        setPhase('done');
      } catch (error) {
        if (isRefusedToken(error)) {
          setPhase('invalid');
          return;
        }
        if (error instanceof ApiError) {
          const code = violationsByName(error).get('new_password');
          if (code) {
            setPasswordError(
              code === 'too_long'
                ? 'identity.err.password.tooLong'
                : 'identity.err.password.tooShort',
            );
            focusFirstInvalid();
            return;
          }
        }
        if (error instanceof ApiError && error.code === 'RATE_LIMITED') {
          countdown.start(error.retryAfter ?? 0);
        }
        report(error);
      }
    },
  });

  if (phase === 'done') {
    return (
      <FocusRegion role="status">
        <VStack gap={4}>
          <Heading level={1}>{t('identity.reset.done.title')}</Heading>
          <Text as="p">{t('identity.reset.done.body')}</Text>
          <Button variant="primary" href="/connexion" label={t('identity.reset.done.signIn')} />
        </VStack>
      </FocusRegion>
    );
  }

  if (phase === 'invalid') {
    return (
      <FocusRegion role="alert">
        <VStack gap={4}>
          <Heading level={1}>{t('identity.code.invalid.title')}</Heading>
          <Text as="p">{t('identity.reset.invalid.body')}</Text>
          <Button
            variant="primary"
            href="/mot-de-passe-oublie"
            label={t('identity.reset.invalid.retry')}
          />
        </VStack>
      </FocusRegion>
    );
  }

  return (
    <VStack gap={5}>
      <Heading level={1}>{t('identity.reset.title')}</Heading>
      <Text as="p">{t('identity.reset.intro')}</Text>
      {failure ? (
        failure.error instanceof ApiError && failure.error.code === 'RATE_LIMITED' ? (
          <RateLimitNotice key={failure.n} />
        ) : (
          <FailureNotice key={failure.n} error={failure.error} />
        )
      ) : null}
      <form
        ref={formRef}
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          void form.handleSubmit();
        }}
      >
        <VStack gap={4}>
          <form.Field
            name="token"
            validators={{
              onBlur: ({ value }) => checkToken(value),
              onSubmit: ({ value }) => checkToken(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0];
              return (
                <TextInput
                  label={t('identity.field.code')}
                  description={t('identity.field.code.description')}
                  value={field.state.value}
                  htmlName={field.name}
                  autoComplete="one-time-code"
                  autoCapitalize="none"
                  hasAutoCorrect={false}
                  hasSpellCheck={false}
                  onBlur={field.handleBlur}
                  onChange={field.handleChange}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="password"
            validators={{
              onBlur: ({ value }) => checkNewPassword(value),
              onSubmit: ({ value }) => checkNewPassword(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? passwordError;
              return (
                <VStack gap={1}>
                  <PasswordField
                    label={t('identity.field.newPassword')}
                    value={field.state.value}
                    name={field.name}
                    autoComplete="new-password"
                    onBlur={field.handleBlur}
                    onChange={(value) => {
                      setPasswordError(undefined);
                      field.handleChange(value);
                    }}
                    error={key ? t(key) : undefined}
                  />
                  <PasswordRule value={field.state.value} />
                </VStack>
              );
            }}
          </form.Field>
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <Button
                type="submit"
                variant="primary"
                width="100%"
                label={
                  countdown.seconds > 0
                    ? t('identity.limit.wait', { seconds: countdown.seconds })
                    : t('identity.reset.submit')
                }
                isLoading={isSubmitting}
                isDisabled={countdown.seconds > 0}
              />
            )}
          </form.Subscribe>
        </VStack>
      </form>
    </VStack>
  );
}
