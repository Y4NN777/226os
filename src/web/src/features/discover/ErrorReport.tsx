import { Button } from '@astryxdesign/core/Button';
import { Icon } from '@astryxdesign/core/Icon';
import { Selector } from '@astryxdesign/core/Selector';
import { Text } from '@astryxdesign/core/Text';
import { TextArea } from '@astryxdesign/core/TextArea';
import { VStack } from '@astryxdesign/core/VStack';
import { HStack } from '@astryxdesign/core/HStack';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import { ApiError } from '../../api/http';
import { useSession } from '../../app/session';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FailureNotice, FocusRegion } from '../identity/notices';
import { useFailure } from '../identity/useFailure';
import { useInvalidFocus } from '../identity/useInvalidFocus';
import { errorStatus } from '../identity/validation';
import { violationsByName } from '../identity/violations';
import { signInHref } from '../solve/signIn';
import { useCreateErrorReport, useResourceSources, type ErrorReportCreate } from './queries';

const kinds = [
  'broken_link',
  'outdated',
  'incorrect',
  'unsupported_country',
  'other',
] as const satisfies readonly ErrorReportCreate['kind'][];

const kindKey = {
  broken_link: 'discover.report.kind.broken_link',
  outdated: 'discover.report.kind.outdated',
  incorrect: 'discover.report.kind.incorrect',
  unsupported_country: 'discover.report.kind.unsupported_country',
  other: 'discover.report.kind.other',
} as const satisfies Record<ErrorReportCreate['kind'], MessageKey>;

const BODY_MAX = 2000;

function isKind(value: string): value is ErrorReportCreate['kind'] {
  return kinds.some((kind) => kind === value);
}

function checkBody(value: string): MessageKey | undefined {
  const size = Array.from(value.trim()).length;
  return size < 1 || size > BODY_MAX ? 'discover.report.err.body' : undefined;
}

function Form({ resourceId, onSent }: { resourceId: string; onSent: () => void }) {
  const { t } = useI18n();
  const create = useCreateErrorReport(resourceId);
  const sources = useResourceSources(resourceId);
  const { failure, report, clear } = useFailure();
  const { formRef, focusFirstInvalid } = useInvalidFocus();
  const [serverInvalid, setServerInvalid] = useState<ReadonlySet<string>>(new Set());

  const form = useForm({
    defaultValues: {
      kind: null as ErrorReportCreate['kind'] | null,
      body: '',
      source: null as string | null,
    },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      clear();
      setServerInvalid(new Set());
      if (value.kind === null) return;
      const body: ErrorReportCreate = {
        kind: value.kind,
        body: value.body.trim(),
        ...(value.source ? { source_id: value.source } : {}),
      };
      try {
        await create.mutateAsync(body);
        onSent();
      } catch (error) {
        const names =
          error instanceof ApiError && error.status === 422 ? violationsByName(error) : null;
        if (names && (names.has('body') || names.has('kind'))) {
          setServerInvalid(new Set(names.keys()));
          focusFirstInvalid();
        } else {
          report(error);
        }
      }
    },
  });

  const sourceOptions = (sources.data?.items ?? []).map((source) => ({
    value: source.id,
    label: source.title ?? source.url,
  }));

  return (
    <VStack gap={3}>
      {failure ? <FailureNotice key={failure.n} error={failure.error} /> : null}
      <form
        method="post"
        ref={formRef}
        noValidate
        aria-label={t('discover.report.action')}
        onSubmit={(event) => {
          event.preventDefault();
          void form.handleSubmit();
        }}
      >
        <VStack gap={3}>
          <form.Field
            name="kind"
            validators={{
              onSubmit: ({ value }) => (value === null ? 'discover.report.err.kind' : undefined),
            }}
          >
            {(field) => {
              const key =
                field.state.meta.errors[0] ??
                (serverInvalid.has('kind') ? 'discover.report.err.kind' : undefined);
              return (
                <Selector
                  label={t('discover.report.kind')}
                  placeholder={t('discover.report.kind.placeholder')}
                  options={kinds.map((kind) => ({ value: kind, label: t(kindKey[kind]) }))}
                  {...(field.state.value ? { value: field.state.value } : {})}
                  onChange={(value) => {
                    field.handleChange(isKind(value) ? value : null);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="body"
            validators={{
              onBlur: ({ value }) => checkBody(value),
              onSubmit: ({ value }) => checkBody(value),
            }}
          >
            {(field) => {
              const key =
                field.state.meta.errors[0] ??
                (serverInvalid.has('body') ? 'discover.report.err.body' : undefined);
              return (
                <TextArea
                  label={t('discover.report.body')}
                  rows={4}
                  value={field.state.value}
                  htmlName={field.name}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          {sourceOptions.length > 0 ? (
            <form.Field name="source">
              {(field) => (
                <Selector
                  label={t('discover.report.source')}
                  isOptional
                  hasClear
                  placeholder={t('discover.report.source.placeholder')}
                  options={sourceOptions}
                  value={field.state.value}
                  onChange={(value) => {
                    field.handleChange(value);
                  }}
                />
              )}
            </form.Field>
          ) : null}
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <div>
                <Button
                  type="submit"
                  variant="primary"
                  label={t('discover.report.submit')}
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

/** « Signaler une erreur » on a listing: a quiet action that opens the form in place. */
export function ErrorReport({ resourceId }: { resourceId: string }) {
  const { t } = useI18n();
  const session = useSession();
  const [isOpen, setOpen] = useState(false);
  const [sent, setSent] = useState(false);

  return (
    <VStack gap={3} hAlign="start">
      <Button
        variant="ghost"
        size="sm"
        label={t('discover.report.action')}
        icon={<Icon icon="warning" />}
        aria-expanded={isOpen}
        onClick={() => {
          setOpen((open) => !open);
        }}
      />
      {isOpen && sent ? (
        <FocusRegion role="status">
          <Text as="p">{t('discover.report.thanks')}</Text>
        </FocusRegion>
      ) : null}
      {isOpen && !sent && session.data?.authenticated !== true ? (
        <HStack gap={3} vAlign="center" wrap="wrap">
          <Text as="p">{t('discover.report.signIn')}</Text>
          <Button
            label={t('discover.report.signIn.action')}
            href={signInHref(`/discover/${resourceId}`)}
          />
        </HStack>
      ) : null}
      {isOpen && !sent && session.data?.authenticated === true ? (
        <Form
          resourceId={resourceId}
          onSent={() => {
            setSent(true);
          }}
        />
      ) : null}
    </VStack>
  );
}
