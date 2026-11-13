import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useCallback, useEffect, useRef, useState } from 'react';
import { api, ApiError } from '../../api/http';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { FailureNotice, FocusRegion, RateLimitNotice } from './notices';
import { useCountdown } from './useCountdown';
import { useFailure } from './useFailure';
import { useHashToken } from './useHashToken';
import { useInvalidFocus } from './useInvalidFocus';
import { isRefusedToken } from './violations';
import { checkToken, errorStatus } from './validation';

type Phase = 'form' | 'checking' | 'done' | 'invalid';

export function VerifyEmailPage() {
  const { t } = useI18n();
  const hashToken = useHashToken();
  const [phase, setPhase] = useState<Phase>(hashToken === null ? 'form' : 'checking');
  const { failure, report, clear } = useFailure();
  const countdown = useCountdown();
  const startCountdown = countdown.start;
  const { formRef, focusFirstInvalid } = useInvalidFocus();

  const confirm = useCallback(
    async (token: string) => {
      clear();
      try {
        await api.call(() => api.client.POST('/auth/verify-email', { body: { token } }));
        setPhase('done');
      } catch (error) {
        if (isRefusedToken(error)) {
          setPhase('invalid');
          return;
        }
        if (error instanceof ApiError && error.code === 'RATE_LIMITED') {
          startCountdown(error.retryAfter ?? 0);
        }
        setPhase('form');
        report(error);
      }
    },
    [clear, report, startCountdown],
  );

  const form = useForm({
    defaultValues: { token: hashToken ?? '' },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      if (countdown.seconds > 0) return;
      await confirm(value.token.trim());
    },
  });

  const autoSubmitted = useRef(false);
  useEffect(() => {
    if (hashToken === null || autoSubmitted.current) return;
    autoSubmitted.current = true;
    void confirm(hashToken.trim());
  }, [hashToken, confirm]);

  if (phase === 'checking') {
    return (
      <VStack gap={4}>
        <Heading level={1}>{t('identity.verify.title')}</Heading>
        <LoadingState label="identity.verify.checking" />
      </VStack>
    );
  }

  if (phase === 'done') {
    return (
      <FocusRegion role="status">
        <VStack gap={4}>
          <Heading level={1}>{t('identity.verify.done.title')}</Heading>
          <Text as="p">{t('identity.verify.done.body')}</Text>
          <Button
            variant="primary"
            href="/connexion?next=/bienvenue"
            label={t('identity.verify.done.signIn')}
          />
        </VStack>
      </FocusRegion>
    );
  }

  if (phase === 'invalid') {
    return (
      <FocusRegion role="alert">
        <VStack gap={4}>
          <Heading level={1}>{t('identity.code.invalid.title')}</Heading>
          <Text as="p">{t('identity.verify.invalid.body')}</Text>
          <HStack gap={3} wrap="wrap">
            <Button
              variant="primary"
              label={t('identity.verify.invalid.retry')}
              onClick={() => {
                form.reset();
                setPhase('form');
              }}
            />
            <Button href="/inscription" label={t('identity.verify.invalid.signUp')} />
          </HStack>
        </VStack>
      </FocusRegion>
    );
  }

  return (
    <VStack gap={5}>
      <Heading level={1}>{t('identity.verify.title')}</Heading>
      <Text as="p">{t('identity.verify.intro')}</Text>
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
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <Button
                type="submit"
                variant="primary"
                width="100%"
                label={
                  countdown.seconds > 0
                    ? t('identity.limit.wait', { seconds: countdown.seconds })
                    : t('identity.verify.submit')
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
