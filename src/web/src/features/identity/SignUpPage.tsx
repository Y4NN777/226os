import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useRouter, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import { api, ApiError } from '../../api/http';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FailureNotice, RateLimitNotice } from './notices';
import { PasswordField } from './PasswordField';
import { PasswordRule } from './PasswordRule';
import { safeNext } from './routeSearch';
import { useCountdown } from './useCountdown';
import { useFailure } from './useFailure';
import { useInvalidFocus } from './useInvalidFocus';
import { violationsByName } from './violations';
import { checkEmail, checkNewPassword, errorStatus, USERNAME_RE } from './validation';

const NAME_MAX = 80;

type FieldName = 'name' | 'username' | 'email' | 'password';
type ServerErrors = Partial<Record<FieldName, MessageKey>>;

const fieldOfName: Record<string, FieldName> = {
  display_name: 'name',
  username: 'username',
  email: 'email',
  password: 'password',
};

function violationKey(field: FieldName, code: string): MessageKey {
  switch (field) {
    case 'name':
      return 'identity.signUp.err.name.tooLong';
    case 'username':
      return code === 'required'
        ? 'identity.signUp.err.username.required'
        : 'identity.signUp.err.username.format';
    case 'email':
      return code === 'required' ? 'identity.err.email.required' : 'identity.err.email.format';
    case 'password':
      if (code === 'required') return 'identity.err.password.required';
      return code === 'too_long'
        ? 'identity.err.password.tooLong'
        : 'identity.err.password.tooShort';
  }
}

function serverFieldErrors(error: ApiError): ServerErrors {
  const result: ServerErrors = {};
  for (const [name, code] of violationsByName(error)) {
    const field = fieldOfName[name];
    if (field) result[field] = violationKey(field, code);
  }
  return result;
}

function checkUsername(value: string): MessageKey | undefined {
  if (!value) return 'identity.signUp.err.username.required';
  return USERNAME_RE.test(value) ? undefined : 'identity.signUp.err.username.format';
}

function checkName(value: string): MessageKey | undefined {
  return Array.from(value.trim()).length > NAME_MAX
    ? 'identity.signUp.err.name.tooLong'
    : undefined;
}

export function SignUpPage() {
  const { t, locale } = useI18n();
  const router = useRouter();
  const next = safeNext(useSearch({ strict: false }).next);
  const { failure, report, clear } = useFailure();
  const countdown = useCountdown();
  const { formRef, focusFirstInvalid } = useInvalidFocus();
  const [serverErrors, setServerErrors] = useState<ServerErrors>({});

  const form = useForm({
    defaultValues: { name: '', username: '', email: '', password: '' },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      if (countdown.seconds > 0) return;
      clear();
      setServerErrors({});
      const name = value.name.trim();
      try {
        await api.call(() =>
          api.client.POST('/auth/sign-up', {
            body: {
              email: value.email.trim(),
              password: value.password,
              username: value.username,
              ...(name ? { display_name: name } : {}),
              ui_lang: locale,
            },
          }),
        );
        // Neutral by design: the API answers 202 whether or not the address or username is taken.
        router.history.push('/verifier-email');
      } catch (error) {
        if (error instanceof ApiError && error.code === 'RATE_LIMITED') {
          countdown.start(error.retryAfter ?? 0);
        }
        const fields = error instanceof ApiError ? serverFieldErrors(error) : {};
        if (Object.keys(fields).length > 0) {
          setServerErrors(fields);
          focusFirstInvalid();
        } else {
          report(error);
        }
      }
    },
  });

  const forget = (field: FieldName) => {
    setServerErrors((current) => ({ ...current, [field]: undefined }));
  };

  return (
    <VStack gap={5}>
      <Heading level={1}>{t('identity.signUp.title')}</Heading>
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
            name="name"
            validators={{
              onBlur: ({ value }) => checkName(value),
              onSubmit: ({ value }) => checkName(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.name;
              return (
                <TextInput
                  label={t('identity.signUp.name.label')}
                  isOptional
                  value={field.state.value}
                  htmlName={field.name}
                  autoComplete="name"
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('name');
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="username"
            validators={{
              onBlur: ({ value }) => checkUsername(value),
              onSubmit: ({ value }) => checkUsername(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.username;
              return (
                <TextInput
                  label={t('identity.signUp.username.label')}
                  description={t('identity.signUp.username.description')}
                  value={field.state.value}
                  htmlName={field.name}
                  autoComplete="username"
                  autoCapitalize="none"
                  hasAutoCorrect={false}
                  hasSpellCheck={false}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('username');
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="email"
            validators={{
              onBlur: ({ value }) => checkEmail(value),
              onSubmit: ({ value }) => checkEmail(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.email;
              return (
                <TextInput
                  label={t('identity.field.email')}
                  type="email"
                  value={field.state.value}
                  htmlName={field.name}
                  autoComplete="email"
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('email');
                    field.handleChange(value);
                  }}
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
              const key = field.state.meta.errors[0] ?? serverErrors.password;
              return (
                <VStack gap={1}>
                  <PasswordField
                    label={t('identity.field.password')}
                    value={field.state.value}
                    name={field.name}
                    autoComplete="new-password"
                    onBlur={field.handleBlur}
                    onChange={(value) => {
                      forget('password');
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
                    : t('identity.signUp.submit')
                }
                isLoading={isSubmitting}
                isDisabled={countdown.seconds > 0}
              />
            )}
          </form.Subscribe>
        </VStack>
      </form>
      <Text as="p">
        {t('identity.signUp.hasAccount')}{' '}
        <Link href={next ? `/connexion?next=${encodeURIComponent(next)}` : '/connexion'}>
          {t('identity.signUp.signIn')}
        </Link>
      </Text>
    </VStack>
  );
}
