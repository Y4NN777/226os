import { Button } from '@astryxdesign/core/Button';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { TextArea } from '@astryxdesign/core/TextArea';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import { ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { useI18n } from '../../i18n/I18nProvider';
import { FailureNotice } from '../identity/notices';
import { useFailure } from '../identity/useFailure';
import { useInvalidFocus } from '../identity/useInvalidFocus';
import { errorStatus } from '../identity/validation';
import { violationsByName } from '../identity/violations';
import { useIdempotencyKey } from './idempotency';
import { useCreateAnswer } from './queries';
import { checkAnswer } from './validate';

type Lang = components['schemas']['Lang'];

/** Answer box: Markdown body and language; no snippet on answers (S3 decision). */
export function AnswerForm({ questionId }: { questionId: string }) {
  const { t, locale } = useI18n();
  const create = useCreateAnswer(questionId);
  const idempotency = useIdempotencyKey();
  const { failure, report, clear } = useFailure();
  const { formRef, focusFirstInvalid } = useInvalidFocus();
  const [serverInvalid, setServerInvalid] = useState(false);

  const form = useForm({
    defaultValues: { body: '', lang: locale },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      clear();
      setServerInvalid(false);
      const body = { body_md: value.body.trim(), lang: value.lang };
      try {
        await create.mutateAsync({ body, key: idempotency.keyFor(body) });
        idempotency.settle();
        form.reset();
      } catch (error) {
        if (error instanceof ApiError && violationsByName(error).has('body_md')) {
          setServerInvalid(true);
          focusFirstInvalid();
        } else {
          report(error);
        }
      }
    },
  });

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
            name="body"
            validators={{
              onBlur: ({ value }) => checkAnswer(value),
              onSubmit: ({ value }) => checkAnswer(value),
            }}
          >
            {(field) => {
              const key =
                field.state.meta.errors[0] ?? (serverInvalid ? 'solve.err.answer' : undefined);
              return (
                <TextArea
                  label={t('solve.answer.field')}
                  rows={6}
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
          <form.Field name="lang">
            {(field) => (
              <SegmentedControl
                label={t('solve.form.lang')}
                value={field.state.value}
                onChange={(value) => {
                  field.handleChange(value as Lang);
                }}
              >
                <SegmentedControlItem value="fr" label={t('solve.lang.fr')} />
                <SegmentedControlItem value="en" label={t('solve.lang.en')} />
              </SegmentedControl>
            )}
          </form.Field>
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <div>
                <Button
                  type="submit"
                  variant="primary"
                  label={t('solve.answer.submit')}
                  isLoading={isSubmitting}
                />
              </div>
            )}
          </form.Subscribe>
        </VStack>
      </form>
    </VStack>
  );
}
