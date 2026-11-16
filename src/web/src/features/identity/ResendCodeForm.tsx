import { Button } from '@astryxdesign/core/Button';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import { api, ApiError } from '../../api/http';
import { useI18n } from '../../i18n/I18nProvider';
import { CardHead, SwitchLine } from './AuthCard';
import { FailureNotice, FocusRegion, RateLimitNotice } from './notices';
import { useCountdown } from './useCountdown';
import { useFailure } from './useFailure';
import { useInvalidFocus } from './useInvalidFocus';
import { checkEmail, errorStatus } from './validation';

/** Asks for a new verification code; `onBack` returns to the code field. */
export function ResendCodeForm({ onBack }: { onBack: () => void }) {
  const { t } = useI18n();
  const [sent, setSent] = useState(false);
  const { failure, report, clear } = useFailure();
  const countdown = useCountdown();
  const { formRef, focusFirstInvalid } = useInvalidFocus();

  const form = useForm({
    defaultValues: { email: '' },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      if (countdown.seconds > 0) return;
      clear();
      try {
        await api.call(() =>
          api.client.POST('/auth/verify-email/resend', { body: { email: value.email.trim() } }),
        );
        // Same answer whether or not the address waits for a confirmation.
        setSent(true);
      } catch (error) {
        if (error instanceof ApiError && error.code === 'RATE_LIMITED') {
          countdown.start(error.retryAfter ?? 0);
        }
        report(error);
      }
    },
  });

  if (sent) {
    return (
      <FocusRegion role="status">
        <VStack gap={4}>
          <CardHead title={t('identity.resend.done.title')} lede={t('identity.resend.done.body')} />
          <Button variant="primary" label={t('identity.resend.done.link')} onClick={onBack} />
        </VStack>
      </FocusRegion>
    );
  }

  return (
    <VStack gap={5}>
      <CardHead title={t('identity.resend.title')} lede={t('identity.resend.intro')} />
      {failure ? (
        failure.error instanceof ApiError && failure.error.code === 'RATE_LIMITED' ? (
          <RateLimitNotice key={failure.n} />
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
            name="email"
            validators={{
              onBlur: ({ value }) => checkEmail(value),
              onSubmit: ({ value }) => checkEmail(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0];
              return (
                <TextInput
                  label={t('identity.field.email')}
                  type="email"
                  value={field.state.value}
                  htmlName={field.name}
                  autoComplete="email"
                  onBlur={field.handleBlur}
                  onChange={field.handleChange}
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
                width="100%"
                label={
                  countdown.seconds > 0
                    ? t('identity.limit.wait', { seconds: countdown.seconds })
                    : t('identity.resend.submit')
                }
                isLoading={isSubmitting}
                isDisabled={countdown.seconds > 0}
              />
            )}
          </form.Subscribe>
        </VStack>
      </form>
      <SwitchLine>
        <Button variant="ghost" label={t('identity.resend.back')} onClick={onBack} />
      </SwitchLine>
    </VStack>
  );
}
