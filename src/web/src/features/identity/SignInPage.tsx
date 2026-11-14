import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useRouter, useSearch } from '@tanstack/react-router';
import { api, ApiError } from '../../api/http';
import { useApplySession } from '../../app/session';
import { useI18n } from '../../i18n/I18nProvider';
import { FailureNotice, FocusRegion, RateLimitNotice } from './notices';
import { PasswordField } from './PasswordField';
import { safeNext } from './routeSearch';
import { useCountdown } from './useCountdown';
import { useFailure } from './useFailure';
import { useInvalidFocus } from './useInvalidFocus';
import { checkEmail, checkPasswordPresent, errorStatus } from './validation';

function SignInNotice({ error }: { error: unknown }) {
  const { t } = useI18n();
  if (error instanceof ApiError) {
    if (error.code === 'INVALID_CREDENTIALS' || error.code === 'UNAUTHENTICATED') {
      // One message for every credential failure: it never says which part was wrong.
      return (
        <FocusRegion role="alert">
          <Banner
            status="error"
            title={t('identity.signIn.bad.title')}
            description={t('identity.signIn.bad.body')}
            collapsible={false}
          />
        </FocusRegion>
      );
    }
    if (error.code === 'RATE_LIMITED') return <RateLimitNotice />;
  }
  return <FailureNotice error={error} />;
}

export function SignInPage() {
  const { t } = useI18n();
  const router = useRouter();
  const applySession = useApplySession();
  const next = safeNext(useSearch({ strict: false }).next);
  const { failure, report, clear } = useFailure();
  const countdown = useCountdown();
  const { formRef, focusFirstInvalid } = useInvalidFocus();

  const form = useForm({
    defaultValues: { email: '', password: '' },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value, formApi }) => {
      if (countdown.seconds > 0) return;
      clear();
      try {
        const state = await api.call(() =>
          api.client.POST('/auth/sign-in', {
            body: { email: value.email.trim(), password: value.password },
          }),
        );
        applySession(state);
        router.history.push(next ?? '/');
      } catch (error) {
        if (error instanceof ApiError) {
          if (error.code === 'RATE_LIMITED') countdown.start(error.retryAfter ?? 0);
          if (error.status === 401) formApi.setFieldValue('password', '');
        }
        report(error);
      }
    },
  });

  const withNext = (path: string) => (next ? `${path}?next=${encodeURIComponent(next)}` : path);

  return (
    <VStack gap={5}>
      <Heading level={1}>{t('identity.signIn.title')}</Heading>
      {failure ? <SignInNotice key={failure.n} error={failure.error} /> : null}
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
                  autoComplete="username"
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
              onBlur: ({ value }) => checkPasswordPresent(value),
              onSubmit: ({ value }) => checkPasswordPresent(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0];
              return (
                <PasswordField
                  label={t('identity.field.password')}
                  value={field.state.value}
                  name={field.name}
                  autoComplete="current-password"
                  onBlur={field.handleBlur}
                  onChange={field.handleChange}
                  error={key ? t(key) : undefined}
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
                    : t('identity.signIn.submit')
                }
                isLoading={isSubmitting}
                isDisabled={countdown.seconds > 0}
              />
            )}
          </form.Subscribe>
        </VStack>
      </form>
      <VStack gap={2}>
        <Link href="/mot-de-passe-oublie">{t('identity.signIn.forgot')}</Link>
        <Text as="p">
          {t('identity.signIn.noAccount')}{' '}
          <Link href={withNext('/inscription')}>{t('identity.signIn.signUp')}</Link>
        </Text>
      </VStack>
    </VStack>
  );
}
